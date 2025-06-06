# news-aggregator
News collecting code from various sites
craeted by Weo Sikho Fuzile

# News Aggregator 🗞️

Keep it a stack - staying informed is crucial but who's got time to browse 50 different news sites? This Go-powered news aggregator pulls everything you need to know about any topic, person, or company in one clean feed.

## What It Does

This joint hits the NewsAPI to fetch the latest articles based on whatever you're curious about. Type in anything - "Tesla", "crypto", "climate change", whatever you're trying to stay up on - and it'll serve up the 18 most recent stories, sorted fresh to old.

## Features

- Real-time news fetching from NewsAPI's global sources
- Clean article display with titles, descriptions, and direct links
- Auto-sorting by publish date so you get the latest first
- English language focus (but you can mod that if you want)
- Caps at 18 articles to keep it digestible

## Before You Even Start

Look, this whole thing runs on NewsAPI, so you need a key - non-negotiable. Head to [NewsAPI](https://newsapi.org/register) and get yours (it's free, don't trip). Without this, you might as well try to run a car with no gas.

## Setup

1. First things first, cop Go from [golang.org](https://golang.org/dl/) if you ain't got it
2. Clone this repo (you know the vibes):
```bash
git clone https://github.com/your-username/news-aggregator.git
cd news-aggregator
```
3. Remember that API key we talked about? Drop it in the code where it says `YOUR_NEWS_API_KEY`. Don't skip this - the whole thing's gonna flame out if you try running it with the default placeholder

## Running It

Just hit it with:
```bash
go run main.go
```

Then type whatever you're trying to read about when it asks. Simple as that.

## Structure

- `Article` struct: Holds the news piece details
- `NewsResponse` struct: Manages the API response

## Pro Tips

- No API key = no news, simple math. Don't even try running this without setting up your key first
- The free API tier got limits like a club bouncer, so don't spam it
- Keep your searches specific - "SpaceX latest launch" will get you better results than just "space"
- Want more than 18 articles? Just tweak that limit in the code
- API key acting up? Double-check you pasted it right

## Contributing

See something you could make better? Pull requests are welcome. No bureaucracy - just make it work better.



Built with Go, powered by NewsAPI, and zero bloat. Questions? Drop an issue in the repo.

## GUI mode changes coming soon bro's

to make it mire fresh and seamless like Thanos glove:
1. entering API key on Gui Mode also on console mode
2. add up sort function on the console
3. get access to more news site where reading articles cost $free.99c
4. make the thing a bit pretty
5. have bookmarks
6. measure memory when running
7. api to web page - dread
8. build and package it
9. Ai summary + prompt to figure out narative or sumn like that
