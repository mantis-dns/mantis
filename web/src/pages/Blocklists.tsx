import { useState } from 'react'
import DataTable from '../components/DataTable'
import Modal from '../components/Modal'
import Toggle from '../components/Toggle'
import { useBlocklists, useCreateBlocklist, useUpdateBlocklist, useDeleteBlocklist } from '../api/blocklists'

export default function Blocklists() {
  const [showAdd, setShowAdd] = useState(false)
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [format, setFormat] = useState('hosts')

  const { data: blocklists, isLoading } = useBlocklists()
  const createMutation = useCreateBlocklist()
  const updateMutation = useUpdateBlocklist()
  const deleteMutation = useDeleteBlocklist()

  const handleAdd = () => {
    createMutation.mutate({ name, url, format }, {
      onSuccess: () => {
        setShowAdd(false)
        setName('')
        setUrl('')
      },
    })
  }

  const columns = [
    { key: 'name', header: 'Name' },
    {
      key: 'url',
      header: 'URL',
      render: (row: Record<string, unknown>) => (
        <span className="font-mono text-xs text-slate-400 truncate max-w-xs block">{String(row.url)}</span>
      ),
    },
    { key: 'format', header: 'Format' },
    { key: 'domainCount', header: 'Domains' },
    {
      key: 'lastStatus',
      header: 'Status',
      render: (row: Record<string, unknown>) => {
        const colors: Record<string, string> = { success: 'text-emerald-400', error: 'text-red-400', pending: 'text-amber-400' }
        const status = String(row.lastStatus)
        return <span className={colors[status] || ''}>{status}</span>
      },
    },
    {
      key: 'enabled',
      header: 'Enabled',
      render: (row: Record<string, unknown>) => (
        <Toggle
          checked={Boolean(row.enabled)}
          onChange={(checked) => updateMutation.mutate({ id: String(row.id), enabled: checked })}
        />
      ),
    },
    {
      key: 'actions',
      header: '',
      render: (row: Record<string, unknown>) => (
        <button
          onClick={() => { if (confirm('Delete this blocklist?')) deleteMutation.mutate(String(row.id)) }}
          className="text-xs text-red-400 hover:text-red-300"
        >
          Delete
        </button>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Blocklists</h1>
        <button
          onClick={() => setShowAdd(true)}
          className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors"
        >
          Add Blocklist
        </button>
      </div>
      <DataTable columns={columns} data={(blocklists ?? [])} loading={isLoading} />

      <Modal open={showAdd} onClose={() => setShowAdd(false)} title="Add Blocklist">
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">Name</label>
            <input value={name} onChange={(e) => setName(e.target.value)} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">URL</label>
            <input value={url} onChange={(e) => setUrl(e.target.value)} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Format</label>
            <select value={format} onChange={(e) => setFormat(e.target.value)} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500">
              <option value="hosts">Hosts file</option>
              <option value="domains">Domain list</option>
              <option value="adblock">Adblock</option>
            </select>
          </div>
          <button onClick={handleAdd} disabled={!name || !url} className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium disabled:opacity-50">
            Add
          </button>
        </div>
      </Modal>
    </div>
  )
}
