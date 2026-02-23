// Stock and Index extraction utility
// Extracts ticker symbols and index names from article text
// Supports: US (NYSE/NASDAQ), JSE (South Africa), ASX (Australia), LSE (UK), HKEX, and more

// Major Global Stock Indices
const indices = [
  // US Indices
  'S&P 500', 'SPX', 'Dow Jones', 'DJIA', 'NASDAQ', 'NDX', 'Russell 2000', 'RUT',
  'S&P/TSX', 'TSX Composite',
  // European Indices
  'FTSE 100', 'FTSE 250', 'DAX', 'CAC 40', 'AEX', 'IBEX 35', 'FTSE MIB', 'SMI',
  'Euro Stoxx 50', 'STOXX 600',
  // Asian Indices
  'Nikkei 225', 'N225', 'Hang Seng', 'HSI', 'Shanghai Composite', 'SSE', 'Shenzhen',
  'KOSPI', 'KOSDAQ', 'Taiwan Weighted', 'TWSE', 'Straits Times', 'STI',
  'BSE Sensex', 'Nifty 50', 'Nifty Bank',
  // Pacific Indices
  'ASX 200', 'XJO', 'All Ordinaries', 'XAO', 'NZX 50',
  // African Indices
  'JSE All Share', 'JALSH', 'FTSE JSE', 'Top 40', 'FTSE/JSE Top 40',
  'JSE Mid Cap', 'JSE Small Cap', 'JSE Resource 20', 'JSE Financial 15',
  'JSE Industrial 25', 'FTSE/JSE Mid Cap', 'EGX 30', 'MASI',
  // Other Emerging Markets
  'Bovespa', 'IBOV', 'IPC Mexico', 'MERVAL', 'IPSA', 'COLCAP',
  'MOEX', 'RTS Index', 'WIG20', 'BUX', 'BET Index',
  // Volatility & Special
  'VIX', 'Volatility Index', 'VVIX',
  // Bond & Commodity Indices
  'DXY', 'Dollar Index', 'US Dollar', 'Gold Index', 'Oil Index',
];

// Common stock ticker patterns
const tickerPatterns = [
  /\b[A-Z]{1,5}\b/g, // Standard tickers (1-5 uppercase letters)
];

