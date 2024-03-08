package main

import (
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
)

func main() {
	initiateRedisClient()
	startServer()
}

func initiateRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func startServer() {
	log.Println("Started Server")
	http.Handle("/ping", rateLimiter(endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening on port :8080", err)
	}
}
