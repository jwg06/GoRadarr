import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Search, Plus, Loader2, Film } from 'lucide-react';
import api from '../lib/api';
import type { Movie } from '../lib/types';

export default function AddMovie() {
  const [term, setTerm] = useState('');
  const [submitted, setSubmitted] = useState('');
  const queryClient = useQueryClient();

  const { data: results = [], isLoading, isFetching } = useQuery<Movie[]>({
    queryKey: ['lookup', submitted],
    queryFn: () => api.get(`/movie/lookup?term=${encodeURIComponent(submitted)}`).then((r) => r.data),
    enabled: submitted.length > 0,
  });

  const addMutation = useMutation({
    mutationFn: (movie: Partial<Movie> & { tmdbId: number }) =>
      api.post('/movie', {
        ...movie,
        monitored: true,
        minimumAvailability: 'released',
        qualityProfileId: 1,
        rootFolderPath: '/movies',
        addOptions: { searchForMovie: true },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['movies'] });
    },
  });

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (term.trim()) setSubmitted(term.trim());
  }

  return (
    <div className="p-4 md:p-6 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Add Movie</h1>
        <p className="text-sm text-gray-400 mt-0.5">Search TMDb to find and add movies</p>
      </div>

      {/* Search form */}
      <form onSubmit={handleSearch} className="flex gap-3 mb-8">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search for a movie..."
            value={term}
            onChange={(e) => setTerm(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-9 pr-4 py-2.5 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-yellow-400/50 transition-colors"
            autoFocus
          />
        </div>
        <button
          type="submit"
          disabled={!term.trim()}
          className="bg-yellow-400 hover:bg-yellow-300 disabled:opacity-50 disabled:cursor-not-allowed text-gray-900 font-medium px-5 py-2.5 rounded-lg text-sm transition-colors"
        >
          Search
        </button>
      </form>

      {/* Loading */}
      {(isLoading || isFetching) && (
        <div className="flex items-center justify-center py-20 gap-3 text-gray-400">
          <Loader2 size={20} className="animate-spin" />
          <span>Searching…</span>
        </div>
      )}

      {/* Results */}
      {!isLoading && !isFetching && results.length > 0 && (
        <div className="space-y-3">
          {results.map((movie) => (
            <div
              key={movie.tmdbId}
              className="flex gap-4 bg-gray-900 border border-gray-800 rounded-lg p-4 hover:border-gray-700 transition-colors"
            >
              {/* Poster */}
              <div className="w-14 h-20 bg-gray-800 rounded flex-shrink-0 overflow-hidden">
                {movie.remotePoster ? (
                  <img src={movie.remotePoster} alt={movie.title} className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Film size={20} className="text-gray-600" />
                  </div>
                )}
              </div>

              {/* Info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <h3 className="font-semibold text-gray-100">
                      {movie.title} <span className="text-gray-400 font-normal">({movie.year})</span>
                    </h3>
                    {movie.studio && <p className="text-xs text-gray-500 mt-0.5">{movie.studio}</p>}
                  </div>
                  <button
                    onClick={() => addMutation.mutate({ ...movie })}
                    disabled={addMutation.isPending}
                    className="flex items-center gap-1.5 bg-yellow-400 hover:bg-yellow-300 disabled:opacity-60 text-gray-900 text-sm font-medium px-3 py-1.5 rounded-lg transition-colors shrink-0"
                  >
                    {addMutation.isPending ? (
                      <Loader2 size={14} className="animate-spin" />
                    ) : (
                      <Plus size={14} />
                    )}
                    Add
                  </button>
                </div>
                {movie.overview && (
                  <p className="text-xs text-gray-400 mt-2 line-clamp-2">{movie.overview}</p>
                )}
                {movie.genres && movie.genres.length > 0 && (
                  <div className="flex gap-1.5 mt-2 flex-wrap">
                    {movie.genres.slice(0, 3).map((g) => (
                      <span key={g} className="text-xs bg-gray-800 text-gray-400 px-2 py-0.5 rounded">
                        {g}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {!isLoading && !isFetching && submitted && results.length === 0 && (
        <div className="text-center py-20 text-gray-500">
          <Film size={40} className="mx-auto mb-3 opacity-30" />
          <p>No results for "{submitted}"</p>
        </div>
      )}

      {addMutation.isSuccess && (
        <div className="fixed bottom-6 right-6 bg-green-500 text-white px-4 py-3 rounded-lg shadow-lg text-sm font-medium">
          Movie added successfully!
        </div>
      )}
      {addMutation.isError && (
        <div className="fixed bottom-6 right-6 bg-red-500 text-white px-4 py-3 rounded-lg shadow-lg text-sm font-medium">
          Failed to add movie
        </div>
      )}
    </div>
  );
}
