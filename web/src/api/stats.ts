import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'
import { REFETCH_INTERVAL } from '../utils/constants'

interface StatsSummary {
  totalQueries: number
  blockedCount: number
  cachedCount: number
  blockedPercent: number
  cacheHitRatio: number
}

interface StatsPoint {
  timestamp: string
  totalQueries: number
  blockedCount: number
  cachedCount: number
}

interface TopDomain {
  domain: string
  count: number
}

interface TopClient {
  ip: string
  hostname?: string
  count: number
}

export function useStatsSummary() {
  return useQuery({
    queryKey: ['stats', 'summary'],
    queryFn: () => apiClient.get<{ data: StatsSummary }>('/stats/summary').then((r) => r.data),
    refetchInterval: REFETCH_INTERVAL,
  })
}

export function useStatsOvertime() {
  return useQuery({
    queryKey: ['stats', 'overtime'],
    queryFn: () => apiClient.get<{ data: StatsPoint[] }>('/stats/overtime').then((r) => r.data),
    refetchInterval: REFETCH_INTERVAL,
  })
}

export function useTopDomains() {
  return useQuery({
    queryKey: ['stats', 'top-domains'],
    queryFn: () => apiClient.get<{ data: { allowed: TopDomain[]; blocked: TopDomain[] } }>('/stats/top-domains').then((r) => r.data),
    refetchInterval: REFETCH_INTERVAL,
  })
}

export function useTopClients() {
  return useQuery({
    queryKey: ['stats', 'top-clients'],
    queryFn: () => apiClient.get<{ data: TopClient[] }>('/stats/top-clients').then((r) => r.data),
    refetchInterval: REFETCH_INTERVAL,
  })
}
