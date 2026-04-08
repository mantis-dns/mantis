import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  List,
  ShieldBan,
  ListChecks,
  Network,
  Wifi,
  Settings,
  BarChart3,
  LogOut,
} from 'lucide-react'
import { useAuthStore } from '../store/authStore'

const links = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/queries', icon: List, label: 'Query Log' },
  { to: '/blocklists', icon: ShieldBan, label: 'Blocklists' },
  { to: '/rules', icon: ListChecks, label: 'Rules' },
  { to: '/dhcp', icon: Wifi, label: 'DHCP' },
  { to: '/network', icon: Network, label: 'Network' },
  { to: '/statistics', icon: BarChart3, label: 'Statistics' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

export default function Sidebar({ open }: { open: boolean }) {
  const logout = useAuthStore((s) => s.logout)

  return (
    <aside
      className={`${
        open ? 'w-60' : 'w-0 -ml-60'
      } transition-all duration-200 bg-slate-900 border-r border-slate-800 flex flex-col h-full overflow-hidden lg:w-60 lg:ml-0`}
    >
      <div className="h-14 flex items-center px-5 border-b border-slate-800">
        <span className="text-lg font-bold text-emerald-400">Mantis</span>
      </div>

      <nav className="flex-1 py-4 space-y-1 px-3">
        {links.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                isActive
                  ? 'bg-emerald-500/10 text-emerald-400'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
              }`
            }
          >
            <Icon size={18} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="p-3 border-t border-slate-800">
        <button
          onClick={() => logout()}
          className="flex items-center gap-3 px-3 py-2 w-full rounded-lg text-sm text-slate-400 hover:text-red-400 hover:bg-slate-800 transition-colors"
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </aside>
  )
}
