import { NavLink } from 'react-router-dom'
import { Activity, CalendarDays, Clock3, Film, HardDriveDownload, PlusCircle, Server, Settings, Telescope } from 'lucide-react'
import { clsx } from 'clsx'
import { useLiveFeedStore } from '../stores/liveFeed'

const links = [
  { to: '/', icon: Film, label: 'Movies', end: true },
  { to: '/add', icon: PlusCircle, label: 'Add Movie' },
  { to: '/calendar', icon: CalendarDays, label: 'Calendar' },
  { to: '/history', icon: Clock3, label: 'History' },
  { to: '/queue', icon: HardDriveDownload, label: 'Queue' },
  { to: '/activity', icon: Activity, label: 'Activity' },
  { to: '/wanted', icon: Telescope, label: 'Wanted' },
  { to: '/settings', icon: Settings, label: 'Settings' },
  { to: '/system/status', icon: Server, label: 'System' },
]

export default function Sidebar() {
  const connected = useLiveFeedStore((state) => state.connected)

  return (
    <aside className="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r border-gray-800 bg-gray-950/95 px-4 py-5 backdrop-blur md:flex">
      <div className="flex items-center gap-3 px-2">
        <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-yellow-400/15 text-yellow-400">
          <Film size={20} />
        </div>
        <div>
          <p className="text-sm font-semibold text-gray-50">GoRadarr</p>
          <p className="text-xs text-gray-500">Fast movie operations</p>
        </div>
      </div>

      <nav className="mt-8 flex flex-1 flex-col gap-1">
        {links.map(({ to, icon: Icon, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              clsx(
                'flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition',
                isActive
                  ? 'bg-yellow-400/10 text-yellow-400'
                  : 'text-gray-400 hover:bg-gray-900 hover:text-gray-100',
              )
            }
          >
            <Icon size={17} />
            <span>{label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="panel-muted mt-4 flex items-center gap-3 px-3 py-3">
        <span className={clsx('h-2.5 w-2.5 rounded-full', connected ? 'bg-green-400' : 'bg-gray-600')} />
        <div>
          <p className="text-sm font-medium text-gray-200">{connected ? 'Realtime active' : 'Realtime offline'}</p>
          <p className="text-xs text-gray-500">Scheduler, queue, and movie updates</p>
        </div>
      </div>
    </aside>
  )
}
