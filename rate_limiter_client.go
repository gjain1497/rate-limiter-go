package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

//SEQUENTIAL
//func main() {
//	log.Println("Reached hit API")
//	// URL of the API endpoint
//	ipAddresses := []string{"192.168.1.1", "192.168.1.2",
//		"192.168.1.3", "192.168.1.1", "192.168.1.1", "192.168.1.2",
//		"192.168.1.3", "192.168.1.3", "192.168.1.3"}
//
//	// Iterate over the IP addresses
//	for _, ipAddress := range ipAddresses {
//		// URL of the API endpoint
//		url := "http://localhost:8080/ping?ip=" + ipAddress
//
//		// Make a GET request
//		resp, err := http.Get(url)
//		if err != nil {
//			fmt.Printf("Error for IP %s: %s\n", ipAddress, err)
//			continue
//		}
//		defer resp.Body.Close()
//		body, _ := ioutil.ReadAll(resp.Body)
//		fmt.Println("Response for IP " + ipAddress + ": " + resp.Status)
//		fmt.Println("Message:", string(body)) // Directly print the string body
//	}
//}

//CONCURRENT
func main() {
	// URL of the API endpoint
	ipAddresses := []string{"192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.1", "192.168.1.1", "192.168.1.2",
		"192.168.1.3", "192.168.1.3", "192.168.1.3"}

	var wg sync.WaitGroup // Use a WaitGroup to wait for all goroutines to finish

	for _, ipAddress := range ipAddresses {
		wg.Add(1)            // Increment the WaitGroup counter
		go func(ip string) { // Pass the current ipAddress to the goroutine
			defer wg.Done() // Decrement the counter when the goroutine completes
			// URL of the API endpoint
			url := "http://localhost:8080/ping?ip=" + ip

			// Make a GET request
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error for IP %s: %s\n", ip, err)
				return
			}
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("Response for IP " + ip + ": " + resp.Status)
			fmt.Println("Message:", string(body)) // Directly print the string body
		}(ipAddress) // Pass the current ipAddress as an argument to the goroutine
	}

	wg.Wait() // Wait for all goroutines to finish
}
