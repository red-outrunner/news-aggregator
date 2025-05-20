package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PublishedAt string `json:"publishedAt"`
}

type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

func sortByTime(articles []Article, ascending bool) {
	sort.Slice(articles, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, articles[i].PublishedAt)
		t2, _ := time.Parse(time.RFC3339, articles[j].PublishedAt)
		if ascending {
			return t1.Before(t2)
		}
		return t1.After(t2)
	})
}

func humanTime(tStr string) string {
	t, err := time.Parse(time.RFC3339, tStr)
	if err != nil {
		return tStr
	}
	dur := time.Since(t)
	switch {
		case dur < time.Minute:
			return "just now"
		case dur < time.Hour:
			return fmt.Sprintf("%dm ago", int(dur.Minutes()))
		case dur < 24*time.Hour:
			return fmt.Sprintf("%dh ago", int(dur.Hours()))
		case dur < 7*24*time.Hour:
			return fmt.Sprintf("%dd ago", int(dur.Hours()/24))
		default:
			return t.Format("Jan 2")
	}
}

func fetchNews(apiKey, query string) ([]Article, error) {
	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&sortBy=publishedAt&language=en&apiKey=%s", query, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var newsResponse NewsResponse
	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, err
	}
	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("API error: %s", body)
	}
	return newsResponse.Articles, nil
}

func loadSavedKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".config", "apikey.txt")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveAPIKey(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config")
	os.MkdirAll(dir, 0700)
	path := filepath.Join(dir, "apikey.txt")
	return ioutil.WriteFile(path, []byte(key), 0600)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("News Aggregator")
	myWindow.Resize(fyne.NewSize(700, 600))

	keyInput := widget.NewEntry()
	keyInput.SetPlaceHolder("Enter your NewsAPI key...")
	apiKey := loadSavedKey()
	if apiKey != "" {
		keyInput.SetText(apiKey)
	}

	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search for news topics...")

	results := container.NewVBox()
	scroll := container.NewVScroll(results)

	sortAsc := false
	sortBtn := widget.NewButton("Sort: New → Old", nil)

	search := func() {
		key := keyInput.Text
		query := queryInput.Text
		if key == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("⚠️ API key required.")}
			results.Refresh()
			return
		}
		articles, err := fetchNews(key, query)
		if err != nil {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("❌ Failed to fetch news.")}
			results.Refresh()
			return
		}
		if len(articles) == 0 {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("🔍 No results found — try tweaking your search query.")}
			results.Refresh()
			return
		}
		sortByTime(articles, sortAsc)
		if len(articles) > 18 {
			articles = articles[:18]
		}
		results.Objects = nil
		for _, a := range articles {
			link, _ := url.Parse(a.URL)
			vbox := container.NewVBox(
				widget.NewHyperlink(fmt.Sprintf("[%s] %s", humanTime(a.PublishedAt), a.Title), link),
						  widget.NewLabel(a.Description),
			)
			results.Add(vbox)
		}
		results.Refresh()
	}

	searchBtn := widget.NewButton("Search", search)

	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Old → New")
		} else {
			sortBtn.SetText("Sort: New → Old")
		}
		search()
	}

	useOnceBtn := widget.NewButton("Use Once", func() {
		search()
	})

	keepBtn := widget.NewButton("Keep Key", func() {
		err := saveAPIKey(keyInput.Text)
		if err != nil {
			r := container.NewVBox(widget.NewLabel("❌ Failed to save API key."))
			results.Objects = r.Objects
			results.Refresh()
		} else {
			r := container.NewVBox(widget.NewLabel("✅ API key saved."))
			results.Objects = r.Objects
			results.Refresh()
		}
	})

	topBar := container.NewVBox(
		container.New(layout.NewFormLayout(),
			      widget.NewLabel("API Key:"), keyInput,
			      widget.NewLabel(""), container.NewHBox(useOnceBtn, keepBtn),
			      widget.NewLabel("Query:"), queryInput,
			      widget.NewLabel(""), container.NewHBox(searchBtn, sortBtn),
		),
	)

	content := container.NewBorder(topBar, nil, nil, nil, scroll)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
