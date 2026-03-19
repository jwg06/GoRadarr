import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, SlidersHorizontal } from 'lucide-react';
import api from '../lib/api';
import type { Movie } from '../lib/types';
import MovieCard from '../components/MovieCard';

type Filter = 'all' | 'monitored' | 'unmonitored' | 'downloaded' | 'missing';

export default function Movies() {
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<Filter>('all');

  const { data: movies = [], isLoading, error } = useQuery<Movie[]>({
    queryKey: ['movies'],
    queryFn: () => api.get('/movie').then((r) => r.data),
  });

  const filtered = useMemo(() => {
    let list = movies;
    if (search) {
      const q = search.toLowerCase();
      list = list.filter((m) => m.title.toLowerCase().includes(q) || String(m.year).includes(q));
    }
    switch (filter) {
      case 'monitored': return list.filter((m) => m.monitored);
      case 'unmonitored': return list.filter((m) => !m.monitored);
      case 'downloaded': return list.filter((m) => m.hasFile);
      case 'missing': return list.filter((m) => m.monitored && !m.hasFile);
      default: return list;
    }
  }, [movies, search, filter]);

  const filterButtons: { value: Filter; label: string }[] = [
    { value: 'all', label: 'All' },
    { value: 'monitored', label: 'Monitored' },
    { value: 'unmonitored', label: 'Unmonitored' },
    { value: 'downloaded', label: 'Downloaded' },
    { value: 'missing', label: 'Missing' },
  ];

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">Movies</h1>
          <p className="text-sm text-gray-400 mt-0.5">{movies.length} total</p>
        </div>
      </div>

      {/* Filter bar */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        {/* Search */}
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search movies..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-9 pr-4 py-2.5 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-yellow-400/50 transition-colors"
          />
        </div>

        {/* Filter pills */}
        <div className="flex items-center gap-1.5 flex-wrap">
          <SlidersHorizontal size={14} className="text-gray-500 mr-1" />
          {filterButtons.map((btn) => (
            <button
              key={btn.value}
              onClick={() => setFilter(btn.value)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                filter === btn.value
                  ? 'bg-yellow-400 text-gray-900'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200'
              }`}
            >
              {btn.label}
            </button>
          ))}
        </div>
      </div>

      {/* Content */}
      {isLoading && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="bg-gray-900 rounded-lg overflow-hidden border border-gray-800 animate-pulse">
              <div className="aspect-[2/3] bg-gray-800" />
              <div className="p-2 space-y-2">
                <div className="h-3 bg-gray-800 rounded w-3/4" />
                <div className="h-3 bg-gray-800 rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      )}

      {error && (
        <div className="text-center py-20 text-red-400">
          <p className="text-lg font-medium">Failed to load movies</p>
          <p className="text-sm mt-1 text-gray-500">Check that the backend is running on port 7878</p>
        </div>
      )}

      {!isLoading && !error && filtered.length === 0 && (
        <div className="text-center py-20 text-gray-500">
          <p className="text-lg">No movies found</p>
          <p className="text-sm mt-1">Try adjusting your filters or adding movies</p>
        </div>
      )}

      {!isLoading && !error && filtered.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {filtered.map((movie) => (
            <MovieCard key={movie.id} movie={movie} />
          ))}
        </div>
      )}
    </div>
  );
}
