import DataTable from '../components/DataTable'

export default function Rules() {
  const columns = [
    { key: 'domain', header: 'Domain' },
    { key: 'type', header: 'Type' },
    { key: 'comment', header: 'Comment' },
    { key: 'created', header: 'Created' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Custom Rules</h1>
        <button className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
          Add Rule
        </button>
      </div>
      <DataTable columns={columns} data={[]} />
    </div>
  )
}
