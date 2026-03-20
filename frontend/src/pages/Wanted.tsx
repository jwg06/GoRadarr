import { useMemo, useState } from 'react'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Search } from 'lucide-react'
import MovieCard from '../components/MovieCard'
import api from '../lib/api'
import { formatShortDate } from '../lib/format'
import type { Movie } from '../lib/types'

export default function Wanted() {
  const [search, setSearch] = useState('')
  const movies = useQuery<Movie[]>({
    queryKey: ['wanted'],
    queryFn: () => api.get('/movie').then((res) => res.data),
  })

  const searchMovie = useMutation({
    mutationFn: (id: number) => api.post(`/movie/${id}/command`, { name: 'MoviesSearch' }),
  })

  const wanted = useMemo(() => (movies.data ?? [])
    .filter((movie) => movie.monitored && !movie.hasFile)
    .filter((movie) => !search || movie.title.toLowerCase().includes(search.toLowerCase())), [movies.data, search])

  return (
    <div className="space-y-6">
      <section className="panel p-5">
        <div className="relative max-w-xl">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
          <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search missing movies" className="field pl-9" />
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
        {wanted.map((movie) => (
          <MovieCard
            key={movie.id}
            movie={movie}
            action={
              <div className="flex items-center justify-between gap-3 text-xs text-gray-400">
                <span>{movie.digitalRelease ? `Digital ${formatShortDate(movie.digitalRelease)}` : 'No release date yet'}</span>
                <button className="btn-primary px-3 py-2" onClick={() => searchMovie.mutate(movie.id)}>Search</button>
              </div>
            }
          />
        ))}
      </section>

      {!movies.isLoading && wanted.length === 0 ? <div className="panel p-10 text-center text-gray-500">No wanted movies right now.</div> : null}
    </div>
  )
}
