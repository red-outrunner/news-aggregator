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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	// "fyne.io/fyne/v2/storage"
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

const apiKeyFile = "apikey.txt"

func saveAPIKeyToFile(key string) error {
	path, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	filePath := filepath.Join(path, apiKeyFile)
	return ioutil.WriteFile(filePath, []byte(key), 0600)
}

func loadAPIKeyFromFile() (string, error) {
	path, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(path, apiKeyFile)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
	myApp := app.New()
	myWindow := myApp.NewWindow("Ubomvu News")
	myWindow.Resize(fyne.NewSize(700, 600))

	var apiKey string
	storedKey, err := loadAPIKeyFromFile()
	if err == nil && storedKey != "" {
		apiKey = storedKey
	}

	results := container.NewVBox()
	scroll := container.NewVScroll(results)

	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search for news topics...")

	sortAsc := false
	sortBtn := widget.NewButton("Sort: New → Old", nil)

	searchBtn := widget.NewButton("Search", nil)

	searchFunc := func() {
		query := queryInput.Text
		if apiKey == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("⚠️ API key required.")}
			results.Refresh()
			return
		}
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
	}

	searchBtn.OnTapped = searchFunc

	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Old → New")
		} else {
			sortBtn.SetText("Sort: New → Old")
		}
		searchFunc()
	}

	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("Enter your NewsAPI key")

	useOnceBtn := widget.NewButton("Use Once", func() {
		apiKey = apiKeyEntry.Text
		if apiKey == "" {
			dialog.ShowError(fmt.Errorf("API key cannot be empty"), myWindow)
			return
		}
		dialog.ShowInformation("Ready", "API key will be used for this session only.", myWindow)
	})

	keepBtn := widget.NewButton("Keep Key", func() {
		apiKey = apiKeyEntry.Text
		if apiKey == "" {
			dialog.ShowError(fmt.Errorf("API key cannot be empty"), myWindow)
			return
		}
		err := saveAPIKeyToFile(apiKey)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		dialog.ShowInformation("Saved", "API key saved and will be used automatically.", myWindow)
	})

	apiKeyBox := container.NewVBox(
		widget.NewLabel("Enter NewsAPI Key:"),
				       apiKeyEntry,
				       container.NewHBox(useOnceBtn, keepBtn),
	)

	topBar := container.NewVBox(
		apiKeyBox,
		container.New(layout.NewFormLayout(),
			      widget.NewLabel(""), queryInput,
			      widget.NewLabel(""), container.NewHBox(searchBtn, sortBtn),
		),
	)

	content := container.NewBorder(topBar, nil, nil, nil, scroll)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
