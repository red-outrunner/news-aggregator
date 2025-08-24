package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Article represents a news article
type Article struct {
	Title             string `json:"title"`
	URLToImage        string `json:"urlToImage"`
	ImpactScore       int    `json:"impactScore,omitempty"`
	PolicyProbability int    `json:"policyProbability,omitempty"`
	Description       string `json:"description"`
	URL               string `json:"url"`
	PublishedAt       string `json:"publishedAt"`
	SentimentScore    int    `json:"sentimentScore,omitempty"`
}

// NewsAPI response structure
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Config holds application configuration
type Config struct {
	APIKey      string `json:"api_key"`
	MaxArticles int    `json:"max_articles"`
}

var (
	positiveKeywordsSet map[string]struct{}
	safeFilenameRegex   *regexp.Regexp
	negativeKeywordsSet map[string]struct{}
	positivePhrases     []string
	negativePhrases     []string

	articles []Article
	config   Config
)

// Positive and negative keywords for sentiment analysis
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

// Phrases for more accurate sentiment analysis
var positivePhrasesList = []string{
	"strong results", "exceeded expectations", "record high", "beats estimates",
	"outperforms market", "positive outlook", "upbeat forecast", "robust growth",
	"solid performance", "impressive gains", "significant improvement",
}

var negativePhrasesList = []string{
	"fell short", "missed expectations", "record low", "disappointing results",
	"underperforms market", "negative outlook", "bleak forecast", "steep decline",
	"poor performance", "significant losses", "sharp drop", "market crash",
}

// init function to initialize keyword sets
func init() {
	positiveKeywordsSet = make(map[string]struct{}, len(positiveKeywords))
	for _, k := range positiveKeywords {
		positiveKeywordsSet[strings.ToLower(k)] = struct{}{}
	}

	negativeKeywordsSet = make(map[string]struct{}, len(negativeKeywords))
	for _, k := range negativeKeywords {
		negativeKeywordsSet[strings.ToLower(k)] = struct{}{}
	}

	positivePhrases = positivePhrasesList
	negativePhrases = negativePhrasesList

	safeFilenameRegex = regexp.MustCompile(`[^\w-]`)
}

// calculateSentimentScore calculates a sentiment score based on keywords and phrases
func calculateSentimentScore(text string) int {
	score := 0
	textLower := strings.ToLower(text)

	// Check for positive phrases first
	for _, phrase := range positivePhrases {
		score += strings.Count(textLower, phrase) * 15
	}

	// Check for negative phrases
	for _, phrase := range negativePhrases {
		score -= strings.Count(textLower, phrase) * 15
	}

	// Check individual words
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

	// Cap the score
	if score > 100 {
		score = 100
	}
	if score < -100 {
		score = -100
	}
	return score
}

// calculateImpactScore calculates an impact score based on important words
func calculateImpactScore(text string) int {
	score := 0
	textLower := strings.ToLower(text)

	// Impactful words that indicate importance
	impactfulWords := []string{
		"major", "significant", "important", "critical", "breaking", "urgent",
		"massive", "huge", "substantial", "considerable", "remarkable",
		"dramatic", "drastic", "severe", "extreme", "exceptional",
	}

	for _, word := range impactfulWords {
		score += strings.Count(textLower, word) * 5
	}

	// Cap the score
	if score > 100 {
		return 100
	}
	return score
}

// sortByLatest sorts articles by publication date (newest first)
func sortByLatest(articles []Article) {
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PublishedAt > articles[j].PublishedAt
	})
}

// sortBySentiment sorts articles by sentiment score
func sortBySentiment(articles []Article, ascending bool) {
	sort.Slice(articles, func(i, j int) bool {
		if ascending {
			return articles[i].SentimentScore < articles[j].SentimentScore
		}
		return articles[i].SentimentScore > articles[j].SentimentScore
	})
}

// sortByImpact sorts articles by impact score
func sortByImpact(articles []Article, ascending bool) {
	sort.Slice(articles, func(i, j int) bool {
		if ascending {
			return articles[i].ImpactScore < articles[j].ImpactScore
		}
		return articles[i].ImpactScore > articles[j].ImpactScore
	})
}

