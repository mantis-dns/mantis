import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

interface CustomRule {
  id: string
  domain: string
  type: 'block' | 'allow'
  comment?: string
  created: string
}

export function useRules() {
  return useQuery({
    queryKey: ['rules'],
    queryFn: () => apiClient.get<{ data: CustomRule[] }>('/rules').then((r) => r.data),
  })
}

export function useCreateRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { domain: string; type: 'block' | 'allow'; comment?: string }) =>
      apiClient.post('/rules', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rules'] }),
  })
}

export function useDeleteRule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/rules/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['rules'] }),
  })
}

export type { CustomRule }
