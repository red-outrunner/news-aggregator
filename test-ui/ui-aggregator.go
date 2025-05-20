package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	"fyne.io/fyne/v2/theme"
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

func fetchNews(apiKey, query string, page int) ([]Article, int, error) {
	url := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&sortBy=publishedAt&language=en&pageSize=18&page=%d&apiKey=%s", query, page, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var newsResponse NewsResponse
	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, 0, err
	}
	if newsResponse.Status != "ok" {
		return nil, 0, fmt.Errorf("API error: %s", body)
	}
	return newsResponse.Articles, newsResponse.TotalResults, nil
}

func loadSavedKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".config", "apikey.txt")
	data, err := os.ReadFile(path)
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
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "apikey.txt")
	return os.WriteFile(path, []byte(key), 0600)
}

func loadThemePreference() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false // Default to light theme
	}
	path := filepath.Join(home, ".config", "news_theme.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return false // Default to light theme
	}
	return strings.TrimSpace(string(data)) == "dark"
}

func saveThemePreference(isDark bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "news_theme.txt")
	theme := "light"
	if isDark {
		theme = "dark"
	}
	return os.WriteFile(path, []byte(theme), 0600)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("News Aggregator")
	myWindow.Resize(fyne.NewSize(700, 600))

	// Load theme preference
	isDarkTheme := loadThemePreference()
	if isDarkTheme {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

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

	// Load More button
	loadMoreBtn := widget.NewButton("Load More", nil)
	loadMoreBtn.Hide() // Hidden until search results are loaded
	loadMoreContainer := container.NewCenter(loadMoreBtn)

	// State for pagination
	var currentPage = 1
	var totalResults = 0
	var allArticles []Article
	var lastQuery string
	var lastSortAsc bool

	sortAsc := false
	sortBtn := widget.NewButton("Sort: New → Old", nil)

	// Theme toggle button
	themeBtn := widget.NewButton("Toggle Dark/Light", func() {
		isDarkTheme = !isDarkTheme
		if isDarkTheme {
			myApp.Settings().SetTheme(theme.DarkTheme())
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
		}
		saveThemePreference(isDarkTheme)
	})

	// Define searchBtn before use
	searchBtn := widget.NewButton("Search", func() {
		key := keyInput.Text
		query := queryInput.Text
		if key == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("⚠️ API key required.")}
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		// Reset pagination for new search
		currentPage = 1
		allArticles = nil
		lastQuery = query
		lastSortAsc = sortAsc
		articles, total, err := fetchNews(key, query, currentPage)
		if err != nil {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("❌ Failed to fetch news.")}
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		if len(articles) == 0 {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("🔍 No results found — try tweaking your search query.")}
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		totalResults = total
		allArticles = articles
		sortByTime(allArticles, sortAsc)
		results.Objects = nil
		for _, a := range allArticles {
			link, _ := url.Parse(a.URL)
			vbox := container.NewVBox(
				widget.NewHyperlink(fmt.Sprintf("[%s] %s", humanTime(a.PublishedAt), a.Title), link),
						  widget.NewLabel(a.Description),
			)
			results.Add(vbox)
		}
		// Show Load More button if more results are available
		if len(allArticles) < totalResults {
			loadMoreBtn.Show()
		} else {
			loadMoreBtn.Hide()
		}
		results.Refresh()
	})

	loadMoreBtn.OnTapped = func() {
		currentPage++
		articles, total, err := fetchNews(keyInput.Text, lastQuery, currentPage)
		if err != nil {
			results.Add(widget.NewLabel("❌ Failed to load more news."))
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		if len(articles) == 0 {
			results.Add(widget.NewLabel("📪 No more articles available."))
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		totalResults = total
		allArticles = append(allArticles, articles...)
		sortByTime(allArticles, lastSortAsc)
		results.Objects = nil
		for _, a := range allArticles {
			link, _ := url.Parse(a.URL)
			vbox := container.NewVBox(
				widget.NewHyperlink(fmt.Sprintf("[%s] %s", humanTime(a.PublishedAt), a.Title), link),
						  widget.NewLabel(a.Description),
			)
			results.Add(vbox)
		}
		// Update Load More button visibility
		if len(allArticles) < totalResults {
			loadMoreBtn.Show()
		} else {
			loadMoreBtn.Hide()
		}
		results.Refresh()
	}

	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Old → New")
		} else {
			sortBtn.SetText("Sort: New → Old")
		}
		lastSortAsc = sortAsc
		if len(allArticles) > 0 {
			sortByTime(allArticles, sortAsc)
			results.Objects = nil
			for _, a := range allArticles {
				link, _ := url.Parse(a.URL)
				vbox := container.NewVBox(
					widget.NewHyperlink(fmt.Sprintf("[%s] %s", humanTime(a.PublishedAt), a.Title), link),
							  widget.NewLabel(a.Description),
				)
				results.Add(vbox)
			}
			results.Refresh()
		}
	}

	useOnceBtn := widget.NewButton("Use Once", func() {
		searchBtn.OnTapped()
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
			      widget.NewLabel(""), container.NewHBox(searchBtn, sortBtn, themeBtn),
		),
	)

	content := container.NewBorder(topBar, loadMoreContainer, nil, nil, scroll)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
