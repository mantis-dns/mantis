import { useQuery } from '@tanstack/react-query'
import { apiClient } from './client'

interface QueryLogEntry {
  id: number
  timestamp: string
  clientIp: string
  domain: string
  queryType: number
  result: string
  upstream?: string
  latencyUs: number
  answer?: string
}

interface QueryLogResponse {
  data: QueryLogEntry[]
  meta: {
    page: number
    perPage: number
    total: number
  }
}

interface QueryLogParams {
  page?: number
  perPage?: number
  domain?: string
  client?: string
  result?: string
}

export function useQueryLog(params: QueryLogParams = {}) {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', String(params.page))
  if (params.perPage) searchParams.set('perPage', String(params.perPage))
  if (params.domain) searchParams.set('domain', params.domain)
  if (params.client) searchParams.set('client', params.client)
  if (params.result) searchParams.set('result', params.result)

  const qs = searchParams.toString()
  const path = `/queries${qs ? `?${qs}` : ''}`

  return useQuery({
    queryKey: ['queries', params],
    queryFn: () => apiClient.get<QueryLogResponse>(path),
  })
}

export type { QueryLogEntry }
