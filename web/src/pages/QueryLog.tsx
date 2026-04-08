import { useState } from 'react'
import DataTable from '../components/DataTable'
import SearchInput from '../components/SearchInput'

export default function QueryLog() {
  const [search, setSearch] = useState('')

  const columns = [
    { key: 'timestamp', header: 'Time' },
    { key: 'domain', header: 'Domain' },
    { key: 'clientIp', header: 'Client' },
    { key: 'queryType', header: 'Type' },
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
        return <span className={colors[result] || ''}>{result}</span>
      },
    },
    { key: 'latencyUs', header: 'Latency' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Query Log</h1>
      </div>
      <SearchInput value={search} onChange={setSearch} placeholder="Search domains..." />
      <DataTable columns={columns} data={[]} />
    </div>
  )
}
