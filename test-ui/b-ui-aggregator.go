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

// Article struct updated to include URLToImage
type Article struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	URL               string `json:"url"`
	URLToImage        string `json:"urlToImage"` // Added for image thumbnail
	PublishedAt       string `json:"publishedAt"`
	ImpactScore       int    `json:"impactScore,omitempty"`
	PolicyProbability int    `json:"policyProbability,omitempty"`
	// IsBookmarked bool // Not strictly needed in struct if we check against a separate list by URL
}

// NewsResponse struct remains the same
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

var (
	bookmarkedArticles []Article
	bookmarksMutex     sync.Mutex // To protect concurrent access to bookmarkedArticles
	bookmarksFilePath  string
)

const bookmarksFilename = "news_aggregator_bookmarks.json"

// sortByTime function remains the same
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

// humanTime function remains the same
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
		return t.Format("Jan 2, 2006") // Slightly more informative default
	}
}

// fetchNews function remains the same (URLToImage will be unmarshalled automatically)
func fetchNews(apiKey, query string, page int) ([]Article, int, error) {
	apiURL := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&sortBy=publishedAt&language=en&pageSize=18&page=%d&apiKey=%s", url.QueryEscape(query), page, apiKey)
	resp, err := http.Get(apiURL)
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
	}

	return newsResponse.Articles, newsResponse.TotalResults, nil
}

// loadSavedKey function remains the same
func loadSavedKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home dir:", err)
		return ""
	}
	path := filepath.Join(home, ".config", "news_aggregator_apikey.txt") // More specific filename
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error reading API key:", err)
		}
		return ""
	}
	return strings.TrimSpace(string(data))
}

// saveAPIKey function remains the same
func saveAPIKey(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home dir: %w", err)
	}
	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	path := filepath.Join(dir, "news_aggregator_apikey.txt") // More specific filename
	return os.WriteFile(path, []byte(key), 0600)
}

// loadThemePreference function remains the same
func loadThemePreference() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home dir for theme:", err)
		return false // Default to light theme on error
	}
	path := filepath.Join(home, ".config", "news_aggregator_theme.txt") // More specific filename
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error reading theme preference:", err)
		}
		return false // Default to light theme
	}
	return strings.TrimSpace(string(data)) == "dark"
}

// saveThemePreference function remains the same
func saveThemePreference(isDark bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home dir for theme: %w", err)
	}
	dir := filepath.Join(home, ".config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error creating config directory for theme: %w", err)
	}
	path := filepath.Join(dir, "news_aggregator_theme.txt") // More specific filename
	theme := "light"
	if isDark {
		theme = "dark"
	}
	return os.WriteFile(path, []byte(theme), 0600)
}

// --- Bookmark Functions ---
func setupBookmarksPath() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home dir for bookmarks:", err)
		// Fallback to current directory if home cannot be determined
		bookmarksFilePath = bookmarksFilename
		return
	}
	configDir := filepath.Join(home, ".config", "newsaggregator")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Println("Error creating config directory for bookmarks:", err)
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
		if os.IsNotExist(err) {
			bookmarkedArticles = []Article{} // Initialize if file doesn't exist
			return
		}
		fmt.Println("Error reading bookmarks file:", err)
		bookmarkedArticles = []Article{}
		return
	}
	if err := json.Unmarshal(data, &bookmarkedArticles); err != nil {
		fmt.Println("Error unmarshalling bookmarks:", err)
		bookmarkedArticles = []Article{}
	}
}