// Known tickers to look for (major companies across global markets)
const knownTickers = new Set([
  // === US MARKET (NYSE/NASDAQ) ===
  // Tech
  'AAPL', 'MSFT', 'GOOGL', 'GOOG', 'AMZN', 'META', 'TSLA', 'NVDA', 'AMD', 'INTC',
  'NFLX', 'ORCL', 'CRM', 'ADBE', 'CSCO', 'AVGO', 'QCOM', 'TXN', 'IBM', 'NOW',
  'INTU', 'AMAT', 'MU', 'LRCX', 'KLAC', 'SNPS', 'CDNS', 'MCHP', 'ADI', 'NXPI',
  // Finance
  'JPM', 'BAC', 'WFC', 'GS', 'MS', 'C', 'BLK', 'SCHW', 'AXP', 'V', 'MA', 'PYPL',
  'COF', 'USB', 'PNC', 'TFC', 'MTB', 'FITB', 'HBAN', 'RF', 'CFG', 'KEY',
  // Healthcare
  'JNJ', 'UNH', 'PFE', 'MRK', 'ABBV', 'TMO', 'ABT', 'DHR', 'BMY', 'LLY', 'GILD',
  'AMGN', 'ISRG', 'VRTX', 'REGN', 'ZTS', 'MRNA', 'BIIB', 'ILMN', 'ALXN', 'BMRN',
  // Consumer
  'WMT', 'PG', 'KO', 'PEP', 'COST', 'HD', 'MCD', 'NKE', 'SBUX', 'TGT', 'LOW',
  'TJX', 'ROST', 'DG', 'DLTR', 'BBY', 'AMZN', 'EBAY', 'ETSY', 'W', 'CHWY',
  // Energy
  'XOM', 'CVX', 'COP', 'SLB', 'EOG', 'MPC', 'PSX', 'VLO', 'OXY', 'HAL',
  'BKR', 'DVN', 'FANG', 'APA', 'HES', 'KMI', 'WMB', 'OKE', 'TRGP', 'LNG',
  // Industrial
  'CAT', 'BA', 'HON', 'UPS', 'GE', 'MMM', 'LMT', 'RTX', 'DE', 'UNP',
  'CSX', 'NSC', 'FDX', 'DAL', 'UAL', 'AAL', 'LUV', 'JBLU', 'ALK', 'SAVE',
  // Telecom & Media
  'T', 'VZ', 'TMUS', 'CHTR', 'CMCSA', 'DIS', 'NFLX', 'PARA', 'WBD', 'FOXA', 'FOX',
  'NWSA', 'NWS', 'NYT', 'GOOG', 'META', 'SNAP', 'PINS', 'TWTR', 'RDDT',
  // Real Estate
  'AMT', 'PLD', 'CCI', 'EQIX', 'SPG', 'PSA', 'WELL', 'DLR', 'O', 'VICI', 'AVB',
  'EQR', 'VTR', 'ESS', 'MAA', 'UDR', 'CPT', 'HST', 'RHP', 'REG', 'FRT',
  // Materials
  'LIN', 'APD', 'SHW', 'ECL', 'FCX', 'NEM', 'DOW', 'DD', 'PPG', 'EMN', 'CE',
  'ALB', 'FMC', 'LYB', 'CF', 'MOS', 'IFF', 'RPM', 'AXTA', 'NEU', 'OLN',
  // ETFs
  'SPY', 'QQQ', 'DIA', 'IWM', 'VTI', 'VOO', 'VEA', 'VWO', 'AGG', 'BND',
  'GLD', 'SLV', 'USO', 'UNG', 'TLT', 'HYG', 'LQD', 'XLF', 'XLE', 'XLK', 'XLV',
  'XLI', 'XLP', 'XLY', 'XLU', 'XLRE', 'XLB', 'XLV', 'VGT', 'VHT', 'VFH',

  // === JSE (JOHANNESBURG STOCK EXCHANGE) ===
  // Mining Resources
  'AGL', 'ANG', 'ARI', 'BHP', 'IMP', 'KIO', 'EXX', 'SOL', 'HAR', 'GFI',
  'GOLD', 'DRM', 'WDL', 'TAU', 'MDO', 'CML', 'RRL', 'ZIM', 'BSR', 'MUG',
  // Financials
  'FSR', 'SBK', 'CPI', 'NPN', 'PRX', 'ABG', 'BVT', 'INL', 'INP', 'RCL',
  'REI', 'TBS', 'UGP', 'VKE', 'GFI', 'ANG', 'SHP', 'MRP', 'WHL', 'TFG',
  // Industrials
  'NPN', 'PRX', 'CFR', 'REM', 'RMI', 'BVT', 'TBS', 'UGP', 'VKE', 'ZMP',
  'CLS', 'EVT', 'MNP', 'NPN', 'OMU', 'PIK', 'PPC', 'SLM', 'SSW', 'TCG',
  // Technology
  'NPN', 'PRX', 'TRU', 'EVT', 'MNP', 'APN', 'DBI', 'VKE', 'ZMP', 'CLS',
  // Retail Consumer
  'SHP', 'MRP', 'WHL', 'TFG', 'CPI', 'TBS', 'UGP', 'VKE', 'ZMP', 'CLS',
  'NPN', 'PRX', 'REM', 'RMI', 'BVT', 'INL', 'INP', 'RCL', 'REI', 'SLM',
  // Telecommunications
  'VOD', 'MTN', 'TKG', 'BCOM', 'DTC', 'EVT', 'MNP', 'APN', 'DBI', 'VKE',
  // Property REITs
  'GRT', 'RDF', 'HYB', 'NEP', 'VUK', 'WTC', 'ACT', 'CPT', 'DIP', 'EMS',
  'FAIR', 'GCP', 'GRO', 'HYP', 'IRV', 'LMP', 'OCE', 'REE', 'RES', 'URB',

  // === ASX (AUSTRALIAN SECURITIES EXCHANGE) ===
  // Banks Financials
  'CBA', 'WBC', 'NAB', 'ANZ', 'MQG', 'QBE', 'WES', 'FMG', 'STO', 'WDS',
  'COL', 'WOW', 'TLS', 'TCL', 'ALL', 'IAG', 'SUN', 'AMP', 'BEN', 'BOQ',
  // Mining Resources
  'BHP', 'RIO', 'FMG', 'NCM', 'EVN', 'RRL', 'SFR', 'MIN', 'WDS', 'STO',
  'ORG', 'WPL', 'IGO', 'S32', 'AWC', 'LYC', 'ILU', 'PDN', 'CHN', 'EMN',
  // Healthcare
  'CSL', 'RMD', 'COH', 'SHL', 'ANN', 'SOL', 'PRY', 'AVH', 'HLS', 'CLV',
  // Technology
  'XRO', 'WTC', 'NEC', 'CPU', 'APT', 'ZIP', 'NWL', 'TNE', '360', 'ELD',
  // Consumer Retail
  'WES', 'COL', 'WOW', 'JBH', 'HVN', 'LOV', 'AVJ', 'KGN', 'MTS', 'NMT',
  // Industrial Conglomerates
  'WES', 'TLS', 'TCL', 'QAN', 'SYD', 'APA', 'ALL', 'ALX', 'AMC', 'AZJ',

  // === LSE (LONDON STOCK EXCHANGE) ===
  // FTSE 100 Major
  'HSBA', 'BP', 'SHEL', 'AZN', 'GSK', 'DGE', 'ULVR', 'RIO', 'AAL', 'BHP',
  'LSEG', 'LLOY', 'BARC', 'VOD', 'BT.A', 'NG', 'GLEN', 'PRU', 'AHT', 'BA',
  'CRH', 'FLTR', 'REL', 'EXPN', 'WPP', 'IAG', 'RR', 'CCH', 'TSCO', 'MRO',
  // FTSE 250
  'RTO', 'SMDS', 'BRBY', 'JD', 'OCDO', 'PSON', 'WTB', 'MNDI', 'HWDN', 'BDEV',
  'CRST', 'RMV', 'POLY', 'TW', 'FCIT', 'SCIN', 'BGS', 'JMAT', 'IMI', 'SGE',

  // === HKEX (HONG KONG EXCHANGE) ===
  '0700', '9988', '0005', '0941', '0388', '1299', '0883', '0939', '1398', '2318',
  '0857', '0001', '0002', '0003', '0011', '0012', '0016', '0017', '0019', '0027',
  '0066', '0083', '0101', '0144', '0151', '0175', '0241', '0267', '0268', '0285',
  '0288', '0291', '0316', '0322', '0386', '0388', '0390', '0392', '0489', '0522',
  '0669', '0688', '0700', '0728', '0762', '0823', '0836', '0857', '0881', '0883',
  '0902', '0914', '0916', '0939', '0941', '0960', '0968', '0981', '0992', '0998',
  '1024', '1038', '1044', '1088', '1093', '1099', '1109', '1113', '1177', '1209',
  '1211', '1288', '1299', '1336', '1337', '1347', '1359', '1378', '1398', '1548',
  '1810', '1876', '1919', '1928', '1997', '2007', '2013', '2015', '2018', '2020',
  '2269', '2313', '2318', '2319', '2331', '2382', '2388', '2600', '2601', '2628',
  '2899', '3328', '3690', '3968', '3988', '6098', '6618', '6862', '9618', '9633', '9888', '9988', '9999',

  // === TSX (TORONTO STOCK EXCHANGE) ===
  'RY', 'TD', 'BNS', 'BMO', 'CM', 'CNR', 'CP', 'ENB', 'TRP', 'PPL',
  'FTS', 'AQN', 'SU', 'CNQ', 'IMO', 'CVE', 'ARX', 'WCP', 'MEG', 'TOU',
  'ABX', 'NEM', 'K', 'AEM', 'WPM', 'FNV', 'OR', 'CG', 'B2GOLD', 'ELD',
  'SHOP', 'CSU', 'L', 'ATD', 'MRU', 'WCN', 'QSR', 'TIH', 'DOL', 'LSPD',

  // === EURONEXT (EURONEXT PARIS/AMSTERDAM/BRUSSELS) ===
  'MC', 'OR', 'SAN', 'AIR', 'TTE', 'EL', 'DG', 'KER', 'SU', 'BNP',
  'ACA', 'GLE', 'CS', 'DSY', 'CAP', 'URW', 'STM', 'VIE', 'PUB', 'RI',
  'ASML', 'PHIA', 'UNA', 'HEIA', 'AD', 'PRX', 'WKL', 'RAND', 'APAM', 'AKZA',
  'UCB', 'KBC', 'ARGX', 'COFB', 'GBLB', 'SOF', 'UCB', 'PROX', 'ELI', 'MBLY',

  // === GERMAN XETRA ===
  'SAP', 'SIE', 'ALV', 'BAS', 'MBG', 'VOW3', 'BMW', 'DTE', 'MUV2', 'ADS',
  'DBK', 'DB1', 'IFX', 'LIN', 'HEN3', 'BEI', 'CON', 'FRE', 'HEI', 'RHM',
  '1COV', 'DPW', 'FME', 'HFG', 'LEG', 'MTX', 'PUM', 'QIA', 'RWE', 'SHL',
  'SY1', 'TEG', 'VNA', 'ZAL', 'AIR', 'BC8', 'BOSS', 'COK', 'DEQ', 'DHL',

  // === JAPAN TSE ===
  '7203', '6758', '9984', '9432', '6861', '6954', '7974', '4063', '4502', '4503',
  '8035', '8058', '8306', '8316', '8411', '9020', '9022', '9433', '9434', '9435',
  '9983', '9984', '6098', '4689', '2432', '3659', '4755', '6501', '6503', '6594',
  '6701', '6702', '6723', '6724', '6752', '6753', '6758', '6762', '6857', '6861',
  '6902', '6954', '6971', '6981', '7201', '7202', '7203', '7267', '7269', '7270',
  '7733', '7741', '7751', '7832', '7974', '8001', '8002', '8031', '8035', '8053',
  '8058', '8088', '8136', '8233', '8252', '8267', '8306', '8316', '8411', '8604',
  '8766', '8801', '8802', '8830', '9007', '9008', '9020', '9022', '9062', '9064',
  '9101', '9104', '9143', '9301', '9432', '9433', '9434', '9435', '9501', '9502',
  '9503', '9531', '9532', '9613', '9697', '9735', '9766', '9983', '9984', '9985',
]);

