import Chart from '../components/Chart'
import { useStatsOvertime, useTopDomains, useTopClients } from '../api/stats'

export default function Statistics() {
  const { data: overtime } = useStatsOvertime()
  const { data: topDomains } = useTopDomains()
  const { data: topClients } = useTopClients()

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Statistics</h1>

      <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
        <h2 className="text-lg font-semibold mb-4">Queries Over Time</h2>
        <Chart
          data={overtime ?? []}
          xKey="timestamp"
          bars={[
            { key: 'totalQueries', color: '#10B981', name: 'Total' },
            { key: 'blockedCount', color: '#F87171', name: 'Blocked' },
          ]}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
          <h2 className="text-lg font-semibold mb-4">Top Domains</h2>
          {(topDomains?.allowed ?? []).length === 0 ? (
            <p className="text-slate-500 text-sm">No data yet</p>
          ) : (
            <div className="space-y-2">
              {(topDomains?.allowed ?? []).map((d) => (
                <div key={d.domain} className="flex items-center justify-between">
                  <span className="font-mono text-sm text-slate-300 truncate">{d.domain}</span>
                  <span className="text-sm text-slate-400">{d.count}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
          <h2 className="text-lg font-semibold mb-4">Top Clients</h2>
          {(topClients ?? []).length === 0 ? (
            <p className="text-slate-500 text-sm">No data yet</p>
          ) : (
            <div className="space-y-2">
              {(topClients ?? []).map((c) => (
                <div key={c.ip} className="flex items-center justify-between">
                  <span className="font-mono text-sm text-slate-300">{c.ip}</span>
                  <span className="text-sm text-slate-400">{c.count}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
