import DataTable from '../components/DataTable'

export default function Dhcp() {
  const columns = [
    { key: 'mac', header: 'MAC' },
    { key: 'ip', header: 'IP' },
    { key: 'hostname', header: 'Hostname' },
    { key: 'leaseEnd', header: 'Lease End' },
    { key: 'isStatic', header: 'Static' },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">DHCP Leases</h1>
        <button className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
          Add Static Lease
        </button>
      </div>
      <DataTable columns={columns} data={[]} />
    </div>
  )
}
