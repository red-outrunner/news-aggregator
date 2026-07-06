// Maps ticker symbols to the exchange/region they most likely trade on.
// Lookup sets are built once at module load.

const jseSymbols = new Set([
  'AGL', 'ANG', 'ARI', 'BHP', 'BVT', 'CFR', 'CLS', 'CPI', 'DRM', 'EXX', 'FSR', 'GFI', 'GOLD', 'HAR', 'IMP', 'INL', 'INP', 'KIO', 'LHC', 'MNP', 'MRP', 'NPN', 'OMU', 'PIK', 'PPC', 'REM', 'RMI', 'SHP', 'SLM', 'SOL', 'SSW', 'TCG', 'TFG', 'VOD', 'WHL', 'WKP', 'ZMP', 'MTN', 'TKG', 'SBK', 'NED', 'DSY', 'BID', 'BAW', 'DCP', 'MEH', 'NTC', 'APN', 'ADI', 'MSM', 'PPK', 'JDG', 'LEW', 'MTC', 'GRT', 'RDF', 'HYB', 'NEP', 'VKE', 'ACT', 'CAP', 'DIP', 'EMS', 'FAIR',
]);

const asxSymbols = new Set([
  'CBA', 'WBC', 'NAB', 'ANZ', 'MQG', 'QBE', 'WES', 'FMG', 'STO', 'WDS', 'COL', 'WOW', 'TLS', 'TCL', 'ALL', 'IAG', 'SUN', 'AMP', 'BEN', 'BOQ', 'CSL', 'RMD', 'COH', 'SHL', 'ANN', 'SGP', 'MGR', 'XRO', 'WTC', 'NEC', 'CPU', 'APT', 'ZIP', 'QAN', 'SYD', 'APA', 'ALX', 'AZJ', 'BXB', 'JHX', 'LLC', 'LYC', 'MIN', 'NCM', 'NST', 'ORG', 'PLS', 'SEK', 'TAH', 'TWE', 'VEA', 'WHC', 'YAL',
]);

const lseSymbols = new Set([
  'HSBA', 'BP', 'SHEL', 'AZN', 'GSK', 'DGE', 'ULVR', 'RIO', 'AAL', 'LSEG', 'LLOY', 'BARC', 'BT.A', 'NG', 'GLEN', 'PRU', 'AHT', 'BA', 'CRH', 'FLTR', 'REL', 'EXPN', 'WPP', 'RR', 'CCH', 'TSCO', 'MRO', 'RTO', 'SMDS', 'BRBY', 'JD', 'OCDO', 'PSON', 'WTB', 'MNDI', 'HWDN', 'BDEV',
]);

// HKEX and TSE both use 4-digit codes; without an exchange suffix the
// symbol alone can't distinguish them
const numericCode = /^\d{4}$/;

/**
 * Best-effort market/region label for a ticker symbol
 */
export function getMarket(symbol: string): string {
  if (numericCode.test(symbol)) return 'HKEX/TSE';
  if (jseSymbols.has(symbol)) return 'JSE';
  if (asxSymbols.has(symbol)) return 'ASX';
  if (lseSymbols.has(symbol)) return 'LSE';
  return 'NYSE/NASDAQ';
}
