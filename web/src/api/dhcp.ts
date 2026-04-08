import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from './client'

interface DhcpLease {
  mac: string
  ip: string
  hostname?: string
  leaseEnd: string
  isStatic: boolean
}

export function useLeases() {
  return useQuery({
    queryKey: ['dhcp', 'leases'],
    queryFn: () => apiClient.get<{ data: DhcpLease[] }>('/dhcp/leases').then((r) => r.data),
  })
}

export function useCreateStaticLease() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { mac: string; ip: string; hostname?: string }) =>
      apiClient.post('/dhcp/leases/static', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dhcp'] }),
  })
}

export function useDeleteStaticLease() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (mac: string) => apiClient.delete(`/dhcp/leases/static/${mac}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['dhcp'] }),
  })
}

export type { DhcpLease }
