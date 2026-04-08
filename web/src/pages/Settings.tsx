import { useState, useEffect } from 'react'
import { useSettings, useUpdateSettings } from '../api/settings'

export default function Settings() {
  const { data: settings } = useSettings()
  const updateMutation = useUpdateSettings()
  const [values, setValues] = useState<Record<string, string>>({})
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (settings) {
      const map: Record<string, string> = {}
      for (const s of settings) {
        map[s.key] = s.value
      }
      setValues(map)
    }
  }, [settings])

  const handleSave = () => {
    updateMutation.mutate(values, {
      onSuccess: () => {
        setSaved(true)
        setTimeout(() => setSaved(false), 2000)
      },
    })
  }

  const set = (key: string, value: string) => setValues((v) => ({ ...v, [key]: value }))

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Settings</h1>

      <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
        <h2 className="text-lg font-semibold mb-4">DNS</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">Upstream DNS Servers</label>
            <input
              value={values['dns.upstreams'] ?? '1.1.1.1, 8.8.8.8'}
              onChange={(e) => set('dns.upstreams', e.target.value)}
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500"
            />
          </div>
          <div>
            <label className="block text-sm text-slate-400 mb-1">Blocking Mode</label>
            <select
              value={values['dns.blocking_mode'] ?? 'null'}
              onChange={(e) => set('dns.blocking_mode', e.target.value)}
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500"
            >
              <option value="null">Null (0.0.0.0)</option>
              <option value="nxdomain">NXDOMAIN</option>
            </select>
          </div>
        </div>
      </div>

      <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
        <h2 className="text-lg font-semibold mb-4">Logging</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm text-slate-400 mb-1">Log Level</label>
            <select
              value={values['logging.level'] ?? 'info'}
              onChange={(e) => set('logging.level', e.target.value)}
              className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-slate-200 focus:outline-none focus:border-emerald-500"
            >
              <option value="debug">Debug</option>
              <option value="info">Info</option>
              <option value="warn">Warn</option>
              <option value="error">Error</option>
            </select>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <button onClick={handleSave} className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg font-medium transition-colors">
          Save Settings
        </button>
        {saved && <span className="text-sm text-emerald-400">Saved!</span>}
      </div>
    </div>
  )
}
