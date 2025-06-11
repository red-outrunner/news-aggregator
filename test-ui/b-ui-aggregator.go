package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Article struct updated to include URLToImage and SentimentScore
type Article struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	URL               string   `json:"url"`
	URLToImage        string   `json:"urlToImage"`
	PublishedAt       string   `json:"publishedAt"`
	ImpactScore       int      `json:"impactScore,omitempty"`
	PolicyProbability int      `json:"policyProbability,omitempty"`
	SentimentScore    int      `json:"sentimentScore,omitempty"` // Added for sentiment analysis
	Source            struct { // NewsAPI often nests source info
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
}

// NewsResponse struct remains the same
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

var (
	bookmarkedArticles []Article
	bookmarksMutex     sync.Mutex
	bookmarksFilePath  string
	readArticles       map[string]bool
	readArticlesMutex  sync.Mutex

	// Keyword sets for efficient lookup
	positiveKeywordsSet map[string]struct{}
	negativeKeywordsSet map[string]struct{}

	// Regex for sanitizing filenames, pre-compiled for efficiency
	safeFilenameRegex *regexp.Regexp
)

const bookmarksFilename = "news_aggregator_bookmarks.json"

// Sentiment keywords (simple lists for demonstration)
var positiveKeywords = []string{
	// General Positive
	"good", "great", "excellent", "positive", "success", "improve", "benefit", "effective", "strong", "happy", "joy", "love", "optimistic", "favorable", "promising", "encouraging",
	// Growth & Expansion
	"grow", "growth", "expansion", "expand", "increase", "surge", "rise", "upward", "upturn", "boom", "accelerate", "augment", "boost", "rally", "recover", "recovery",
	// Achievement & Performance
	"achieve", "achieved", "outperform", "exceed", "beat", "record", "profitable", "profit", "gains", "earnings", "revenue", "dividend", "surplus",
	// Innovation & Advancement
	"innovative", "innovation", "breakthrough", "advance", "launch", "new", "develop", "upgrade", "leading", "cutting-edge",
	// Market Sentiment & Confidence
	"bullish", "optimism", "confidence", "stable", "stability", "support", "demand", "hot", "high", "robust",
	// Deals & Approvals
	"acquire", "acquisition", "merger", "partnership", "agreement", "approve", "approved", "endorse", "confirm",
}
var negativeKeywords = []string{
	// General Negative
	"bad", "poor", "terrible", "negative", "fail", "failure", "weak", "adverse", "sad", "angry", "fear", "pessimistic", "unfavorable", "discouraging",
	// Decline & Contraction
	"decline", "decrease", "drop", "fall", "slump", "downturn", "recession", "contraction", "reduce", "cut", "loss", "losses", "deficit", "shrink", "erode", "weaken",
	// Problems & Risks
	"crisis", "disaster", "risk", "warn", "warning", "threat", "problem", "issue", "concern", "challenge", "obstacle", "difficulty", "uncertainty", "volatile", "volatility",
	// Poor Performance
	"underperform", "miss", "shortfall", "struggle", "stagnate", "delay", "halt",
	// Market Sentiment & Lack of Confidence
	"bearish", "pessimism", "doubt", "skepticism", "unstable", "instability", "pressure", "low", "oversupply", "bubble",
	// Legal & Regulatory Issues
	"investigation", "lawsuit", "penalty", "fine", "sanction", "ban", "fraud", "scandal", "recall", "dispute", "reject", "denied", "downgrade",
}

// Market-driving keywords for impact score calculation
var impactScoreKeywords = []string{
	// General Market Drivers
	"recession", "inflation", "interest rates", "market crash", "trade war", "supply chain", "corporate earnings", "acquisition", "ipo", "federal reserve", "economic growth", "unemployment", "government stimulus", "new regulation", "geopolitical risk", "tariff", "sanction", "deficit", "surplus", "bankruptcy", "takeover", "merger", "venture capital", "private equity", "stock market", "bond market", "currency fluctuation", "commodity prices", "consumer spending", "housing market", "energy crisis", "financial crisis", "debt ceiling", "quantitative easing", "fiscal policy", "monetary policy", "trade deal", "market sentiment", "volatility", "correction", "bear market", "bull market", "earnings report", "profit warning", "economic forecast", "global economy", "emerging markets",
	// EU Specific Market Drivers
	"ecb", "european central bank", "eurozone", "brexit impact", "eu stimulus", "recovery fund", "dax", "cac 40", "ftse mib", "euro stoxx 50", "sovereign debt", "eu bailout", "esm", "european stability mechanism",
	// South African Specific Market Drivers
	"sarb", "south african reserve bank", "jse", "johannesburg stock exchange", "rand", "load shedding", "eskom", "mining sector sa", "sa budget", "credit rating south africa", "foreign direct investment sa", "state owned enterprises sa", "bee impact",
}

