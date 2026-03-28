package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gitlab.com/threetopia/envgo"
)

type RegisterRequest struct {
	Registrant struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Gender   string `json:"gender"`
		TicketID string `json:"ticket_id"`
	} `json:"registrant"`
	Attendees []struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Gender   string `json:"gender"`
		TicketID string `json:"ticket_id"`
	} `json:"attendees"`
}

type TestResult struct {
	TotalRequests     int
	SuccessCount      int32
	FailedCount       int32
	InsufficientStock int32
	DuplicateKey      int32
	NetworkError      int32
	SuccessRate       float64
	Duration          time.Duration
}

func main() {
	envgo.LoadDotEnv("./.env")

	baseURL := envgo.GetString("STRESS_TEST_BASE_URL", "http://localhost:8000")
	ticketID := envgo.GetString("STRESS_TEST_TICKET_ID", "")
	concurrency := 50
	duration := 30 * time.Second

	if ticketID == "" {
		fmt.Println("STRESS_TEST_TICKET_ID is required")
		os.Exit(1)
	}

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           TICKET BOOKING STRESS TEST                         ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Base URL:      %-45s║\n", baseURL)
	fmt.Printf("║ Ticket ID:     %-45s║\n", ticketID[:8]+"...")
	fmt.Printf("║ Concurrency:   %-45d║\n", concurrency)
	fmt.Printf("║ Duration:      %-45v║\n", duration)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	results := runStressTest(baseURL, ticketID, concurrency, duration)

	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      TEST RESULTS                            ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Total Requests:     %-38d║\n", results.TotalRequests)
	fmt.Printf("║ Successful:          %-38d║\n", results.SuccessCount)
	fmt.Printf("║ Failed:              %-38d║\n", results.FailedCount)
	fmt.Printf("║ Insufficient Stock:  %-38d║\n", results.InsufficientStock)
	fmt.Printf("║ Duplicate Key:       %-38d║\n", results.DuplicateKey)
	fmt.Printf("║ Network Errors:      %-38d║\n", results.NetworkError)
	fmt.Printf("║ Success Rate:        %-38.2f%%║\n", results.SuccessRate)
	fmt.Printf("║ Duration:            %-38v║\n", results.Duration)
	fmt.Printf("║ Requests/Second:     %-38.2f║\n", float64(results.TotalRequests)/results.Duration.Seconds())
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func runStressTest(baseURL, ticketID string, concurrency int, duration time.Duration) TestResult {
	var result TestResult
	result.Duration = duration

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	wg := sync.WaitGroup{}
	sem := make(chan struct{}, concurrency)

	var requestCount int32
	var stop atomic.Bool

	go func() {
		for {
			select {
			case <-ticker.C:
				currentCount := atomic.LoadInt32(&requestCount)
				if currentCount > 0 {
					fmt.Printf("\r[*] Running: %d requests sent | Success: %d | Failed: %d",
						currentCount, result.SuccessCount, result.FailedCount)
				}
			case <-ctx.Done():
				stop.Store(true)
				return
			}
		}
	}()

	for !stop.Load() {
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			atomic.AddInt32(&requestCount, 1)
			makeRequest(baseURL, ticketID, &result)
		}()
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	result.TotalRequests = int(requestCount)
	if result.TotalRequests > 0 {
		result.SuccessRate = float64(result.SuccessCount) / float64(result.TotalRequests) * 100
	}

	return result
}

func makeRequest(baseURL, ticketID string, result *TestResult) {
	email := fmt.Sprintf("test_%s_%d@example.com", uuid.New().String()[:8], time.Now().UnixNano())

	reqBody := RegisterRequest{}
	reqBody.Registrant.Name = "Test User"
	reqBody.Registrant.Email = email
	reqBody.Registrant.Phone = "+6281234567890"
	reqBody.Registrant.Gender = "male"
	reqBody.Registrant.TicketID = ticketID

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		atomic.AddInt32(&result.NetworkError, 1)
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/api/v1/register", bytes.NewBuffer(jsonBody))
	if err != nil {
		atomic.AddInt32(&result.NetworkError, 1)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt32(&result.NetworkError, 1)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		atomic.AddInt32(&result.SuccessCount, 1)
	} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errMsg, ok := errResp["error"]; ok {
			if contains(errMsg, []string{"stock", "habis", "tidak mencukupi"}) {
				atomic.AddInt32(&result.InsufficientStock, 1)
			} else if contains(errMsg, []string{"duplicate", "unique"}) {
				atomic.AddInt32(&result.DuplicateKey, 1)
			}
		}
		atomic.AddInt32(&result.FailedCount, 1)
	} else {
		atomic.AddInt32(&result.FailedCount, 1)
	}
}

func contains(s string, substrs []string) bool {
	for _, sub := range substrs {
		if bytes.Contains([]byte(s), []byte(sub)) {
			return true
		}
	}
	return false
}

func init() {
	log.SetFlags(0)
}
