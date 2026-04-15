package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ─── ANSI Colors ────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	green  = "\033[38;5;82m"
	cyan   = "\033[38;5;51m"
	red    = "\033[38;5;196m"
	yellow = "\033[38;5;226m"
	gray   = "\033[38;5;240m"
	white  = "\033[38;5;255m"
	purple = "\033[38;5;135m"
)

// ─── Banner ──────────────────────────────────────────────────────────────────

func printBanner() {
	fmt.Printf("\n%s%s", cyan, bold)
	fmt.Println(`
  ██████╗ ██╗██████╗ ██╗  ██╗██╗   ██╗███╗   ██╗████████╗
  ██╔══██╗██║██╔══██╗██║  ██║██║   ██║████╗  ██║╚══██╔══╝
  ██║  ██║██║██████╔╝███████║██║   ██║██╔██╗ ██║   ██║   
  ██║  ██║██║██╔══██╗██╔══██║██║   ██║██║╚██╗██║   ██║   
  ██████╔╝██║██║  ██║██║  ██║╚██████╔╝██║ ╚████║   ██║   
  ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝  `)
	fmt.Printf("%s\n", reset)
	fmt.Printf("  %s%s Directory & File Bruteforcer%s", green, bold, reset)
	fmt.Printf("  %s| by Ahan Pahlevi | github.com/AhanDotID/dirhunt%s\n\n", gray, reset)
}

// ─── Result ───────────────────────────────────────────────────────────────────

type Result struct {
	URL        string
	StatusCode int
	Size       int64
	Redirect   string
}

func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return cyan
	case code == 403:
		return yellow
	case code >= 400 && code < 500:
		return red
	case code >= 500:
		return purple
	default:
		return gray
	}
}

func statusLabel(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "FOUND"
	case code >= 300 && code < 400:
		return "REDIR"
	case code == 403:
		return "FORBID"
	case code >= 400 && code < 500:
		return "CLIENT"
	case code >= 500:
		return "SERVER"
	default:
		return "???"
	}
}

func printResult(r Result) {
	col := statusColor(r.StatusCode)
	label := statusLabel(r.StatusCode)
	sizeStr := formatSize(r.Size)

	redir := ""
	if r.Redirect != "" {
		redir = fmt.Sprintf(" %s→ %s%s", gray, r.Redirect, reset)
	}

	fmt.Printf("  %s[%s]%s %s%-6s%s %s%-6s%s %s%s%s%s\n",
		col, label, reset,
		col, fmt.Sprintf("%d", r.StatusCode), reset,
		gray, sizeStr, reset,
		white, r.URL, reset,
		redir,
	)
}

func formatSize(size int64) string {
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%.1fM", float64(size)/1024/1024)
	case size >= 1024:
		return fmt.Sprintf("%.1fK", float64(size)/1024)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

// ─── HTTP Client ──────────────────────────────────────────────────────────────

func newClient(timeout int, followRedirect bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    100,
	}

	redirectPolicy := func(req *http.Request, via []*http.Request) error {
		if !followRedirect {
			return http.ErrUseLastResponse
		}
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	}

	return &http.Client{
		Transport:     transport,
		CheckRedirect: redirectPolicy,
		Timeout:       time.Duration(timeout) * time.Second,
	}
}

// ─── Scanner ──────────────────────────────────────────────────────────────────

func probe(client *http.Client, url string, userAgent string, matchCodes map[int]bool, ignoreCodes map[int]bool) *Result {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	code := resp.StatusCode

	// apply ignore list
	if len(ignoreCodes) > 0 && ignoreCodes[code] {
		return nil
	}

	// apply match list
	if len(matchCodes) > 0 && !matchCodes[code] {
		return nil
	}

	// skip 404 by default unless explicitly matched
	if code == 404 && !matchCodes[404] {
		return nil
	}

	size := resp.ContentLength

	redir := ""
	if code >= 300 && code < 400 {
		redir = resp.Header.Get("Location")
	}

	return &Result{
		URL:        url,
		StatusCode: code,
		Size:       size,
		Redirect:   redir,
	}
}

// ─── Wordlist loader ──────────────────────────────────────────────────────────

func loadWordlist(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}
	return words, scanner.Err()
}

// ─── Parse codes ─────────────────────────────────────────────────────────────

