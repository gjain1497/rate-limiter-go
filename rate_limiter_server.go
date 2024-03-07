package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Status string
	Body   string
}
type IPRateLimitingMappingConfig struct {
	IPLimits map[string][]int `json:"ipLimits"`
}

var redisClient *redis.Client

type TokenBucketRateLimiter struct {
	capacity            int
	refillRatePerSecond int
}

func NewTokenBucketRateLimiter(capacity, refillRate int) TokenBucketRateLimiter {
	return TokenBucketRateLimiter{
		capacity:            capacity,
		refillRatePerSecond: refillRate,
	}
}

func (limiter *TokenBucketRateLimiter) Allow(userId string) bool {
	script := `
		local tokens = tonumber(redis.call("get", KEYS[1]))
		local lastRefill = tonumber(redis.call("get", KEYS[2]))
		

		local now = tonumber(ARGV[1])
		local refillRate = tonumber(ARGV[2])
		local capacity = tonumber(ARGV[3])

		if not lastRefill then
			lastRefill = now
			tokens = capacity
		end

		-- refill logic
		local secondsElapsed = now - lastRefill
		local tokensToAdd = secondsElapsed * refillRate
		tokens = math.min(tokens + tokensToAdd, capacity)
		lastRefill = now

		-- consumption logic
		tokens = math.max(tokens - 1, -1)

		local ttl = 600 -- only for testing
		
		local key = KEYS[1]
		redis.log(redis.LOG_NOTICE, "Value of key: " .. key)
		redis.call("setex", KEYS[1], ttl, tokens)
		redis.call("setex", KEYS[2], ttl, lastRefill)

			return tokens
	`

	userIdTokensKey := fmt.Sprintf("user_id.%s.tokens", userId)
	fmt.Println("key ", userIdTokensKey)
	userIdLastRefillKey := fmt.Sprintf("user_id.%s.last_refill", userId)

	cmd := redisClient.Eval(context.Background(), script,
		[]string{userIdTokensKey, userIdLastRefillKey},
		time.Now().Unix(), limiter.refillRatePerSecond, limiter.capacity)

	if cmd.Err() != nil {
		fmt.Println(cmd.Err().Error())
		return false
	}

	tokenCount := cmd.Val().(int64)

	fmt.Println("tokenCount", tokenCount)

	return tokenCount >= 0
}

func main() {
	log.Println("Hello")
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	http.Handle("/ping", rateLimiter(endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening on port :8080", err)
	}
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{
		Status: "Successful",
		Body:   "Hi! You've reached the API. How may I help you?",
	}
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func rateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//FOR GETTING ACTUAL IP ADDRESS IN PROD WE USE THIS
		//ip, _, err := net.SplitHostPort(r.RemoteAddr)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}

		config, err := readIPLateLimitingConfig("ip_rate_limit_mapping_config.json")
		if err != nil {
			log.Fatalf("Error reading config: %s", err)
		}

		//HERE WE ARE PASSING IP ADDRESS AS PARAMETERS
		ipAddress := r.URL.Query().Get("ip")
		var bucketSize int
		var refillRate int

		log.Println("ipAddress passed is ", ipAddress)
		// Attempt to retrieve the rate limiting configuration for the specified IP address
		if limits, ok := config.IPLimits[ipAddress]; ok && len(limits) == 2 {
			bucketSize := limits[0]
			refillRate := limits[1]
			fmt.Printf("Found configuration for IP %s - Bucket Size: %d, Refill Rate: %d\n", ipAddress, bucketSize, refillRate)
		} else {
			fmt.Printf("Configuration for IP %s not found or is invalid.\n", ipAddress)
		}
		limiter := NewTokenBucketRateLimiter(bucketSize, refillRate)
		if !limiter.Allow(ipAddress) {
			log.Println("Reached here ")
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		next(w, r)
	})
}

func readIPLateLimitingConfig(filePath string) (*IPRateLimitingMappingConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &IPRateLimitingMappingConfig{}
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
