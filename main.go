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

type Article struct {
	Title             string `json:"title"`
	URLToImage        string `json:"urlToImage"`
	ImpactScore       int    `json:"impactScore,omitempty"`
	PolicyProbability int    `json:"policyProbability,omitempty"`
	Description       string `json:"description"`
	URL               string `json:"url"`
	PublishedAt       string `json:"publishedAt"`
}

// NewsAPI
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

var (
	positiveKeywordsSet map[string]struct{}
	safeFilenameRegex   *regexp.Regexp
	negativeKeywordsSet map[string]struct{}

	articles []Article
	apiKey   string
)

// sortByLatest
func sortByLatest(articles []Article) {
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].PublishedAt > articles[j].PublishedAt
	})
}

func sortBySentiment(articles []Article, ascending bool) {
	sort.Slice(articles, func(i, j int) bool {
		scoreI := calculateSentimentScore(articles[i].Title + " " + articles[i].Description)
		scoreJ := calculateSentimentScore(articles[j].Title + " " + articles[j].Description)
		if ascending {
			return scoreI < scoreJ
		}
		return scoreI > scoreJ
	})
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
	safeFilenameRegex = regexp.MustCompile(`[^\w-]`)
}

// calculateSentimentScore calculates a sentiment score based on keywords
func calculateSentimentScore(text string) int {
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
	return 0
}

// fetchNews
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
	for i := range newsResponse.Articles {
		newsResponse.Articles[i].ImpactScore = calculateImpactScore(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
		newsResponse.Articles[i].SentimentScore = calculateSentimentScore(newsResponse.Articles[i].Title + " " + newsResponse.Articles[i].Description)
	}

	return newsResponse.Articles, nil
}

// displayArticles
func displayArticles(articles []Article) {
	for i, article := range articles {
		fmt.Printf("%d. [%s] [Sentiment: %d]\nTitle: %s\nDescription: %s\nURL: %s\n\n", i+1, humanTime(article.PublishedAt), article.SentimentScore, article.Title, article.Description, article.URL)

		if i < len(articles)-1 {
			fmt.Println("#####Next article#####")
		}
	}
}

func main() {
	fmt.Println("Welcome to the News Aggregator CLI!")

	// Load API key
	apiKey = loadAPIKey()
	if apiKey == "" {
		fmt.Print("Please enter your NewsAPI key: ")
		fmt.Scanln(&apiKey)
		saveAPIKey(apiKey)
	}

	for {
		fmt.Print("\nEnter a keyword, company name, person, or topic (or 'sort latest', 'sort sentiment', or 'exit'): ")
		var input string
		fmt.Scanln(&input)

		switch strings.ToLower(input) {
		case "sort latest":
			if len(articles) > 0 {
				sortByLatest(articles)
				displayArticles(articles)
			} else {
				fmt.Println("No articles fetched yet. Please perform a search first.")
			}
			continue
		case "sort sentiment":
			if len(articles) > 0 {
				sortBySentiment(articles, false) // Sort by sentiment descending
				displayArticles(articles)
			} else {
				fmt.Println("No articles fetched yet. Please perform a search first.")
			}
			continue
		case "exit":
			fmt.Println("Exiting...")
			return
		}

		articles, err := fetchNews(apiKey, input)
		if err != nil {
			fmt.Printf("Error fetching news: %v\n", err)
			continue
		}

		if len(articles) > 0 {
			sortByLatest(articles) // Default sort by latest
			if len(articles) > 18 {
				articles = articles[:18]
			}
			fmt.Printf("Here are the latest %d articles:\n\n", len(articles))
			displayArticles(articles)
		} else {
			fmt.Println("No articles found for the given query.")
		}
	}
}

// loadAPIKey loads the API key from a file in the user's config directory.
func loadAPIKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return ""
	}

	configDir := filepath.Join(home, ".config", "news-aggregator")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return ""
	}

	keyFile := filepath.Join(configDir, "api_key.txt")
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return ""
	}

	data, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println("Error reading API key file:", err)
		return ""
	}

	return strings.TrimSpace(string(data))
}

// saveAPIKey saves the API key to a file in the user's config directory.
func saveAPIKey(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user's home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "news-aggregator")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	keyFile := filepath.Join(configDir, "api_key.txt")

	if apiKey == "YOUR_NEWS_API_KEY" {
		fmt.Println("Please replace 'YOUR_NEWS_API_KEY' with your actual NewsAPI key.")
		os.Exit(1)
	}

	return os.WriteFile(keyFile, []byte(key), 0600)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
