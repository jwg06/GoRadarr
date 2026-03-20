import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, SlidersHorizontal } from 'lucide-react'
import MovieCard from '../components/MovieCard'
import api from '../lib/api'
import type { Movie, QualityProfile } from '../lib/types'

export default function Movies() {
  const [search, setSearch] = useState('')
  const [stateFilter, setStateFilter] = useState<'all' | 'monitored' | 'unmonitored' | 'downloaded' | 'missing'>('all')
  const [profileFilter, setProfileFilter] = useState('all')

  const { data: movies = [], isLoading, error } = useQuery<Movie[]>({
    queryKey: ['movies'],
    queryFn: () => api.get('/movie').then((res) => res.data),
  })

  const { data: profiles = [] } = useQuery<QualityProfile[]>({
    queryKey: ['quality-profiles'],
    queryFn: () => api.get('/qualityprofile').then((res) => res.data),
  })

  const filtered = useMemo(() => {
    return movies.filter((movie) => {
      const matchesSearch = !search || [movie.title, movie.overview, String(movie.year)]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(search.toLowerCase())

      const matchesState = (() => {
        switch (stateFilter) {
          case 'monitored': return movie.monitored
          case 'unmonitored': return !movie.monitored
          case 'downloaded': return movie.hasFile
          case 'missing': return movie.monitored && !movie.hasFile
          default: return true
        }
      })()

      const matchesProfile = profileFilter === 'all' || String(movie.qualityProfileId) === profileFilter
      return matchesSearch && matchesState && matchesProfile
    })
  }, [movies, profileFilter, search, stateFilter])

  const stats = {
    total: movies.length,
    monitored: movies.filter((movie) => movie.monitored).length,
    available: movies.filter((movie) => movie.hasFile).length,
    missing: movies.filter((movie) => movie.monitored && !movie.hasFile).length,
  }

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-4">
        {[
          ['Total Movies', stats.total],
          ['Monitored', stats.monitored],
          ['Available', stats.available],
          ['Wanted', stats.missing],
        ].map(([label, value]) => (
          <div key={label} className="panel p-5">
            <p className="text-sm text-gray-400">{label}</p>
            <p className="stat-value mt-2">{value}</p>
          </div>
        ))}
      </section>

      <section className="panel p-4 md:p-5">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search title, year, or overview"
              className="field pl-9"
            />
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="inline-flex items-center gap-2 px-2 text-sm text-gray-500">
              <SlidersHorizontal size={14} />
              Filters
            </span>
            {(['all', 'monitored', 'unmonitored', 'downloaded', 'missing'] as const).map((value) => (
              <button
                key={value}
                onClick={() => setStateFilter(value)}
                className={stateFilter === value ? 'btn-primary px-3 py-2' : 'btn-secondary px-3 py-2'}
              >
                {value[0].toUpperCase() + value.slice(1)}
              </button>
            ))}
            <select className="field min-w-44" value={profileFilter} onChange={(event) => setProfileFilter(event.target.value)}>
              <option value="all">All quality profiles</option>
              {profiles.map((profile) => (
                <option key={profile.id} value={profile.id}>{profile.name}</option>
              ))}
            </select>
          </div>
        </div>
      </section>

      {error ? <div className="panel p-6 text-sm text-red-300">{(error as Error).message}</div> : null}

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 8 }).map((_, index) => <div key={index} className="panel aspect-[0.72] animate-pulse bg-gray-900/40" />)}
        </div>
      ) : null}

      {!isLoading && filtered.length === 0 ? (
        <div className="panel p-10 text-center text-gray-400">No movies match the current filters.</div>
      ) : null}

      {!isLoading && filtered.length > 0 ? (
        <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {filtered.map((movie) => <MovieCard key={movie.id || movie.tmdbId} movie={movie} />)}
        </section>
      ) : null}
    </div>
  )
}
