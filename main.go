package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "sort"
)


type Article struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    URL         string `json:"url"`
    PublishedAt string `json:"publishedAt"`
}

// NewsAPI
type NewsResponse struct {
    Status  string    `json:"status"`
    TotalResults int   `json:"totalResults"`
    Articles []Article `json:"articles"`
}

// sortByLatest 
func sortByLatest(articles []Article) {
    sort.Slice(articles, func(i, j int) bool {
        return articles[i].PublishedAt > articles[j].PublishedAt
    })
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

    return newsResponse.Articles, nil
}

// displayArticles 
func displayArticles(articles []Article) {
    for i, article := range articles {
        fmt.Printf("%d. [%s]\nTitle: %s\nDescription: %s\nURL: %s\n\n", i+1, article.PublishedAt, article.Title, article.Description, article.URL)

        if i < len(articles)-1 {
        	fmt.Println("#####Next article#####")
        }
    }
}

func main() {
    fmt.Println("Welcome to the News-aggregator program!")
    fmt.Print("Enter a keyword, company name, person, or topic: ")
    var query string
    fmt.Scanln(&query)

    // Replace with your own NewsAPI key
    apiKey := "YOUR_NEWS_API_KEY"

    if apiKey == "YOUR_NEWS_API_KEY" {
        fmt.Println("Please replace 'YOUR_NEWS_API_KEY' with your actual NewsAPI key.")
        os.Exit(1)
    }

    // fetchNews 
    articles, err := fetchNews(apiKey, query)
    if err != nil {
        fmt.Printf("Error fetching news: %v\n", err)
        os.Exit(1)
    }

    // Sort activate
    sortByLatest(articles)

    // Limit of articles
    if len(articles) > 18 {
        articles = articles[:18]
    }

    //  articles display
    if len(articles) == 0 {
        fmt.Println("No articles found for the given query.")
    } else {
        fmt.Printf("Here are the latest %d articles:\n\n", len(articles))
        displayArticles(articles)
    }
}
