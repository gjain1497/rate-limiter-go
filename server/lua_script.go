package main

import (
	"context"
	"fmt"
	"time"
)

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

	fmt.Println("tokenCount for ip address", tokenCount, userId)

	return tokenCount >= 0
}
