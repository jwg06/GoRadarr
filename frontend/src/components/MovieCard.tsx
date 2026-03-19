import { clsx } from 'clsx';
import { CheckCircle, Download, XCircle, Eye, EyeOff } from 'lucide-react';
import type { Movie } from '../lib/types';

interface Props {
  movie: Movie;
  onClick?: () => void;
}

function StatusBadge({ movie }: { movie: Movie }) {
  if (movie.hasFile) {
    return (
      <span className="flex items-center gap-1 bg-green-500/20 text-green-400 text-xs px-2 py-0.5 rounded-full">
        <CheckCircle size={10} /> Downloaded
      </span>
    );
  }
  if (movie.monitored) {
    return (
      <span className="flex items-center gap-1 bg-blue-500/20 text-blue-400 text-xs px-2 py-0.5 rounded-full">
        <Download size={10} /> Monitored
      </span>
    );
  }
  return (
    <span className="flex items-center gap-1 bg-gray-600/40 text-gray-400 text-xs px-2 py-0.5 rounded-full">
      <XCircle size={10} /> Unmonitored
    </span>
  );
}

export default function MovieCard({ movie, onClick }: Props) {
  return (
    <div
      onClick={onClick}
      className={clsx(
        'group bg-gray-900 rounded-lg overflow-hidden border border-gray-800',
        'hover:border-yellow-400/50 hover:shadow-lg hover:shadow-yellow-400/5',
        'transition-all duration-200 cursor-pointer'
      )}
    >
      {/* Poster */}
      <div className="relative aspect-[2/3] bg-gray-800 overflow-hidden">
        {movie.remotePoster ? (
          <img
            src={movie.remotePoster}
            alt={movie.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <span className="text-gray-600 text-4xl font-bold">
              {movie.title.charAt(0)}
            </span>
          </div>
        )}

        {/* Monitored indicator */}
        <div className="absolute top-2 right-2">
          {movie.monitored ? (
            <Eye size={14} className="text-yellow-400 drop-shadow" />
          ) : (
            <EyeOff size={14} className="text-gray-500" />
          )}
        </div>

        {/* Year overlay */}
        <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/80 to-transparent px-2 py-1">
          <span className="text-xs text-gray-300">{movie.year}</span>
        </div>
      </div>

      {/* Info */}
      <div className="p-2 space-y-1.5">
        <h3 className="text-sm font-medium text-gray-100 truncate leading-tight" title={movie.title}>
          {movie.title}
        </h3>
        <StatusBadge movie={movie} />
        {movie.genres && movie.genres.length > 0 && (
          <p className="text-xs text-gray-500 truncate">{movie.genres.slice(0, 2).join(', ')}</p>
        )}
      </div>
    </div>
  );
}
