import { Outlet, useLocation } from 'react-router-dom'
import { Activity, Plus } from 'lucide-react'
import Sidebar from './Sidebar'
import { useLiveFeed } from '../hooks/useLiveFeed'
import { useLiveFeedStore } from '../stores/liveFeed'

const titles: Record<string, { title: string; subtitle: string }> = {
  '/': { title: 'Movies', subtitle: 'Library health, quality, and discovery at a glance.' },
  '/add': { title: 'Add Movie', subtitle: 'Search TMDB and add titles in a single flow.' },
  '/calendar': { title: 'Release Calendar', subtitle: 'Track theatrical, digital, and physical releases.' },
  '/history': { title: 'History', subtitle: 'Audit grabs, imports, and failures.' },
  '/queue': { title: 'Queue', subtitle: 'Monitor active downloads and retry stuck items.' },
  '/wanted': { title: 'Wanted', subtitle: 'Find monitored movies still missing from the library.' },
  '/settings': { title: 'Settings', subtitle: 'Manage quality, clients, indexers, and notifications.' },
  '/system/status': { title: 'System', subtitle: 'Operational status, storage, and live events.' },
}

export default function Layout() {
  useLiveFeed()
  const location = useLocation()
  const connected = useLiveFeedStore((state) => state.connected)
  const latestEvent = useLiveFeedStore((state) => state.events[0])
  const meta = titles[location.pathname] ?? titles['/']

  return (
    <div className="flex min-h-screen bg-transparent text-gray-50">
      <Sidebar />
      <div className="flex min-h-screen flex-1 flex-col md:ml-64">
        <header className="sticky top-0 z-30 border-b border-gray-800/80 bg-gray-950/85 backdrop-blur">
          <div className="flex flex-col gap-4 px-4 py-4 md:px-8 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.3em] text-yellow-400">GoRadarr</p>
              <h1 className="mt-1 text-2xl font-semibold text-gray-50">{meta.title}</h1>
              <p className="mt-1 text-sm text-gray-400">{meta.subtitle}</p>
            </div>
            <div className="flex flex-col gap-3 md:flex-row md:items-center">
              <div className="panel-muted flex items-center gap-3 px-4 py-3 text-sm">
                <Activity size={16} className={connected ? 'text-green-400' : 'text-gray-500'} />
                <div>
                  <p className="font-medium text-gray-200">{connected ? 'Live feed connected' : 'Live feed reconnecting'}</p>
                  <p className="text-xs text-gray-500">
                    {latestEvent ? latestEvent.eventType : 'Waiting for scheduler and library events'}
                  </p>
                </div>
              </div>
              <a href="/add" className="btn-primary">
                <Plus size={16} />
                Add Movie
              </a>
            </div>
          </div>
        </header>
        <main className="flex-1 px-4 py-6 md:px-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