// Policy-related keywords for probability calculation
var policyKeywords = []string{
	// General & US Focused (many are broadly applicable)
	"policy", "regulation", "law", "government", "legislation", "bill", "congress", "senate", "parliament", "decree", "treaty", "court", "ruling", "initiative", "mandate", "executive order", "tariff", "sanction", "subsidy", "public policy", "compliance", "enforcement", "oversight", "hearing", "testimony", "budget", "appropriation", "act", "statute", "ordinance", "directive", "guideline", "framework", "accord", "pact", "resolution", "referendum", "lobbying", "advocacy", "think tank", "white paper", "federal", "state", "local government", "agency", "commission", "authority", "irs", "federal reserve", "supreme court", "white house", "capitol hill", "reform", "governance", "judiciary",
	// EU Specific
	"european parliament", "european commission", "council of the european union", "eu directive", "eu regulation", "mep", "ecj", "eurozone", "brussels", "ecb", "european central bank", "single market", "schengen", "brexit", "article 50", "eusl", "gdpr",
	// UK Specific (examples, as often related to EU context or distinct major economy)
	"parliament uk", "house of commons", "house of lords", "downing street", "hmrc", "bank of england", "chancellor",
	// South African Specific
	"parliament sa", "national assembly sa", "national council of provinces", "ncop", "sars", "south african revenue service", "constitutional court sa", "concourt", "provincial government sa", "cabinet sa", "cosatu", "public protector sa", "union buildings", "south african reserve bank", "sarb", "anc", "da", "eff", "state capture", "bee", "black economic empowerment", "land reform", "national development plan", "ndp", "municipal",
	// Other International Bodies/Terms
	"united nations", "unsc", "world bank", "imf", "wto", "who", "icc", "g7", "g20", "oecd",
	// Generic legislative body names from other major regions (examples)
	"bundestag", // Germany
	"diet",      // Japan
	"duma",      // Russia
}

func init() {
	positiveKeywordsSet = make(map[string]struct{}, len(positiveKeywords))
	for _, k := range positiveKeywords {
		positiveKeywordsSet[strings.ToLower(k)] = struct{}{}
	}

	negativeKeywordsSet = make(map[string]struct{}, len(negativeKeywords))
	for _, k := range negativeKeywords {
		negativeKeywordsSet[strings.ToLower(k)] = struct{}{}
	}
	// For impactScoreKeywords and policyKeywords, they contain phrases and use strings.Contains.
	// The main optimization (ToLower once on text) is already in place for those functions.

	// Pre-compile regex for filename sanitization
	safeFilenameRegex = regexp.MustCompile(`[^\w-]`)
}

// --- Utility Functions (Sorting, Time, etc.) ---
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

