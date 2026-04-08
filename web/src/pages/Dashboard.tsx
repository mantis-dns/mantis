import { Activity, ShieldOff, Database, Zap } from 'lucide-react'
import StatCard from '../components/StatCard'

export default function Dashboard() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard title="Total Queries" value="0" icon={Activity} color="text-emerald-400" />
        <StatCard title="Blocked" value="0" icon={ShieldOff} color="text-red-400" subtitle="0%" />
        <StatCard title="Cache Hits" value="0" icon={Database} color="text-blue-400" subtitle="0%" />
        <StatCard title="Latency" value="0ms" icon={Zap} color="text-amber-400" subtitle="avg" />
      </div>

      <div className="bg-slate-900 rounded-xl border border-slate-800 p-6">
        <h2 className="text-lg font-semibold mb-4">Queries Over Time</h2>
        <div className="h-64 flex items-center justify-center text-slate-500">
          Chart will appear when queries are processed
        </div>
      </div>
    </div>
  )
}
