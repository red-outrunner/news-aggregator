# red-news-aggregator
News collecting code from various sites
craeted by Weo Sikho Fuzile

## What It Does

This joint hits the NewsAPI to fetch the latest articles based on whatever you're curious about. Type in anything - "Tesla", "crypto", "climate change", whatever you're trying to stay up on - and it'll serve up the 18 most recent stories, sorted fresh to old.

## Structure
Main folder is cli based version - way behind in updates
> `test-ui/` for ui version
> - the `ui-aggregator.go` is for main tested version I recommend that build for stable type
> - the `b-ui-aggregator.go` is for the new feature I am tryin like the workaround ollama

## Features
* **Search & Filter**
    * Search news with any keyword or phrase.
    * Filter by a date range (`YYYY-MM-DD`).

* **Analysis Tools**
    * **Sentiment Score:** Automatically see if news coverage is positive, negative, or neutral.
    * **Keyword Highlighting:** Your search terms are highlighted directly in the titles and descriptions.
    * **Trend Analysis:** Click the "Trend" button to see a simple chart of how often your topic appeared each day.
    * **Impact & Policy Scores:** Get at-a-glance scores for an article's potential market impact or policy relevance.

* **Manage & Export**
    * **Bookmarks:** Save important articles to a separate list to view later. Your bookmarks are saved between sessions.
    * **Read/Unread Status:** Keep track of articles you've already reviewed during a session.
    * **Export to Document:** Save your current article list to a formatted `.rtf` document compatible with MS Word and other processors.
    * **Copy to Clipboard:** Copy a summary of the articles for easy pasting.

* **Customization**
    * **Light & Dark Mode:** Choose your preferred theme. Your choice is saved automatically.
    * **Sorting Options:** Sort articles by time (newest/oldest) or by sentiment score (highest/lowest).



## 1. Before You Even Start

Look, this whole thing runs on NewsAPI, so you need a key - non-negotiable. Head to [NewsAPI](https://newsapi.org/register) and get yours (it's free, don't trip). Without this, you might as well try to run a car with no gas.
Have Ollama installed for ai summary - the only memory demanding task

## 2. Setup

1. First things first, cop Go from [golang.org](https://golang.org/dl/) if you ain't got it
2. Clone this repo (you know the vibes):
```bash
git clone https://github.com/your-username/news-aggregator.git
cd news-aggregator
```
## Running It

Just hit it with:
```bash
go run main.go
```
Insert your API key

Then type whatever you're trying to read about when it asks. Simple as that.

## Pro Tips

- No API key = no news, simple math. Don't even try running this without setting up your key first
- The free API tier got limits like a club bouncer, so don't spam it
- Keep your searches specific - "JSE acquisition" will get you better results than just "JSE"
- Want more than 18 articles? Just tweak that limit in the code
- API key acting up? Double-check you pasted it right

## Contributing

See something you could make better? Pull requests are welcome. No bureaucracy - just make it work better.

Built with Go, powered by NewsAPI, and zero bloat. Questions? Drop an issue in the repo.

## GUI mode changes coming soon bro's

to make it mire fresh and seamless like Thanos glove:
1. entering API key on Gui Mode also on console mode
2. add up sort function on the console - Done
3. get access to more news site where reading articles cost $free.99c - on pipeline
4. make the thing a bit pretty - semi-done
5. have bookmarks -DOne
6. measure memory when running - current
7. api to web page - dread
8. build and package it
9. Ai summary + prompt to figure out narative or sumn like that
10. A work around using ollama only for ai summary - in discovery & research and testing

## License
 
This project is licensed under the **Mozilla Public License 2.0**. See the full license text [here](https://www.mozilla.org/en-US/MPL/2.0/).
 
