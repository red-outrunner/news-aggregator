# News Aggregator Web

A modern web-based news aggregator with sentiment analysis, impact scoring, and policy probability tracking.

## Features

- **Search & Fetch**: Search NewsAPI for articles on any topic
- **Sentiment Analysis**: Automatically scores articles from -100 (negative) to +100 (positive)
- **Impact Scoring**: Measures article importance based on impactful keywords
- **Policy Probability**: Calculates policy relevance percentage
- **Bookmarks**: Save articles for later (persisted in localStorage)
- **Dark/Light Mode**: Toggle between themes (preference saved)
- **Sorting**: Sort by latest, oldest, sentiment, or impact
- **Responsive Design**: Works on desktop and mobile

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand (with localStorage persistence)
- **API**: Next.js API Routes (server-side NewsAPI proxy)

## Getting Started

### Prerequisites

- Node.js 18+ installed
- npm or yarn
- A free NewsAPI key from [newsapi.org](https://newsapi.org/register)

### Installation

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
npm install
```

3. Create environment file and add your API key:
```bash
cp .env.local.example .env.local
```

Edit `.env.local` and add your NewsAPI key:
```
NEWS_API_KEY=your_actual_api_key_here
```

4. Run the development server:
```bash
npm run dev
```

5. Open [http://localhost:3000](http://localhost:3000) in your browser

## Project Structure

```
web/
├── src/
│   ├── app/
│   │   ├── api/
│   │   │   └── news/
│   │   │       └── route.ts      # NewsAPI proxy endpoint
│   │   ├── page.tsx              # Main page component
│   │   ├── layout.tsx            # Root layout
│   │   └── globals.css           # Global styles
│   ├── components/
│   │   ├── SearchBar.tsx         # Search input component
│   │   ├── ArticleCard.tsx       # Individual article display
│   │   ├── ArticleList.tsx       # List of articles
│   │   ├── SortFilter.tsx        # Sorting controls
│   │   ├── BookmarksModal.tsx    # Bookmarks popup
│   │   ├── ThemeToggle.tsx       # Dark/light mode toggle
│   │   └── index.ts              # Component exports
│   ├── lib/
│   │   ├── sentiment.ts          # Sentiment analysis algorithm
│   │   ├── impact.ts             # Impact scoring algorithm
│   │   ├── policy.ts             # Policy probability algorithm
│   │   ├── analyzer.ts           # Combined analyzer
│   │   ├── types.ts              # TypeScript types
│   │   └── utils.ts              # Utility functions
│   └── store/
│       └── newsStore.ts          # Zustand state management
├── .env.local                    # Environment variables (gitignored)
├── .env.local.example            # Environment template
└── package.json
```

## Algorithms

All algorithms are migrated from the original Go implementation:

### Sentiment Analysis
- Checks positive/negative phrases (+/-15 points each)
- Checks positive/negative keywords (+/-10 points each)
- Score range: -100 to +100

### Impact Score
- Counts impactful words like "crisis", "breakthrough", "major" (+5 points each)
- Score range: 0 to 100

### Policy Probability
- Counts policy-related keywords like "regulation", "congress", "legislation" (+10 points each)
- Score range: 0 to 100 (percentage)

## Deployment

### Vercel (Recommended)

1. Push code to GitHub
2. Import project in [Vercel](https://vercel.com)
3. Add `NEWS_API_KEY` environment variable in Vercel dashboard
4. Deploy

### Docker

```bash
docker build -t news-aggregator .
docker run -p 3000:3000 -e NEWS_API_KEY=your_key news-aggregator
```

## Comparison with Go Version

| Feature | Go CLI | Go Fyne GUI | Web (Next.js) |
|---------|--------|-------------|---------------|
| Platform | Terminal | Desktop | Web/Mobile |
| Installation | Go binary | Go binary | Node.js/npm |
| API Key Security | Local config | Local config | Server-side |
| Bookmarks | JSON file | JSON file | localStorage |
| Theme | Light/Dark | Light/Dark | Light/Dark |
| Sentiment Analysis | ✅ | ✅ | ✅ |
| Impact Score | ✅ | ✅ | ✅ |
| Policy Score | ❌ | ✅ | ✅ |
| AI Summary (Ollama) | ✅ | ✅ | ❌ (optional) |
| Deployment | Local only | Local only | Anywhere |

## License

Mozilla Public License 2.0 (same as original)

## Credits

Migrated from the original Go news aggregator by Weo Sikho Fuzile
