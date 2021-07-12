package main

import (
	"fmt"
	"log"
	"time"

	"mycurl/internal/config"
	"mycurl/internal/http"
)

func main() {
	cfg := config.GetConfig()
	cfg.Validate()

	if cfg.Verbose {
		fmt.Printf(
			"REQUEST\nMethod: %s\nURL: %s\nContent type: %s\n\n", cfg.Method, cfg.URL, cfg.ContentType)
	}

	resp, err := doRequest(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Verbose {
		fmt.Printf(
			"RESPONSE\nStatus: %d\nContent type: %s\nBody len: %d\nBody: \n", resp.Status, resp.Headers["Content-Type"], len(resp.Body))
	}
	fmt.Println(string(resp.Body))
}

func doRequest(cfg *config.Config) (*http.Response, error) {
	client, err := http.NewHttpClient(5 * time.Second)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.Method(cfg.Method), cfg.URL, cfg.ContentType, []byte(cfg.Body))
	if err != nil {
		return nil, err
	}

	req.SetHeader("User-Agent", "mycurl/0.0.1")
	return client.Do(req)
}
