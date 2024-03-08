package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func NewTokenBucketRateLimiter(capacity, refillRate int) TokenBucketRateLimiter {
	return TokenBucketRateLimiter{
		capacity:            capacity,
		refillRatePerSecond: refillRate,
	}
}

//RATE LIMITER MIDDLEWARE LAYER
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

		// Retrieve the rate limiting configuration for the specified IP address
		if limits, ok := config.IPLimits[ipAddress]; ok && len(limits) == 2 {
			bucketSize = limits[0]
			refillRate = limits[1]
			fmt.Printf("Found configuration for IP %s - Bucket Size: %d, Refill Rate: %d\n", ipAddress, bucketSize, refillRate)
		} else {
			fmt.Printf("Configuration for IP %s not found or is invalid.\n", ipAddress)
		}
		limiter := NewTokenBucketRateLimiter(bucketSize, refillRate)
		if !limiter.Allow(ipAddress) {
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
