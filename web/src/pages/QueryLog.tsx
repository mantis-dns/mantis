import { useState } from 'react'
import DataTable from '../components/DataTable'
import SearchInput from '../components/SearchInput'
import { useQueryLog } from '../api/queries'
import { useDebounce } from '../hooks/useDebounce'
import { QUERY_TYPES } from '../utils/constants'
import { formatLatency, formatRelativeTime } from '../utils/format'

export default function QueryLog() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const debouncedSearch = useDebounce(search, 300)

  const { data, isLoading } = useQueryLog({
    page,
    perPage: 50,
    domain: debouncedSearch || undefined,
  })

  const entries = data?.data ?? []
  const total = data?.meta?.total ?? 0

  const columns = [
    {
      key: 'timestamp',
      header: 'Time',
      render: (row: Record<string, unknown>) => (
        <span className="text-xs font-mono text-slate-400">{formatRelativeTime(String(row.timestamp))}</span>
      ),
    },
    {
      key: 'domain',
      header: 'Domain',
      render: (row: Record<string, unknown>) => (
        <span className="font-mono text-sm">{String(row.domain)}</span>
      ),
    },
    { key: 'clientIp', header: 'Client' },
    {
      key: 'queryType',
      header: 'Type',
      render: (row: Record<string, unknown>) => (
        <span className="text-xs bg-slate-800 px-2 py-0.5 rounded">
          {QUERY_TYPES[Number(row.queryType)] ?? String(row.queryType)}
        </span>
      ),
    },
    {
      key: 'result',
      header: 'Result',
      render: (row: Record<string, unknown>) => {
        const colors: Record<string, string> = {
          allowed: 'text-emerald-400',
          blocked: 'text-red-400',
          cached: 'text-blue-400',
          error: 'text-amber-400',
        }
        const result = String(row.result)
        return <span className={`text-xs font-medium ${colors[result] || ''}`}>{result}</span>
      },
    },
    {
      key: 'latencyUs',
      header: 'Latency',
      render: (row: Record<string, unknown>) => (
        <span className="text-xs text-slate-400">{formatLatency(Number(row.latencyUs))}</span>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Query Log</h1>
        <span className="text-sm text-slate-400">{total} total</span>
      </div>
      <SearchInput value={search} onChange={setSearch} placeholder="Search domains..." />
      <DataTable columns={columns} data={entries} loading={isLoading} />
      {total > 50 && (
        <div className="flex justify-center gap-2">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-3 py-1 bg-slate-800 rounded text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="px-3 py-1 text-sm text-slate-400">Page {page}</span>
          <button
            onClick={() => setPage((p) => p + 1)}
            disabled={page * 50 >= total}
            className="px-3 py-1 bg-slate-800 rounded text-sm disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  )
}
