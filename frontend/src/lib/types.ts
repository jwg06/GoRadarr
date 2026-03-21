export interface Ratings {
  value: number
  votes: number
}

export interface Movie {
  id: number
  title: string
  sortTitle: string
  tmdbId: number
  imdbId?: string
  overview?: string
  status: string
  year: number
  runtime: number
  studio?: string
  qualityProfileId: number
  rootFolderPath?: string
  path?: string
  monitored: boolean
  minimumAvailability: string
  hasFile: boolean
  added: string
  inCinemas?: string
  physicalRelease?: string
  digitalRelease?: string
  remotePoster?: string
  ratings?: Ratings
  genres?: string[]
}

export interface QualityProfile {
  id: number
  name: string
  upgradeAllowed: boolean
  cutoff: number
}

export interface QualityDefinition {
  id: number
  title: string
  minSize: number
  maxSize: number
  preferredSize: number
  quality: { id: number; name: string }
}

export interface HistoryRecord {
  id: number
  movieId: number
  sourceTitle: string
  quality?: { quality?: { name?: string } }
  date: string
  eventType: string
  data?: Record<string, string>
  movie?: Movie
}

export interface HistoryResponse {
  page: number
  pageSize: number
  totalRecords: number
  records: HistoryRecord[]
}

export type CalendarMovie = Movie

export interface SystemStatus {
  version: string
  buildTime: string
  appData: string
  startupPath: string
  osName: string
  osVersion: string
  branch: string
  authentication: string
  sqliteVersion: string
  urlBase: string
  runtimeVersion: string
  runtimeName: string
  startTime: string
}

export interface DiskSpace {
  path: string
  label: string
  freeSpace: number
  totalSpace: number
}

export interface IndexerConfig {
  id: number
  name: string
  implementation: string
  configContract: string
  enableRss: boolean
  enableAutomaticSearch: boolean
  enableInteractiveSearch: boolean
  priority: number
  fields?: Record<string, unknown>
}

export interface DownloadClientConfig {
  id: number
  name: string
  enable: boolean
  protocol: string
  priority: number
  implementation: string
  configContract: string
  fields?: Record<string, unknown>
}

export interface NotificationConfig {
  id: number
  name: string
  implementation: string
  configContract: string
  fields?: Record<string, unknown>
  onGrab: boolean
  onDownload: boolean
  onUpgrade: boolean
  onRename: boolean
}

export interface GeneralConfig {
  bindAddress: string
  port: number
  urlBase: string
  enableSsl: boolean
  launchBrowser: boolean
  authenticationMethod: string
  analyticsEnabled: boolean
  logLevel: string
  logTarget: string      // stderr | stdout | file | syslog
  logFile: string
  syslogAddress: string
  syslogPort: number
  syslogNetwork: string  // udp | tcp | unix
  branch: string
  updateAutomatically: boolean
  updateMechanism: string
}

export interface QueueRecord {
  id: number
  movieId: number
  title: string
  size: number
  sizeleft: number
  status: string
  trackedDownloadStatus: string
  trackedDownloadState: string
  downloadId: string
  protocol: string
  downloadClient: string
}

export interface QueueResponse {
  page: number
  pageSize: number
  totalRecords: number
  records: QueueRecord[]
}

export interface QueueStatus {
  totalCount: number
  count: number
  warnings: boolean
  errors: boolean
}

export interface FeedEvent {
  eventType: string
  data: unknown
  receivedAt: string
}

export interface SchemaOption {
  implementation: string
  configContract: string
  protocol?: string
  fields?: Array<{ name: string; label: string; type: string; value?: unknown }>
}
