package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
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

func main() {
	apiKey := "YOUR_NEWS_API_KEY"
	if apiKey == "YOUR_NEWS_API_KEY" {
		fmt.Println("⚠️  Please replace 'YOUR_NEWS_API_KEY' with your actual NewsAPI key.")
		return
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("News Aggregator")
	myWindow.Resize(fyne.NewSize(700, 600))

	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search for news topics...")

	results := container.NewVBox()
	scroll := container.NewVScroll(results)

	sortAsc := false
	sortBtn := widget.NewButton("Sort: New → Old", nil)

	searchBtn := widget.NewButton("Search", func() {
		query := queryInput.Text
		articles, err := fetchNews(apiKey, query)
		if err != nil {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("❌ Failed to fetch news.")}
			results.Refresh()
			return
		}
		if len(articles) > 18 {
			articles = articles[:18]
		}
		sortByTime(articles, sortAsc)

		results.Objects = nil
		for _, a := range articles {
			link, _ := url.Parse(a.URL)
			results.Add(widget.NewHyperlink(fmt.Sprintf("[%s] %s", a.PublishedAt[:10], a.Title), link))
		}
		results.Refresh()
	})

	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Old → New")
		} else {
			sortBtn.SetText("Sort: New → Old")
		}
		searchBtn.OnTapped()
	}

	topBar := container.New(layout.NewFormLayout(),
		widget.NewLabel(""), queryInput,  // expand input full width
		widget.NewLabel(""), container.NewHBox(searchBtn, sortBtn),
	)


	content := container.NewBorder(topBar, nil, nil, nil, scroll)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
