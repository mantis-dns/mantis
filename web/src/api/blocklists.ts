import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

interface BlocklistSource {
  id: string
  name: string
  url: string
  enabled: boolean
  format: string
  domainCount: number
  lastUpdated?: string
  lastStatus: string
}

export function useBlocklists() {
  return useQuery({
    queryKey: ['blocklists'],
    queryFn: () => apiClient.get<{ data: BlocklistSource[] }>('/blocklists').then((r) => r.data),
  })
}

export function useCreateBlocklist() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; url: string; format: string }) =>
      apiClient.post('/blocklists', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blocklists'] }),
  })
}

export function useUpdateBlocklist() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; name?: string; url?: string; enabled?: boolean; format?: string }) =>
      apiClient.put(`/blocklists/${id}`, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blocklists'] }),
  })
}

export function useDeleteBlocklist() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/blocklists/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['blocklists'] }),
  })
}

export type { BlocklistSource }
