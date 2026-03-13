package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	latency time.Duration
	status  int
	err     error
}

func main() {
	var (
		rawURL      string
		n           int
		concurrency int
		timeout     time.Duration
		keepAlive   bool
	)

	flag.StringVar(&rawURL, "url", "http://localhost:8081/movies?director=Christopher+Nolan&min_rating=4", "Target URL")
	flag.IntVar(&n, "n", 1000, "Total number of requests")
	flag.IntVar(&concurrency, "c", 1000, "Concurrency")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Per-request timeout")
	flag.BoolVar(&keepAlive, "keepalive", true, "Use HTTP keep-alive")
	flag.Parse()

	u, err := url.Parse(rawURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid url: %v\n", err)
		os.Exit(1)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          2000,
		MaxIdleConnsPerHost:   2000,
		MaxConnsPerHost:       0,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     !keepAlive,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	if concurrency > n {
		concurrency = n
	}

	results := make([]result, n)

	var (
		startAll = time.Now()
		wg       sync.WaitGroup
		idx      int64 = -1
	)

	wg.Add(concurrency)
	for w := 0; w < concurrency; w++ {
		go func() {
			defer wg.Done()
			for {
				i := int(atomic.AddInt64(&idx, 1))
				if i >= n {
					return
				}

				req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), nil)

				t0 := time.Now()
				resp, e := client.Do(req)
				dt := time.Since(t0)

				r := result{latency: dt, err: e}
				if e == nil && resp != nil {
					r.status = resp.StatusCode
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
				results[i] = r
			}
		}()
	}

	wg.Wait()
	totalWall := time.Since(startAll)

	var (
		okCount      int
		errCount     int
		latencies    []float64
		statusCount  = map[int]int{}
		sumLatencyMS float64
		minLatencyMS = math.Inf(1)
		maxLatencyMS float64
	)

	for _, r := range results {
		if r.err != nil {
			errCount++
			continue
		}
		okCount++
		statusCount[r.status]++

		ms := float64(r.latency) / float64(time.Millisecond)
		latencies = append(latencies, ms)
		sumLatencyMS += ms
		if ms < minLatencyMS {
			minLatencyMS = ms
		}
		if ms > maxLatencyMS {
			maxLatencyMS = ms
		}
	}

	sort.Float64s(latencies)

	mean := 0.0
	if len(latencies) > 0 {
		mean = sumLatencyMS / float64(len(latencies))
	} else {
		minLatencyMS = 0
	}

	p50 := percentile(latencies, 0.50)
	p95 := percentile(latencies, 0.95)
	p99 := percentile(latencies, 0.99)

	rps := float64(n) / totalWall.Seconds()

	fmt.Println("==== Load Test Results ====")
	fmt.Printf("URL: %s\n", rawURL)
	fmt.Printf("Requests: %d\n", n)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Printf("Timeout: %s\n", timeout)
	fmt.Printf("Keep-Alive: %v\n\n", keepAlive)

	fmt.Printf("Total time: %s\n", totalWall)
	fmt.Printf("Throughput: %.2f req/s\n", rps)
	fmt.Printf("Success: %d\n", okCount)
	fmt.Printf("Errors: %d\n\n", errCount)

	if len(latencies) > 0 {
		fmt.Println("Latency (ms):")
		fmt.Printf("min: %.2f\n", minLatencyMS)
		fmt.Printf("p50: %.2f\n", p50)
		fmt.Printf("p95: %.2f\n", p95)
		fmt.Printf("p99: %.2f\n", p99)
		fmt.Printf("max: %.2f\n", maxLatencyMS)
		fmt.Printf("avg: %.2f\n\n", mean)
	} else {
		fmt.Println("No successful requests to compute latency.")
	}

	fmt.Println("Status codes:")
	var codes []int
	for c := range statusCount {
		codes = append(codes, c)
	}
	sort.Ints(codes)
	for _, c := range codes {
		fmt.Printf("%d: %d\n", c, statusCount[c])
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	rank := int(math.Ceil(p*float64(len(sorted)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}
