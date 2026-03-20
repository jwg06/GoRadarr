import { create } from 'zustand'
import type { FeedEvent } from '../lib/types'

interface LiveFeedState {
  connected: boolean
  events: FeedEvent[]
  setConnected: (connected: boolean) => void
  pushEvent: (event: FeedEvent) => void
}

export const useLiveFeedStore = create<LiveFeedState>((set) => ({
  connected: false,
  events: [],
  setConnected: (connected) => set({ connected }),
  pushEvent: (event) => set((state) => ({
    events: [event, ...state.events].slice(0, 50),
  })),
}))