// Company name to ticker mapping (for text mentions across global markets)
const companyToTicker: Record<string, string> = {
  // === US COMPANIES ===
  'Apple': 'AAPL', 'Microsoft': 'MSFT', 'Google': 'GOOGL', 'Amazon': 'AMZN',
  'Meta': 'META', 'Facebook': 'META', 'Tesla': 'TSLA', 'Nvidia': 'NVDA',
  'JPMorgan': 'JPM', 'Bank of America': 'BAC', 'Wells Fargo': 'WFC',
  'Goldman Sachs': 'GS', 'Morgan Stanley': 'MS', 'BlackRock': 'BLK',
  'Johnson & Johnson': 'JNJ', 'Pfizer': 'PFE', 'Merck': 'MRK',
  'Walmart': 'WMT', 'Procter & Gamble': 'PG', 'Coca-Cola': 'KO',
  'ExxonMobil': 'XOM', 'Chevron': 'CVX', 'ConocoPhillips': 'COP',
  'Boeing': 'BA', 'Caterpillar': 'CAT', 'General Electric': 'GE',
  'AT&T': 'T', 'Verizon': 'VZ', 'Comcast': 'CMCSA', 'Disney': 'DIS',
  
  // === JSE (SOUTH AFRICA) COMPANIES ===
  'Gold Fields': 'GFI', 'Anglo American PLC': 'AGL', 'Naspers': 'NPN',
  'Prosus': 'PRX', 'Richemont': 'CFR', 'FirstRand': 'FSR',
  'Standard Bank': 'SBK', 'Capitec': 'CPI', 'Vodacom': 'VOD',
  'MTN': 'MTN', 'Telkom': 'TKG', 'Eskom': 'ESK',
  'Sasol': 'SOL', 'AngloGold Ashanti': 'ANG', 'Harmony Gold': 'HAR',
  'Impala Platinum': 'IMP', 'Kumba Iron Ore': 'KIO', 'Exxaro': 'EXX',
  'Shoprite': 'SHP', 'Mr Price': 'MRP', 'Woolworths': 'WHL',
  'The Foschini Group': 'TFG', 'TFG Limited': 'TFG', 'Capitec Bank': 'CPI',
  'Nedbank': 'NED', 'Investec': 'INL', 'Discovery': 'DSY',
  'Discovery Holdings': 'DSY', 'Bid Corporation': 'BID', 'Bidvest': 'BVT',
  'Barloworld': 'BAW', 'Brait': 'BRT', 'Clicks Group': 'CLS',
  'Dis-Chem': 'DCP', 'Mediclinic': 'MEH', 'Netcare': 'NTC',
  'Life Healthcare': 'LHC', 'Aspen': 'APN', 'Adcock Ingram': 'ADI',
  'Massmart': 'MSM', 'Pick n Pay': 'PIK', 'Boxer': 'BOX',
  'Pepkor': 'PPK', 'Ackermans': 'ACK', 'JD Group': 'JDG',
  'Lewis': 'LEW', 'Unitrans': 'UTR', 'Superbalist': 'SPL',
  'Takealot': 'TKT', 'Jumia': 'JMIA', 'MultiChoice': 'MTC',
  'Showmax': 'SHW', 'eMedia': 'MEA', 'e.tv': 'ETV',
  'SABC': 'SAB', 'Primedia': 'PMD', 'Arena Holdings': 'ARE',
  'Vukile Property': 'VKE', 'Growthpoint': 'GRT', 'Redefine': 'RDF',
  'Hyprop': 'HYB', 'Newpark': 'NEP', 'Witwatersrand': 'WTC',
  'Atterbury': 'ACT', 'Capital Appreciation': 'CAP', 'Divercity': 'DIP',
  'Emira': 'EMS', 'Fairvest': 'FAIR', 'Growthpoint Properties': 'GRT',
  'Hyprop Investments': 'HYB', 'IRV': 'IRV', 'LMP': 'LMP',
  
  // === ASX (AUSTRALIA) COMPANIES ===
  'Commonwealth Bank': 'CBA', 'Westpac': 'WBC', 'National Australia Bank': 'NAB',
  'ANZ': 'ANZ', 'Macquarie': 'MQG', 'QBE Insurance': 'QBE',
  'Wesfarmers': 'WES', 'Fortescue': 'FMG', 'Santos': 'STO',
  'Woodside': 'WDS', 'Coles': 'COL', 'Woolworths AU': 'WOW',
  'Telstra': 'TLS', 'Transurban': 'TCL', 'Allianz': 'ALL',
  'Insurance Australia': 'IAG', 'Suncorp': 'SUN', 'AMP': 'AMP',
  'Bendigo Bank': 'BEN', 'Bank of Queensland': 'BOQ', 'CSL': 'CSL',
  'ResMed': 'RMD', 'Cochlear': 'COH', 'Sonic Healthcare': 'SHL',
  'Ansell': 'ANN', 'Stockland': 'SGP', 'Mirvac': 'MGR',
  'Xero': 'XRO', 'WiseTech': 'WTC', 'NEC': 'NEC',
  'Computershare': 'CPU', 'Afterpay': 'APT', 'Zip Co': 'ZIP',
  'Qantas': 'QAN', 'Sydney Airport': 'SYD', 'APA Group': 'APA',
  'Atlas Arteria': 'ALX', 'Aurizon': 'AZJ', 'Brambles': 'BXB',
  'James Hardie': 'JHX', 'Lendlease': 'LLC', 'Lynas': 'LYC',
  'Mineral Resources': 'MIN', 'Newcrest': 'NCM', 'Northern Star': 'NST',
  'Origin Energy': 'ORG', 'Pilbara': 'PLS', 'Ramius': 'RMD',
  'Seek': 'SEK', 'Tabcorp': 'TAH', 'Treasury Wine': 'TWE',
  'Viva Energy': 'VEA', 'Whitehaven': 'WHC', 'Yancoal': 'YAL',
  
  // === LSE (UK) COMPANIES ===
  'HSBC': 'HSBA', 'BP': 'BP', 'Shell': 'SHEL', 'AstraZeneca': 'AZN',
  'GSK': 'GSK', 'Diageo': 'DGE', 'Unilever': 'ULVR', 'Rio Tinto': 'RIO',
  'Anglo American': 'AAL', 'BHP Group': 'BHP', 'London Stock Exchange': 'LSEG',
  'Lloyds': 'LLOY', 'Barclays': 'BARC', 'Vodafone': 'VOD', 'BT Group': 'BT.A',
  'National Grid': 'NG', 'Glencore': 'GLEN', 'Prudential': 'PRU',
  'Ashtead': 'AHT', 'BAE Systems': 'BA', 'CRH': 'CRH',
  'Flutter': 'FLTR', 'RELX': 'REL', 'Experian': 'EXPN', 'WPP': 'WPP',
  'International Airlines': 'IAG', 'Rolls-Royce': 'RR', 'Coca-Cola Europacific': 'CCH',
  'Tesco': 'TSCO', 'Melrose': 'MRO', 'Rotork': 'ROR',
  'Reckitt': 'RKT', 'Compass': 'CPG', 'Associated British': 'ABF',
  'Croda': 'CRDA', 'Halma': 'HLMA', 'Hikma': 'HIK',
  'Howden Joinery': 'HWDN', 'Intermediate Capital': 'ICP', 'JD Sports': 'JD',
  'Kingfisher': 'KGF', 'Land Securities': 'LAND', 'Legal & General': 'LGEN',
  'LondonMetric': 'LMP', 'M&G': 'MNG', 'Marks & Spencer': 'MKS',
  'Next': 'NXT', 'Ocado': 'OCDO', 'Pearson': 'PSON',
  'Persimmon': 'PSN', 'Phoenix': 'PHNX', 'Rightmove': 'RMV',
  'Rentokil': 'RTO', 'Sage': 'SGE', 'Severn Trent': 'SVT',
  'Smith & Nephew': 'SN', 'Spirax-Sarco': 'SPX', 'Standard Chartered': 'STAN',
  'Taylor Wimpey': 'TW', 'Unite': 'UTG', 'United Utilities': 'UU',
  'Weir': 'WEIR', 'Whitbread': 'WTB', '3i Group': 'III',
  
  // === HKEX (HONG KONG/CHINA) COMPANIES ===
  'Tencent': '0700', 'Alibaba': '9988', 'HSBC Holdings': '0005',
  'China Mobile': '0941', 'HKEX': '0388', 'AIA': '1299',
  'CNOOC': '0883', 'CCB': '0939', 'ICBC': '1398', 'Ping An': '2318',
  'PetroChina': '0857', 'CK Hutchison': '0001', 'CLP': '0002',
  'HK & China Gas': '0003', 'Hang Seng Bank': '0011', 'Henderson Land': '0012',
  'CK Asset': '0151', 'Galaxy Entertainment': '0027', 'MTR': '0066',
  'Hang Lung': '0083', 'Sino Land': '0083', 'China Unicom': '0762',
  'Link REIT': '0823', 'China Resources': '0291', 'Sinopec': '0386',
  'China Life': '2628', 'BYD': '1211', 'Xiaomi': '1810',
  'Meituan': '3690', 'JD': '9618', 'Pinduoduo': 'PDD',
  'Baidu': '9888', 'NetEase': '9999', 'Bilibili': '9626',
  'Li Auto': '2015', 'NIO': '9866', 'XPeng': '9868',
  'Semiconductor': '0981', 'Lenovo': '0992', 'CITIC': '0267',
  'Sun Hung Kai': '0016', 'New World': '0017', 'Swire Pacific': '0019',
  'Wheelock': '0020', 'Hang Lung Prop': '0101', 'Kerry Prop': '0688',
  'China Overseas': '0688', 'Longfor': '0960', 'Country Garden': '2007',
  'Evergrande': '3333', 'Sunac': '1918', 'Agile': '3383',
  
  // === TSX (CANADA) COMPANIES ===
  'Royal Bank': 'RY', 'TD Bank': 'TD', 'Scotiabank': 'BNS',
  'BMO': 'BMO', 'CIBC': 'CM', 'Canadian National': 'CNR',
  'Canadian Pacific': 'CP', 'Enbridge': 'ENB', 'TC Energy': 'TRP',
  'Power Corp': 'POW', 'Fortis': 'FTS', 'Algonquin': 'AQN',
  'Suncor': 'SU', 'Canadian Natural': 'CNQ', 'Imperial Oil': 'IMO',
  'Cenovus': 'CVE', 'Arc Resources': 'ARX', 'Whitecap': 'WCP',
  'MEG Energy': 'MEG', 'Tourmaline': 'TOU', 'Barrick Gold': 'ABX',
  'Newmont': 'NEM', 'Agnico Eagle': 'AEM', 'Wheaton': 'WPM',
  'Franco-Nevada': 'FNV', 'Osisko': 'OR', 'B2Gold': 'BTO',
  'Eldorado': 'ELD', 'Shopify': 'SHOP', 'Constellation': 'CSU',
  'Loblaw': 'L', 'Alimentation Couche-Tard': 'ATD', 'Metro': 'MRU',
  'Waste Connections': 'WCN', 'Restaurant Brands': 'QSR', 'Toromont': 'TIH',
  'Dollarama': 'DOL', 'Lightspeed': 'LSPD', 'CGI': 'GIB.A',
  'Open Text': 'OTEX', 'Kinross': 'K', 'Teck': 'TECK.B',
  'Hudbay': 'HBM', 'First Quantum': 'FM', 'Lundin': 'LUN',
  
  // === EUROPEAN COMPANIES ===
  'LVMH': 'MC', "L'Oreal": 'OR', 'Sanofi': 'SAN', 'Airbus': 'AIR',
  'TotalEnergies': 'TTE', 'EssilorLuxottica': 'EL', 'Danone': 'BN',
  'Kering': 'KER', 'Schneider': 'SU', 'BNP Paribas': 'BNP',
  'Credit Agricole': 'ACA', 'Societe Generale': 'GLE', 'Carrefour': 'CA',
  'Dassault Systemes': 'DSY', 'Capgemini': 'CAP', 'Unibail-Rodamco': 'URW',
  'STMicroelectronics': 'STM', 'Veolia': 'VIE', 'Publicis': 'PUB',
  'ASML Holding': 'ASML', 'Philips': 'PHIA', 'Unilever NL': 'UNA',
  'Heineken': 'HEIA', 'Adyen': 'ADYEN', 'Prosus NV': 'PRX',
  'Wolters Kluwer': 'WKL', 'Randstad': 'RAND', 'Aperam': 'APAM',
  'AkzoNobel': 'AKZA', 'UCB': 'UCB', 'KBC Group': 'KBC',
  'Argenx': 'ARGX', 'SAP': 'SAP', 'Siemens': 'SIE',
  'Allianz SE': 'ALV', 'BASF': 'BAS', 'Mercedes-Benz Group': 'MBG',
  'Volkswagen': 'VOW3', 'BMW': 'BMW', 'Deutsche Telekom': 'DTE',
  'Munich Re': 'MUV2', 'Adidas': 'ADS', 'Deutsche Bank': 'DBK',
  'Infineon': 'IFX', 'Linde': 'LIN', 'Henkel': 'HEN3',
  'Beiersdorf': 'BEI', 'Continental': 'CON', 'Fresenius': 'FRE',
  'HeidelbergCement': 'HEI', 'Rheinmetall': 'RHM', 'Covestro': '1COV',
  'Deutsche Post': 'DPW', 'Fresenius Medical Care': 'FME', 'Hannover Re': 'HNR1',
  'LEG Immobilien': 'LEG', 'MTU Aero Engines': 'MTX',
  'Puma': 'PUM', 'QIAGEN': 'QIA', 'RWE': 'RWE',
  'Siemens Healthineers': 'SHL', 'Symrise': 'SY1', 'Vonovia': 'VNA',
  'Zalando': 'ZAL', 'Ferrari': 'RACE',
  'Prysmian': 'PRY', 'Moncler': 'MONC', "Hermes International": 'RMS',
  'Chanel': 'CHANEL', 'Cartier': 'CFR',
};

