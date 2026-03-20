import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { LoaderCircle, RefreshCw, Trash2 } from 'lucide-react'
import api from '../lib/api'
import { formatBytes } from '../lib/format'
import type { QueueResponse, QueueStatus } from '../lib/types'

function progressPercent(size: number, sizeLeft: number) {
  if (!size) return 0
  return Math.max(0, Math.min(100, Math.round(((size - sizeLeft) / size) * 100)))
}

export default function Queue() {
  const queryClient = useQueryClient()
  const queue = useQuery<QueueResponse>({
    queryKey: ['queue'],
    queryFn: () => api.get('/queue?page=1&pageSize=50').then((res) => res.data),
  })
  const status = useQuery<QueueStatus>({
    queryKey: ['queue-status'],
    queryFn: () => api.get('/queue/status').then((res) => res.data),
  })

  const remove = useMutation({
    mutationFn: (id: number) => api.delete(`/queue/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['queue'] })
      queryClient.invalidateQueries({ queryKey: ['queue-status'] })
    },
  })

  const retry = useMutation({
    mutationFn: (id: number) => api.put(`/queue/${id}/grab`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue'] }),
  })

  return (
    <div className="space-y-6">
      <section className="grid gap-4 md:grid-cols-3">
        <div className="panel p-5"><p className="text-sm text-gray-400">Total queue items</p><p className="stat-value mt-2">{status.data?.count ?? 0}</p></div>
        <div className="panel p-5"><p className="text-sm text-gray-400">Warnings</p><p className="stat-value mt-2">{status.data?.warnings ? 'Yes' : 'No'}</p></div>
        <div className="panel p-5"><p className="text-sm text-gray-400">Errors</p><p className="stat-value mt-2">{status.data?.errors ? 'Yes' : 'No'}</p></div>
      </section>

      <section className="space-y-3">
        {queue.isLoading ? <div className="panel p-10 text-center text-gray-400">Loading queue…</div> : null}
        {(queue.data?.records ?? []).map((item) => {
          const percent = progressPercent(item.size, item.sizeleft)
          return (
            <article key={item.id} className="panel p-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <p className="font-semibold text-gray-100">{item.title}</p>
                  <p className="mt-1 text-sm text-gray-500">{item.downloadClient || 'Unknown client'} · {item.protocol || 'mixed'} · {item.status || 'queued'}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button className="btn-secondary px-3 py-2" onClick={() => retry.mutate(item.id)}>
                    {retry.isPending ? <LoaderCircle size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                    Retry
                  </button>
                  <button className="btn-secondary px-3 py-2" onClick={() => remove.mutate(item.id)}>
                    {remove.isPending ? <LoaderCircle size={16} className="animate-spin" /> : <Trash2 size={16} />}
                    Remove
                  </button>
                </div>
              </div>
              <div className="mt-4">
                <div className="mb-2 flex justify-between text-xs text-gray-500">
                  <span>{formatBytes(item.size - item.sizeleft)} downloaded</span>
                  <span>{formatBytes(item.size)} total</span>
                </div>
                <div className="h-2 rounded-full bg-gray-800">
                  <div className="h-2 rounded-full bg-yellow-400" style={{ width: `${percent}%` }} />
                </div>
              </div>
            </article>
          )
        })}
        {!queue.isLoading && !(queue.data?.records.length) ? <div className="panel p-10 text-center text-gray-500">Queue is clear.</div> : null}
      </section>
    </div>
  )
}
