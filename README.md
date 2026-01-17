
# Gopher-Get 

**Gopher-Get** is a high-performance, concurrent CLI downloader built with Go. It doesn't just download files; it manages system resources efficiently and ensures data integrity with "Smart Resume" capabilities.

## Why Gopher-Get?

Most simple downloaders start from scratch if the connection drops. **Gopher-Get** is designed to pick up exactly where it left off, saving you time and bandwidth.

### 🛠 Key Engineering Features
* **Worker Pool Architecture:** Uses a fixed number of goroutines to prevent system resource exhaustion.
* **Idempotent Downloads:** Implements `HTTP 206 Partial Content` via Range headers for seamless resuming.
* **Streamed I/O:** Leverages `io.MultiWriter` to pipe data directly to disk, maintaining a constant memory footprint even for multi-gigabyte files.
* **Graceful Handling:** Integrated `context` management for network timeouts and cancellations.

Pass your URLs as arguments. Control the level of concurrency using the -c flag (default is 3).
---

## Installation Instructions

```bash
# Clone the project
git clone [https://github.com/AhrnishDahal/gopher-get.git](https://github.com/AhrnishDahal/gopher-get.git)

# Install dependencies
go mod tidy

# Build the executable
go build -o gopher-get

