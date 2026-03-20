import { useQuery } from '@tanstack/react-query'
import { HardDrive, ServerCog } from 'lucide-react'
import api from '../lib/api'
import { formatBytes, formatUptime } from '../lib/format'
import type { DiskSpace, SystemStatus } from '../lib/types'
import { useLiveFeedStore } from '../stores/liveFeed'

export default function System() {
  const status = useQuery<SystemStatus>({ queryKey: ['system-status'], queryFn: () => api.get('/system/status').then((res) => res.data) })
  const disks = useQuery<DiskSpace[]>({ queryKey: ['disk-space'], queryFn: () => api.get('/system/diskspace').then((res) => res.data) })
  const connected = useLiveFeedStore((state) => state.connected)
  const events = useLiveFeedStore((state) => state.events)

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-4">
        <div className="panel p-5"><p className="text-sm text-gray-400">Version</p><p className="stat-value mt-2">{status.data?.version ?? '—'}</p></div>
        <div className="panel p-5"><p className="text-sm text-gray-400">Uptime</p><p className="stat-value mt-2">{status.data ? formatUptime(status.data.startTime) : '—'}</p></div>
        <div className="panel p-5"><p className="text-sm text-gray-400">Runtime</p><p className="stat-value mt-2">{status.data?.runtimeName ?? '—'}</p></div>
        <div className="panel p-5"><p className="text-sm text-gray-400">Realtime</p><p className="stat-value mt-2">{connected ? 'Live' : 'Offline'}</p></div>
      </section>

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-6">
          <div className="panel p-5">
            <h3 className="flex items-center gap-2 text-lg font-semibold text-gray-100"><ServerCog size={18} /> Runtime details</h3>
            <dl className="mt-4 grid gap-3 md:grid-cols-2">
              {[
                ['OS', status.data?.osName ?? '—'],
                ['Arch', status.data?.osVersion ?? '—'],
                ['SQLite', status.data?.sqliteVersion ?? '—'],
                ['App Data', status.data?.appData ?? '—'],
                ['Auth', status.data?.authentication ?? '—'],
                ['Base URL', status.data?.urlBase || '/'],
              ].map(([label, value]) => (
                <div key={label} className="panel-muted p-4">
                  <dt className="text-xs uppercase tracking-[0.2em] text-gray-500">{label}</dt>
                  <dd className="mt-2 text-sm text-gray-200">{value}</dd>
                </div>
              ))}
            </dl>
          </div>

          <div className="panel p-5">
            <h3 className="flex items-center gap-2 text-lg font-semibold text-gray-100"><HardDrive size={18} /> Disk space</h3>
            <div className="mt-4 space-y-4">
              {disks.data?.map((disk) => {
                const used = disk.totalSpace - disk.freeSpace
                const percent = disk.totalSpace ? Math.round((used / disk.totalSpace) * 100) : 0
                return (
                  <div key={disk.path} className="panel-muted p-4">
                    <div className="mb-2 flex items-center justify-between gap-3 text-sm">
                      <span className="font-medium text-gray-100">{disk.label}</span>
                      <span className="text-gray-500">{formatBytes(disk.freeSpace)} free</span>
                    </div>
                    <div className="h-2 rounded-full bg-gray-800">
                      <div className="h-2 rounded-full bg-yellow-400" style={{ width: `${percent}%` }} />
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        <aside className="panel p-5">
          <h3 className="text-lg font-semibold text-gray-100">Live event feed</h3>
          <p className="mt-1 text-sm text-gray-500">Scheduler heartbeats, file imports, and movie mutations.</p>
          <div className="mt-4 max-h-[640px] space-y-2 overflow-y-auto">
            {events.length === 0 ? <div className="panel-muted p-4 text-sm text-gray-500">No events received yet.</div> : null}
            {events.map((event, index) => (
              <div key={`${event.receivedAt}-${index}`} className="panel-muted p-3 font-mono text-xs text-gray-300">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-yellow-400">{event.eventType}</span>
                  <span className="text-gray-500">{new Date(event.receivedAt).toLocaleTimeString()}</span>
                </div>
                <pre className="mt-2 overflow-x-auto whitespace-pre-wrap text-gray-400">{JSON.stringify(event.data, null, 2)}</pre>
              </div>
            ))}
          </div>
        </aside>
      </section>
    </div>
  )
}
