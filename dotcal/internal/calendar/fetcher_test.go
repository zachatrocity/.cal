package calendar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewFetcher(t *testing.T) {
	fetcher := NewFetcher()
	if fetcher == nil {
		t.Error("Expected non-nil fetcher")
	}
	if fetcher.client == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestFetch(t *testing.T) {
	fetcher := NewFetcher()

	t.Run("fetch from URL", func(t *testing.T) {
		// Create test server
		testData := "BEGIN:VCALENDAR\nEND:VCALENDAR"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testData))
		}))
		defer server.Close()

		feed := Feed{
			Source: server.URL,
			IsURL:  true,
		}

		data, err := fetcher.Fetch(feed)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if string(data) != testData {
			t.Errorf("Expected %q, got %q", testData, string(data))
		}
	})

	t.Run("fetch from file", func(t *testing.T) {
		// Create temporary test file
		testData := "BEGIN:VCALENDAR\nEND:VCALENDAR"
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.ics")
		if err := os.WriteFile(tmpFile, []byte(testData), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		feed := Feed{
			Source: tmpFile,
			IsURL:  false,
		}

		data, err := fetcher.Fetch(feed)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if string(data) != testData {
			t.Errorf("Expected %q, got %q", testData, string(data))
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		feed := Feed{
			Source: "http://invalid.example.com",
			IsURL:  true,
		}

		_, err := fetcher.Fetch(feed)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		feed := Feed{
			Source: "/nonexistent/file.ics",
			IsURL:  false,
		}

		_, err := fetcher.Fetch(feed)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("HTTP error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		feed := Feed{
			Source: server.URL,
			IsURL:  true,
		}

		_, err := fetcher.Fetch(feed)
		if err == nil {
			t.Error("Expected error for HTTP 404 response")
		}
	})

	t.Run("HTTP server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		feed := Feed{
			Source: server.URL,
			IsURL:  true,
		}

		_, err := fetcher.Fetch(feed)
		if err == nil {
			t.Error("Expected error for HTTP 500 response")
		}
	})

	t.Run("malformed URL", func(t *testing.T) {
		feed := Feed{
			Source: "not-a-url",
			IsURL:  true,
		}

		_, err := fetcher.Fetch(feed)
		if err == nil {
			t.Error("Expected error for malformed URL")
		}
	})
}