// fetchNews fetches news articles from NewsAPI
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

	// Calculate scores for each article
	for i := range newsResponse.Articles {
		content := newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description
		newsResponse.Articles[i].ImpactScore = calculateImpactScore(content)
		newsResponse.Articles[i].SentimentScore = calculateSentimentScore(content)
	}

	return newsResponse.Articles, nil
}

// displayArticles displays articles in a formatted way
func displayArticles(articles []Article) {
	if len(articles) == 0 {
		fmt.Println("No articles to display.")
		return
	}

	for i, article := range articles {
		fmt.Printf("%d. [%s] [Sentiment: %d] [Impact: %d]\n", 
			i+1, humanTime(article.PublishedAt), article.SentimentScore, article.ImpactScore)
		fmt.Printf("Title: %s\n", article.Title)
		if article.Description != "" {
			fmt.Printf("Description: %s\n", article.Description)
		}
		fmt.Printf("URL: %s\n", article.URL)
		fmt.Println()

		if i < len(articles)-1 {
			fmt.Println("#####Next article#####")
		}
	}
}

// filterBySentiment filters articles by sentiment
func filterBySentiment(articles []Article, positive bool) []Article {
	var filtered []Article
	for _, article := range articles {
		if positive && article.SentimentScore > 0 {
			filtered = append(filtered, article)
		} else if !positive && article.SentimentScore < 0 {
			filtered = append(filtered, article)
		}
	}
	return filtered
}

// filterByImpact filters articles by impact score
func filterByImpact(articles []Article, threshold int) []Article {
	var filtered []Article
	for _, article := range articles {
		if article.ImpactScore >= threshold {
			filtered = append(filtered, article)
		}
	}
	return filtered
}

// humanTime converts RFC3339 time to human-readable format
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

// loadConfig loads configuration from file
func loadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "news-aggregator")
	configFile := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{MaxArticles: 18}, nil
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Set defaults if not specified
	if cfg.MaxArticles == 0 {
		cfg.MaxArticles = 18
	}

	return &cfg, nil
}

// saveConfig saves configuration to file
func saveConfig(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user's home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "news-aggregator")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	return ioutil.WriteFile(configFile, data, 0600)
}

// loadAPIKey loads the API key from config
func loadAPIKey() string {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return ""
	}
	config = *cfg
	return config.APIKey
}

// saveAPIKey saves the API key to config
func saveAPIKey(key string) error {
	if key == "" || len(key) < 30 {
		return fmt.Errorf("invalid API key format")
	}

	config.APIKey = key
	return saveConfig(&config)
}

