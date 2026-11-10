package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/anaskhan96/soup"
	"golang.org/x/net/html/charset"
)

const (
	newsURL = "https://www.thebearden.se/news?page=%d"

	// How long a successful scrape is reused before thebearden.se is scraped again.
	refreshInterval = 10 * time.Minute
	// How long to wait before retrying after a failed scrape.
	retryInterval = time.Minute

	// Limits on what a single scrape may do, so a misbehaving remote site can't
	// make it run forever or eat all memory.
	fetchTimeout = 15 * time.Second
	maxPageBytes = 5 << 20
	maxPages     = 1000
)

var httpClient = &http.Client{Timeout: fetchTimeout}

func main() {
	port := flag.String("port", "8080", "Port to bind to")
	flag.Parse()

	funcMap := template.FuncMap{
		"sizeDecrease": func(size float64) float64 {
			return size * 0.98
		},
	}
	t := template.Must(template.New("koll.gohtml").Funcs(funcMap).ParseFiles("./html/koll.gohtml"))

	var history historyCache
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/league" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}
		res, err := history.get()
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not fetch news from thebearden.se, try again later.", http.StatusServiceUnavailable)
			return
		}
		err = t.Execute(w, res)
		if err != nil {
			log.Println(err)
		}
	}
	http.HandleFunc("/", handler)

	server := &http.Server{
		Addr:              ":" + *port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		// A request that arrives while the cache is cold waits for a full scrape.
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}
	log.Fatal(server.ListenAndServe())
}

// historyCache holds the latest scrape result so that incoming requests don't
// each trigger a full scrape of thebearden.se.
type historyCache struct {
	mu          sync.Mutex
	data        *result
	nextRefresh time.Time
}

func (c *historyCache) get() (*result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.nextRefresh) {
		if c.data == nil {
			return nil, errors.New("no data yet, waiting before retrying scrape")
		}
		return c.data, nil
	}

	// Set before scraping, so a scrape that fails or panics isn't retried by
	// every following request.
	c.nextRefresh = time.Now().Add(retryInterval)
	data, err := fetchHistory()
	if err != nil {
		if c.data == nil {
			return nil, err
		}
		log.Printf("serving cached data, scrape failed: %v", err)
		return c.data, nil
	}
	c.data = data
	c.nextRefresh = time.Now().Add(refreshInterval)
	return c.data, nil
}

type result struct {
	Pending       bool
	CurrentStreak int
	BestStreak    int
	Weeks         []week
}

type week struct {
	Year    int
	Week    int
	Entries []entry
}

type entry struct {
	Title string
	date  time.Time
}

func fetchPage(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	body, err := charset.NewReader(io.LimitReader(resp.Body, maxPageBytes), resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func fetchHistory() (*result, error) {
	var entries []entry
	for page := 1; ; page++ {
		if page > maxPages {
			return nil, fmt.Errorf("gave up after %d pages", maxPages)
		}
		url := fmt.Sprintf(newsURL, page)
		response, err := fetchPage(url)
		if err != nil {
			return nil, err
		}
		parsedHtml := soup.HTMLParse(response)
		if parsedHtml.Error != nil {
			return nil, fmt.Errorf("parsing %s: %w", url, parsedHtml.Error)
		}
		for _, node := range parsedHtml.FindAll("div", "id", "newsbox") {
			titleNode := node.Find("h3")
			dateNode := node.Find("p", "class", "inline")
			if titleNode.Error != nil || dateNode.Error != nil {
				log.Printf("skipping malformed news entry on %s", url)
				continue
			}
			date, err := time.ParseInLocation("02/01/2006", dateNode.Text(), time.Local)
			if err != nil {
				log.Printf("skipping news entry with bad date on %s: %v", url, err)
				continue
			}
			entries = append(entries, entry{titleNode.Text(), date})
		}
		if parsedHtml.Find("input", "value", "Older").Error != nil {
			break
		}
	}
	if len(entries) == 0 {
		return nil, errors.New("no news entries found on thebearden.se")
	}

	res := &result{}
	now := time.Now()
	for then, entryIndex := entries[len(entries)-1].date, 0; then.Before(now) || sameWeek(then, now); now = now.Add(-7 * 24 * time.Hour) {
		year, weeknum := now.ISOWeek()
		var entriesForWeek []entry
		for ; entryIndex < len(entries) && sameWeek(entries[entryIndex].date, now); entryIndex++ {
			entriesForWeek = append(entriesForWeek, entries[entryIndex])
		}
		res.Weeks = append(res.Weeks, week{Year: year, Week: weeknum, Entries: entriesForWeek})
	}
	res.Pending = len(res.Weeks[0].Entries) == 0
	countingStreak := 0
	brokenStreak := false
	for i, week := range res.Weeks {
		if len(week.Entries) == 0 {
			countingStreak = 0
			if i != 0 {
				brokenStreak = true
			}
		} else {
			if !brokenStreak {
				res.CurrentStreak++
			}
			countingStreak++
			if countingStreak > res.BestStreak {
				res.BestStreak = countingStreak
			}
		}
	}
	return res, nil
}

func sameWeek(a, b time.Time) bool {
	aYear, aWeek := a.ISOWeek()
	bYear, bWeek := b.ISOWeek()
	return aYear == bYear && aWeek == bWeek
}