func saveBookmarks() {
	bookmarksMutex.Lock()
	defer bookmarksMutex.Unlock()

	data, err := json.MarshalIndent(bookmarkedArticles, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling bookmarks:", err)
		return
	}
	if err := os.WriteFile(bookmarksFilePath, data, 0600); err != nil {
		fmt.Println("Error writing bookmarks file:", err)
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
	// Unlock before calling saveBookmarks, then re-lock for the defer
	// This is to avoid deadlock as saveBookmarks also locks.
	// A more robust solution might involve passing the lock or using channels.
	
	found := false
	var updatedBookmarks []Article
	for _, bm := range bookmarkedArticles {
		if bm.URL == article.URL {
			found = true
			// Skip adding it to updatedBookmarks to remove it
		} else {
			updatedBookmarks = append(updatedBookmarks, bm)
		}
	}

	if !found {
		// Add the article if it wasn't found (i.e., it's a new bookmark)
		isAlreadyThere := false
		for _, ub := range updatedBookmarks { 
			if ub.URL == article.URL {
				isAlreadyThere = true
				break
			}
		}
		if !isAlreadyThere {
			updatedBookmarks = append(updatedBookmarks, article)
		}
	}
	bookmarkedArticles = updatedBookmarks
	bookmarksMutex.Unlock() // Unlock before save
	saveBookmarks() 
	// No need to re-lock here if the function is ending.
	// If more operations followed that needed the lock, then re-locking would be necessary.
}


// summarizeText function remains the same
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


// calculateImpactScore function remains the same
func calculateImpactScore(text string) int {
	keywords := []string{"crisis", "breakthrough", "disaster", "economy", "war", "pandemic", "reform", "urgent", "major", "global", "election", "protest", "conflict", "threat"}
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(textLower, k) {
			score += 7 
		}
	}
	return min(100, score) 
}

// calculatePolicyProbability function remains the same
func calculatePolicyProbability(text string) int {
	keywords := []string{"policy", "regulation", "law", "government", "legislation", "bill", "congress", "senate", "parliament", "decree", "treaty", "court", "ruling", "initiative"}
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(textLower, k) {
			score += 10 
		}
	}
	return min(100, score) 
}

// min function remains the same
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// askAI function. Simplified "and more" logic.
func askAI(question string, articles []Article) string {
	searchQuery := strings.ToLower(strings.TrimSpace(question))
	if searchQuery == "" {
		return "Please ask a specific question."
	}

	var relevantArticlesOutput []string
	questionWords := strings.Fields(searchQuery)
	initialRelevantCount := 0 

	for _, a := range articles {
		content := strings.ToLower(a.Title + " " + a.Description)
		matchCount := 0
		for _, qWord := range questionWords {
			if len(qWord) > 2 && strings.Contains(content, qWord) {
				matchCount++
			}
		}
		if matchCount > 0 {
			summary := summarizeText(a.Description)
			if summary == "No content available to summarize." && a.Title != "" {
				summary = a.Title
			}
			relevantArticlesOutput = append(relevantArticlesOutput, fmt.Sprintf("- %s: %s", a.Title, summary))
			initialRelevantCount++
		}
	}

	if len(relevantArticlesOutput) == 0 {
		return "No relevant information found in the currently loaded articles for your question: '" + question + "'"
	}

	if initialRelevantCount > 5 {
		relevantArticlesOutput = relevantArticlesOutput[:5] 
		relevantArticlesOutput = append(relevantArticlesOutput, fmt.Sprintf("\n(And %d more relevant articles found...)", initialRelevantCount-5))
	}

	return "Based on the articles, here's some information related to '" + question + "':\n\n" + strings.Join(relevantArticlesOutput, "\n\n")
}


