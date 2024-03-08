package main

import (
	"context"
	"fmt"
	"time"
)

func (limiter *TokenBucketRateLimiter) Allow(clientId string) bool {
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

			return {tokens, secondsElapsed, tokensToAdd}
	`

	userIdTokensKey := fmt.Sprintf("client_id.%s.tokens", clientId)
	fmt.Println("key ", userIdTokensKey)
	userIdLastRefillKey := fmt.Sprintf("client_id.%s.last_refill", clientId)

	cmd := redisClient.Eval(context.Background(), script,
		[]string{userIdTokensKey, userIdLastRefillKey},
		time.Now().Unix(), limiter.refillRatePerSecond, limiter.capacity)

	if cmd.Err() != nil {
		fmt.Println(cmd.Err().Error())
		return false
	}
	// Parse the result into a []interface{}
	result, err := cmd.Result()
	if err != nil {
		fmt.Println("Error parsing result:", err)
		return false
	}

	// Parse the array returned by the Lua script
	vals, ok := result.([]interface{})
	if !ok || len(vals) != 3 {
		fmt.Println("Invalid result format")
		return false
	}

	tokenCount, ok := vals[0].(int64)
	if !ok {
		fmt.Println("Error parsing token count")
		return false
	}

	secondsElapsed, ok := vals[1].(int64)
	if !ok {
		fmt.Println("Error parsing secondsElapsed")
		return false
	}

	tokensToAdd, ok := vals[2].(int64)
	if !ok {
		fmt.Println("Error parsing tokensToAdd")
		return false
	}

	fmt.Printf("Token count: %d, Time elapsed: %d, Tokens to be Added : %d\n", tokenCount, secondsElapsed, tokensToAdd)

	fmt.Sprintf("tokenCount for ip address %s is %d", clientId, tokenCount)

	return tokenCount >= 0
}
