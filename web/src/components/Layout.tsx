import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import { useUIStore } from '../store/uiStore'
import { Menu } from 'lucide-react'

export default function Layout() {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)
  const toggleSidebar = useUIStore((s) => s.toggleSidebar)

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar open={sidebarOpen} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-14 flex items-center px-4 bg-slate-900 border-b border-slate-800 lg:hidden">
          <button onClick={toggleSidebar} className="p-2 hover:bg-slate-800 rounded">
            <Menu size={20} />
          </button>
          <span className="ml-3 font-semibold text-emerald-400">Mantis</span>
        </header>
        <main className="flex-1 overflow-auto p-6 bg-slate-950">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