func parseCodes(raw string) map[int]bool {
	m := make(map[int]bool)
	if raw == "" {
		return m
	}
	for _, p := range strings.Split(raw, ",") {
		var code int
		fmt.Sscanf(strings.TrimSpace(p), "%d", &code)
		if code > 0 {
			m[code] = true
		}
	}
	return m
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	printBanner()

	// flags
	target      := flag.String("u", "", "Target URL (e.g. https://example.com)")
	wordlist    := flag.String("w", "", "Path to wordlist file")
	threads     := flag.Int("t", 50, "Number of concurrent threads")
	timeout     := flag.Int("timeout", 10, "HTTP request timeout in seconds")
	extensions  := flag.String("x", "", "Extensions to append (e.g. php,html,txt)")
	matchStr    := flag.String("mc", "", "Match HTTP status codes (e.g. 200,301,403)")
	ignoreStr   := flag.String("fc", "404", "Filter/ignore HTTP status codes (e.g. 404,400)")
	output      := flag.String("o", "", "Output file to save results")
	userAgent   := flag.String("ua", "DirHunt/1.0 (github.com/AhanDotID/dirhunt)", "User-Agent header")
	noFollow    := flag.Bool("no-follow", false, "Do not follow redirects")
	flag.Parse()

	// validation
	if *target == "" || *wordlist == "" {
		fmt.Printf("%s[!]%s Usage: dirhunt -u <URL> -w <wordlist> [options]\n\n", red, reset)
		flag.PrintDefaults()
		fmt.Println()
		os.Exit(1)
	}

	// normalize target
	base := strings.TrimRight(*target, "/")
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}

	// load wordlist
	words, err := loadWordlist(*wordlist)
	if err != nil {
		fmt.Printf("%s[!]%s Cannot read wordlist: %v\n", red, reset, err)
		os.Exit(1)
	}

	// build extensions
	exts := []string{""}
	if *extensions != "" {
		for _, e := range strings.Split(*extensions, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				if !strings.HasPrefix(e, ".") {
					e = "." + e
				}
				exts = append(exts, e)
			}
		}
	}

	// build URL list
	var urls []string
	for _, w := range words {
		for _, ext := range exts {
			urls = append(urls, fmt.Sprintf("%s/%s%s", base, w, ext))
		}
	}

	// parse codes
	matchCodes  := parseCodes(*matchStr)
	ignoreCodes := parseCodes(*ignoreStr)

	// info header
	fmt.Printf("  %s[*]%s Target    : %s%s%s\n", cyan, reset, white, base, reset)
	fmt.Printf("  %s[*]%s Wordlist  : %s%s%s (%d words)\n", cyan, reset, white, *wordlist, reset, len(words))
	fmt.Printf("  %s[*]%s Extensions: %s%s%s\n", cyan, reset, white, func() string {
		if *extensions == "" { return "none" }
		return *extensions
	}(), reset)
	fmt.Printf("  %s[*]%s Threads   : %s%d%s\n", cyan, reset, white, *threads, reset)
	fmt.Printf("  %s[*]%s Requests  : %s%d%s\n", cyan, reset, white, len(urls), reset)
	fmt.Printf("\n  %s%s%-8s %-6s %-8s %s%s\n", gray, bold, "STATUS", "CODE", "SIZE", "URL", reset)
	fmt.Printf("  %s%s%s\n\n", gray, strings.Repeat("─", 70), reset)

	// output file
	var outFile *os.File
	if *output != "" {
		outFile, err = os.Create(*output)
		if err != nil {
			fmt.Printf("%s[!]%s Cannot create output file: %v\n", red, reset, err)
			os.Exit(1)
		}
		defer outFile.Close()
	}

	// run scan
	client := newClient(*timeout, !*noFollow)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		found   int32
		scanned int32
		total   = int32(len(urls))
	)

	sem := make(chan struct{}, *threads)
	startTime := time.Now()

	// progress ticker
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				s := atomic.LoadInt32(&scanned)
				f := atomic.LoadInt32(&found)
				pct := float64(s) / float64(total) * 100
				fmt.Printf("\r  %s[~]%s Progress: %s%.1f%%%s (%d/%d) | Found: %s%d%s    ",
					gray, reset, cyan, pct, reset, s, total, green, f, reset)
			case <-done:
				return
			}
		}
	}()

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()

			result := probe(client, url, *userAgent, matchCodes, ignoreCodes)
			atomic.AddInt32(&scanned, 1)

			if result != nil {
				atomic.AddInt32(&found, 1)
				mu.Lock()
				fmt.Printf("\r%s\n", strings.Repeat(" ", 80)) // clear progress line
				printResult(*result)
				if outFile != nil {
					fmt.Fprintf(outFile, "[%d] %s (size: %d)\n", result.StatusCode, result.URL, result.Size)
				}
				mu.Unlock()
			}
		}(u)
	}

	wg.Wait()
	ticker.Stop()
	close(done)

	// summary
	elapsed := time.Since(startTime)
	fmt.Printf("\r%s\n", strings.Repeat(" ", 80))
	fmt.Printf("\n  %s%s%s\n", gray, strings.Repeat("─", 70), reset)
	fmt.Printf("  %s[✓]%s Done! Found %s%d%s results in %s%.2fs%s\n",
		green, reset, green, found, reset, cyan, elapsed.Seconds(), reset)
	if *output != "" {
		fmt.Printf("  %s[✓]%s Saved to: %s%s%s\n", green, reset, white, *output, reset)
	}
	fmt.Println()
}
