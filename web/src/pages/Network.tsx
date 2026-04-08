import DataTable from '../components/DataTable'
import { useTopClients } from '../api/stats'

export default function Network() {
  const { data: clients, isLoading } = useTopClients()

  const columns = [
    { key: 'ip', header: 'IP', render: (row: Record<string, unknown>) => <span className="font-mono text-sm">{String(row.ip)}</span> },
    { key: 'hostname', header: 'Hostname' },
    { key: 'count', header: 'Queries' },
  ]

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Network Clients</h1>
      <DataTable columns={columns} data={(clients ?? [])} loading={isLoading} />
    </div>
  )
}
