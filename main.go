package main

import (
	"flag"
	"fmt"
	"time"

	"log"
	"net/http"
	"text/template"

	"github.com/anaskhan96/soup"
)

func main() {
	port := flag.String("port", "8080", "Port to bind to")
	flag.Parse()

	soup.Headers["User-agent"] = "Mozilla/5.0"
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/league" {
			http.Error(w, "404 not found.", http.StatusNotFound)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}
		funcMap := template.FuncMap{
			"sizeDecrease": func(size float64) float64 {
				return size * 0.98
			},
		}
		t, err := template.New("koll.gohtml").Funcs(funcMap).ParseFiles("./html/koll.gohtml")
		if err != nil {
			log.Println(err)
		}
		pending, currentStreak, bestStreak, weeks := fetchHistory()
		err = t.Execute(w, result{Pending: pending, CurrentStreak: currentStreak, BestStreak: bestStreak, Weeks: weeks})
		if err != nil {
			fmt.Println(err)
		}
	}
	http.HandleFunc("/", handler)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
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

func fetchHistory() (pending bool, currentStreak, bestStreak int, weeks []week) {
	morePages := true
	var entries []entry
	for page := 1; morePages; page++ {
		response, err := soup.Get(fmt.Sprintf("http://www.thebearden.se/news?page=%d", page))
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		parsedHtml := soup.HTMLParse(response)
		for _, node := range parsedHtml.FindAll("div", "id", "newsbox") {
			title := node.Find("h3").Text()
			date, _ := time.ParseInLocation("02/01/2006", node.Find("p", "class", "inline").Text(), time.Local)
			entries = append(entries, entry{title, date})
		}
		older := parsedHtml.Find("input", "value", "Older")
		if older.Error != nil {
			morePages = false
		}
	}

	now := time.Now()
	for then, entryIndex := entries[len(entries)-1].date, 0; then.Before(now) || sameWeek(then, now); now = now.Add(-7 * 24 * time.Hour) {
		year, weeknum := now.ISOWeek()
		var entriesForWeek []entry
		for ; entryIndex < len(entries) && sameWeek(entries[entryIndex].date, now); entryIndex++ {
			entriesForWeek = append(entriesForWeek, entries[entryIndex])
		}
		weeks = append(weeks, week{Year: year, Week: weeknum, Entries: entriesForWeek})
	}
	pending = len(weeks[0].Entries) == 0
	countingStreak := 0
	brokenStreak := false
	for i, week := range weeks {
		if len(week.Entries) == 0 {
			countingStreak = 0
			if i != 0 {
				brokenStreak = true
			}
		} else {
			if !brokenStreak {
				currentStreak++
			}
			countingStreak++
			if countingStreak > bestStreak {
				bestStreak = countingStreak
			}
		}
	}
	return
}

func sameWeek(a, b time.Time) bool {
	aYear, aWeek := a.ISOWeek()
	bYear, bWeek := b.ISOWeek()
	return aYear == bYear && aWeek == bWeek
}