func sortBySentiment(articles []Article, ascending bool) { // ascending true means Low to High
	sort.Slice(articles, func(i, j int) bool {
		if ascending {
			return articles[i].SentimentScore < articles[j].SentimentScore
		}
		return articles[i].SentimentScore > articles[j].SentimentScore
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
		return t.Format("Jan 2, 2006")
	}
}

func summarizeText(text string) string {
	if strings.TrimSpace(text) == "" {
		return "No content available to summarize."
	}
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	if len(sentences) == 0 {
		maxLength := 150
		if len(text) <= maxLength {
			return text
		}
		if idx := strings.LastIndex(text[:maxLength], " "); idx != -1 {
			return text[:idx] + "..."
		}
		return text[:maxLength] + "..."
	}
	var summary strings.Builder
	sentenceCount := 0
	desiredSentences := 2
	originalTextIndex := 0
	for _, s := range sentences {
		trimmedSentence := strings.TrimSpace(s)
		if trimmedSentence != "" {
			actualSentenceStart := strings.Index(text[originalTextIndex:], trimmedSentence)
			if actualSentenceStart != -1 {
				actualSentenceEnd := actualSentenceStart + len(trimmedSentence)
				summary.WriteString(text[originalTextIndex+actualSentenceStart : originalTextIndex+actualSentenceEnd])
				if originalTextIndex+actualSentenceEnd < len(text) {
					punctuation := text[originalTextIndex+actualSentenceEnd]
					if punctuation == '.' || punctuation == '!' || punctuation == '?' {
						summary.WriteRune(rune(punctuation))
					} else {
						summary.WriteString(".")
					}
				} else {
					summary.WriteString(".")
				}
				originalTextIndex += actualSentenceEnd + 1
			} else {
				summary.WriteString(trimmedSentence)
				summary.WriteString(".")
			}
			summary.WriteString(" ")
			sentenceCount++
			if sentenceCount >= desiredSentences {
				break
			}
		}
	}
	result := strings.TrimSpace(summary.String())
	if result == "" {
		maxLength := 150
		if len(text) <= maxLength {
			return text
		}
		if idx := strings.LastIndex(text[:maxLength], " "); idx != -1 {
			return text[:idx] + "..."
		}
		return text[:maxLength] + "..."
	}
	return result
}

// --- Core Logic (Fetch, Sentiment, etc.) ---
func calculateSentimentScore(text string) int {
	if text == "" {
		return 0
	}
	score := 0
	textLower := strings.ToLower(text)
	// Using a simple word splitter. For more complex scenarios, a more robust tokenizer might be needed.
	words := strings.FieldsFunc(textLower, func(r rune) bool {
		// Split on anything not a letter or digit
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	for _, word := range words {
		// word is already lowercase due to textLower and splitter behavior
		if _, found := positiveKeywordsSet[word]; found {
			score += 10
		}
		if _, found := negativeKeywordsSet[word]; found {
			score -= 10
		}
	}
	if score > 100 {
		score = 100
	}
	if score < -100 {
		score = -100
	}
	return score
}

func fetchNews(apiKey, query, fromDate, toDate string, page int) ([]Article, int, error) {
	baseApiURL := "https://newsapi.org/v2/everything"
	queryParams := url.Values{}
	queryParams.Add("q", query)
	queryParams.Add("sortBy", "publishedAt")
	queryParams.Add("language", "en")
	queryParams.Add("pageSize", "18")
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("apiKey", apiKey)

	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err == nil {
			queryParams.Add("from", fromDate)
		} else {
			fmt.Printf("Warning: Invalid 'from' date format: %s. Ignoring.\n", fromDate)
		}
	}
	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err == nil {
			queryParams.Add("to", toDate)
		} else {
			fmt.Printf("Warning: Invalid 'to' date format: %s. Ignoring.\n", toDate)
		}
	}

	fullApiURL := fmt.Sprintf("%s?%s", baseApiURL, queryParams.Encode())

	resp, err := http.Get(fullApiURL)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to connect to news service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var newsResponse NewsResponse
	if err := json.Unmarshal(body, &newsResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to parse news JSON: %w. Response: %s", err, string(body))
	}
	if newsResponse.Status != "ok" {
		errMsg := newsResponse.Status
		var rawResponse map[string]interface{}
		if json.Unmarshal(body, &rawResponse) == nil {
			if message, ok := rawResponse["message"].(string); ok {
				errMsg = message
			}
		} else if len(newsResponse.Articles) > 0 && newsResponse.Articles[0].Title != "" &&
			(strings.Contains(strings.ToLower(newsResponse.Articles[0].Title), "error") || newsResponse.Articles[0].Description == "") {
			errMsg = newsResponse.Articles[0].Title
		}
		return nil, 0, fmt.Errorf("API error: %s. Full response: %s", errMsg, string(body))
	}

	for i := range newsResponse.Articles {
		newsResponse.Articles[i].ImpactScore = calculateImpactScore(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
		newsResponse.Articles[i].PolicyProbability = calculatePolicyProbability(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
		newsResponse.Articles[i].SentimentScore = calculateSentimentScore(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
	}

	return newsResponse.Articles, newsResponse.TotalResults, nil
}

// --- Config & Persistence ---
func loadSavedKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home dir for API key: %v\n", err)
		return ""
	}
	path := filepath.Join(home, ".config", "news_aggregator_apikey.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error reading API key file %s: %v\n", path, err)
		}
		return ""
	}
	return strings.TrimSpace(string(data))
}
func saveAPIKey(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home dir for API key: %w", err)
	}
	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creating config directory %s: %w", dir, err)
	}
	path := filepath.Join(dir, "news_aggregator_apikey.txt")
	if err := os.WriteFile(path, []byte(key), 0600); err != nil {
		return fmt.Errorf("error writing API key to %s: %w", path, err)
	}
	return nil
}
func loadThemePreference() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home dir for theme: %v\n", err)
		return false // Default to light theme on error
	}
	path := filepath.Join(home, ".config", "news_aggregator_theme.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error reading theme preference file %s: %v\n", path, err)
		}
		return false
	}
	return strings.TrimSpace(string(data)) == "dark"
}
func saveThemePreference(isDark bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home dir for theme: %w", err)
	}
	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creating config directory %s for theme: %w", dir, err)
	}
	path := filepath.Join(dir, "news_aggregator_theme.txt")
	theme := "light"
	if isDark {
		theme = "dark"
	}
	if err := os.WriteFile(path, []byte(theme), 0600); err != nil {
		return fmt.Errorf("error writing theme preference to %s: %w", path, err)
	}
	return nil
}
func setupBookmarksPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home dir for bookmarks: %v. Using default path.\n", err)
		bookmarksFilePath = bookmarksFilename
		return
	}
	configDir := filepath.Join(home, ".config", "newsaggregator_v3")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory %s for bookmarks: %v. Using default path.\n", configDir, err)
		bookmarksFilePath = bookmarksFilename
		return
	}
	bookmarksFilePath = filepath.Join(configDir, bookmarksFilename)
}

