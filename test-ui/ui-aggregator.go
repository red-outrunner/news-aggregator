package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

func sortByLatest(articles []Article) {
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PublishedAt > articles[j].PublishedAt
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
	err = json.Unmarshal(body, &newsResponse)
	if err != nil {
		return nil, err
	}

	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("failed to fetch news: %s", string(body))
	}

	return newsResponse.Articles, nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("News Aggregator")
	myWindow.Resize(fyne.NewSize(600, 500))

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter topic...")

	results := container.NewVBox()

	scroll := container.NewVScroll(results)
	searchButton := widget.NewButton("Search", func() {
		apiKey := "YOUR_NEWS_API_KEY"
		if apiKey == "YOUR_NEWS_API_KEY" {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("Replace API key in code!")}
			results.Refresh()
			return
		}
		articles, err := fetchNews(apiKey, input.Text)
		if err != nil {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("Error fetching news.")}
			results.Refresh()
			return
		}
		sortByLatest(articles)
		if len(articles) > 10 {
			articles = articles[:10]
		}
		results.Objects = nil
		for _, article := range articles {
			link, _ := url.Parse(article.URL)
			linkButton := widget.NewHyperlink(article.Title, link)
			results.Add(linkButton)
		}
		results.Refresh()
	})

	content := container.NewBorder(
		container.NewVBox(input, searchButton),
		nil, nil, nil,
		scroll,
	)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
