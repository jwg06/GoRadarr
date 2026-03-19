export interface Ratings {
  value: number;
  votes: number;
}

export interface Movie {
  id: number;
  title: string;
  sortTitle: string;
  tmdbId: number;
  imdbId?: string;
  overview?: string;
  status: string;
  year: number;
  runtime: number;
  studio?: string;
  qualityProfileId: number;
  rootFolderPath?: string;
  path?: string;
  monitored: boolean;
  minimumAvailability: string;
  hasFile: boolean;
  added: string;
  inCinemas?: string;
  physicalRelease?: string;
  digitalRelease?: string;
  remotePoster?: string;
  ratings?: Ratings;
  genres?: string[];
}

export interface QualityProfile {
  id: number;
  name: string;
  upgradeAllowed: boolean;
  cutoff: number;
}

export interface HistoryRecord {
  id: number;
  movieId: number;
  sourceTitle: string;
  quality: { quality: { name: string } };
  date: string;
  eventType: string;
  data?: Record<string, string>;
  movie?: Movie;
}

export interface HistoryResponse {
  page: number;
  pageSize: number;
  totalRecords: number;
  records: HistoryRecord[];
}

export interface CalendarMovie {
  id: number;
  title: string;
  year: number;
  inCinemas?: string;
  physicalRelease?: string;
  digitalRelease?: string;
  remotePoster?: string;
  monitored: boolean;
  hasFile: boolean;
}

export interface SystemStatus {
  version: string;
  buildTime: string;
  isDebug: boolean;
  isProduction: boolean;
  isAdmin: boolean;
  isUserInteractive: boolean;
  startupPath: string;
  appData: string;
  osName: string;
  osVersion: string;
  isMonoRuntime: boolean;
  isMono: boolean;
  isLinux: boolean;
  isOsx: boolean;
  isWindows: boolean;
  branch: string;
  authentication: string;
  sqliteVersion: string;
  urlBase: string;
  runtimeVersion: string;
  runtimeName: string;
  startTime: string;
}

export interface DiskSpace {
  path: string;
  label: string;
  freeSpace: number;
  totalSpace: number;
}

export interface IndexerConfig {
  id: number;
  name: string;
  enableRss: boolean;
  enableAutomaticSearch: boolean;
  enableInteractiveSearch: boolean;
  supportsRss: boolean;
  supportsSearch: boolean;
  protocol: string;
  priority: number;
  fields?: { name: string; value: unknown }[];
}

export interface DownloadClientConfig {
  id: number;
  name: string;
  enable: boolean;
  protocol: string;
  priority: number;
  fields?: { name: string; value: unknown }[];
}

export interface NotificationConfig {
  id: number;
  name: string;
  onGrab: boolean;
  onDownload: boolean;
  onUpgrade: boolean;
  onRename: boolean;
  supportsOnGrab: boolean;
  supportsOnDownload: boolean;
  supportsOnUpgrade: boolean;
  supportsOnRename: boolean;
  fields?: { name: string; value: unknown }[];
}

export interface GeneralConfig {
  bindAddress: string;
  port: number;
  urlBase: string;
  enableSsl: boolean;
  launchBrowser: boolean;
  authenticationMethod: string;
  analyticsEnabled: boolean;
  logLevel: string;
  branch: string;
  updateAutomatically: boolean;
  updateMechanism: string;
}
