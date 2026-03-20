import { CheckCircle2, CircleDot, Eye, EyeOff, Sparkles } from 'lucide-react'
import { clsx } from 'clsx'
import type { Movie } from '../lib/types'

export default function MovieCard({ movie, action }: { movie: Movie; action?: React.ReactNode }) {
  const badge = movie.hasFile
    ? { label: 'Available', className: 'bg-green-500/15 text-green-300', icon: CheckCircle2 }
    : movie.monitored
      ? { label: 'Monitored', className: 'bg-blue-500/15 text-blue-300', icon: CircleDot }
      : { label: 'Paused', className: 'bg-gray-700 text-gray-300', icon: EyeOff }
  const BadgeIcon = badge.icon

  return (
    <article className="panel group overflow-hidden transition hover:border-yellow-400/40 hover:shadow-lg hover:shadow-yellow-400/5">
      <div className="relative aspect-[2/3] bg-gray-800">
        {movie.remotePoster ? (
          <img src={movie.remotePoster} alt={movie.title} className="h-full w-full object-cover transition duration-300 group-hover:scale-[1.02]" />
        ) : (
          <div className="flex h-full w-full items-center justify-center text-4xl font-semibold text-gray-700">{movie.title[0]}</div>
        )}
        <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/90 to-transparent px-4 pb-4 pt-10">
          <div className="flex items-center justify-between gap-3">
            <span className="rounded-full bg-black/50 px-2 py-1 text-xs text-gray-200">{movie.year || 'TBD'}</span>
            <div className="flex items-center gap-2 text-xs text-gray-300">
              {movie.monitored ? <Eye size={13} className="text-yellow-400" /> : <EyeOff size={13} />}
              {movie.ratings?.value ? <span className="flex items-center gap-1"><Sparkles size={12} className="text-yellow-400" />{movie.ratings.value.toFixed(1)}</span> : null}
            </div>
          </div>
        </div>
      </div>
      <div className="space-y-3 p-4">
        <div>
          <h3 className="line-clamp-1 text-sm font-semibold text-gray-100">{movie.title}</h3>
          <p className="mt-1 line-clamp-2 text-xs text-gray-500">{movie.overview || 'Metadata will populate after the first refresh.'}</p>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span className={clsx('inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium', badge.className)}>
            <BadgeIcon size={12} />
            {badge.label}
          </span>
          <span className="text-xs text-gray-500">Q{movie.qualityProfileId}</span>
        </div>
        {action ? <div>{action}</div> : null}
      </div>
    </article>
  )
}
