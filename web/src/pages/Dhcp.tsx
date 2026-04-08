import { useState } from 'react'
import DataTable from '../components/DataTable'
import Modal from '../components/Modal'
import { useLeases, useCreateStaticLease, useDeleteStaticLease } from '../api/dhcp'
import { formatRelativeTime } from '../utils/format'

export default function Dhcp() {
  const [showAdd, setShowAdd] = useState(false)
  const [mac, setMac] = useState('')
  const [ip, setIp] = useState('')
  const [hostname, setHostname] = useState('')

  const { data: leases, isLoading } = useLeases()
  const createMutation = useCreateStaticLease()
  const deleteMutation = useDeleteStaticLease()

  const handleAdd = () => {
    createMutation.mutate({ mac, ip, hostname: hostname || undefined }, {
      onSuccess: () => {
        setShowAdd(false)
        setMac('')
        setIp('')
        setHostname('')
      },
    })
  }

  const columns = [
    { key: 'mac', header: 'MAC', render: (row: Record<string, unknown>) => <span className="font-mono text-xs">{String(row.mac)}</span> },
    { key: 'ip', header: 'IP', render: (row: Record<string, unknown>) => <span className="font-mono text-sm">{String(row.ip)}</span> },
    { key: 'hostname', header: 'Hostname' },
    {
      key: 'leaseEnd',
      header: 'Lease End',
      render: (row: Record<string, unknown>) => <span className="text-xs text-slate-400">{formatRelativeTime(String(row.leaseEnd))}</span>,
    },
    {
      key: 'isStatic',
      header: 'Static',
      render: (row: Record<string, unknown>) => row.isStatic ? <span className="text-emerald-400 text-xs">Yes</span> : <span className="text-slate-500 text-xs">No</span>,
    },
    {
      key: 'actions',
      header: '',
      render: (row: Record<string, unknown>) => row.isStatic ? (
        <button onClick={() => deleteMutation.mutate(String(row.mac))} className="text-xs text-red-400 hover:text-red-300">Delete</button>
      ) : null,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">DHCP Leases</h1>
        <button onClick={() => setShowAdd(true)} className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
          Add Static Lease
        </button>
      </div>
      <DataTable columns={columns} data={(leases ?? [])} loading={isLoading} />

      <Modal open={showAdd} onClose={() => setShowAdd(false)} title="Add Static Lease">
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">MAC Address</label>
            <input value={mac} onChange={(e) => setMac(e.target.value)} placeholder="AA:BB:CC:DD:EE:FF" className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">IP Address</label>
            <input value={ip} onChange={(e) => setIp(e.target.value)} placeholder="192.168.1.100" className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Hostname (optional)</label>
            <input value={hostname} onChange={(e) => setHostname(e.target.value)} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <button onClick={handleAdd} disabled={!mac || !ip} className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium disabled:opacity-50">Add</button>
        </div>
      </Modal>
    </div>
  )
}
