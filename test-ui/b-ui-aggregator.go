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

// Article struct remains the same
type Article struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	URL               string `json:"url"`
	PublishedAt       string `json:"publishedAt"`
	ImpactScore       int    `json:"impactScore,omitempty"`
	PolicyProbability int    `json:"policyProbability,omitempty"`
}

// NewsResponse struct remains the same
type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

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

// fetchNews function remains the same
func fetchNews(apiKey, query string, page int) ([]Article, int, error) {
	// Construct the URL for the NewsAPI request
	// Fetches 18 articles per page, sorted by publication date in English
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
		// NewsAPI often includes a 'message' field on error
		errMsg := newsResponse.Status
		if len(newsResponse.Articles) > 0 && newsResponse.Articles[0].Title != "" { // Heuristic for error message in API
			errMsg = newsResponse.Articles[0].Title
		}
		return nil, 0, fmt.Errorf("API error: %s. Full response: %s", errMsg, string(body))
	}

	// Calculate impact and policy scores for each article
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

// summarizeText function remains the same
// It creates a summary by taking the first few sentences of the text.
func summarizeText(text string) string {
	if strings.TrimSpace(text) == "" {
		return "No content available to summarize."
	}
	// Split by common sentence delimiters. Handle cases like "Mr. Jones" correctly.
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})

	if len(sentences) == 0 {
		// If no sentences found (e.g. text without punctuation), return a snippet
		maxLength := 150
		if len(text) <= maxLength {
			return text
		}
		return text[:maxLength] + "..."
	}

	var summary strings.Builder
	sentenceCount := 0
	desiredSentences := 2 // Number of sentences for the summary

	for i, s := range sentences {
		trimmedSentence := strings.TrimSpace(s)
		if trimmedSentence != "" {
			summary.WriteString(trimmedSentence)
			// Add back the delimiter, checking the original text if possible or assuming '.'
			// This part is tricky without knowing the exact delimiter. For simplicity, add '.'
			// A more robust solution would involve looking at the character after s in the original text.
			if i < len(text) && (text[strings.Index(text, s)+len(s)] == '.' || text[strings.Index(text, s)+len(s)] == '!' || text[strings.Index(text, s)+len(s)] == '?') {
				summary.WriteByte(text[strings.Index(text, s)+len(s)])
			} else {
				summary.WriteString(".") // Default delimiter
			}

			summary.WriteString(" ") // Add space between sentences
			sentenceCount++
			if sentenceCount >= desiredSentences {
				break
			}
		}
	}

	result := strings.TrimSpace(summary.String())
	if result == "" { // Fallback if all sentences were whitespace
		maxLength := 150
		if len(text) <= maxLength {
			return text
		}
		return text[:maxLength] + "..."
	}
	return result
}


// calculateImpactScore function remains the same
func calculateImpactScore(text string) int {
	keywords := []string{"crisis", "breakthrough", "disaster", "economy", "war", "pandemic", "reform", "urgent", "major", "global"}
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(textLower, k) {
			score += 10 // Adjusted score per keyword
		}
	}
	return min(100, score) // Ensure score doesn't exceed 100
}

// calculatePolicyProbability function remains the same
func calculatePolicyProbability(text string) int {
	keywords := []string{"policy", "regulation", "law", "government", "legislation", "bill", "congress", "senate", "parliament", "decree"}
	score := 0
	textLower := strings.ToLower(text)
	for _, k := range keywords {
		if strings.Contains(textLower, k) {
			score += 15 // Adjusted score per keyword
		}
	}
	return min(100, score) // Ensure score doesn't exceed 100
}