func loadBookmarks() {
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()
	data, err := os.ReadFile(bookmarksFilePath)
	if err != nil {
		bookmarkedArticles = []Article{}
		return
	}
	if err := json.Unmarshal(data, &bookmarkedArticles); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling bookmarks from %s: %v\n", bookmarksFilePath, err)
		bookmarkedArticles = []Article{}
	}
}

func saveBookmarks() {
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()
	data, err := json.MarshalIndent(bookmarkedArticles, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling bookmarks: %v\n", err)
		return
	}
	if err := os.WriteFile(bookmarksFilePath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing bookmarks file %s: %v\n", bookmarksFilePath, err)
	}
}

func isBookmarked(articleURL string) bool {
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()
	for _, bm := range bookmarkedArticles {
		if bm.URL == articleURL {
			return true
		}
	}
	return false
}

func toggleBookmark(article Article) {
	bookmarksMutex.Lock()
	found := false
	var updatedBookmarks []Article
	for _, bm := range bookmarkedArticles {
		if bm.URL == article.URL {
			found = true
		} else {
			updatedBookmarks = append(updatedBookmarks, bm)
		}
	}
	if !found {
		updatedBookmarks = append(updatedBookmarks, article)
	}
	bookmarkedArticles = updatedBookmarks
	bookmarksMutex.Unlock()
	saveBookmarks()
}

func markAsRead(articleURL string) {
	readArticlesMutex.Lock()
	defer readArticlesMutex.Unlock()
	if readArticles == nil {
		readArticles = make(map[string]bool)
	}
	readArticles[articleURL] = true
}

func isRead(articleURL string) bool {
	readArticlesMutex.Lock()
	defer readArticlesMutex.Unlock()
	if readArticles == nil {
		return false
	}
	return readArticles[articleURL]
}

func calculateImpactScore(text string) int {
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range impactScoreKeywords {
		if strings.Contains(textLower, k) {
			score += 7
		}
	}
	return min(100, score)
}

