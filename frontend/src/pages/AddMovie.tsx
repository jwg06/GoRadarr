import { useState } from 'react'
import type { FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { LoaderCircle, Plus, Search } from 'lucide-react'
import MovieCard from '../components/MovieCard'
import api, { apiErrorMessage } from '../lib/api'
import type { Movie } from '../lib/types'

export default function AddMovie() {
  const [term, setTerm] = useState('')
  const [submitted, setSubmitted] = useState('')
  const queryClient = useQueryClient()

  const lookup = useQuery<Movie[]>({
    queryKey: ['lookup', submitted],
    enabled: submitted.length > 0,
    queryFn: () => api.get(`/movie/lookup?term=${encodeURIComponent(submitted)}`).then((res) => res.data),
  })

  const addMovie = useMutation({
    mutationFn: (movie: Movie) => api.post('/movie', {
      title: movie.title,
      sortTitle: movie.sortTitle || movie.title,
      tmdbId: movie.tmdbId,
      overview: movie.overview,
      status: movie.status || 'announced',
      year: movie.year,
      inCinemas: movie.inCinemas,
      digitalRelease: movie.digitalRelease,
      runtime: movie.runtime,
      studio: movie.studio,
      monitored: true,
      minimumAvailability: 'released',
      qualityProfileId: 1,
      rootFolderPath: '/movies',
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['movies'] })
      queryClient.invalidateQueries({ queryKey: ['wanted'] })
    },
  })

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitted(term.trim())
  }

  return (
    <div className="space-y-6">
      <section className="panel p-5">
        <form onSubmit={handleSubmit} className="flex flex-col gap-3 md:flex-row">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input
              autoFocus
              value={term}
              onChange={(event) => setTerm(event.target.value)}
              placeholder="Search TMDB by title"
              className="field pl-9"
            />
          </div>
          <button type="submit" className="btn-primary" disabled={!term.trim() || lookup.isFetching}>
            {lookup.isFetching ? <LoaderCircle size={16} className="animate-spin" /> : <Search size={16} />}
            Search
          </button>
        </form>
        <p className="mt-3 text-sm text-gray-500">Set `GORADARR_METADATA_TMDB_API_KEY` to enable real TMDB search results.</p>
      </section>

      {lookup.error ? <div className="panel p-5 text-sm text-red-300">{apiErrorMessage(lookup.error)}</div> : null}

      {lookup.isLoading ? <div className="panel p-10 text-center text-gray-400">Searching TMDB…</div> : null}

      {lookup.data?.length ? (
        <section className="grid gap-4 lg:grid-cols-2">
          {lookup.data.map((movie) => (
            <MovieCard
              key={movie.tmdbId}
              movie={movie}
              action={
                <button className="btn-primary w-full" onClick={() => addMovie.mutate(movie)} disabled={addMovie.isPending}>
                  {addMovie.isPending ? <LoaderCircle size={16} className="animate-spin" /> : <Plus size={16} />}
                  Add to library
                </button>
              }
            />
          ))}
        </section>
      ) : null}

      {submitted && !lookup.isLoading && !lookup.error && lookup.data?.length === 0 ? (
        <div className="panel p-10 text-center text-gray-400">No movies found for “{submitted}”.</div>
      ) : null}
    </div>
  )
}
