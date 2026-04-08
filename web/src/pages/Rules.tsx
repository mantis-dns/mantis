import { useState } from 'react'
import DataTable from '../components/DataTable'
import Modal from '../components/Modal'
import { useRules, useCreateRule, useDeleteRule } from '../api/rules'
import { formatRelativeTime } from '../utils/format'

export default function Rules() {
  const [showAdd, setShowAdd] = useState(false)
  const [domain, setDomain] = useState('')
  const [type, setType] = useState<'block' | 'allow'>('block')
  const [comment, setComment] = useState('')

  const { data: rules, isLoading } = useRules()
  const createMutation = useCreateRule()
  const deleteMutation = useDeleteRule()

  const handleAdd = () => {
    createMutation.mutate({ domain, type, comment: comment || undefined }, {
      onSuccess: () => {
        setShowAdd(false)
        setDomain('')
        setComment('')
      },
    })
  }

  const columns = [
    {
      key: 'domain',
      header: 'Domain',
      render: (row: Record<string, unknown>) => <span className="font-mono text-sm">{String(row.domain)}</span>,
    },
    {
      key: 'type',
      header: 'Type',
      render: (row: Record<string, unknown>) => (
        <span className={row.type === 'block' ? 'text-red-400' : 'text-emerald-400'}>{String(row.type)}</span>
      ),
    },
    { key: 'comment', header: 'Comment' },
    {
      key: 'created',
      header: 'Created',
      render: (row: Record<string, unknown>) => <span className="text-xs text-slate-400">{formatRelativeTime(String(row.created))}</span>,
    },
    {
      key: 'actions',
      header: '',
      render: (row: Record<string, unknown>) => (
        <button onClick={() => deleteMutation.mutate(String(row.id))} className="text-xs text-red-400 hover:text-red-300">Delete</button>
      ),
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Custom Rules</h1>
        <button onClick={() => setShowAdd(true)} className="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
          Add Rule
        </button>
      </div>
      <DataTable columns={columns} data={(rules ?? [])} loading={isLoading} />

      <Modal open={showAdd} onClose={() => setShowAdd(false)} title="Add Rule">
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">Domain</label>
            <input value={domain} onChange={(e) => setDomain(e.target.value)} placeholder="ads.example.com" className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Type</label>
            <select value={type} onChange={(e) => setType(e.target.value as 'block' | 'allow')} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500">
              <option value="block">Block</option>
              <option value="allow">Allow</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Comment (optional)</label>
            <input value={comment} onChange={(e) => setComment(e.target.value)} className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500" />
          </div>
          <button onClick={handleAdd} disabled={!domain} className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium disabled:opacity-50">Add</button>
        </div>
      </Modal>
    </div>
  )
}
