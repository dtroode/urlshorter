//go:build tools

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// Number of URLs to create in each batch
	batchSize = 100
	// Number of concurrent users
	numUsers = 10
	// Number of batches per user
	batchesPerUser = 5
	// Time to wait between requests to avoid overwhelming the server
	requestDelay = 100 * time.Millisecond
)

type createShortURLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type createShortURLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type getUserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func main() {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL environment variable is required")
	}

	// Ensure baseURL ends with a slash
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	start := time.Now()

	// Create a wait group to wait for all users to finish
	var wg sync.WaitGroup
	wg.Add(numUsers)

	// Start concurrent users
	for i := 0; i < numUsers; i++ {
		go func(userNum int) {
			defer wg.Done()
			runUser(userNum, baseURL)
		}(i)
	}

	// Wait for all users to finish
	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Load test completed in %v\n", duration)
}

func runUser(userNum int, baseURL string) {
	// Create a client with a cookie jar to maintain session
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Printf("User %d: Failed to create cookie jar: %v", userNum, err)
		return
	}

	client := &http.Client{
		Jar: jar,
	}

	// First request will create a new user session
	resp, err := client.Get(baseURL)
	if err != nil {
		log.Printf("User %d: Failed to create session: %v", userNum, err)
		return
	}
	defer resp.Body.Close()

	// Get user ID from cookie
	cookies := resp.Cookies()
	var userID string
	for _, cookie := range cookies {
		if cookie.Name == "token" {
			userID = cookie.Value
			break
		}
	}
	if userID == "" {
		log.Printf("User %d: Failed to get user ID from cookie", userNum)
		return
	}
	log.Printf("User %d: Created session with ID %s", userNum, userID)

	// Store created short URLs for later use
	var shortURLs []string

	// Create multiple batches of URLs
	for batchNum := 0; batchNum < batchesPerUser; batchNum++ {
		// Log cookies before each request
		parsedURL, _ := url.Parse(baseURL)
		cookies := jar.Cookies(parsedURL)
		log.Printf("User %d: Cookies before batch %d: %v", userNum, batchNum, cookies)

		// Create batch request with unique URLs for this user
		batch := make([]createShortURLBatchRequest, batchSize)
		for i := 0; i < batchSize; i++ {
			batch[i] = createShortURLBatchRequest{
				CorrelationID: fmt.Sprintf("user%d_batch%d_url%d", userNum, batchNum, i),
				OriginalURL:   fmt.Sprintf("https://example.com/user%d/batch%d/url%d/%d", userNum, batchNum, i, time.Now().UnixNano()),
			}
		}

		// Send batch request
		batchJSON, err := json.Marshal(batch)
		if err != nil {
			log.Printf("User %d: Failed to marshal batch: %v", userNum, err)
			continue
		}

		resp, err := client.Post(
			baseURL+"api/shorten/batch",
			"application/json",
			bytes.NewBuffer(batchJSON),
		)
		if err != nil {
			log.Printf("User %d: Failed to create batch: %v", userNum, err)
			continue
		}

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("User %d (ID: %s): Batch creation failed with status %d: %s", userNum, userID, resp.StatusCode, string(body))
			continue
		}

		// Parse response
		var batchResp []createShortURLBatchResponse
		if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
			resp.Body.Close()
			log.Printf("User %d (ID: %s): Failed to decode batch response: %v", userNum, userID, err)
			continue
		}
		resp.Body.Close()

		// Store short URLs
		for _, url := range batchResp {
			shortURLs = append(shortURLs, url.ShortURL)
		}

		log.Printf("User %d (ID: %s): Successfully created batch of %d URLs", userNum, userID, len(batchResp))

		// Log cookies after batch creation
		cookiesAfterBatch := jar.Cookies(parsedURL)
		log.Printf("User %d: Cookies after batch %d: %v", userNum, batchNum, cookiesAfterBatch)

		// Add a small delay to ensure URLs are saved
		time.Sleep(100 * time.Millisecond)

		// Verify URLs were saved by trying to get them immediately
		verifyResp, err := client.Get(baseURL + "api/user/urls")
		if err != nil {
			log.Printf("User %d (ID: %s): Failed to verify URLs after creation: %v", userNum, userID, err)
			continue
		}

		if verifyResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(verifyResp.Body)
			verifyResp.Body.Close()
			log.Printf("User %d (ID: %s): Verification failed with status %d: %s", userNum, userID, verifyResp.StatusCode, string(body))
			continue
		}

		var verifyURLs []getUserURLResponse
		if err := json.NewDecoder(verifyResp.Body).Decode(&verifyURLs); err != nil {
			verifyResp.Body.Close()
			log.Printf("User %d (ID: %s): Failed to decode verification response: %v", userNum, userID, err)
			continue
		}
		verifyResp.Body.Close()

		log.Printf("User %d (ID: %s): Verified %d URLs after creation", userNum, userID, len(verifyURLs))

		time.Sleep(requestDelay)
	}

	// Get all user URLs
	resp, err = client.Get(baseURL + "api/user/urls")
	if err != nil {
		log.Printf("User %d (ID: %s): Failed to get user URLs: %v", userNum, userID, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusNoContent {
			log.Printf("User %d (ID: %s): No URLs found for user", userNum, userID)
		} else {
			log.Printf("User %d (ID: %s): Failed to get user URLs with status %d: %s", userNum, userID, resp.StatusCode, string(body))
		}
		return
	}

	var userURLs []getUserURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&userURLs); err != nil {
		resp.Body.Close()
		log.Printf("User %d (ID: %s): Failed to decode user URLs response: %v", userNum, userID, err)
		return
	}
	resp.Body.Close()

	// Verify we got all URLs
	if len(userURLs) != len(shortURLs) {
		log.Printf("User %d (ID: %s): Expected %d URLs but got %d", userNum, userID, len(shortURLs), len(userURLs))
	} else {
		log.Printf("User %d (ID: %s): Successfully retrieved %d URLs", userNum, userID, len(userURLs))
	}

	// Delete half of the URLs
	urlsToDelete := shortURLs[:len(shortURLs)/2]

	// Extract short keys from URLs
	shortKeysToDelete := make([]string, len(urlsToDelete))
	for i, fullURL := range urlsToDelete {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			log.Printf("User %d: Failed to parse URL %s: %v", userNum, fullURL, err)
			continue
		}
		// Extract the short key from the path (remove leading slash)
		shortKey := strings.TrimPrefix(parsedURL.Path, "/")
		shortKeysToDelete[i] = shortKey
	}

	deleteJSON, err := json.Marshal(shortKeysToDelete)
	if err != nil {
		log.Printf("User %d: Failed to marshal delete request: %v", userNum, err)
		return
	}

	req, err := http.NewRequest(http.MethodDelete, baseURL+"api/user/urls", bytes.NewBuffer(deleteJSON))
	if err != nil {
		log.Printf("User %d: Failed to create delete request: %v", userNum, err)
		return
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Printf("User %d: Failed to delete URLs: %v", userNum, err)
		return
	}

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("User %d: Failed to delete URLs with status %d: %s", userNum, resp.StatusCode, string(body))
		return
	}
	resp.Body.Close()

	// Wait longer for deletion to complete (async operation)
	time.Sleep(5 * time.Second)

	// Verify URLs were deleted
	resp, err = client.Get(baseURL + "api/user/urls")
	if err != nil {
		log.Printf("User %d: Failed to get user URLs after deletion: %v", userNum, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("User %d: Failed to get user URLs after deletion with status %d: %s", userNum, resp.StatusCode, string(body))
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(&userURLs); err != nil {
		resp.Body.Close()
		log.Printf("User %d: Failed to decode user URLs response after deletion: %v", userNum, err)
		return
	}
	resp.Body.Close()

	// Verify we got the correct number of URLs
	expectedRemaining := len(shortURLs) - len(urlsToDelete)
	if len(userURLs) != expectedRemaining {
		log.Printf("User %d: After deletion, expected %d URLs but got %d", userNum, expectedRemaining, len(userURLs))
	}

	// Try to access some deleted URLs
	for _, shortURL := range urlsToDelete {
		parsedURL, err := url.Parse(shortURL)
		if err != nil {
			log.Printf("User %d: Failed to parse URL %s: %v", userNum, shortURL, err)
			continue
		}

		// Extract the short key from the path
		shortKey := strings.TrimPrefix(parsedURL.Path, "/")

		// Log cookies before making the request
		baseParsedURL, _ := url.Parse(baseURL)
		cookies := jar.Cookies(baseParsedURL)
		log.Printf("User %d: Cookies before accessing deleted URL %s: %v", userNum, shortKey, cookies)

		// Remove trailing slash from baseURL to avoid double slashes
		cleanBaseURL := strings.TrimSuffix(baseURL, "/")
		fullRequestURL := cleanBaseURL + "/" + shortKey
		log.Printf("User %d: Making request to deleted URL: %s", userNum, fullRequestURL)

		resp, err := client.Get(fullRequestURL)
		if err != nil {
			log.Printf("User %d: Failed to get deleted URL %s: %v", userNum, shortURL, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		log.Printf("User %d: Deleted URL %s returned status %d, body: %s", userNum, shortURL, resp.StatusCode, string(body))

		if resp.StatusCode != http.StatusGone {
			log.Printf("User %d: Deleted URL %s returned unexpected status %d: %s", userNum, shortURL, resp.StatusCode, string(body))
			continue
		}

		time.Sleep(requestDelay)
	}
}
