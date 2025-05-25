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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Article struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	PublishedAt     string `json:"publishedAt"`
	ImpactScore     int    `json:"impactScore,omitempty"`
	PolicyProbability int `json:"policyProbability,omitempty"`
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

	for i := range newsResponse.Articles {
		newsResponse.Articles[i].ImpactScore = calculateImpactScore(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
		newsResponse.Articles[i].PolicyProbability = calculatePolicyProbability(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
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
		return false
	}
	path := filepath.Join(home, ".config", "news_theme.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return false
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

func summarizeText(text string) string {
	if text == "" {
		return "No content available."
	}
	sentences := strings.Split(text, ".")
	if len(sentences) == 0 {
		return text
	}
	count := 0
	var summary strings.Builder
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			summary.WriteString(s + ".")
			count++
			if count >= 2 {
				break
			}
		}
	}
	result := summary.String()
	if result == "" {
		return text[:min(100, len(text))] + "..."
	}
	return result
}

func calculateImpactScore(text string) int {
	keywords := []string{"crisis", "breakthrough", "disaster", "economy", "war", "pandemic", "reform"}
	score := 0
	text = strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(text, k) {
			score += 20
		}
	}
	return min(100, score)
}

func calculatePolicyProbability(text string) int {
	keywords := []string{"policy", "regulation", "law", "government", "legislation", "bill", "congress"}
	score := 0
	text = strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(text, k) {
			score += 25
		}
	}
	return min(100, score)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func askAI(question string, articles []Article) string {
	question = strings.ToLower(question)
	var results strings.Builder
	for _, a := range articles {
		content := strings.ToLower(a.Title + " " + a.Description)
		if strings.Contains(content, question) {
			results.WriteString(fmt.Sprintf("- %s: %s\n", a.Title, summarizeText(a.Description)))
		}
	}
	if results.String() == "" {
		return "No relevant information found in the articles."
	}
	return results.String()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("News Aggregator")
	myWindow.Resize(fyne.NewSize(700, 700))

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
	askAIInput := widget.NewEntry()
	askAIInput.SetPlaceHolder("Ask a question about the articles...")

	results := container.NewVBox()
	scroll := container.NewVScroll(results)

	loadMoreBtn := widget.NewButton("Load More", nil)
	loadMoreBtn.Hide()
	loadMoreContainer := container.NewCenter(loadMoreBtn)

	var currentPage = 1
	var totalResults = 0
	var allArticles []Article

	sortAsc := false
	sortBtn := widget.NewButton("Sort: New → Old", nil)
	themeBtn := widget.NewButton("Toggle Dark/Light", func() {
		isDarkTheme = !isDarkTheme
		if isDarkTheme {
			myApp.Settings().SetTheme(theme.DarkTheme())
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
		}
		saveThemePreference(isDarkTheme)
	})

	exportBtn := widget.NewButton("Export to Markdown", func() {
		var sb strings.Builder
		sb.WriteString("# News Articles\n\n")
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n", a.Title))
			sb.WriteString(fmt.Sprintf("- **URL**: %s\n", a.URL))
			sb.WriteString(fmt.Sprintf("- **Published**: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", a.Description))
			sb.WriteString(fmt.Sprintf("- **Impact Score**: %d\n", a.ImpactScore))
			sb.WriteString(fmt.Sprintf("- **Policy Probability**: %d%%\n", a.PolicyProbability))
			sb.WriteString("\n")
		}
		home, _ := os.UserHomeDir()
		path := filepath.Join(home, "news_export.md")
		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
			myApp.SendNotification(&fyne.Notification{
				Title:   "Export Error",
				Content: err.Error(),
			})
			return
		}
		myApp.SendNotification(&fyne.Notification{
			Title:   "Export Success",
			Content: fmt.Sprintf("Exported to %s", path),
		})
	})

	clipboardBtn := widget.NewButton("Copy to Clipboard", func() {
		var sb strings.Builder
		sb.WriteString("# News Articles\n\n")
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n", a.Title))
			sb.WriteString(fmt.Sprintf("- **URL**: %s\n", a.URL))
			sb.WriteString(fmt.Sprintf("- **Published**: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", a.Description))
			sb.WriteString(fmt.Sprintf("- **Impact Score**: %d\n", a.ImpactScore))
			sb.WriteString(fmt.Sprintf("- **Policy Probability**: %d%%\n", a.PolicyProbability))
			sb.WriteString("\n")
		}
		myWindow.Clipboard().SetContent(sb.String())
		myApp.SendNotification(&fyne.Notification{
			Title:   "Clipboard Success",
			Content: "Articles copied to clipboard",
		})
	})

	// Moved showSummaryDialog here to ensure it's defined before use
	showSummaryDialog := func(content string) {
		var dlg *widget.PopUp
		dlg = widget.NewModalPopUp(
			container.NewVBox(
				widget.NewLabelWithStyle("Result:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					  widget.NewLabel(content),
					  widget.NewButton("Close", func() {
						  dlg.Hide()
					  }),
			),
			myWindow.Canvas(),
		)
		dlg.Show()
	}

	askBtn := widget.NewButton("Ask AI", func() {
		question := askAIInput.Text
		if question == "" {
			myApp.SendNotification(&fyne.Notification{
				Title:   "Ask AI Error",
				Content: "Please enter a question",
			})
			return
		}
		answer := askAI(question, allArticles)
		showSummaryDialog(answer)
	})

	var lastQuery string

	refreshResultsUI := func() {
		results.Objects = nil
		for _, a := range allArticles {
			link, _ := url.Parse(a.URL)
			summarizeBtn := widget.NewButtonWithIcon("Summarize", theme.DocumentCreateIcon(), nil)
			articleText := a.Description
			if articleText == "" {
				articleText = a.Title
			}

			summarizeBtn.OnTapped = func() {
				summary := summarizeText(articleText)
				showSummaryDialog(fmt.Sprintf("Summary: %s\nImpact Score: %d\nPolicy Probability: %d%%", summary, a.ImpactScore, a.PolicyProbability))
			}

			vbox := container.NewVBox(
				widget.NewHyperlink(fmt.Sprintf("[%s] %s (Impact: %d, Policy: %d%%)", humanTime(a.PublishedAt), a.Title, a.ImpactScore, a.PolicyProbability), link),
						  widget.NewLabel(a.Description),
						  summarizeBtn,
			)
			results.Add(vbox)
		}
		results.Refresh()
	}

	searchBtn := widget.NewButton("Search", func() {
		key := keyInput.Text
		query := queryInput.Text
		if key == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabel("⚠️ API key required.")}
			results.Refresh()
			loadMoreBtn.Hide()
			return
		}
		currentPage = 1
		allArticles = nil
		lastQuery = query

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
		refreshResultsUI()
		if len(allArticles) < totalResults {
			loadMoreBtn.Show()
		} else {
			loadMoreBtn.Hide()
		}
		saveAPIKey(key)
	})

	loadMoreBtn.OnTapped = func() {
		currentPage++
		key := keyInput.Text
		query := lastQuery
		articles, _, err := fetchNews(key, query, currentPage)
		if err != nil {
			myApp.SendNotification(&fyne.Notification{
				Title:   "Load More Error",
				Content: err.Error(),
			})
			return
		}
		allArticles = append(allArticles, articles...)
		sortByTime(allArticles, sortAsc)
		refreshResultsUI()
		if len(allArticles) >= totalResults {
			loadMoreBtn.Hide()
		}
	}

	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Old → New")
		} else {
			sortBtn.SetText("Sort: New → Old")
		}
		sortByTime(allArticles, sortAsc)
		refreshResultsUI()
	}

	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(keyInput, themeBtn),
				  container.NewHBox(queryInput, searchBtn, sortBtn, exportBtn, clipboardBtn),
				  container.NewHBox(askAIInput, askBtn),
		),
		loadMoreContainer,
		nil,
		nil,
		scroll,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