// min function remains the same
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// askAI function remains the same
// This function simulates an AI by searching for keywords in the loaded articles.
func askAI(question string, articles []Article) string {
	searchQuery := strings.ToLower(strings.TrimSpace(question))
	if searchQuery == "" {
		return "Please ask a specific question."
	}

	var relevantArticles []string
	for _, a := range articles {
		content := strings.ToLower(a.Title + " " + a.Description)
		// A simple check if any word from the question appears in the article content.
		// This can be improved with more sophisticated matching.
		found := false
		for _, qWord := range strings.Fields(searchQuery) {
			if len(qWord) > 2 && strings.Contains(content, qWord) { // Avoid very short words, match part of the word
				found = true
				break
			}
		}
		if found {
			summary := summarizeText(a.Description)
			if summary == "No content available to summarize." && a.Title != "" {
				summary = a.Title // Use title if description is empty
			}
			relevantArticles = append(relevantArticles, fmt.Sprintf("- %s: %s", a.Title, summary))
		}
	}

	if len(relevantArticles) == 0 {
		return "No relevant information found in the currently loaded articles for your question: '" + question + "'"
	}
	if len(relevantArticles) > 5 { // Limit number of results shown
		relevantArticles = relevantArticles[:5]
		relevantArticles = append(relevantArticles, "\n(And more...)")
	}

	return "Based on the articles, here's some information related to '" + question + "':\n\n" + strings.Join(relevantArticles, "\n\n")
}


