package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"time"
)

type (
	MainConfig struct {
		IPAddress []string
	}
)

var (
	MainCfg *MainConfig
)

var GetConfig = func() (*MainConfig, error) {
	if MainCfg == nil {
		return nil, errors.New("Config is empty")
	}
	return MainCfg, nil
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
		tokens = math.max(tokens - 1, 0)

		local ttl = 60 -- only for testing
		
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

	//limiter := NewTokenBucketRateLimiter(10, 2)

	//userId := time.Now().GoString()
	//for i := 1; i <= 40; i++ {
	//	if limiter.Allow(userId) {
	//		fmt.Printf("Request %d allowed\n", i)
	//	} else {
	//		fmt.Printf("Request %d rejected\n", i)
	//	}
	//	time.Sleep(100 * time.Millisecond)
	//	if i > 30 {
	//		time.Sleep(500 * time.Millisecond)
	//	}
	//}
	//viper.SetConfigFile("config.yaml") // Assuming your config file is named config.yaml
	//err := viper.ReadInConfig()
	//if err != nil {
	//	log.Fatalf("Error reading config file: %s", err)
	//}
	//
	//ip1 := viper.GetStringSlice("IPAddress")[0]
	//ip2 := viper.GetStringSlice("IPAddress")[1]
	//ip3 := viper.GetStringSlice("IPAddress")[2]

	http.Handle("/ping", rateLimiter(endpointHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening on port :8080", err)
	}

}

type Message struct {
	Status string
	Body   string
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
	log.Println("Reached in rateLimiter")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//FOR GETTING ACTUAL IP ADDRESS IN PROD WE USE THIS
		//ip, _, err := net.SplitHostPort(r.RemoteAddr)
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}

		//HERE WE ARE PASSING IP ADDRESS AS PARAMETERS
		ip := r.URL.Query().Get("ip")
		log.Println("ip passed is ", ip)
		limiter := NewTokenBucketRateLimiter(3, 2)
		if !limiter.Allow(ip) {
			log.Println("Reached here ")
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}
			log.Println("Reached message ", message)

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}
		next(w, r)
	})
}
