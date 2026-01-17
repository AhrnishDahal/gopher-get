package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Result holds the outcome of a download for final reporting
type Result struct {
	URL string
	Err error
}

func downloadFile(ctx context.Context, url string) error {
	fileName := path.Base(url)
	var startByte int64 = 0

	// 1. Check if file exists locally and get its size
	if info, err := os.Stat(fileName); err == nil {
		startByte = info.Size()
	}

	// 2. Prepare the Request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// 3. If we have some data, ask the server for the rest
	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 4. Handle Status Codes
	// 200 OK = Server doesn't support resume, starting over
	// 206 Partial Content = Server is sending the missing piece
	// 416 Requested Range Not Satisfiable = File is likely already finished
	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return nil // Already downloaded
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned: %s", resp.Status)
	}

	// 5. Open file in Append mode if resuming, otherwise Create/Truncate
	flags := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == http.StatusPartialContent {
		flags |= os.O_APPEND
		fmt.Printf("Resuming %s from byte %d...\n", fileName, startByte)
	} else {
		flags |= os.O_TRUNC
		startByte = 0 // Reset if server forced a full download
	}

	out, err := os.OpenFile(fileName, flags, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	// 6. Progress Bar (Adjusted for remaining bytes)
	bar := progressbar.DefaultBytes(
		resp.ContentLength+startByte,
		"downloading "+fileName,
	)
	bar.Add64(startByte) // Start the bar at the current progress

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}
func worker(ctx context.Context, id int, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		// Each download has a strict 30-second timeout
		downloadCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := downloadFile(downloadCtx, url)
		cancel() // Clean up context resources

		results <- Result{URL: url, Err: err}
	}
}

func main() {
	// 1. Flags: Let user define concurrency and timeout
	workerCount := flag.Int("c", 3, "Number of concurrent downloads")
	flag.Parse()

	// Remaining arguments are the URLs
	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Println("Usage: gopher-get -c 5 <url1> <url2> ...")
		return
	}

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))
	var wg sync.WaitGroup

	// 2. Global context for the entire application
	ctx := context.Background()

	// 3. Start the Worker Pool
	for w := 1; w <= *workerCount; w++ {
		wg.Add(1)
		go worker(ctx, w, jobs, results, &wg)
	}

	// 4. Feed the queue
	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	// 5. Wait for completion in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// 6. Collect and report results
	fmt.Println("\n--- Summary ---")
	for res := range results {
		if res.Err != nil {
			fmt.Printf("[FAIL] %s: %v\n", res.URL, res.Err)
		} else {
			fmt.Printf("[OK]   %s\n", res.URL)
		}
	}
}
