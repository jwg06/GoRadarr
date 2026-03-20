import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useLiveFeedStore } from '../stores/liveFeed'
import type { FeedEvent } from '../lib/types'

const invalidateByEvent: Record<string, Array<string[]>> = {
  'movie.added': [['movies'], ['wanted']],
  'movie.updated': [['movies'], ['wanted'], ['calendar']],
  'movie.deleted': [['movies'], ['wanted']],
  'download.imported': [['movies'], ['queue'], ['wanted']],
  'queue.changed': [['queue'], ['queue-status']],
  'task.completed': [['system-status']],
  'health.changed': [['system-status']],
}

export function useLiveFeed() {
  const queryClient = useQueryClient()
  const setConnected = useLiveFeedStore((state) => state.setConnected)
  const pushEvent = useLiveFeedStore((state) => state.pushEvent)

  useEffect(() => {
    const eventSource = new EventSource('/api/v1/feed')

    eventSource.onopen = () => setConnected(true)
    eventSource.onerror = () => setConnected(false)
    eventSource.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as { eventType: string; data: unknown }
        const entry: FeedEvent = {
          eventType: parsed.eventType,
          data: parsed.data,
          receivedAt: new Date().toISOString(),
        }
        pushEvent(entry)
        for (const queryKey of invalidateByEvent[parsed.eventType] ?? []) {
          queryClient.invalidateQueries({ queryKey })
        }
      } catch {
        // Ignore malformed events.
      }
    }

    return () => eventSource.close()
  }, [pushEvent, queryClient, setConnected])
}
