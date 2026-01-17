package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake file content"))
	}))
	defer server.Close()

	ctx := context.Background()
	testFile := "pkg.png"

	err := downloadFile(ctx, server.URL+"/"+testFile)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Errorf("file %s was not created", testFile)
	}

	os.Remove(testFile)
}

func TestDownloadResume(t *testing.T) {
	content := "full-length-content"
	initialData := "full-l"
	testFile := "resume_test.txt"

	if err := os.WriteFile(testFile, []byte(initialData), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") == "bytes=6-" {
			w.WriteHeader(http.StatusPartialContent)
			w.Write([]byte(content[6:]))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	if err := downloadFile(context.Background(), server.URL+"/"+testFile); err != nil {
		t.Fatalf("resume failed: %v", err)
	}

	finalData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(finalData) != content {
		t.Errorf("got %q, want %q", string(finalData), content)
	}
}
