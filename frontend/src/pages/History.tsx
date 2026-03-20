import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search } from 'lucide-react'
import api from '../lib/api'
import type { HistoryResponse } from '../lib/types'

const eventFilters = ['all', 'grabbed', 'downloadFolderImported', 'downloadFailed'] as const

export default function History() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [eventType, setEventType] = useState<(typeof eventFilters)[number]>('all')

  const history = useQuery<HistoryResponse>({
    queryKey: ['history', page],
    queryFn: () => api.get(`/history?page=${page}&pageSize=25&sortKey=date&sortDirection=descending`).then((res) => res.data),
  })

  const filtered = useMemo(() => {
    return (history.data?.records ?? []).filter((record) => {
      const matchesSearch = !search || [record.sourceTitle, record.movie?.title ?? ''].join(' ').toLowerCase().includes(search.toLowerCase())
      const matchesEvent = eventType === 'all' || record.eventType === eventType
      return matchesSearch && matchesEvent
    })
  }, [eventType, history.data?.records, search])

  return (
    <div className="space-y-4">
      <section className="panel p-4 md:p-5">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Filter source title or movie" className="field pl-9" />
          </div>
          <div className="flex flex-wrap gap-2">
            {eventFilters.map((filter) => (
              <button key={filter} className={eventType === filter ? 'btn-primary px-3 py-2' : 'btn-secondary px-3 py-2'} onClick={() => setEventType(filter)}>
                {filter === 'all' ? 'All events' : filter}
              </button>
            ))}
          </div>
        </div>
      </section>

      <section className="panel overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full text-sm">
            <thead className="border-b border-gray-800 text-left text-xs uppercase tracking-[0.2em] text-gray-500">
              <tr>
                <th className="px-4 py-3">Movie</th>
                <th className="px-4 py-3">Source</th>
                <th className="px-4 py-3">Event</th>
                <th className="px-4 py-3">Date</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((record) => (
                <tr key={record.id} className="border-b border-gray-800/60 last:border-b-0">
                  <td className="px-4 py-3 text-gray-200">{record.movie?.title ?? `Movie #${record.movieId}`}</td>
                  <td className="px-4 py-3 text-gray-400">{record.sourceTitle}</td>
                  <td className="px-4 py-3"><span className="rounded-full bg-gray-800 px-2 py-1 text-xs text-gray-200">{record.eventType}</span></td>
                  <td className="px-4 py-3 text-gray-500">{new Date(record.date).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!filtered.length && !history.isLoading ? <div className="p-10 text-center text-gray-500">No matching history records.</div> : null}
      </section>

      <div className="flex items-center justify-between text-sm text-gray-400">
        <span>Page {history.data?.page ?? page}</span>
        <div className="flex gap-2">
          <button className="btn-secondary px-3 py-2" disabled={page <= 1} onClick={() => setPage((value) => Math.max(1, value - 1))}>Previous</button>
          <button className="btn-secondary px-3 py-2" disabled={page >= Math.ceil((history.data?.totalRecords ?? 0) / 25)} onClick={() => setPage((value) => value + 1)}>Next</button>
        </div>
      </div>
    </div>
  )
}
