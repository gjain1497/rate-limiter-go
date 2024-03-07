package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Reached hit API")
	// URL of the API endpoint
	ipAddresses := []string{"192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.1", "192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.3", "192.168.1.3"}

	// Iterate over the IP addresses
	for _, ipAddress := range ipAddresses {
		// URL of the API endpoint
		url := "http://localhost:8080/ping?ip=" + ipAddress

		// Make a GET request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error for IP %s: %s\n", ipAddress, err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Response for IP %s: %s\n", ipAddress, resp.Status)
	}
}