export interface StockMention {
  symbol: string;
  name: string;
  type: 'stock' | 'index' | 'etf';
  context: string;
}

/**
 * Extracts stock tickers and index mentions from text
 */
export function extractStockMentions(text: string): StockMention[] {
  const mentions: StockMention[] = [];
  const foundSymbols = new Set<string>();
  const textLower = text.toLowerCase();

  // Check for index mentions
  for (const index of indices) {
    if (textLower.includes(index.toLowerCase())) {
      mentions.push({
        symbol: index.toUpperCase(),
        name: index,
        type: 'index',
        context: '',
      });
      foundSymbols.add(index.toUpperCase());
    }
  }

  // Check for company names and map to tickers
  for (const [company, ticker] of Object.entries(companyToTicker)) {
    if (!foundSymbols.has(ticker) && textLower.includes(company.toLowerCase())) {
      mentions.push({
        symbol: ticker,
        name: company,
        type: 'stock',
        context: '',
      });
      foundSymbols.add(ticker);
    }
  }

  // Look for ticker symbols in the text (uppercase 1-5 letters)
  const tickerRegex = /\b[A-Z]{1,5}\b/g;
  const matches = text.match(tickerRegex);
  
  if (matches) {
    for (const match of matches) {
      if (!foundSymbols.has(match) && knownTickers.has(match)) {
        mentions.push({
          symbol: match,
          name: match,
          type: 'stock',
          context: '',
        });
        foundSymbols.add(match);
      }
    }
  }

  // Check for ETF patterns
  const etfPatterns = ['ETF', 'ETN', 'Fund'];
  for (const pattern of etfPatterns) {
    const etfRegex = new RegExp(`\\b([A-Z]{2,4})\\s*${pattern}\\b`, 'gi');
    const etfMatches = text.match(etfRegex);
    if (etfMatches) {
      for (const etfMatch of etfMatches) {
        const symbolMatch = etfMatch.match(/[A-Z]{2,4}/);
        if (symbolMatch && !foundSymbols.has(symbolMatch[0])) {
          mentions.push({
            symbol: symbolMatch[0],
            name: etfMatch,
            type: 'etf',
            context: '',
          });
          foundSymbols.add(symbolMatch[0]);
        }
      }
    }
  }

  return mentions;
}

/**
 * Extracts unique stock symbols from an array of articles
 */
export function extractStocksFromArticles(articles: Array<{ title: string; description: string }>): StockMention[] {
  const allMentions = new Map<string, StockMention>();

  for (const article of articles) {
    const content = `${article.title} ${article.description}`;
    const mentions = extractStockMentions(content);
    
    for (const mention of mentions) {
      if (!allMentions.has(mention.symbol)) {
        allMentions.set(mention.symbol, mention);
      }
    }
  }

  return Array.from(allMentions.values());
}
