import DataTable from '../components/DataTable'

export default function Network() {
  const columns = [
    { key: 'hostname', header: 'Hostname' },
    { key: 'ip', header: 'IP' },
    { key: 'mac', header: 'MAC' },
    { key: 'queryCount', header: 'Queries' },
    { key: 'lastSeen', header: 'Last Seen' },
  ]

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Network Clients</h1>
      <DataTable columns={columns} data={[]} />
    </div>
  )
}
