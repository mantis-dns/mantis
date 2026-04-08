export const API_BASE = '/api/v1'

export const QUERY_TYPES: Record<number, string> = {
  1: 'A',
  2: 'NS',
  5: 'CNAME',
  6: 'SOA',
  15: 'MX',
  16: 'TXT',
  28: 'AAAA',
  33: 'SRV',
  255: 'ANY',
}

export const REFETCH_INTERVAL = 10_000 // 10 seconds
