package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type IpAddressesConfig struct {
	IPAddresses []string `json:"ipAddresses"`
}

//SEQUENTIAL
func main() {
	config, err := readIPConfig("ip_config.json")
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}
	for _, ipAddress := range config.IPAddresses {
		// URL of the API endpoint
		url := "http://localhost:8080/ping?ip=" + ipAddress

		// Make a GET request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error for IP %s: %s\n", ipAddress, err)
			continue
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Response for IP " + ipAddress + ": " + resp.Status)
		fmt.Println("Message:", string(body)) // Directly print the string body
	}
}

//CONCURRENT
//func main() {
//	config, err := readIPConfig("ip_config.json")
//	if err != nil {
//		log.Fatalf("Error reading config: %s", err)
//	}
//	var wg sync.WaitGroup
//
//	for _, ipAddress := range config.IPAddresses {
//		wg.Add(1)
//		go func(ip string) {
//			defer wg.Done()
//			url := "http://localhost:8080/ping?ip=" + ip
//
//			resp, err := http.Get(url)
//			if err != nil {
//				fmt.Printf("Error for IP %s: %s\n", ip, err)
//				return
//			}
//			defer resp.Body.Close()
//
//			body, _ := ioutil.ReadAll(resp.Body)
//			fmt.Println("Response for IP " + ip + ": " + resp.Status)
//			fmt.Println("Message:", string(body))
//		}(ipAddress)
//	}
//
//	wg.Wait()
//}
