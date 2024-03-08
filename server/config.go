package main

import "github.com/go-redis/redis/v8"

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