// Main application function
func main() {
	myApp := app.NewWithID("com.example.newsaggregator") // Added AppID
	myWindow := myApp.NewWindow("News Aggregator Deluxe")
	myWindow.Resize(fyne.NewSize(800, 700)) // Slightly wider for cards

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
	// API Key Input
	keyInput := widget.NewPasswordEntry() // Use PasswordEntry for API keys
	keyInput.SetPlaceHolder("Enter NewsAPI key...")
	apiKey := loadSavedKey()
	if apiKey != "" {
		keyInput.SetText(apiKey)
	}

	// Theme Toggle Button
	themeBtn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), nil) // Icon changes with theme
	updateThemeBtnIcon := func(isDark bool) {
		if isDark {
			themeBtn.SetIcon(theme.LightThemeIcon()) // Icon to switch to Light
			themeBtn.SetText("Light Mode")
		} else {
			themeBtn.SetIcon(theme.DarkThemeIcon()) // Icon to switch to Dark
			themeBtn.SetText("Dark Mode")
		}
	}
	updateThemeBtnIcon(isDarkTheme) // Initial icon
	themeBtn.OnTapped = func() {
		isDarkTheme = !isDarkTheme
		if isDarkTheme {
			myApp.Settings().SetTheme(theme.DarkTheme())
		} else {
			myApp.Settings().SetTheme(theme.LightTheme())
		}
		updateThemeBtnIcon(isDarkTheme)
		saveThemePreference(isDarkTheme)
		myWindow.Content().Refresh() // Refresh entire window to apply theme changes properly
	}

	apiKeyLabel := widget.NewLabel("API Key:")
	apiKeyRow := container.NewBorder(nil, nil, apiKeyLabel, themeBtn, keyInput)


	// Search Input and Button
	queryInput := widget.NewEntry()
	queryInput.SetPlaceHolder("Search news topics (e.g., 'Go programming')")

	// Results container and scroll
	results := container.NewVBox() // This will hold article cards
	// Initial message in results area
	results.Add(widget.NewLabelWithStyle("Enter your API key and a search query to begin.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))

	scroll := container.NewVScroll(results)
	scroll.SetMinSize(fyne.NewSize(200,300)) // Ensure scroll has a min size

	// Loading Indicator
	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide() // Initially hidden

	// Load More Button
	loadMoreBtn := widget.NewButtonWithIcon("Load More Articles", theme.MoreVerticalIcon(), nil)
	loadMoreBtn.Hide() // Initially hidden
	loadMoreContainer := container.NewCenter(loadMoreBtn)


	// Function to refresh the UI with articles
	refreshResultsUI := func() {
		results.Objects = nil // Clear previous results

		if len(allArticles) == 0 {
			results.Add(widget.NewLabelWithStyle("No articles found for your query.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
			results.Refresh()
			return
		}

		for i := range allArticles {
			article := allArticles[i] // Create a new variable for the closure

			parsedURL, err := url.Parse(article.URL)
			if err != nil {
				fmt.Printf("Error parsing URL %s: %v\n", article.URL, err)
				parsedURL, _ = url.Parse("https://example.com/invalid-url") // Fallback URL
			}

			// Description for the card - truncate if too long
			cardDescription := article.Description
			if len(cardDescription) > 150 { // Max length for card snippet
				cardDescription = cardDescription[:150] + "..."
			}
			if strings.TrimSpace(cardDescription) == "" {
				cardDescription = "No description available."
			}

			descriptionLabel := widget.NewLabel(cardDescription)
			descriptionLabel.Wrapping = fyne.TextWrapWord


			// Detailed Summary/Info Button for each card
			detailsBtn := widget.NewButtonWithIcon("Details", theme.InfoIcon(), func() {
				fullSummary := summarizeText(article.Description) // Get 2-sentence summary

				content := container.NewVBox(
					widget.NewLabelWithStyle(article.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewSeparator(),
					widget.NewLabelWithStyle("Full Summary:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					widget.NewLabel(fullSummary),
					widget.NewSeparator(),
					widget.NewLabel(fmt.Sprintf("Impact Score: %d/100", article.ImpactScore)),
					widget.NewLabel(fmt.Sprintf("Policy Probability: %d%%", article.PolicyProbability)),
					widget.NewSeparator(),
					widget.NewHyperlink("Open Original Article", parsedURL),
				)

				var detailPopUp *widget.PopUp
				detailPopUp = widget.NewModalPopUp(
					container.NewBorder(
						nil, // Top
						container.NewHBox(layout.NewSpacer(), widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { detailPopUp.Hide() }), layout.NewSpacer()), // Bottom
						nil, // Left
						nil, // Right
						container.NewVScroll(content), // Center, scrollable
					),
					myWindow.Canvas(),
				)
				detailPopUp.Resize(fyne.NewSize(myWindow.Canvas().Size().Width*0.8, myWindow.Canvas().Size().Height*0.7))
				detailPopUp.Show()
			})

			// Card content
			cardContent := container.NewVBox(
				descriptionLabel,
				widget.NewSeparator(),
				container.NewGridWithColumns(2,
					widget.NewLabel(fmt.Sprintf("Impact: %d", article.ImpactScore)),
					widget.NewLabel(fmt.Sprintf("Policy Chance: %d%%", article.PolicyProbability)),
				),
				widget.NewSeparator(),
				container.NewHBox(
					widget.NewHyperlink("Read Full Article", parsedURL),
					layout.NewSpacer(),
					detailsBtn,
				),
			)

			card := widget.NewCard(
				article.Title,
				fmt.Sprintf("Published: %s", humanTime(article.PublishedAt)),
				cardContent,
			)
			results.Add(card)
			results.Add(widget.NewSeparator()) // Visual separation between cards
		}
		results.Refresh()
		scroll.ScrollToTop()
	}

	// Sort Button
	sortBtn := widget.NewButtonWithIcon("Sort: Newest First", theme.SortAscendingIcon(), nil) // Icon changes
	sortBtn.OnTapped = func() {
		sortAsc = !sortAsc
		if sortAsc {
			sortBtn.SetText("Sort: Oldest First")
			sortBtn.SetIcon(theme.SortDescendingIcon())
		} else {
			sortBtn.SetText("Sort: Newest First")
			sortBtn.SetIcon(theme.SortAscendingIcon())
		}
		sortByTime(allArticles, sortAsc)
		refreshResultsUI()
	}

	// Search Button
	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		key := keyInput.Text
		query := queryInput.Text

		if key == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabelWithStyle("⚠️ API key is required. Please enter it above.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})}
			results.Refresh()
			loadMoreBtn.Hide()
			loadingIndicator.Hide()
			return
		}
		if query == "" {
			results.Objects = []fyne.CanvasObject{widget.NewLabelWithStyle("⚠️ Please enter a search query.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})}
			results.Refresh()
			loadMoreBtn.Hide()
			loadingIndicator.Hide()
			return
		}

		// Show loading indicator, clear results
		results.Objects = nil // Clear previous content
		results.Add(loadingIndicator)
		loadingIndicator.Show()
		results.Refresh()

		loadMoreBtn.Hide()
		currentPage = 1
		allArticles = nil
		lastQuery = query

		// Fetch news in a goroutine to keep UI responsive
		go func(apiKey, searchQuery string) {
			articles, total, err := fetchNews(apiKey, searchQuery, currentPage)

			// Update UI on the main thread
			myWindow.RunTransaction(func() {
				loadingIndicator.Hide()
				results.Objects = nil // Clear loading indicator

				if err != nil {
					results.Add(widget.NewLabelWithStyle(fmt.Sprintf("❌ Error fetching news: %v", err), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
					results.Refresh()
					loadMoreBtn.Hide()
					return
				}
				if len(articles) == 0 {
					results.Add(widget.NewLabelWithStyle("🔍 No results found for your query. Try different keywords.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}))
					results.Refresh()
					loadMoreBtn.Hide()
					return
				}

				totalResults = total
				allArticles = articles
				sortByTime(allArticles, sortAsc) // Apply current sort order
				refreshResultsUI()

				if len(allArticles) < totalResults && len(allArticles) > 0 {
					loadMoreBtn.Show()
				} else {
					loadMoreBtn.Hide()
				}
				saveAPIKey(key) // Save API key on successful search
			})
		}(key, query)
	})

	queryInput.OnSubmitted = func(s string) { // Allow search on Enter key
		searchBtn.OnTapped()
	}

	searchRow := container.NewBorder(nil, nil, queryInput, container.NewHBox(searchBtn, sortBtn))


	// "Ask AI" Input and Button
	askAIInput := widget.NewEntry()
	askAIInput.SetPlaceHolder("Ask a question about loaded articles...")

	// This dialog shows the response from the 'askAI' function.
	showAIResponseDialog := func(title, content string) {
		var popUp *widget.PopUp
		popUp = widget.NewModalPopUp(
			container.NewVBox(
				widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				container.NewVScroll(widget.NewLabel(content)), // Scrollable content
				widget.NewSeparator(),
				widget.NewButtonWithIcon("Close", theme.CancelIcon(), func() { popUp.Hide() }),
			),
			myWindow.Canvas(),
		)
		popUp.Resize(fyne.NewSize(myWindow.Canvas().Size().Width*0.7, myWindow.Canvas().Size().Height*0.6))
		popUp.Show()
	}

	askBtn := widget.NewButtonWithIcon("Ask AI", theme.QuestionIcon(), func() {
		question := askAIInput.Text
		if question == "" {
			showAIResponseDialog("Ask AI Error", "Please enter a question.")
			return
		}
		if len(allArticles) == 0 {
			showAIResponseDialog("Ask AI Info", "Please search for and load some articles before asking a question.")
			return
		}
		answer := askAI(question, allArticles)
		showAIResponseDialog("AI Response", answer)
	})
	askAIInput.OnSubmitted = func(s string) { // Allow ask on Enter key
		askBtn.OnTapped()
	}

	askAIRow := container.NewBorder(nil, nil, askAIInput, askBtn)


	// Export and Clipboard Buttons
	exportBtn := widget.NewButtonWithIcon("Export MD", theme.FileTextIcon(), func() {
		if len(allArticles) == 0 {
			myApp.SendNotification(&fyne.Notification{Title: "Export Info", Content: "No articles to export."})
			return
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# News Articles for Query: %s\n\n", lastQuery))
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n", a.Title))
			sb.WriteString(fmt.Sprintf("- **URL**: %s\n", a.URL))
			sb.WriteString(fmt.Sprintf("- **Published**: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", strings.TrimSpace(a.Description)))
			sb.WriteString(fmt.Sprintf("- **Impact Score**: %d/100\n", a.ImpactScore))
			sb.WriteString(fmt.Sprintf("- **Policy Probability**: %d%%\n", a.PolicyProbability))
			sb.WriteString(fmt.Sprintf("- **Summary**: %s\n", summarizeText(a.Description)))
			sb.WriteString("\n---\n\n")
		}
		home, _ := os.UserHomeDir()
		// Use a more descriptive filename including the query and date
		dateStr := time.Now().Format("2006-01-02")
		safeQuery := strings.ReplaceAll(strings.ToLower(lastQuery), " ", "_")
		if len(safeQuery) > 30 { safeQuery = safeQuery[:30] } // Limit length
		fileName := fmt.Sprintf("news_export_%s_%s.md", safeQuery, dateStr)
		path := filepath.Join(home, "Documents", fileName) // Save to Documents folder by default

		// Ensure Documents directory exists
		docDir := filepath.Join(home, "Documents")
		if _, err := os.Stat(docDir); os.IsNotExist(err) {
			os.MkdirAll(docDir, 0755)
		}

		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
			showAIResponseDialog("Export Error", fmt.Sprintf("Failed to write file: %v", err))
			return
		}
		myApp.SendNotification(&fyne.Notification{Title: "Export Success", Content: fmt.Sprintf("Articles exported to %s", path)})
	})

	clipboardBtn := widget.NewButtonWithIcon("Copy All", theme.ContentCopyIcon(), func() {
		if len(allArticles) == 0 {
			myApp.SendNotification(&fyne.Notification{Title: "Clipboard Info", Content: "No articles to copy."})
			return
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# News Articles for Query: %s\n\n", lastQuery))
		for _, a := range allArticles {
			sb.WriteString(fmt.Sprintf("## %s\n", a.Title))
			sb.WriteString(fmt.Sprintf("- URL: %s\n", a.URL)) // Simpler format for clipboard
			sb.WriteString(fmt.Sprintf("- Published: %s\n", humanTime(a.PublishedAt)))
			sb.WriteString(fmt.Sprintf("- Description: %s\n", summarizeText(a.Description))) // Summarized
			sb.WriteString("\n")
		}
		myWindow.Clipboard().SetContent(sb.String())
		myApp.SendNotification(&fyne.Notification{Title: "Clipboard Success", Content: "Article summaries copied to clipboard."})
	})

	utilityRow := container.NewHBox(layout.NewSpacer(), exportBtn, clipboardBtn, layout.NewSpacer()) // Centered

	// Load More Button OnTapped Action
	loadMoreBtn.OnTapped = func() {
		currentPage++
		key := keyInput.Text
		query := lastQuery // Use the last successful query for loading more

		loadingIndicator.Show() // Show loading indicator near button or in results
		loadMoreBtn.SetText("Loading...")
		loadMoreBtn.Disable()

		go func(apiKey, searchQuery string, pageNum int) {
			articles, _, err := fetchNews(apiKey, searchQuery, pageNum)

			myWindow.RunTransaction(func(){
				loadingIndicator.Hide()
				loadMoreBtn.SetText("Load More Articles")
				loadMoreBtn.Enable()

				if err != nil {
					myApp.SendNotification(&fyne.Notification{Title: "Load More Error", Content: err.Error()})
					// Potentially decrement currentPage if fetch failed, or offer retry
					return
				}
				if len(articles) > 0 {
					allArticles = append(allArticles, articles...)
					sortByTime(allArticles, sortAsc) // Re-sort with new articles
					refreshResultsUI()
				}

				if len(allArticles) >= totalResults || len(articles) == 0 {
					loadMoreBtn.Hide() // Hide if no more articles or page was empty
				} else {
					loadMoreBtn.Show()
				}
			})
		}(key, query, currentPage)
	}

	// --- Layout ---
	topControls := container.NewVBox(
		apiKeyRow,
		searchRow,
		askAIRow,
		utilityRow,
		widget.NewSeparator(),
	)

	content := container.NewBorder(
		topControls,       // Top
		loadMoreContainer, // Bottom
		nil,               // Left
		nil,               // Right
		scroll,            // Center content (results list)
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