func calculatePolicyProbability(text string) int {
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range policyKeywords {
		if strings.Contains(textLower, k) {
			score += 10
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

// --- Main Application ---
func main() {
	myApp := app.NewWithID("com.example.newsaggregator.marketresearch.v2")
	myWindow := myApp.NewWindow("News on Red:market research tool")
	myWindow.Resize(fyne.NewSize(1000, 850))

	setupBookmarksPath()
	loadBookmarks()
	readArticles = make(map[string]bool)

	isDarkTheme := loadThemePreference()
	if isDarkTheme {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	var currentPage = 1
	var totalResults = 0
	var allArticles []Article
	var lastQuery, lastFromDate, lastToDate string

	type SortMode int
	const (
		SortTimeDesc SortMode = iota
		SortTimeAsc
		SortSentimentDesc
		SortSentimentAsc
	)
	currentSortMode := SortTimeDesc

	keyInput := widget.NewPasswordEntry()
	keyInput.SetPlaceHolder("Enter NewsAPI key...")
	apiKey := loadSavedKey()
	if apiKey != "" {
		keyInput.SetText(apiKey)
	}
	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil)
	updateThemeButtonText := func(isDark bool) {
		if isDark {
			themeBtn.SetText("Light Theme")
		} else {
			themeBtn.SetText("Dark Theme")
		}
		themeBtn.Refresh()
	}
	updateThemeButtonText(isDarkTheme)
	themeBtn.OnTapped = func() {
		isDarkTheme = !isDarkTheme
		if isDarkTheme {
			myApp.Settings().SetTheme(theme.DarkTheme())
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
		}
		updateThemeButtonText(isDarkTheme)
		if errSave := saveThemePreference(isDarkTheme); errSave != nil {
			fmt.Fprintf(os.Stderr, "Error saving theme preference: %v\n", errSave)
		}
	}
	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyRow := container.NewBorder(nil, nil, apiKeyLabel, themeBtn, keyInput)
	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search news topics...")
	fromDateEntry := widget.NewEntry()
	fromDateEntry.SetPlaceHolder("From: YYYY-MM-DD")
	toDateEntry := widget.NewEntry()
	toDateEntry.SetPlaceHolder("To: YYYY-MM-DD")
	dateFilterRow := container.NewGridWithColumns(2, container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, fromDateEntry), container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, toDateEntry))
	results := container.NewVBox()
	results.Add(widget.NewLabelWithStyle("Enter API key and search query.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(300, 400))
	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide()
	loadMoreBtn := widget.NewButtonWithIcon("Load More Articles", theme.MoreVerticalIcon(), nil)
	loadMoreBtn.Hide()
	loadMoreContainer := container.NewCenter(loadMoreBtn)

	var refreshResultsUI func()
	var showBookmarksView func()
	var showTrendAnalysisDialog func()

	// Helper for RichText highlighting
	createHighlightedText := func(text, query string, highlightRegex *regexp.Regexp) *widget.RichText {
		richText := widget.NewRichText()
		if query == "" || highlightRegex == nil {
			richText.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: text}}
			return richText
		}

		matches := highlightRegex.FindAllStringIndex(text, -1)
		lastIndex := 0

		for _, match := range matches {
			start, end := match[0], match[1]
			if start > lastIndex {
				richText.Segments = append(richText.Segments, &widget.TextSegment{Text: text[lastIndex:start]})
			}
			highlightStyle := widget.RichTextStyleStrong
			richText.Segments = append(richText.Segments, &widget.TextSegment{Text: text[start:end], Style: highlightStyle})
			lastIndex = end
		}
		if lastIndex < len(text) {
			richText.Segments = append(richText.Segments, &widget.TextSegment{Text: text[lastIndex:]})
		}

		if len(richText.Segments) == 0 && text != "" {
			richText.Segments = []widget.RichTextSegment{&widget.TextSegment{Text: text}}
		}
		return richText
	}

	refreshResultsUI = func() {
		results.Objects = nil
		if len(allArticles) == 0 {
			if lastQuery != "" {
				results.Add(widget.NewLabelWithStyle("No articles for '"+lastQuery+"'.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			} else {
				results.Add(widget.NewLabelWithStyle("Enter API key & search.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			}
			results.Refresh()
			return
		}

		var currentHighlightRegex *regexp.Regexp
		if lastQuery != "" {
			queryWords := strings.Fields(strings.ToLower(lastQuery))
			var reParts []string
			for _, qw := range queryWords {
				if len(qw) > 0 { // Ensure word is not empty
					reParts = append(reParts, regexp.QuoteMeta(qw))
				}
			}
			if len(reParts) > 0 {
				currentHighlightRegex = regexp.MustCompile(`(?i)\b(` + strings.Join(reParts, "|") + `)\b`)
			}
		}

		for i := range allArticles {
			article := allArticles[i]
			parsedURL, _ := url.Parse(article.URL)

			// Pass the pre-compiled regex to the helper
			titleRichText := createHighlightedText(article.Title, lastQuery, currentHighlightRegex)
			var currentTitleStyle fyne.TextStyle
			if isRead(article.URL) {
				currentTitleStyle.Italic = true
			} else {
				currentTitleStyle.Bold = true
			}

			for j := range titleRichText.Segments {
				if ts, ok := titleRichText.Segments[j].(*widget.TextSegment); ok {
					isHighlighted := ts.Style.TextStyle.Bold // Strong sets Bold
					if isHighlighted {
						if isRead(article.URL) {
							ts.Style.TextStyle.Italic = true
						}
					} else {
						ts.Style.TextStyle = currentTitleStyle
					}
				}
			}
			titleRichText.Refresh()

			cardDescription := article.Description
			if len(cardDescription) > 180 {
				cardDescription = cardDescription[:177] + "..."
			}
			if strings.TrimSpace(cardDescription) == "" {
				cardDescription = "No description."
			}
			descriptionRichText := createHighlightedText(cardDescription, lastQuery, currentHighlightRegex)
			descriptionRichText.Wrapping = fyne.TextWrapWord

			imgWidget := canvas.NewImageFromResource(nil)
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.SetMinSize(fyne.NewSize(120, 80))
			if article.URLToImage != "" {
				go func(imgURL string, targetImg *canvas.Image) {
					client := http.Client{Timeout: 15 * time.Second}
					resp, err := client.Get(imgURL)
					if err != nil {
						return
					}
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						return
					}
					imgData, err := io.ReadAll(resp.Body)
					if err != nil {
						return
					}
					_, _, err = image.Decode(bytes.NewReader(imgData))
					if err != nil {
						return
					}
					imgRes := fyne.NewStaticResource(filepath.Base(imgURL), imgData)
					targetImg.Resource = imgRes
					targetImg.Refresh()
				}(article.URLToImage, imgWidget)
			}

			bookmarkBtn := widget.NewButtonWithIcon("", nil, nil)
			updateBookmarkButton := func(btn *widget.Button, articleURL string) {
				if isBookmarked(articleURL) {
					btn.SetIcon(theme.ConfirmIcon())
					btn.SetText("Bookmarked")
				} else {
					btn.SetIcon(theme.ContentAddIcon())
					btn.SetText("Bookmark")
				}
				btn.Refresh()
			}
			updateBookmarkButton(bookmarkBtn, article.URL)
			currentArticleForBookmark := article
			bookmarkBtn.OnTapped = func() {
				toggleBookmark(currentArticleForBookmark)
				updateBookmarkButton(bookmarkBtn, currentArticleForBookmark.URL)
			}

			detailsBtn := widget.NewButtonWithIcon("Details", theme.InfoIcon(), func() {
				currentArticleForDetail := article
				markAsRead(currentArticleForDetail.URL)
				refreshResultsUI()
				fullSummary := summarizeText(currentArticleForDetail.Description)
				if strings.TrimSpace(currentArticleForDetail.Description) == "" {
					fullSummary = "No full description."
				}
				currentArticleParsedURL, _ := url.Parse(currentArticleForDetail.URL)
				content := container.NewVBox(widget.NewLabelWithStyle(currentArticleForDetail.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewSeparator(), widget.NewLabelWithStyle("Summary:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewLabel(fullSummary), widget.NewSeparator(), widget.NewLabel(fmt.Sprintf("Impact: %d", currentArticleForDetail.ImpactScore)), widget.NewLabel(fmt.Sprintf("Policy: %d%%", currentArticleForDetail.PolicyProbability)), widget.NewLabel(fmt.Sprintf("Sentiment: %d", currentArticleForDetail.SentimentScore)), widget.NewSeparator(), widget.NewHyperlink("Open Original", currentArticleParsedURL))
				for _, obj := range content.Objects {
					if lbl, ok := obj.(*widget.Label); ok {
						lbl.Wrapping = fyne.TextWrapWord
					}
				}
				var detailPopUp *widget.PopUp
				closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { detailPopUp.Hide() })
				dialogContainer := container.NewBorder(nil, container.NewCenter(closeButton), nil, nil, container.NewVScroll(content))
				detailPopUp = widget.NewModalPopUp(dialogContainer, myWindow.Canvas())
				detailPopUp.Resize(fyne.NewSize(myWindow.Canvas().Size().Width*0.8, myWindow.Canvas().Size().Height*0.7))
				detailPopUp.Show()
			})

			sentimentLabel := widget.NewLabel(fmt.Sprintf("Sentiment: %d", article.SentimentScore))
			if article.SentimentScore > 20 {
				sentimentLabel.Importance = widget.SuccessImportance
			} else if article.SentimentScore < -20 {
				sentimentLabel.Importance = widget.DangerImportance
			} else {
				sentimentLabel.Importance = widget.MediumImportance
			}

			textContent := container.NewVBox(descriptionRichText, widget.NewSeparator(), container.NewGridWithColumns(3, widget.NewLabel(fmt.Sprintf("Impact: %d", article.ImpactScore)), widget.NewLabel(fmt.Sprintf("Policy: %d%%", article.PolicyProbability)), sentimentLabel), widget.NewSeparator(), container.NewHBox(widget.NewHyperlink("Read Full Article", parsedURL), layout.NewSpacer(), bookmarkBtn, detailsBtn))
			cardContentWithImage := container.NewBorder(nil, nil, imgWidget, nil, textContent)
			cardMainContent := container.NewVBox(titleRichText, cardContentWithImage)
			card := widget.NewCard("", fmt.Sprintf("Published: %s by %s", humanTime(article.PublishedAt), article.Source.Name), cardMainContent)
			results.Add(card)
		}
		results.Refresh()
		if currentSortMode == SortTimeDesc && currentPage == 1 {
			scroll.ScrollToTop()
		}
	}

	showBookmarksView = func() {
		bmWin := myApp.NewWindow("Bookmarked Articles")
		bmWin.Resize(fyne.NewSize(700, 600))
		listContent := container.NewVBox()
		scrollableList := container.NewVScroll(listContent)
		var refreshBookmarksList func()
		refreshBookmarksList = func() {
			listContent.Objects = nil
			bookmarksMutex.Lock()
			currentBookmarks := make([]Article, len(bookmarkedArticles))
			copy(currentBookmarks, bookmarkedArticles)
			bookmarksMutex.Unlock()
			if len(currentBookmarks) == 0 {
				listContent.Add(widget.NewLabelWithStyle("No articles bookmarked.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			} else {
				for _, bmArticle := range currentBookmarks {
					articleForView := bmArticle
					titleLabel := widget.NewLabelWithStyle(articleForView.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
					descLabel := widget.NewLabel(summarizeText(articleForView.Description))
					descLabel.Wrapping = fyne.TextWrapWord
					urlLink, _ := url.Parse(articleForView.URL)
					removeBtn := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
						dialog.ShowConfirm("Remove Bookmark", fmt.Sprintf("Remove '%s'?", articleForView.Title), func(confirm bool) {
							if confirm {
								toggleBookmark(articleForView)
								refreshBookmarksList()
								refreshResultsUI()
							}
						}, bmWin)
					})
					listContent.Add(container.NewVBox(titleLabel, widget.NewLabel(fmt.Sprintf("Published: %s", humanTime(articleForView.PublishedAt))), descLabel, container.NewHBox(widget.NewHyperlink("Open", urlLink), layout.NewSpacer(), removeBtn), widget.NewSeparator()))
				}
			}
			listContent.Refresh()
			scrollableList.Refresh()
		}
		refreshBookmarksList()
		bmWin.SetContent(scrollableList)
		bmWin.Show()
	}

	showTrendAnalysisDialog = func() {
		if len(allArticles) == 0 || lastQuery == "" {
			dialog.ShowInformation("Trend Analysis", "No articles or search query to analyze.", myWindow)
			return
		}

		trendWin := myApp.NewWindow(fmt.Sprintf("Trend for: %s", lastQuery))
		trendWin.Resize(fyne.NewSize(400, 500))

		countsByDate := make(map[string]int) // YYYY-MM-DD -> count
		queryWords := strings.Fields(strings.ToLower(lastQuery))

		for _, article := range allArticles {
			articleTextLower := strings.ToLower(article.Title + " " + article.Description)
			foundKeyword := false
			for _, qw := range queryWords {
				if strings.Contains(articleTextLower, qw) {
					foundKeyword = true
					break
				}
			}
			if foundKeyword {
				t, err := time.Parse(time.RFC3339, article.PublishedAt)
				if err == nil {
					dateStr := t.Format("2006-01-02")
					countsByDate[dateStr]++
				}
			}
		}

		if len(countsByDate) == 0 {
			dialog.ShowInformation("Trend Analysis", "No articles matched the query for trend analysis.", myWindow)
			trendWin.Close()
			return
		}

		var dates []string
		for dateStr := range countsByDate {
			dates = append(dates, dateStr)
		}
		sort.Strings(dates)

		tableContent := container.NewVBox()
		tableContent.Add(container.NewGridWithColumns(2, widget.NewLabelWithStyle("Date", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), widget.NewLabelWithStyle("Article Count", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})))
		tableContent.Add(widget.NewSeparator())

		for _, dateStr := range dates {
			tableContent.Add(container.NewGridWithColumns(2, widget.NewLabel(dateStr), widget.NewLabel(fmt.Sprintf("%d", countsByDate[dateStr]))))
		}

		trendWin.SetContent(container.NewScroll(tableContent))
		trendWin.Show()
	}

	sortBtn := widget.NewButtonWithIcon("Sort: Time ↓", theme.MenuDropDownIcon(), nil)
	sortBtn.OnTapped = func() {
		switch currentSortMode {
		case SortTimeDesc:
			currentSortMode = SortTimeAsc
			sortBtn.SetText("Sort: Time ↑")
			sortBtn.SetIcon(theme.MenuDropUpIcon())
			sortByTime(allArticles, true)
		case SortTimeAsc:
			currentSortMode = SortSentimentDesc
			sortBtn.SetText("Sort: Sentiment ↓")
			sortBtn.SetIcon(theme.MenuDropDownIcon())
			sortBySentiment(allArticles, false)
		case SortSentimentDesc:
			currentSortMode = SortSentimentAsc
			sortBtn.SetText("Sort: Sentiment ↑")
			sortBtn.SetIcon(theme.MenuDropUpIcon())
			sortBySentiment(allArticles, true)
		case SortSentimentAsc:
			currentSortMode = SortTimeDesc
			sortBtn.SetText("Sort: Time ↓")
			sortBtn.SetIcon(theme.MenuDropDownIcon())
			sortByTime(allArticles, false)
		}
		refreshResultsUI()
	}

	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		key := keyInput.Text
		query := queryInput.Text
		fromDate := fromDateEntry.Text
		toDate := toDateEntry.Text
		if key == "" {
			dialog.ShowError(fmt.Errorf("API key is required"), myWindow)
			return
		}
		if query == "" {
			dialog.ShowError(fmt.Errorf("Search query is required"), myWindow)
			return
		}
		results.Objects = nil
		results.Add(loadingIndicator)
		loadingIndicator.Show()
		results.Refresh()
		loadMoreBtn.Hide()
		currentPage = 1
		lastQuery = query
		lastFromDate = fromDate
		lastToDate = toDate

		fetchedArticles, total, err := fetchNews(key, query, fromDate, toDate, currentPage)
		loadingIndicator.Hide()
		results.Objects = nil
		if err != nil {
			dialog.ShowError(err, myWindow)
			allArticles = nil
			results.Refresh()
			return
		}
		if len(fetchedArticles) == 0 {
			dialog.ShowInformation("Search Results", "No articles found.", myWindow)
			allArticles = nil
			results.Refresh()
			return
		}
		totalResults = total
		allArticles = fetchedArticles
		switch currentSortMode {
		case SortTimeDesc:
			sortByTime(allArticles, false)
		case SortTimeAsc:
			sortByTime(allArticles, true)
		case SortSentimentDesc:
			sortBySentiment(allArticles, false)
		case SortSentimentAsc:
			sortBySentiment(allArticles, true)
		}
		refreshResultsUI()
		if len(allArticles) < totalResults && len(allArticles) > 0 {
			loadMoreBtn.Show()
		} else {
			loadMoreBtn.Hide()
		}
		if errSave := saveAPIKey(key); errSave != nil {
			fmt.Fprintf(os.Stderr, "Error saving API key: %v\n", errSave)
			// Optionally, inform the user via a dialog or notification if critical
		}
	})

	queryInput.OnSubmitted = func(s string) { searchBtn.OnTapped() }
	fromDateEntry.OnSubmitted = func(s string) { searchBtn.OnTapped() }
	toDateEntry.OnSubmitted = func(s string) { searchBtn.OnTapped() }
	searchRow := container.NewBorder(nil, nil, nil, container.NewHBox(searchBtn, sortBtn), queryInput)

	exportBtn := widget.NewButtonWithIcon("Export MD", theme.FileTextIcon(), func() {
		if len(allArticles) == 0 {
			myApp.SendNotification(&fyne.Notification{Title: "Export Info", Content: "No articles."})
			return
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# News: %s\n\n", lastQuery))
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n- URL: <%s>\n- Pub: %s\n- Src: %s\n- Desc: %s\n- Impact: %d\n- Policy: %d%%\n- Sentiment: %d\n- Sum: %s\n\n---\n\n", a.Title, a.URL, humanTime(a.PublishedAt), a.Source.Name, strings.TrimSpace(a.Description), a.ImpactScore, a.PolicyProbability, a.SentimentScore, summarizeText(a.Description)))
		}
		home, err := os.UserHomeDir()
		if err != nil {
			dialog.ShowError(fmt.Errorf("could not get home directory: %w", err), myWindow)
			return
		}
		docDir := filepath.Join(home, "Documents")
		os.MkdirAll(docDir, 0755)
		dateStr := time.Now().Format("2006-01-02")
		safeQuery := strings.ReplaceAll(strings.ToLower(lastQuery), " ", "_")
		safeQuery = safeFilenameRegex.ReplaceAllString(safeQuery, "") // Use pre-compiled regex
		if len(safeQuery) > 30 {
			safeQuery = safeQuery[:30]
		}
		fileName := fmt.Sprintf("news_export_%s_%s.md", safeQuery, dateStr)
		path := filepath.Join(docDir, fileName)
		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
			dialog.ShowError(fmt.Errorf("failed to export: %w", err), myWindow)
			return
		}
		myApp.SendNotification(&fyne.Notification{Title: "Export OK", Content: path})
	})
	clipboardBtn := widget.NewButtonWithIcon("Copy All", theme.ContentCopyIcon(), func() {
		if len(allArticles) == 0 {
			myApp.SendNotification(&fyne.Notification{Title: "Clipboard Info", Content: "No articles."})
			return
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Query: %s\n\n", lastQuery))
		for i, a := range allArticles {
			sb.WriteString(fmt.Sprintf("Art %d: %s\n Link: %s\n Pub: %s\n Sum: %s\n\n", i+1, a.Title, a.URL, humanTime(a.PublishedAt), summarizeText(a.Description)))
		}
		myWindow.Clipboard().SetContent(sb.String())
		myApp.SendNotification(&fyne.Notification{Title: "Clipboard OK", Content: fmt.Sprintf("%d arts copied.", len(allArticles))})
	})
	bookmarksBtn := widget.NewButtonWithIcon("Bookmarks", theme.FolderOpenIcon(), func() { showBookmarksView() })
	trendBtn := widget.NewButtonWithIcon("Trend", theme.InfoIcon(), func() { showTrendAnalysisDialog() }) // Corrected: theme.InfoIcon()
	utilityRow := container.NewHBox(layout.NewSpacer(), trendBtn, bookmarksBtn, exportBtn, clipboardBtn, layout.NewSpacer())
	loadMoreBtn.OnTapped = func() {
		currentPage++
		key := keyInput.Text
		fetchedArticles, _, err := fetchNews(key, lastQuery, lastFromDate, lastToDate, currentPage)
		if err != nil {
			myApp.SendNotification(&fyne.Notification{Title: "Load More Error", Content: err.Error()})
			currentPage--
			return
		}
		if len(fetchedArticles) > 0 {
			allArticles = append(allArticles, fetchedArticles...)
			switch currentSortMode {
			case SortTimeDesc:
				sortByTime(allArticles, false)
			case SortTimeAsc:
				sortByTime(allArticles, true)
			case SortSentimentDesc:
				sortBySentiment(allArticles, false)
			case SortSentimentAsc:
				sortBySentiment(allArticles, true)
			}
			refreshResultsUI()
			scroll.ScrollToBottom()
		}
		if len(allArticles) >= totalResults || len(fetchedArticles) == 0 {
			loadMoreBtn.Hide()
		} else {
			loadMoreBtn.Show()
		}
	}
	topControls := container.NewVBox(apiKeyRow, searchRow, dateFilterRow, utilityRow, widget.NewSeparator())
	content := container.NewBorder(topControls, loadMoreContainer, nil, nil, scroll)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
