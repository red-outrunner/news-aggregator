package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	"go.uber.org/zap"
)

// Article struct (unchanged)
type Article struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	URL               string   `json:"url"`
	URLToImage        string   `json:"urlToImage"`
	PublishedAt       string   `json:"publishedAt"`
	ImpactScore       int      `json:"impactScore,omitempty"`
	PolicyProbability int      `json:"policyProbability,omitempty"`
	SentimentScore    int      `json:"sentimentScore,omitempty"`
	Source            struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"source"`
}

// NewsResponse struct (unchanged)
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Config for unified settings
type Config struct {
	APIKey      string `json:"apiKey"`
	IsDarkTheme bool   `json:"isDarkTheme"`
}

// CacheEntry for news caching
type CacheEntry struct {
	Articles     []Article
	TotalResults int
	Timestamp    time.Time
}

// StockData for stock watcher
type StockData struct {
	Ticker      string
	Price       float64
	Change      float64
	ChangePct   float64
	LastUpdated time.Time
}

// NewsProvider interface for multiple APIs
type NewsProvider interface {
	FetchNews(query, fromDate, toDate string, page, pageSize int) ([]Article, int, error)
}

// NewsAPIProvider for NewsAPI.org
type NewsAPIProvider struct {
	APIKey string
}

// AppState for modular state management
type AppState struct {
	APIKey        string
	CurrentPage   int
	TotalResults  int
	Articles      []Article
	LastQuery     string
	LastFromDate  string
	LastToDate    string
	SortMode      SortMode
	IsDarkTheme   bool
	NewsCache     map[string]CacheEntry
	CacheMutex    sync.Mutex
	Providers     []NewsProvider
}

// UIComponents for modular UI
type UIComponents struct {
	Window          fyne.Window
	Results         *container.Scroll
	QueryInput      *widget.Entry
	FromDateEntry   *widget.Entry
	ToDateEntry     *widget.Entry
	KeyInput        *widget.Entry
	SearchBtn       *widget.Button
	SortBtn         *widget.Button
	LoadMoreBtn     *widget.Button
	WatcherBtn      *widget.Button
	TrendBtn        *widget.Button
	BookmarksBtn    *widget.Button
	ExportBtn       *widget.Button
	ClipboardBtn    *widget.Button
	SentimentFilter *widget.Select
}

var (
	// Logger
	logger *zap.Logger

	// Existing globals
	bookmarkedArticles  []Article
	bookmarksMutex      sync.Mutex
	bookmarksFilePath   string
	readArticles        map[string]bool
	readArticlesMutex   sync.Mutex
	imageCacheDir       string
	positiveKeywordsSet map[string]struct{}
	negativeKeywordsSet map[string]struct{}
	safeFilenameRegex   *regexp.Regexp
	tickerRegex         *regexp.Regexp
	cacheTTL            = 10 * time.Minute

	// Stock watcher globals
	watchedStocks     []string
	watchlistMutex    sync.Mutex
	watchlistFilePath string
)

const (
	bookmarksFilename = "news_aggregator_bookmarks.json"
	watchlistFilename = "news_aggregator_watchlist.json"
	configFilename    = "config.json"
)

// Sentiment, Impact, and Policy keywords (unchanged, omitted for brevity)
var positiveKeywords = []string{
	"good", "great", "excellent", "positive", "success", "improve", "benefit", "effective", "strong", "happy", "joy", "love", "optimistic", "favorable", "promising", "encouraging",
	"grow", "growth", "expansion", "expand", "increase", "surge", "rise", "upward", "upturn", "boom", "accelerate", "augment", "boost", "rally", "recover", "recovery",
	"achieve", "achieved", "outperform", "exceed", "beat", "record", "profitable", "profit", "gains", "earnings", "revenue", "dividend", "surplus",
	"innovative", "innovation", "breakthrough", "advance", "launch", "new", "develop", "upgrade", "leading", "cutting-edge",
	"bullish", "optimism", "confidence", "stable", "stability", "support", "demand", "hot", "high", "robust",
	"acquire", "acquisition", "merger", "partnership", "agreement", "approve", "approved", "endorse", "confirm",
}
var negativeKeywords = []string{
	"bad", "poor", "terrible", "negative", "fail", "failure", "weak", "adverse", "sad", "angry", "fear", "pessimistic", "unfavorable", "discouraging",
	"decline", "decrease", "drop", "fall", "slump", "downturn", "recession", "contraction", "reduce", "cut", "loss", "losses", "deficit", "shrink", "erode", "weaken",
	"crisis", "disaster", "risk", "warn", "warning", "threat", "problem", "issue", "concern", "challenge", "obstacle", "difficulty", "uncertainty", "volatile", "volatility",
	"underperform", "miss", "shortfall", "struggle", "stagnate", "delay", "halt",
	"bearish", "pessimism", "doubt", "skepticism", "unstable", "instability", "pressure", "low", "oversupply", "bubble",
	"investigation", "lawsuit", "penalty", "fine", "sanction", "ban", "fraud", "scandal", "recall", "dispute", "reject", "denied", "downgrade",
}
var impactScoreKeywords = []string{
	"recession", "inflation", "interest rates", "market crash", "trade war", "supply chain", "corporate earnings", "acquisition", "ipo", "federal reserve", "economic growth", "unemployment", "government stimulus", "new regulation", "geopolitical risk", "tariff", "sanction", "deficit", "surplus", "bankruptcy", "takeover", "merger", "venture capital", "private equity", "stock market", "bond market", "currency fluctuation", "commodity prices", "consumer spending", "housing market", "energy crisis", "financial crisis", "debt ceiling", "quantitative easing", "fiscal policy", "monetary policy", "trade deal", "market sentiment", "volatility", "correction", "bear market", "bull market", "earnings report", "profit warning", "economic forecast", "global economy", "emerging markets",
	"ecb", "european central bank", "eurozone", "brexit impact", "eu stimulus", "recovery fund", "dax", "cac 40", "ftse mib", "euro stoxx 50", "sovereign debt", "eu bailout", "esm", "european stability mechanism",
	"sarb", "south african reserve bank", "jse", "johannesburg stock exchange", "rand", "load shedding", "eskom", "mining sector sa", "sa budget", "credit rating south africa", "foreign direct investment sa", "state owned enterprises sa", "bee impact",
}
var policyKeywords = []string{
	"policy", "regulation", "law", "government", "legislation", "bill", "congress", "senate", "parliament", "decree", "treaty", "court", "ruling", "initiative", "mandate", "executive order", "tariff", "sanction", "subsidy", "public policy", "compliance", "enforcement", "oversight", "hearing", "testimony", "budget", "appropriation", "act", "statute", "ordinance", "directive", "guideline", "framework", "accord", "pact", "resolution", "referendum", "lobbying", "advocacy", "think tank", "white paper", "federal", "state", "local government", "agency", "commission", "authority", "irs", "federal reserve", "supreme court", "white house", "capitol hill", "reform", "governance", "judiciary",
	"european parliament", "european commission", "council of the european union", "eu directive", "eu regulation", "mep", "ecj", "eurozone", "brussels", "ecb", "european central bank", "single market", "schengen", "brexit", "article 50", "eusl", "gdpr",
	"parliament uk", "house of commons", "house of lords", "downing street", "hmrc", "bank of england", "chancellor",
	"parliament sa", "national assembly sa", "national council of provinces", "ncop", "sars", "south african revenue service", "constitutional court sa", "concourt", "provincial government sa", "cabinet sa", "cosatu", "public protector sa", "union buildings", "south african reserve bank", "sarb", "anc", "da", "eff", "state capture", "bee", "black economic empowerment", "land reform", "national development plan", "ndp", "municipal",
	"united nations", "unsc", "world bank", "imf", "wto", "who", "icc", "g7", "g20", "oecd",
	"bundestag", "diet", "duma",
}

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	positiveKeywordsSet = make(map[string]struct{}, len(positiveKeywords))
	for _, k := range positiveKeywords {
		positiveKeywordsSet[strings.ToLower(k)] = struct{}{}
	}

	negativeKeywordsSet = make(map[string]struct{}, len(negativeKeywords))
	for _, k := range negativeKeywords {
		negativeKeywordsSet[strings.ToLower(k)] = struct{}{}
	}
	safeFilenameRegex = regexp.MustCompile(`[^\w-]`)
	tickerRegex = regexp.MustCompile(`^[A-Z0-9]{1,5}(\.[A-Z]{1,3})?$`)
}

// --- Utility Functions ---
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

func sortBySentiment(articles []Article, ascending bool) {
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

// --- Core Logic ---
func calculateSentimentScore(text string) int {
	if text == "" {
		return 0
	}
	score := 0
	textLower := strings.ToLower(text)
	words := strings.FieldsFunc(textLower, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	for _, word := range words {
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

// --- Config & Persistence ---
func loadConfig() Config {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".config", "newsaggregator_v3", configFilename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Warn("Failed to read config file", zap.Error(err))
		return Config{}
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Warn("Failed to parse config file", zap.Error(err))
		return Config{}
	}
	return config
}

func saveConfig(config Config) error {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "newsaggregator_v3")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		logger.Warn("Failed to create config directory", zap.Error(err))
		return err
	}
	configPath := filepath.Join(configDir, configFilename)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal config", zap.Error(err))
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func setupImageCacheDir() {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Warn("Failed to get home directory", zap.Error(err))
		imageCacheDir = "image_cache"
		os.MkdirAll(imageCacheDir, 0700)
		return
	}
	configDir := filepath.Join(home, ".config", "newsaggregator_v3")
	imageCacheDir = filepath.Join(configDir, "image_cache")
	if err := os.MkdirAll(imageCacheDir, 0700); err != nil {
		logger.Warn("Could not create image cache directory", zap.Error(err))
		imageCacheDir = "image_cache"
		os.MkdirAll(imageCacheDir, 0700)
	}
}

func getCachePathForURL(imageURL string) string {
	hasher := sha256.New()
	hasher.Write([]byte(imageURL))
	hash := hex.EncodeToString(hasher.Sum(nil))
	ext := filepath.Ext(imageURL)
	if ext == "" {
		ext = ".jpg"
	}
	return filepath.Join(imageCacheDir, hash+ext)
}

func setupBookmarksPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Warn("Failed to get home directory for bookmarks", zap.Error(err))
		bookmarksFilePath = bookmarksFilename
		return
	}
	configDir := filepath.Join(home, ".config", "newsaggregator_v3")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		logger.Warn("Could not create bookmarks config directory", zap.Error(err))
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
		logger.Warn("Failed to read bookmarks file", zap.Error(err))
		bookmarkedArticles = []Article{}
		return
	}
	if err := json.Unmarshal(data, &bookmarkedArticles); err != nil {
		logger.Warn("Failed to parse bookmarks file", zap.Error(err))
		bookmarkedArticles = []Article{}
	}
}

func saveBookmarks() {
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()
	data, err := json.MarshalIndent(bookmarkedArticles, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal bookmarks", zap.Error(err))
		return
	}
	if err := os.WriteFile(bookmarksFilePath, data, 0600); err != nil {
		logger.Error("Failed to write bookmarks file", zap.Error(err))
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

func setupWatchlistPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Warn("Failed to get home directory for watchlist", zap.Error(err))
		watchlistFilePath = watchlistFilename
		return
	}
	configDir := filepath.Join(home, ".config", "newsaggregator_v3")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		logger.Warn("Could not create watchlist config directory", zap.Error(err))
		watchlistFilePath = watchlistFilename
		return
	}
	watchlistFilePath = filepath.Join(configDir, watchlistFilename)
}

func loadWatchlist() {
	watchlistMutex.Lock()
	defer watchlistMutex.Unlock()
	data, err := os.ReadFile(watchlistFilePath)
	if err != nil {
		logger.Warn("Failed to read watchlist file", zap.Error(err))
		watchedStocks = []string{}
		return
	}
	if err := json.Unmarshal(data, &watchedStocks); err != nil {
		logger.Warn("Failed to parse watchlist file", zap.Error(err))
		watchedStocks = []string{}
	}
}

func saveWatchlist() {
	watchlistMutex.Lock()
	defer watchlistMutex.Unlock()
	data, err := json.MarshalIndent(watchedStocks, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal watchlist", zap.Error(err))
		return
	}
	if err := os.WriteFile(watchlistFilePath, data, 0600); err != nil {
		logger.Error("Failed to write watchlist file", zap.Error(err))
	}
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

func isValidTicker(ticker string) bool {
	return tickerRegex.MatchString(ticker)
}

// --- News Fetching ---
func (p *NewsAPIProvider) FetchNews(query, fromDate, toDate string, page, pageSize int) ([]Article, int, error) {
	baseApiURL := "https://newsapi.org/v2/everything"
	queryParams := url.Values{}
	queryParams.Add("q", query)
	queryParams.Add("sortBy", "publishedAt")
	queryParams.Add("language", "en")
	queryParams.Add("pageSize", fmt.Sprintf("%d", pageSize))
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("apiKey", p.APIKey)

	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err == nil {
			queryParams.Add("from", fromDate)
		} else {
			logger.Warn("Invalid 'from' date format", zap.String("date", fromDate))
		}
	}
	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err == nil {
			queryParams.Add("to", toDate)
		} else {
			logger.Warn("Invalid 'to' date format", zap.String("toDate", toDate))
		}
	}

	fullApiURL := fmt.Sprintf("%s?%s", baseApiURL, queryParams.Encode())

	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := http.Get(fullApiURL)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to connect to news service: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			logger.Warn("Rate limit hit, retrying", zap.Int("attempt", attempt))
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
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
	return nil, 0, fmt.Errorf("exceeded retry attempts for API request")
}

func (app *AppState) fetchNewsFromProviders(query, fromDate, toDate string, page, pageSize int) ([]Article, int, error) {
	app.CacheMutex.Lock()
	cacheKey := fmt.Sprintf("%s:%s:%s:%d:%d", query, fromDate, toDate, page, pageSize)
	if entry, found := app.NewsCache[cacheKey]; found && time.Since(entry.Timestamp) < cacheTTL {
		app.CacheMutex.Unlock()
		return entry.Articles, entry.TotalResults, nil
	}
	app.CacheMutex.Unlock()

	var allArticles []Article
	totalResults := 0
	for _, provider := range app.Providers {
		articles, count, err := provider.FetchNews(query, fromDate, toDate, page, pageSize)
		if err != nil {
			logger.Warn("Failed to fetch from provider", zap.Error(err))
			continue
		}
		allArticles = append(allArticles, articles...)
		totalResults += count
	}

	if len(allArticles) == 0 && totalResults == 0 {
		return nil, 0, fmt.Errorf("no articles fetched from any provider")
	}

	app.CacheMutex.Lock()
	app.NewsCache[cacheKey] = CacheEntry{
		Articles:     allArticles,
		TotalResults: totalResults,
		Timestamp:    time.Now(),
	}
	app.CacheMutex.Unlock()

	go app.preloadNextPage(query, fromDate, toDate, page, pageSize)

	return allArticles, totalResults, nil
}

func (app *AppState) preloadNextPage(query, fromDate, toDate string, currentPage, pageSize int) {
	go func() {
		cacheKey := fmt.Sprintf("%s:%s:%s:%d:%d", query, fromDate, toDate, currentPage+1, pageSize)
		app.CacheMutex.Lock()
		if _, found := app.NewsCache[cacheKey]; found {
			app.CacheMutex.Unlock()
			return
		}
		app.CacheMutex.Unlock()

		articles, total, err := app.fetchNewsFromProviders(query, fromDate, toDate, currentPage+1, pageSize)
		if err != nil {
			logger.Warn("Failed to preload next page", zap.Error(err))
			return
		}
		app.CacheMutex.Lock()
		app.NewsCache[cacheKey] = CacheEntry{Articles: articles, TotalResults: total, Timestamp: time.Now()}
		app.CacheMutex.Unlock()
	}()
}

func fetchStockData(ticker string) (*StockData, error) {
	// Placeholder for Alpha Vantage API (replace with actual key and logic)
	apiKey := "YOUR_ALPHA_VANTAGE_KEY"
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", ticker, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		logger.Error("Failed to fetch stock data", zap.String("ticker", ticker), zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	// Mocked response for now
	return &StockData{
		Ticker:      ticker,
		Price:       150.25,
		Change:      2.5,
		ChangePct:   1.7,
		LastUpdated: time.Now(),
	}, nil
}

// --- AppState Methods ---
func NewAppState() *AppState {
	config := loadConfig()
	return &AppState{
		APIKey:      config.APIKey,
		IsDarkTheme: config.IsDarkTheme,
		NewsCache:   make(map[string]CacheEntry),
		SortMode:    SortTimeDesc,
		Providers:   []NewsProvider{&NewsAPIProvider{APIKey: config.APIKey}},
	}
}

func (app *AppState) SaveConfig() {
	config := Config{
		APIKey:      app.APIKey,
		IsDarkTheme: app.IsDarkTheme,
	}
	if err := saveConfig(config); err != nil {
		logger.Error("Failed to save config", zap.Error(err))
	}
}

type SortMode int

const (
	SortTimeDesc SortMode = iota
	SortTimeAsc
	SortSentimentDesc
	SortSentimentAsc
)

func (app *AppState) SetupUI(myApp fyne.App) *UIComponents {
	myWindow := myApp.NewWindow("News on Red: Market Research Tool")
	myWindow.Resize(fyne.NewSize(1000, 850))

	// Setup configurations
	setupBookmarksPath()
	setupImageCacheDir()
	setupWatchlistPath()
	loadBookmarks()
	loadWatchlist()
	readArticles = make(map[string]bool)

	// Theme
	if app.IsDarkTheme {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	// UI Components
	keyInput := widget.NewPasswordEntry()
	keyInput.SetPlaceHolder("Enter NewsAPI key...")
	if app.APIKey != "" {
		keyInput.SetText(app.APIKey)
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
	updateThemeButtonText(app.IsDarkTheme)
	themeBtn.OnTapped = func() {
		app.IsDarkTheme = !app.IsDarkTheme
		if app.IsDarkTheme {
			myApp.Settings().SetTheme(theme.DarkTheme())
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
		}
		updateThemeButtonText(app.IsDarkTheme)
		app.SaveConfig()
	}

	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyLabel.TextStyle = fyne.TextStyle{Monospace: true, Size: 14}
	apiKeyRow := container.NewBorder(nil, nil, apiKeyLabel, themeBtn, keyInput)

	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search news topics...")
	fromDateEntry := widget.NewEntry()
	fromDateEntry.SetPlaceHolder("From: YYYY-MM-DD")
	toDateEntry := widget.NewEntry()
	toDateEntry.SetPlaceHolder("To: YYYY-MM-DD")
	dateFilterRow := container.NewGridWithColumns(2,
		container.NewBorder(nil, nil, widget.NewLabel("From:"), nil, fromDateEntry),
		container.NewBorder(nil, nil, widget.NewLabel("To:"), nil, toDateEntry),
	)

	results := container.NewVBox()
	results.Add(widget.NewLabelWithStyle("Enter API key and search query.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(300, 400))

	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide()
	loadMoreBtn := widget.NewButtonWithIcon("Load More Articles", theme.MoreVerticalIcon(), nil)
	loadMoreBtn.Hide()
	loadMoreContainer := container.NewCenter(loadMoreBtn)

	sentimentFilter := widget.NewSelect([]string{"All", "Positive (>20)", "Neutral (-20 to 20)", "Negative (<-20)"}, nil)

	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), nil)
	sortBtn := widget.NewButtonWithIcon("Sort: Time ↓", theme.MenuDropDownIcon(), nil)
	exportBtn := widget.NewButtonWithIcon("Export", theme.FileTextIcon(), nil)
	clipboardBtn := widget.NewButtonWithIcon("Copy All", theme.ContentCopyIcon(), nil)
	bookmarksBtn := widget.NewButtonWithIcon("Bookmarks", theme.FolderOpenIcon(), nil)
	trendBtn := widget.NewButtonWithIcon("Trend", theme.SettingsIcon(), nil)
	watcherBtn := widget.NewButtonWithIcon("Watcher", theme.VisibilityIcon(), nil)

	return &UIComponents{
		Window:          myWindow,
		Results:         scroll,
		QueryInput:      queryInput,
		FromDateEntry:   fromDateEntry,
		ToDateEntry:     toDateEntry,
		KeyInput:        keyInput,
		SearchBtn:       searchBtn,
		SortBtn:         sortBtn,
		LoadMoreBtn:     loadMoreBtn,
		WatcherBtn:      watcherBtn,
		TrendBtn:        trendBtn,
		BookmarksBtn:    bookmarksBtn,
		ExportBtn:       exportBtn,
		ClipboardBtn:    clipboardBtn,
		SentimentFilter: sentimentFilter,
	}
}

func (app *AppState) HandleSearch(ui *UIComponents, myApp fyne.App) {
	key := ui.KeyInput.Text
	query := ui.QueryInput.Text
	fromDate := ui.FromDateEntry.Text
	toDate := ui.ToDateEntry.Text
	if key == "" {
		dialog.ShowError(fmt.Errorf("API key is required"), ui.Window)
		return
	}
	if query == "" {
		dialog.ShowError(fmt.Errorf("Search query is required"), ui.Window)
		return
	}

	ui.Results.Content.(*container.VBox).Objects = nil
	ui.Results.Content.(*container.VBox).Add(ui.LoadMoreBtn)
	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Show()
	ui.Results.Content.(*container.VBox).Add(loadingIndicator)
	ui.Results.Refresh()
	ui.LoadMoreBtn.Hide()
	app.CurrentPage = 1
	app.LastQuery = query
	app.LastFromDate = fromDate
	app.LastToDate = toDate
	app.APIKey = key
	app.Providers = []NewsProvider{&NewsAPIProvider{APIKey: key}}

	fetchedArticles, total, err := app.fetchNewsFromProviders(query, fromDate, toDate, app.CurrentPage, 18)
	loadingIndicator.Hide()
	ui.Results.Content.(*container.VBox).Objects = nil
	if err != nil {
		dialog.ShowError(err, ui.Window)
		app.Articles = nil
		ui.Results.Refresh()
		return
	}
	if len(fetchedArticles) == 0 {
		dialog.ShowInformation("Search Results", "No articles found.", ui.Window)
		app.Articles = nil
		ui.Results.Refresh()
		return
	}
	app.TotalResults = total
	app.Articles = fetchedArticles
	switch app.SortMode {
	case SortTimeDesc:
		sortByTime(app.Articles, false)
	case SortTimeAsc:
		sortByTime(app.Articles, true)
	case SortSentimentDesc:
		sortBySentiment(app.Articles, false)
	case SortSentimentAsc:
		sortBySentiment(app.Articles, true)
	}
	app.RefreshResultsUI(ui)
	if len(app.Articles) < app.TotalResults && len(app.Articles) > 0 {
		ui.LoadMoreBtn.Show()
	} else {
		ui.LoadMoreBtn.Hide()
	}
	app.SaveConfig()
}

func (app *AppState) HandleSort(ui *UIComponents) {
	switch app.SortMode {
	case SortTimeDesc:
		app.SortMode = SortTimeAsc
		ui.SortBtn.SetText("Sort: Time ↑")
		ui.SortBtn.SetIcon(theme.MenuDropUpIcon())
		sortByTime(app.Articles, true)
	case SortTimeAsc:
		app.SortMode = SortSentimentDesc
		ui.SortBtn.SetText("Sort: Sentiment ↓")
		ui.SortBtn.SetIcon(theme.MenuDropDownIcon())
		sortBySentiment(app.Articles, false)
	case SortSentimentDesc:
		app.SortMode = SortSentimentAsc
		ui.SortBtn.SetText("Sort: Sentiment ↑")
		ui.SortBtn.SetIcon(theme.MenuDropUpIcon())
		sortBySentiment(app.Articles, true)
	case SortSentimentAsc:
		app.SortMode = SortTimeDesc
		ui.SortBtn.SetText("Sort: Time ↓")
		ui.SortBtn.SetIcon(theme.MenuDropDownIcon())
		sortByTime(app.Articles, false)
	}
	app.RefreshResultsUI(ui)
}

func (app *AppState) HandleSentimentFilter(ui *UIComponents, filter string) {
	filteredArticles := []Article{}
	for _, a := range app.Articles {
		switch filter {
		case "Positive (>20)":
			if a.SentimentScore > 20 {
				filteredArticles = append(filteredArticles, a)
			}
		case "Neutral (-20 to 20)":
			if a.SentimentScore >= -20 && a.SentimentScore <= 20 {
				filteredArticles = append(filteredArticles, a)
			}
		case "Negative (<-20)":
			if a.SentimentScore < -20 {
				filteredArticles = append(filteredArticles, a)
			}
		default:
			filteredArticles = append(filteredArticles, a)
		}
	}
	app.Articles = filteredArticles
	app.RefreshResultsUI(ui)
}

func (app *AppState) HandleExport(ui *UIComponents, myApp fyne.App) {
	if len(app.Articles) == 0 {
		myApp.SendNotification(&fyne.Notification{Title: "Export Info", Content: "No articles."})
		return
	}
	formatSelect := widget.NewSelect([]string{"Markdown", "CSV", "JSON"}, func(format string) {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Error("Could not get home directory", zap.Error(err))
			dialog.ShowError(fmt.Errorf("could not get home directory: %w", err), ui.Window)
			return
		}
		docDir := filepath.Join(home, "Documents")
		os.MkdirAll(docDir, 0755)
		dateStr := time.Now().Format("2006-01-02")
		safeQuery := strings.ReplaceAll(strings.ToLower(app.LastQuery), " ", "_")
		safeQuery = safeFilenameRegex.ReplaceAllString(safeQuery, "")
		if len(safeQuery) > 30 {
			safeQuery = safeQuery[:30]
		}
		fileName := fmt.Sprintf("news_export_%s_%s", safeQuery, dateStr)
		var content []byte
		switch format {
		case "CSV":
			fileName += ".csv"
			var sb strings.Builder
			sb.WriteString("Title,URL,Published,Source,Description,Impact,Policy,Sentiment\n")
			for _, a := range app.Articles {
				sb.WriteString(fmt.Sprintf("%q,%q,%q,%q,%q,%d,%d,%d\n", a.Title, a.URL, humanTime(a.PublishedAt), a.Source.Name, a.Description, a.ImpactScore, a.PolicyProbability, a.SentimentScore))
			}
			content = []byte(sb.String())
		case "JSON":
			fileName += ".json"
			content, _ = json.MarshalIndent(app.Articles, "", "  ")
		default: // Markdown
			fileName += ".md"
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("# News: %s\n\n", app.LastQuery))
			for _, a := range app.Articles {
				sb.WriteString(fmt.Sprintf("## %s\n- URL: <%s>\n- Pub: %s\n- Src: %s\n- Desc: %s\n- Impact: %d\n- Policy: %d%%\n- Sentiment: %d\n- Sum: %s\n\n---\n\n", a.Title, a.URL, humanTime(a.PublishedAt), a.Source.Name, strings.TrimSpace(a.Description), a.ImpactScore, a.PolicyProbability, a.SentimentScore, summarizeText(a.Description)))
			}
			content = []byte(sb.String())
		}
		path := filepath.Join(docDir, fileName)
		if err := os.WriteFile(path, content, 0644); err != nil {
			logger.Error("Failed to export", zap.Error(err))
			dialog.ShowError(fmt.Errorf("failed to export: %w", err), ui.Window)
			return
		}
		myApp.SendNotification(&fyne.Notification{Title: "Export OK", Content: path})
	})
	dialog.NewCustom("Select Export Format", "Cancel", formatSelect, ui.Window).Show()
}

func (app *AppState) HandleClipboard(ui *UIComponents, myApp fyne.App) {
	if len(app.Articles) == 0 {
		myApp.SendNotification(&fyne.Notification{Title: "Clipboard Info", Content: "No articles."})
		return
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Query: %s\n\n", app.LastQuery))
	for i, a := range app.Articles {
		sb.WriteString(fmt.Sprintf("Art %d: %s\n Link: %s\n Pub: %s\n Sum: %s\n\n", i+1, a.Title, a.URL, humanTime(a.PublishedAt), summarizeText(a.Description)))
	}
	ui.Window.Clipboard().SetContent(sb.String())
	myApp.SendNotification(&fyne.Notification{Title: "Clipboard OK", Content: fmt.Sprintf("%d arts copied.", len(app.Articles))})
}

func (app *AppState) RefreshResultsUI(ui *UIComponents) {
	results := ui.Results.Content.(*container.VBox)
	results.Objects = nil
	if len(app.Articles) == 0 {
		if app.LastQuery != "" {
			results.Add(widget.NewLabelWithStyle("No articles for '"+app.LastQuery+"'.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
		} else {
			results.Add(widget.NewLabelWithStyle("Enter API key & search.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
		}
		results.Refresh()
		return
	}

	var currentHighlightRegex *regexp.Regexp
	if app.LastQuery != "" {
		queryWords := strings.Fields(strings.ToLower(app.LastQuery))
		var reParts []string
		for _, qw := range queryWords {
			if len(qw) > 0 {
				reParts = append(reParts, regexp.QuoteMeta(qw))
			}
		}
		if len(reParts) > 0 {
			currentHighlightRegex = regexp.MustCompile(`(?i)\b(` + strings.Join(reParts, "|") + `)\b`)
		}
	}

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

	for i := range app.Articles {
		article := app.Articles[i]
		parsedURL, _ := url.Parse(article.URL)

		titleRichText := createHighlightedText(article.Title, app.LastQuery, currentHighlightRegex)
		var currentTitleStyle fyne.TextStyle
		if isRead(article.URL) {
			currentTitleStyle.Italic = true
		} else {
			currentTitleStyle.Bold = true
		}

		for j := range titleRichText.Segments {
			if ts, ok := titleRichText.Segments[j].(*widget.TextSegment); ok {
				isHighlighted := ts.Style.TextStyle.Bold
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
		descriptionRichText := createHighlightedText(cardDescription, app.LastQuery, currentHighlightRegex)
		descriptionRichText.Wrapping = fyne.TextWrapWord

		imgWidget := canvas.NewImageFromResource(nil)
		imgWidget.FillMode = canvas.ImageFillContain
		imgWidget.SetMinSize(fyne.NewSize(120, 80))
		if article.URLToImage != "" {
			go func(imgURL string, targetImg *canvas.Image) {
				cachePath := getCachePathForURL(imgURL)
				if _, err := os.Stat(cachePath); err == nil {
					imgData, err := os.ReadFile(cachePath)
					if err == nil {
						imgRes := fyne.NewStaticResource(filepath.Base(imgURL), imgData)
						targetImg.Resource = imgRes
						targetImg.Refresh()
						return
					}
				}
				client := http.Client{Timeout: 15 * time.Second}
				resp, err := client.Get(imgURL)
				if err != nil {
					logger.Warn("Failed to fetch image", zap.String("url", imgURL), zap.Error(err))
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					logger.Warn("Image fetch failed", zap.String("url", imgURL), zap.Int("status", resp.StatusCode))
					return
				}
				imgData, err := io.ReadAll(resp.Body)
				if err != nil {
					logger.Warn("Failed to read image data", zap.Error(err))
					return
				}
				_, _, err = image.Decode(bytes.NewReader(imgData))
				if err != nil {
					logger.Warn("Failed to decode image", zap.Error(err))
					return
				}
				if err := os.WriteFile(cachePath, imgData, 0644); err != nil {
					logger.Warn("Failed to write image to cache", zap.Error(err))
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
			app.RefreshResultsUI(ui)
			fullSummary := summarizeText(currentArticleForDetail.Description)
			if strings.TrimSpace(currentArticleForDetail.Description) == "" {
				fullSummary = "No full description."
			}
			currentArticleParsedURL, _ := url.Parse(currentArticleForDetail.URL)
			content := container.NewVBox(
				widget.NewLabelWithStyle(currentArticleForDetail.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabelWithStyle("Summary:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fullSummary),
				widget.NewSeparator(),
				widget.NewLabel(fmt.Sprintf("Impact: %d", currentArticleForDetail.ImpactScore)),
				widget.NewLabel(fmt.Sprintf("Policy: %d%%", currentArticleForDetail.PolicyProbability)),
				widget.NewLabel(fmt.Sprintf("Sentiment: %d", currentArticleForDetail.SentimentScore)),
				widget.NewSeparator(),
				widget.NewHyperlink("Open Original", currentArticleParsedURL),
			)
			for _, obj := range content.Objects {
				if lbl, ok := obj.(*widget.Label); ok {
					lbl.Wrapping = fyne.TextWrapWord
				}
			}
			var detailPopUp *widget.PopUp
			closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { detailPopUp.Hide() })
			dialogContainer := container.NewBorder(nil, container.NewCenter(closeButton), nil, nil, container.NewVScroll(content))
			detailPopUp = widget.NewModalPopUp(dialogContainer, ui.Window.Canvas())
			detailPopUp.Resize(fyne.NewSize(ui.Window.Canvas().Size().Width*0.8, ui.Window.Canvas().Size().Height*0.7))
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

		textContent := container.NewVBox(
			descriptionRichText,
			widget.NewSeparator(),
			container.NewGridWithColumns(3,
				widget.NewLabel(fmt.Sprintf("Impact: %d", article.ImpactScore)),
				widget.NewLabel(fmt.Sprintf("Policy: %d%%", article.PolicyProbability)),
				sentimentLabel,
			),
			widget.NewSeparator(),
			container.NewHBox(
				widget.NewHyperlink("Read Full Article", parsedURL),
				layout.NewSpacer(),
				bookmarkBtn,
				detailsBtn,
			),
		)
		cardContentWithImage := container.NewBorder(nil, nil, imgWidget, nil, textContent)
		cardMainContent := container.NewVBox(titleRichText, cardContentWithImage)
		card := widget.NewCard("", fmt.Sprintf("Published: %s by %s", humanTime(article.PublishedAt), article.Source.Name), cardMainContent)
		results.Add(card)
	}
	results.Refresh()
	if app.SortMode == SortTimeDesc && app.CurrentPage == 1 {
		ui.Results.ScrollToTop()
	}
}

func (app *AppState) ShowBookmarksView(ui *UIComponents, myApp fyne.App) {
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
							app.RefreshResultsUI(ui)
						}
					}, bmWin)
				})
				listContent.Add(container.NewVBox(
					titleLabel,
					widget.NewLabel(fmt.Sprintf("Published: %s", humanTime(articleForView.PublishedAt))),
					descLabel,
					container.NewHBox(widget.NewHyperlink("Open", urlLink), layout.NewSpacer(), removeBtn),
					widget.NewSeparator(),
				))
			}
		}
		listContent.Refresh()
		scrollableList.Refresh()
	}
	refreshBookmarksList()
	bmWin.SetContent(scrollableList)
	bmWin.Show()
}

func (app *AppState) ShowTrendAnalysisDialog(ui *UIComponents, myApp fyne.App) {
	if len(app.Articles) == 0 || app.LastQuery == "" {
		dialog.ShowInformation("Trend Analysis", "No articles or search query to analyze.", ui.Window)
		return
	}

	trendWin := myApp.NewWindow(fmt.Sprintf("Trend for: %s", app.LastQuery))
	trendWin.Resize(fyne.NewSize(600, 500))

	countsByDate := make(map[string]int)
	queryWords := strings.Fields(strings.ToLower(app.LastQuery))
	for _, article := range app.Articles {
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
		dialog.ShowInformation("Trend Analysis", "No articles matched the query for trend analysis.", ui.Window)
		trendWin.Close()
		return
	}

	var dates []string
	var counts []float64
	for dateStr := range countsByDate {
		dates = append(dates, dateStr)
	}
	sort.Strings(dates)
	for _, dateStr := range dates {
		counts = append(counts, float64(countsByDate[dateStr]))
	}

	chartJSON := fmt.Sprintf(`{
		"type": "line",
		"data": {
			"labels": %s,
			"datasets": [{
				"label": "Article Count",
				"data": %s,
				"borderColor": "#2196F3",
				"backgroundColor": "rgba(33, 150, 243, 0.2)",
				"fill": true
			}]
		},
		"options": {
			"scales": {
				"y": {
					"beginAtZero": true
				}
			},
			"plugins": {
				"title": {
					"display": true,
					"text": "Article Count Over Time"
				}
			}
		}
	}`, toJSON(dates), toJSON(counts))

	tableContent := container.NewVBox()
	tableContent.Add(container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Date", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Article Count", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	))
	tableContent.Add(widget.NewSeparator())
	for _, dateStr := range dates {
		tableContent.Add(container.NewGridWithColumns(2,
			widget.NewLabel(dateStr),
			widget.NewLabel(fmt.Sprintf("%d", countsByDate[dateStr])),
		))
	}

	content := container.NewVBox(widget.NewLabel(chartJSON), tableContent)
	trendWin.SetContent(container.NewScroll(content))
	trendWin.Show()
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func (app *AppState) ShowStockWatcherView(ui *UIComponents, myApp fyne.App) {
	watcherWin := myApp.NewWindow("Stock Watcher")
	watcherWin.Resize(fyne.NewSize(600, 500))

	var refreshWatcherList func()
	tickerList := container.NewVBox()
	scrollableList := container.NewVScroll(tickerList)
	newTickerEntry := widget.NewEntry()
	newTickerEntry.SetPlaceHolder("e.g., AAPL, NPN.JO, VOD.L...")

	addTickerBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		ticker := strings.ToUpper(strings.TrimSpace(newTickerEntry.Text))
		if ticker == "" || !isValidTicker(ticker) {
			dialog.ShowError(fmt.Errorf("Invalid ticker format"), watcherWin)
			return
		}

		watchlistMutex.Lock()
		found := false
		for _, t := range watchedStocks {
			if t == ticker {
				found = true
				break
			}
		}
		if !found {
			watchedStocks = append(watchedStocks, ticker)
			sort.Strings(watchedStocks)
		}
		watchlistMutex.Unlock()

		if !found {
			saveWatchlist()
			refreshWatcherList()
			newTickerEntry.SetText("")
		}
	})
	newTickerEntry.OnSubmitted = func(s string) {
		addTickerBtn.OnTapped()
	}

	topContent := container.NewBorder(nil, nil, nil, addTickerBtn, newTickerEntry)
	content := container.NewBorder(topContent, nil, nil, nil, scrollableList)

	refreshWatcherList = func() {
		tickerList.Objects = nil
		watchlistMutex.Lock()
		currentWatchlist := make([]string, len(watchedStocks))
		copy(currentWatchlist, watchedStocks)
		watchlistMutex.Unlock()

		for _, ticker := range currentWatchlist {
			tickerForClosure := ticker
			newsContainer := container.NewVBox()
			loading := widget.NewProgressBarInfinite()
			newsContainer.Add(loading)

			removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				dialog.ShowConfirm("Remove Stock", fmt.Sprintf("Remove '%s' from watchlist?", tickerForClosure), func(confirm bool) {
					if confirm {
						watchlistMutex.Lock()
						var updatedWatchlist []string
						for _, t := range watchedStocks {
							if t != tickerForClosure {
								updatedWatchlist = append(updatedWatchlist, t)
							}
						}
						watchedStocks = updatedWatchlist
						watchlistMutex.Unlock()
						saveWatchlist()
						refreshWatcherList()
					}
				}, watcherWin)
			})
			removeBtn.Importance = widget.LowImportance

			var headerContent *container.Box
			go func() {
				stockData, err := fetchStockData(tickerForClosure)
				if err != nil {
					logger.Warn("Failed to fetch stock data", zap.String("ticker", tickerForClosure), zap.Error(err))
					headerContent = container.NewHBox(
						widget.NewLabelWithStyle(tickerForClosure, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						widget.NewLabel("Stock data unavailable"),
					)
				} else {
					headerContent = container.NewHBox(
						widget.NewLabelWithStyle(tickerForClosure, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
						widget.NewLabel(fmt.Sprintf("Price: $%.2f", stockData.Price)),
						widget.NewLabel(fmt.Sprintf("Change: %.2f (%.2f%%)", stockData.Change, stockData.ChangePct)),
					)
				}
				header := container.NewBorder(nil, nil, headerContent, removeBtn)
				cardContent := container.NewVBox(header, widget.NewSeparator(), newsContainer)
				tickerList.Objects = append(tickerList.Objects, widget.NewCard("", "", cardContent))
				tickerList.Refresh()
			}()

			go func() {
				articles, _, err := app.fetchNewsFromProviders(tickerForClosure, "", "", 1, 5)
				loading.Hide()
				newsContainer.Objects = nil

				if err != nil {
					newsContainer.Add(widget.NewLabel("Error fetching news: " + err.Error()))
					newsContainer.Refresh()
					return
				}

				if len(articles) == 0 {
					newsContainer.Add(widget.NewLabel("No recent articles found."))
				} else {
					for _, article := range articles {
						titleLabel := widget.NewLabel(article.Title)
						titleLabel.Wrapping = fyne.TextWrapWord
						urlLink, _ := url.Parse(article.URL)
						link := widget.NewHyperlink("Read", urlLink)
						articleRow := container.NewBorder(nil, nil, nil, link, titleLabel)
						newsContainer.Add(articleRow)
					}
				}
				newsContainer.Refresh()
			}()
		}
		tickerList.Refresh()
	}

	watcherWin.SetContent(content)
	watcherWin.Show()
	refreshWatcherList()
}

func main() {
	defer logger.Sync()
	myApp := app.NewWithID("com.example.newsaggregator.marketresearch.v3")
	appState := NewAppState()
	ui := appState.SetupUI(myApp)

	// Event handlers
	var lastInputTime time.Time
	ui.QueryInput.OnChanged = func(s string) {
		if len(s) < 3 || time.Since(lastInputTime) < 500*time.Millisecond {
			return
		}
		lastInputTime = time.Now()
		go func() {
			suggestions := []string{"market", "stocks", "economy"} // Placeholder
			suggestionList := widget.NewList(
				func() int { return len(suggestions) },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(i widget.ListItemID, o fyne.CanvasObject) {
					o.(*widget.Label).SetText(suggestions[i])
				},
			)
			suggestionList.OnSelected = func(id widget.ListItemID) {
				ui.QueryInput.SetText(suggestions[id])
				appState.HandleSearch(ui, myApp)
			}
			popUp := widget.NewPopUp(suggestionList, ui.Window.Canvas())
			popUp.ShowAtPosition(ui.QueryInput.Position().Add(fyne.NewPos(0, ui.QueryInput.Size().Height)))
		}()
	}

	ui.SearchBtn.OnTapped = func() { appState.HandleSearch(ui, myApp) }
	ui.SortBtn.OnTapped = func() { appState.HandleSort(ui) }
	ui.SentimentFilter.OnChanged = func(s string) { appState.HandleSentimentFilter(ui, s) }
	ui.ExportBtn.OnTapped = func() { appState.HandleExport(ui, myApp) }
	ui.ClipboardBtn.OnTapped = func() { appState.HandleClipboard(ui, myApp) }
	ui.BookmarksBtn.OnTapped = func() { appState.ShowBookmarksView(ui, myApp) }
	ui.TrendBtn.OnTapped = func() { appState.ShowTrendAnalysisDialog(ui, myApp) }
	ui.WatcherBtn.OnTapped = func() { appState.ShowStockWatcherView(ui, myApp) }

	ui.LoadMoreBtn.OnTapped = func() {
		appState.CurrentPage++
		fetchedArticles, _, err := appState.fetchNewsFromProviders(appState.LastQuery, appState.LastFromDate, appState.LastToDate, appState.CurrentPage, 18)
		if err != nil {
			logger.Error("Failed to load more articles", zap.Error(err))
			myApp.SendNotification(&fyne.Notification{Title: "Load More Error", Content: err.Error()})
			appState.CurrentPage--
			return
		}
		if len(fetchedArticles) > 0 {
			appState.Articles = append(appState.Articles, fetchedArticles...)
			switch appState.SortMode {
			case SortTimeDesc:
				sortByTime(appState.Articles, false)
			case SortTimeAsc:
				sortByTime(appState.Articles, true)
			case SortSentimentDesc:
				sortBySentiment(appState.Articles, false)
			case SortSentimentAsc:
				sortBySentiment(appState.Articles, true)
			}
			appState.RefreshResultsUI(ui)
			ui.Results.ScrollToBottom()
		}
		if len(appState.Articles) >= appState.TotalResults || len(fetchedArticles) == 0 {
			ui.LoadMoreBtn.Hide()
		} else {
			ui.LoadMoreBtn.Show()
		}
	}

	ui.QueryInput.OnSubmitted = func(s string) { appState.HandleSearch(ui, myApp) }
	ui.FromDateEntry.OnSubmitted = func(s string) { appState.HandleSearch(ui, myApp) }
	ui.ToDateEntry.OnSubmitted = func(s string) { appState.HandleSearch(ui, myApp) }

	utilityRow := container.NewHBox(layout.NewSpacer(), ui.WatcherBtn, ui.TrendBtn, ui.BookmarksBtn, ui.ExportBtn, ui.ClipboardBtn, layout.NewSpacer())
	searchRow := container.NewBorder(nil, nil, nil, container.NewHBox(ui.SearchBtn, ui.SortBtn), ui.QueryInput)
	topControls := container.NewVBox(ui.SentimentFilter, ui.KeyInput, searchRow, dateFilterRow, utilityRow, widget.NewSeparator())
	content := container.NewBorder(topControls, ui.LoadMoreBtn, nil, nil, ui.Results)
	ui.Window.SetContent(content)
	ui.Window.ShowAndRun()
}
