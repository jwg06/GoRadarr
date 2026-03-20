import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import api from '../lib/api'
import { formatShortDate } from '../lib/format'
import type { CalendarMovie } from '../lib/types'

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

function dateKey(date: Date) {
  return date.toISOString().split('T')[0]
}

export default function CalendarPage() {
  const [current, setCurrent] = useState(new Date())
  const [selected, setSelected] = useState<string>('')

  const start = useMemo(() => new Date(current.getFullYear(), current.getMonth(), 1), [current])
  const end = useMemo(() => new Date(current.getFullYear(), current.getMonth() + 1, 0), [current])

  const { data: movies = [] } = useQuery<CalendarMovie[]>({
    queryKey: ['calendar', current.getFullYear(), current.getMonth()],
    queryFn: () => api.get(`/calendar?start=${dateKey(start)}&end=${dateKey(end)}`).then((res) => res.data),
  })

  const moviesByDay = useMemo(() => {
    const map = new Map<string, CalendarMovie[]>()
    for (const movie of movies) {
      for (const date of [movie.inCinemas, movie.physicalRelease, movie.digitalRelease]) {
        if (!date) continue
        const key = date.split('T')[0]
        map.set(key, [...(map.get(key) ?? []), movie])
      }
    }
    return map
  }, [movies])

  const cells = useMemo(() => {
    const blanks = Array.from({ length: start.getDay() }, () => null)
    const monthDays = Array.from({ length: end.getDate() }, (_, index) => index + 1)
    const all = [...blanks, ...monthDays]
    while (all.length % 7 !== 0) all.push(null)
    return all
  }, [end, start])

  const selectedMovies = selected ? moviesByDay.get(selected) ?? [] : []

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
      <section className="panel overflow-hidden">
        <div className="flex items-center justify-between border-b border-gray-800 px-5 py-4">
          <h2 className="text-lg font-semibold text-gray-100">
            {current.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })}
          </h2>
          <div className="flex gap-2">
            <button className="btn-secondary px-3 py-2" onClick={() => setCurrent(new Date(current.getFullYear(), current.getMonth() - 1, 1))}><ChevronLeft size={16} /></button>
            <button className="btn-secondary px-3 py-2" onClick={() => setCurrent(new Date(current.getFullYear(), current.getMonth() + 1, 1))}><ChevronRight size={16} /></button>
          </div>
        </div>
        <div className="grid grid-cols-7 border-b border-gray-800 text-center text-xs uppercase tracking-[0.2em] text-gray-500">
          {days.map((day) => <div key={day} className="py-3">{day}</div>)}
        </div>
        <div className="grid grid-cols-7">
          {cells.map((day, index) => {
            const key = day ? `${current.getFullYear()}-${String(current.getMonth() + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}` : ''
            const entries = key ? moviesByDay.get(key) ?? [] : []
            return (
              <button
                key={`${key}-${index}`}
                disabled={!day}
                onClick={() => day && setSelected(key)}
                className={`min-h-28 border-b border-r border-gray-800 px-2 py-2 text-left transition ${day ? 'hover:bg-gray-800/40' : 'bg-gray-950/40'} ${selected === key ? 'bg-yellow-400/8' : ''}`}
              >
                {day ? <div className="text-xs text-gray-400">{day}</div> : null}
                <div className="mt-2 space-y-1">
                  {entries.slice(0, 3).map((movie) => (
                    <div key={`${key}-${movie.id}-${movie.tmdbId}`} className="truncate rounded-lg bg-yellow-400/10 px-2 py-1 text-xs text-yellow-200">
                      {movie.title}
                    </div>
                  ))}
                  {entries.length > 3 ? <div className="px-2 text-xs text-gray-500">+{entries.length - 3} more</div> : null}
                </div>
              </button>
            )
          })}
        </div>
      </section>

      <aside className="panel p-5">
        <h3 className="text-lg font-semibold text-gray-100">{selected ? formatShortDate(selected) : 'Select a day'}</h3>
        <p className="mt-1 text-sm text-gray-500">Release details for theatrical, digital, and physical availability.</p>
        <div className="mt-4 space-y-3">
          {selectedMovies.length === 0 ? <div className="panel-muted p-4 text-sm text-gray-500">No releases scheduled for the selected day.</div> : null}
          {selectedMovies.map((movie) => (
            <div key={`${selected}-${movie.id}-${movie.tmdbId}`} className="panel-muted p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-medium text-gray-100">{movie.title}</p>
                  <p className="text-sm text-gray-500">{movie.year || 'TBD'} · {movie.status}</p>
                </div>
                <span className="rounded-full bg-gray-800 px-2 py-1 text-xs text-gray-300">{movie.hasFile ? 'In library' : 'Upcoming'}</span>
              </div>
              <div className="mt-3 flex flex-wrap gap-2 text-xs text-gray-300">
                {movie.inCinemas ? <span className="rounded-full bg-blue-500/15 px-2 py-1">In Cinemas {formatShortDate(movie.inCinemas)}</span> : null}
                {movie.digitalRelease ? <span className="rounded-full bg-purple-500/15 px-2 py-1">Digital {formatShortDate(movie.digitalRelease)}</span> : null}
                {movie.physicalRelease ? <span className="rounded-full bg-green-500/15 px-2 py-1">Physical {formatShortDate(movie.physicalRelease)}</span> : null}
              </div>
            </div>
          ))}
        </div>
      </aside>
    </div>
  )
}