// Main application function
func main() {
	myApp := app.NewWithID("com.example.newsaggregator.deluxe.v2") // Updated AppID
	myWindow := myApp.NewWindow("News Aggregator Deluxe")
	myWindow.Resize(fyne.NewSize(900, 800)) // Slightly larger for new button

	setupBookmarksPath()
	loadBookmarks() // Load bookmarks at startup

	// Load theme preference
	isDarkTheme := loadThemePreference()
	if isDarkTheme {
		myApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		myApp.Settings().SetTheme(theme.LightTheme())
	}

	// --- Data Variables ---
	var currentPage = 1
	var totalResults = 0
	var allArticles []Article
	var lastQuery string
	sortAsc := false // Default sort: Newest first

	// --- UI Elements ---
	keyInput := widget.NewPasswordEntry() 
	keyInput.SetPlaceHolder("Enter NewsAPI key...")
	apiKey := loadSavedKey()
	if apiKey != "" {
		keyInput.SetText(apiKey)
	}

	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil) 
	updateThemeButtonText := func(isDark bool) {
		if isDark {
			themeBtn.SetText("Set Light Theme")
		} else {
			themeBtn.SetText("Set Dark Theme")
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
		saveThemePreference(isDarkTheme)
	}
	
	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyRow := container.NewBorder(nil, nil, apiKeyLabel, themeBtn, keyInput)

	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search news topics (e.g., 'Go programming', 'AI breakthroughs')")

	results := container.NewVBox() 
	results.Add(widget.NewLabelWithStyle("Enter your API key and a search query to begin exploring news.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))

	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(300,400)) 

	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide() 

	loadMoreBtn := widget.NewButtonWithIcon("Load More Articles", theme.MoreVerticalIcon(), nil)
	loadMoreBtn.Hide() 
	loadMoreContainer := container.NewCenter(loadMoreBtn)

	var refreshResultsUI func() // Declare for mutual recursion with showBookmarksView if needed
	var showBookmarksView func() // Declare for mutual recursion

	refreshResultsUI = func() {
		results.Objects = nil 

		if len(allArticles) == 0 {
			if lastQuery != "" { 
				results.Add(widget.NewLabelWithStyle("No articles found for your query: '"+lastQuery+"'.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			} else {
				results.Add(widget.NewLabelWithStyle("Enter your API key and a search query to begin exploring news.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			}
			results.Refresh()
			return
		}

		for i := range allArticles {
			article := allArticles[i] 

			parsedURL, err := url.Parse(article.URL)
			if err != nil {
				fmt.Printf("Error parsing URL %s: %v\n", article.URL, err)
				parsedURL, _ = url.Parse("https://example.com/invalid-url") 
			}

			cardDescription := article.Description
			if len(cardDescription) > 180 { 
				cardDescription = cardDescription[:180]
				if idx := strings.LastIndex(cardDescription, " "); idx != -1 {
					cardDescription = cardDescription[:idx] + "..."
				} else {
					cardDescription += "..."
				}
			}
			if strings.TrimSpace(cardDescription) == "" {
				cardDescription = "No description available for this article."
			}
			
			descriptionLabel := widget.NewLabel(cardDescription)
			descriptionLabel.Wrapping = fyne.TextWrapWord
			
			imgWidget := canvas.NewImageFromResource(nil) 
			imgWidget.FillMode = canvas.ImageFillContain
			imgWidget.SetMinSize(fyne.NewSize(120, 80)) 

			if article.URLToImage != "" {
				// Removed appRef parameter from goroutine signature and call
				go func(imgURL string, targetImg *canvas.Image) { 
					client := http.Client{ Timeout: 15 * time.Second }
					resp, err := client.Get(imgURL)
					if err != nil { fmt.Printf("Error fetching image %s: %v\n", imgURL, err); return }
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK { fmt.Printf("Error fetching image %s: status %s\n", imgURL, resp.Status); return }
					imgData, err := io.ReadAll(resp.Body)
					if err != nil { fmt.Printf("Error reading image data %s: %v\n", imgURL, err); return }
					_, _, err = image.Decode(bytes.NewReader(imgData))
					if err != nil { fmt.Printf("Error decoding image %s: %v\n", imgURL, err); return }
					imgRes := fyne.NewStaticResource(filepath.Base(imgURL), imgData)
					
					// Directly update and refresh the image widget from the goroutine
					// This is a workaround if RunOnMain APIs are consistently failing.
					// It might not be perfectly thread-safe in all Fyne versions/scenarios
					// but is the simplest alternative if main thread execution calls are problematic.
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
					// Corrected: Use theme.StarIcon() as a fallback for theme.BookmarkIcon()
					btn.SetIcon(theme.StarIcon()) 
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
				fullSummary := summarizeText(currentArticleForDetail.Description)
				if strings.TrimSpace(currentArticleForDetail.Description) == "" {
					fullSummary = "Full description is not available."
				}
				currentArticleParsedURL, _ := url.Parse(currentArticleForDetail.URL)
				content := container.NewVBox(
					widget.NewLabelWithStyle(currentArticleForDetail.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewSeparator(),
					widget.NewLabelWithStyle("Summary:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabel(fullSummary), 
					widget.NewSeparator(),
					widget.NewLabel(fmt.Sprintf("Impact Score: %d/100", currentArticleForDetail.ImpactScore)),
					widget.NewLabel(fmt.Sprintf("Policy Relevance: %d%%", currentArticleForDetail.PolicyProbability)), 
					widget.NewSeparator(),
					widget.NewHyperlink("Open Original Article", currentArticleParsedURL),
				)
				for _, obj := range content.Objects {
					if lbl, ok := obj.(*widget.Label); ok { lbl.Wrapping = fyne.TextWrapWord }
				}
				var detailPopUp *widget.PopUp 
				closeButton := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { detailPopUp.Hide() }) 
				dialogContainer := container.NewBorder(nil, container.NewCenter(closeButton), nil, nil, container.NewVScroll(content))
				detailPopUp = widget.NewModalPopUp(dialogContainer, myWindow.Canvas())
				detailPopUp.Resize(fyne.NewSize(myWindow.Canvas().Size().Width*0.8, myWindow.Canvas().Size().Height*0.7))
				detailPopUp.Show()
			})

			textContent := container.NewVBox(
				descriptionLabel,
				widget.NewSeparator(),
				container.NewGridWithColumns(2,
					widget.NewLabel(fmt.Sprintf("Impact: %d", article.ImpactScore)),
					widget.NewLabel(fmt.Sprintf("Policy: %d%%", article.PolicyProbability)),
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
			card := widget.NewCard(
				article.Title,
				fmt.Sprintf("Published: %s", humanTime(article.PublishedAt)),
				cardContentWithImage, 
			)
			results.Add(card)
		}
		results.Refresh()
		if !sortAsc && currentPage == 1 { 
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
				listContent.Add(widget.NewLabelWithStyle("No articles bookmarked yet.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			} else {
				for _, bmArticle := range currentBookmarks {
					articleForView := bmArticle

					titleLabel := widget.NewLabelWithStyle(articleForView.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
					descLabel := widget.NewLabel(summarizeText(articleForView.Description))
					descLabel.Wrapping = fyne.TextWrapWord
					urlLink, _ := url.Parse(articleForView.URL)

					removeBtn := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), func() {
						dialog.ShowConfirm("Remove Bookmark", fmt.Sprintf("Are you sure you want to remove '%s'?", articleForView.Title), func(confirm bool) {
							if confirm {
								toggleBookmark(articleForView) 
								refreshBookmarksList()         
								refreshResultsUI()             
							}
						}, bmWin)
					})

					articleEntry := container.NewVBox(
						titleLabel,
						widget.NewLabel(fmt.Sprintf("Published: %s", humanTime(articleForView.PublishedAt))),
						descLabel,
						container.NewHBox(widget.NewHyperlink("Open Article", urlLink), layout.NewSpacer(), removeBtn),
						widget.NewSeparator(),
					)
					listContent.Add(articleEntry)
				}
			}
			listContent.Refresh()
			scrollableList.Refresh() 
		}

		refreshBookmarksList() 

		bmWin.SetContent(scrollableList)
		bmWin.Show()
	}


	sortBtn := widget.NewButtonWithIcon("Sort: Newest First", theme.MenuDropDownIcon(), nil) 
	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Oldest First")
			sortBtn.SetIcon(theme.MenuDropUpIcon()) 
		} else {
			sortBtn.SetText("Sort: Newest First")
			sortBtn.SetIcon(theme.MenuDropDownIcon()) 
		}
		sortByTime(allArticles, sortAsc)
		refreshResultsUI()
	}
	
	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		key := keyInput.Text
		query := queryInput.Text

		if key == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabelWithStyle("⚠️ API key is required. Please enter it above.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})}
			results.Refresh(); loadMoreBtn.Hide(); loadingIndicator.Hide(); return
		}
		if query == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabelWithStyle("⚠️ Please enter a search query.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})}
			results.Refresh(); loadMoreBtn.Hide(); loadingIndicator.Hide(); return
		}

		results.Objects = nil 
		results.Add(loadingIndicator)
		loadingIndicator.Show()
		results.Refresh()
		
		loadMoreBtn.Hide()
		currentPage = 1
		lastQuery = query
		
		fetchedArticles, total, err := fetchNews(key, query, currentPage)
		
		loadingIndicator.Hide()
		results.Objects = nil 

		if err != nil {
			results.Add(widget.NewLabelWithStyle(fmt.Sprintf("❌ Error fetching news: %v", err), fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}))
			results.Refresh(); loadMoreBtn.Hide(); allArticles = nil; return
		}
		if len(fetchedArticles) == 0 {
			results.Add(widget.NewLabelWithStyle("🔍 No results found for your query: '"+query+"'. Try different keywords.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			results.Refresh(); loadMoreBtn.Hide(); allArticles = nil; return
		}

		totalResults = total
		allArticles = fetchedArticles 
		sortByTime(allArticles, sortAsc) 
		refreshResultsUI()

		if len(allArticles) < totalResults && len(allArticles) > 0 {
			loadMoreBtn.Show()
		} else {
			loadMoreBtn.Hide()
		}
		if errSave := saveAPIKey(key); errSave != nil { 
			fmt.Println("Error saving API key:", errSave) 
		}
	})
	
	queryInput.OnSubmitted = func(s string) { 
		searchBtn.OnTapped()
	}

	searchRow := container.NewBorder(nil, nil, nil, container.NewHBox(searchBtn, sortBtn), queryInput)

	askAIInput := widget.NewEntry()
	askAIInput.SetPlaceHolder("Ask a question about loaded articles...")
	
	showAIResponseDialog := func(title, content string) {
		contentLabel := widget.NewLabel(content)
		contentLabel.Wrapping = fyne.TextWrapWord
		dialogVBox := container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			container.NewVScroll(contentLabel), 
			widget.NewSeparator(),
		)
		var modal *widget.PopUp
		closeDialogBtn := widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { modal.Hide() })
		modal = widget.NewModalPopUp(
			container.NewBorder(nil, container.NewCenter(closeDialogBtn), nil, nil, dialogVBox),
			myWindow.Canvas(),
		)
		modal.Resize(fyne.NewSize(myWindow.Canvas().Size().Width*0.7, myWindow.Canvas().Size().Height*0.6))
		modal.Show()
	}

	askBtn := widget.NewButtonWithIcon("Ask AI", theme.QuestionIcon(), func() {
		question := askAIInput.Text
		if question == "" { showAIResponseDialog("Ask AI Error", "Please enter a question to ask the AI."); return }
		if len(allArticles) == 0 { showAIResponseDialog("Ask AI Info", "No articles loaded. Please perform a search first."); return }
		answer := askAI(question, allArticles) 
		showAIResponseDialog("AI Response", answer)
	})
	askAIInput.OnSubmitted = func(s string) { 
		askBtn.OnTapped()
	}
	
	askAIRow := container.NewBorder(nil, nil, nil, askBtn, askAIInput) 

	exportBtn := widget.NewButtonWithIcon("Export MD", theme.FileTextIcon(), func() {
		if len(allArticles) == 0 { myApp.SendNotification(&fyne.Notification{Title: "Export Info", Content: "No articles to export."}); return }
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# News Articles for Query: %s\n\n", lastQuery))
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n", a.Title))
			sb.WriteString(fmt.Sprintf("- **URL**: <%s>\n", a.URL)) 
			sb.WriteString(fmt.Sprintf("- **Published**: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", strings.TrimSpace(a.Description)))
			sb.WriteString(fmt.Sprintf("- **Impact Score**: %d/100\n", a.ImpactScore))
			sb.WriteString(fmt.Sprintf("- **Policy Relevance**: %d%%\n", a.PolicyProbability))
			sb.WriteString(fmt.Sprintf("- **Summary (2 sentences)**: %s\n", summarizeText(a.Description)))
			sb.WriteString("\n---\n\n")
		}
		home, errHome := os.UserHomeDir()
		if errHome != nil { showAIResponseDialog("Export Error", "Could not determine user home directory."); return }
		docDir := filepath.Join(home, "Documents")
		if _, err := os.Stat(docDir); os.IsNotExist(err) {
			if errMkdir := os.MkdirAll(docDir, 0755); errMkdir != nil { showAIResponseDialog("Export Error", fmt.Sprintf("Could not create Documents directory: %v", errMkdir)); return }
		}
		dateStr := time.Now().Format("2006-01-02")
		safeQuery := strings.ReplaceAll(strings.ToLower(lastQuery), " ", "_")
		safeQuery = strings.ReplaceAll(safeQuery, "/", "_") 
		safeQuery = strings.ReplaceAll(safeQuery, "\\", "_")
		if len(safeQuery) > 30 { safeQuery = safeQuery[:30] } 
		fileName := fmt.Sprintf("news_export_%s_%s.md", safeQuery, dateStr)
		path := filepath.Join(docDir, fileName) 
		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil { showAIResponseDialog("Export Error", fmt.Sprintf("Failed to write file to %s: %v", path, err)); return }
		myApp.SendNotification(&fyne.Notification{Title: "Export Success", Content: fmt.Sprintf("Articles exported to %s", path)})
	})

	clipboardBtn := widget.NewButtonWithIcon("Copy All", theme.ContentCopyIcon(), func() {
		if len(allArticles) == 0 { myApp.SendNotification(&fyne.Notification{Title: "Clipboard Info", Content: "No articles to copy."}); return }
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("News Query: %s\n\n", lastQuery))
		for i, a := range allArticles {
			sb.WriteString(fmt.Sprintf("Article %d: %s\n", i+1, a.Title))
			sb.WriteString(fmt.Sprintf("  Link: %s\n", a.URL))
			sb.WriteString(fmt.Sprintf("  Published: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("  Summary: %s\n", summarizeText(a.Description)))
			sb.WriteString("\n")
		}
		myWindow.Clipboard().SetContent(sb.String())
		myApp.SendNotification(&fyne.Notification{Title: "Clipboard Success", Content: fmt.Sprintf("%d article summaries copied.", len(allArticles))})
	})
	
	// New Bookmarks Button
	bookmarksBtn := widget.NewButtonWithIcon("Bookmarks", theme.FolderOpenIcon(), func() {
		showBookmarksView()
	})

	utilityRow := container.NewHBox(layout.NewSpacer(), bookmarksBtn, exportBtn, clipboardBtn, layout.NewSpacer()) 

	loadMoreBtn.OnTapped = func() {
		currentPage++
		key := keyInput.Text
		query := lastQuery 
		originalBtnText := loadMoreBtn.Text
		loadMoreBtn.SetText("Loading More...")
		loadMoreBtn.Disable()
		fetchedArticles, _, err := fetchNews(key, query, currentPage)
		loadMoreBtn.SetText(originalBtnText)
		loadMoreBtn.Enable()
		if err != nil { myApp.SendNotification(&fyne.Notification{Title: "Load More Error", Content: err.Error()}); currentPage--; return }
		if len(fetchedArticles) > 0 { 
			allArticles = append(allArticles, fetchedArticles...) 
			sortByTime(allArticles, sortAsc) 
			refreshResultsUI()
			scroll.ScrollToBottom() 
		}
		if len(allArticles) >= totalResults || len(fetchedArticles) == 0 { 
			loadMoreBtn.Hide() 
		} else {
			loadMoreBtn.Show()
		}
	}
	
	topControls := container.NewVBox(
		apiKeyRow,
		searchRow, 
		askAIRow,  
		utilityRow,
		widget.NewSeparator(), 
	)

	content := container.NewBorder(
		topControls,       
		loadMoreContainer, 
		nil,               
		nil,               
		scroll,            
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
