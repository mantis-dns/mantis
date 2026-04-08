import { Activity, ShieldOff, Database, Zap } from 'lucide-react'
import StatCard from '../components/StatCard'
import Chart from '../components/Chart'
import { useStatsSummary, useStatsOvertime } from '../api/stats'
import { formatNumber } from '../utils/format'

export default function Dashboard() {
  const { data: summary } = useStatsSummary()
  const { data: overtime } = useStatsOvertime()

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Queries"
          value={formatNumber(summary?.totalQueries ?? 0)}
          icon={Activity}
          color="text-emerald-400"
        />
        <StatCard
          title="Blocked"
          value={formatNumber(summary?.blockedCount ?? 0)}
          icon={ShieldOff}
          color="text-red-400"
          subtitle={`${(summary?.blockedPercent ?? 0).toFixed(1)}%`}
        />
        <StatCard
          title="Cache Hits"
          value={formatNumber(summary?.cachedCount ?? 0)}
          icon={Database}
          color="text-blue-400"
          subtitle={`${(summary?.cacheHitRatio ?? 0).toFixed(1)}%`}
        />
        <StatCard title="Domains Blocked" value="--" icon={Zap} color="text-amber-400" />
      </div>

      <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
        <h2 className="text-lg font-semibold mb-4">Queries Over Time</h2>
        <Chart
          data={overtime ?? []}
          xKey="timestamp"
          lines={[
            { key: 'totalQueries', color: '#10B981', name: 'Total' },
            { key: 'blockedCount', color: '#F87171', name: 'Blocked' },
            { key: 'cachedCount', color: '#60A5FA', name: 'Cached' },
          ]}
        />
      </div>
    </div>
  )
}