// showHelp displays help information
func showHelp() {
	fmt.Println("News Aggregator CLI - Help")
	fmt.Println("==========================")
	fmt.Println("Commands:")
	fmt.Println("  <search term>        - Search for news articles")
	fmt.Println("  sort latest          - Sort by publication date (newest first)")
	fmt.Println("  sort sentiment       - Sort by sentiment score (most positive first)")
	fmt.Println("  sort impact          - Sort by impact score (highest first)")
	fmt.Println("  filter positive      - Show only positive sentiment articles")
	fmt.Println("  filter negative      - Show only negative sentiment articles")
	fmt.Println("  filter impact <num>  - Show articles with impact score >= num")
	fmt.Println("  config max <num>     - Set maximum articles to display")
	fmt.Println("  help                 - Show this help message")
	fmt.Println("  exit                 - Exit the application")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tesla                - Search for articles about Tesla")
	fmt.Println("  sort sentiment       - Sort current articles by sentiment")
	fmt.Println("  filter impact 50     - Show high-impact articles")
	fmt.Println("  config max 10        - Display maximum 10 articles")
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("Welcome to the News Aggregator CLI!")
	fmt.Println("Type 'help' for available commands.")

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Warning: Could not load config:", err)
		config = Config{MaxArticles: 18}
	} else {
		config = *cfg
	}

	// Load API key
	apiKey := loadAPIKey()
	if apiKey == "" {
		fmt.Print("Please enter your NewsAPI key: ")
		fmt.Scanln(&apiKey)
		if err := saveAPIKey(apiKey); err != nil {
			fmt.Printf("Error saving API key: %v\n", err)
			return
		}
		fmt.Println("API key saved successfully!")
	}

	for {
		fmt.Print("\nEnter command or search term: ")
		var input string
		fmt.Scanln(&input)

		// Parse input for commands with arguments
		parts := strings.Fields(strings.ToLower(input))
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "sort":
			if len(parts) < 2 {
				fmt.Println("Usage: sort [latest|sentiment|impact]")
				continue
			}
			if len(articles) == 0 {
				fmt.Println("No articles fetched yet. Please perform a search first.")
				continue
			}

			switch parts[1] {
			case "latest":
				sortByLatest(articles)
				displayArticles(articles[:min(len(articles), config.MaxArticles)])
			case "sentiment":
				sortBySentiment(articles, false) // Descending (most positive first)
				displayArticles(articles[:min(len(articles), config.MaxArticles)])
			case "impact":
				sortByImpact(articles, false) // Descending (highest impact first)
				displayArticles(articles[:min(len(articles), config.MaxArticles)])
			default:
				fmt.Println("Invalid sort option. Use: latest, sentiment, or impact")
			}

		case "filter":
			if len(parts) < 2 {
				fmt.Println("Usage: filter [positive|negative|impact <number>]")
				continue
			}
			if len(articles) == 0 {
				fmt.Println("No articles fetched yet. Please perform a search first.")
				continue
			}

			switch parts[1] {
			case "positive":
				filtered := filterBySentiment(articles, true)
				if len(filtered) == 0 {
					fmt.Println("No positive sentiment articles found.")
				} else {
					fmt.Printf("Showing %d positive sentiment articles:\n\n", len(filtered))
					displayArticles(filtered[:min(len(filtered), config.MaxArticles)])
				}
			case "negative":
				filtered := filterBySentiment(articles, false)
				if len(filtered) == 0 {
					fmt.Println("No negative sentiment articles found.")
				} else {
					fmt.Printf("Showing %d negative sentiment articles:\n\n", len(filtered))
					displayArticles(filtered[:min(len(filtered), config.MaxArticles)])
				}
			case "impact":
				if len(parts) < 3 {
					fmt.Println("Usage: filter impact <number>")
					continue
				}
				threshold := 0
				fmt.Sscanf(parts[2], "%d", &threshold)
				filtered := filterByImpact(articles, threshold)
				if len(filtered) == 0 {
					fmt.Printf("No articles found with impact score >= %d.\n", threshold)
				} else {
					fmt.Printf("Showing %d articles with impact score >= %d:\n\n", len(filtered), threshold)
					displayArticles(filtered[:min(len(filtered), config.MaxArticles)])
				}
			default:
				fmt.Println("Invalid filter option. Use: positive, negative, or impact")
			}

		case "config":
			if len(parts) < 3 {
				fmt.Println("Usage: config max <number>")
				continue
			}
			if parts[1] == "max" {
				var max int
				if _, err := fmt.Sscanf(parts[2], "%d", &max); err != nil || max <= 0 {
					fmt.Println("Invalid number. Please enter a positive integer.")
					continue
				}
				config.MaxArticles = max
				if err := saveConfig(&config); err != nil {
					fmt.Printf("Error saving config: %v\n", err)
				} else {
					fmt.Printf("Maximum articles set to %d\n", max)
				}
			} else {
				fmt.Println("Invalid config option. Use: max")
			}

		case "help":
			showHelp()

		case "exit":
			fmt.Println("Exiting...")
			return

		default:
			// Treat as search query
			searchQuery := strings.Join(parts, " ")
			fmt.Printf("Searching for: %s\n", searchQuery)

			newArticles, err := fetchNews(apiKey, searchQuery)
			if err != nil {
				fmt.Printf("Error fetching news: %v\n", err)
				continue
			}

			if len(newArticles) > 0 {
				articles = newArticles
				sortByLatest(articles) // Default sort by latest
				if len(articles) > config.MaxArticles {
					articles = articles[:config.MaxArticles]
				}
				fmt.Printf("Found %d articles. Showing latest %d:\n\n", len(newArticles), len(articles))
				displayArticles(articles)
			} else {
				fmt.Println("No articles found for the given query.")
			}
		}
	}
}
