import DataTable from '../components/DataTable'

export default function Blocklists() {
  const columns = [
    { key: 'name', header: 'Name' },
    { key: 'url', header: 'URL' },
    { key: 'format', header: 'Format' },
    { key: 'domainCount', header: 'Domains' },
    { key: 'lastStatus', header: 'Status' },
    { key: 'enabled', header: 'Enabled' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Blocklists</h1>
        <button className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
          Add Blocklist
        </button>
      </div>
      <DataTable columns={columns} data={[]} />
    </div>
  )
}
