import { useMemo, useEffect, useRef, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import api from '../lib/api'

type LogLevel = 'ALL' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'

interface LogEntry {
  time: string
  level: string
  msg: string
  attrs?: Record<string, unknown>
}

interface HistoryRecord {
  id: number
  movieId: number
  sourceTitle: string
  quality: unknown
  date: string
  eventType: string
  downloadId?: string
}

const LEVEL_PILLS: LogLevel[] = ['ALL', 'DEBUG', 'INFO', 'WARN', 'ERROR']

const levelBadgeClass: Record<string, string> = {
  DEBUG: 'bg-gray-700 text-gray-300',
  INFO: 'bg-blue-900/60 text-blue-300',
  WARN: 'bg-yellow-900/60 text-yellow-300',
  WARNING: 'bg-yellow-900/60 text-yellow-300',
  ERROR: 'bg-red-900/60 text-red-300',
}

const PAGE_SIZE = 20

export default function Activity() {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'logs' | 'tasks'>('logs')
  const [levelFilter, setLevelFilter] = useState<LogLevel>('ALL')
  const [liveMode, setLiveMode] = useState(false)
  const [liveLogs, setLiveLogs] = useState<LogEntry[]>([])
  const [historyPage, setHistoryPage] = useState(1)
  const bottomRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)

  const { data: allLogs = [], refetch: refetchLogs } = useQuery<LogEntry[]>({
    queryKey: ['logs'],
    queryFn: () => api.get('/log').then((res) => res.data),
  })

  const { data: history = [], refetch: refetchHistory } = useQuery<HistoryRecord[]>({
    queryKey: ['activity-history'],
    queryFn: () => api.get('/history/recent').then((res) => res.data),
    enabled: tab === 'tasks',
  })

  // SSE live log subscription
  useEffect(() => {
    if (!liveMode) {
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }
      return
    }

    const es = new EventSource('/api/v1/log/stream')
    esRef.current = es

    es.onmessage = (event) => {
      try {
        const entry: LogEntry = JSON.parse(event.data)
        setLiveLogs((prev) => [...prev, entry].slice(-200))
      } catch {
        // ignore malformed events
      }
    }

    return () => {
      es.close()
      esRef.current = null
    }
  }, [liveMode])

  // Auto-scroll when new live entries arrive
  useEffect(() => {
    if (liveMode && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [liveLogs, liveMode])

  const displayLogs = liveMode ? liveLogs : allLogs

  const visibleLogs = useMemo(() => {
    if (levelFilter === 'ALL') return displayLogs
    return displayLogs.filter((l) => l.level === levelFilter || l.level === levelFilter + 'ING')
  }, [displayLogs, levelFilter])

  const totalHistoryPages = Math.max(1, Math.ceil(history.length / PAGE_SIZE))
  const pagedHistory = history.slice((historyPage - 1) * PAGE_SIZE, historyPage * PAGE_SIZE)

  function handleLiveModeToggle() {
    if (!liveMode) {
      setLiveLogs([...allLogs])
    }
    setLiveMode((v) => !v)
  }

  return (
    <div className="space-y-4">
      {/* Tabs */}
      <div className="flex border-b border-gray-800">
        {(['logs', 'tasks'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-5 py-3 text-sm font-medium capitalize transition ${
              tab === t
                ? 'border-b-2 border-yellow-400 text-yellow-400'
                : 'text-gray-400 hover:text-gray-100'
            }`}
          >
            {t === 'logs' ? 'Logs' : 'Tasks'}
          </button>
        ))}
      </div>

      {/* ── Logs Tab ── */}
      {tab === 'logs' && (
        <div className="space-y-3">
          {/* Controls */}
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex flex-wrap gap-2" role="group" aria-label="Level filter">
              {LEVEL_PILLS.map((lvl) => (
                <button
                  key={lvl}
                  onClick={() => setLevelFilter(lvl)}
                  className={`rounded-full px-3 py-1 text-xs font-semibold transition ${
                    levelFilter === lvl
                      ? 'bg-yellow-400 text-gray-900'
                      : 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200'
                  }`}
                >
                  {lvl}
                </button>
              ))}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => {
                  refetchLogs()
                  queryClient.invalidateQueries({ queryKey: ['logs'] })
                }}
                className="btn-secondary flex items-center gap-2 px-3 py-2"
              >
                <RefreshCw size={14} />
                Refresh
              </button>
              <button
                onClick={handleLiveModeToggle}
                className={`flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition ${
                  liveMode ? 'bg-yellow-400 text-gray-900' : 'btn-secondary'
                }`}
              >
                <span
                  className={`h-2 w-2 rounded-full ${
                    liveMode ? 'animate-pulse bg-gray-900' : 'bg-gray-500'
                  }`}
                />
                Live
              </button>
            </div>
          </div>

          {/* Log table */}
          <section className="panel overflow-hidden">
            <div className="max-h-[60vh] overflow-auto">
              <table className="min-w-full text-sm">
                <thead className="sticky top-0 z-10 border-b border-gray-800 bg-gray-900 text-left text-xs uppercase tracking-widest text-gray-500">
                  <tr>
                    <th className="px-4 py-3">Time</th>
                    <th className="px-4 py-3">Level</th>
                    <th className="px-4 py-3">Message</th>
                  </tr>
                </thead>
                <tbody>
                  {visibleLogs.map((log, i) => (
                    <tr
                      key={i}
                      className="border-b border-gray-800/60 font-mono text-xs last:border-b-0"
                    >
                      <td className="whitespace-nowrap px-4 py-2 text-gray-500">
                        {log.time ? new Date(log.time).toLocaleTimeString() : '—'}
                      </td>
                      <td className="px-4 py-2">
                        <span
                          className={`rounded px-1.5 py-0.5 text-xs font-medium ${
                            levelBadgeClass[log.level] ?? 'bg-gray-800 text-gray-400'
                          }`}
                        >
                          {log.level}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-gray-200">{log.msg}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {visibleLogs.length === 0 && (
                <p className="p-10 text-center text-gray-500">No log entries.</p>
              )}
              <div ref={bottomRef} />
            </div>
          </section>
        </div>
      )}

      {/* ── Tasks Tab ── */}
      {tab === 'tasks' && (
        <div className="space-y-3">
          <div className="flex justify-end">
            <button
              onClick={() => refetchHistory()}
              className="btn-secondary flex items-center gap-2 px-3 py-2"
            >
              <RefreshCw size={14} />
              Refresh
            </button>
          </div>

          <section className="panel overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead className="border-b border-gray-800 text-left text-xs uppercase tracking-widest text-gray-500">
                  <tr>
                    <th className="px-4 py-3">Date</th>
                    <th className="px-4 py-3">Source</th>
                    <th className="px-4 py-3">Event</th>
                    <th className="px-4 py-3">Download ID</th>
                  </tr>
                </thead>
                <tbody>
                  {pagedHistory.map((rec) => (
                    <tr
                      key={rec.id}
                      className="border-b border-gray-800/60 last:border-b-0"
                    >
                      <td className="px-4 py-3 text-gray-500">
                        {new Date(rec.date).toLocaleString()}
                      </td>
                      <td className="px-4 py-3 text-gray-200">{rec.sourceTitle}</td>
                      <td className="px-4 py-3">
                        <span className="rounded-full bg-gray-800 px-2 py-1 text-xs text-gray-300">
                          {rec.eventType}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-400">{rec.downloadId || '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {pagedHistory.length === 0 && (
                <p className="p-10 text-center text-gray-500">No task history.</p>
              )}
            </div>
          </section>

          {totalHistoryPages > 1 && (
            <div className="flex items-center justify-between text-sm text-gray-400">
              <span>
                Page {historyPage} of {totalHistoryPages}
              </span>
              <div className="flex gap-2">
                <button
                  className="btn-secondary px-3 py-2"
                  disabled={historyPage <= 1}
                  onClick={() => setHistoryPage((p) => Math.max(1, p - 1))}
                >
                  Previous
                </button>
                <button
                  className="btn-secondary px-3 py-2"
                  disabled={historyPage >= totalHistoryPages}
                  onClick={() => setHistoryPage((p) => Math.min(totalHistoryPages, p + 1))}
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
