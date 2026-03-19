import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Server, HardDrive, Clock, GitBranch, Cpu, Wifi, WifiOff } from 'lucide-react';
import api from '../lib/api';
import type { SystemStatus, DiskSpace } from '../lib/types';

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

function formatUptime(startTime: string) {
  const diff = Date.now() - new Date(startTime).getTime();
  const h = Math.floor(diff / 3600000);
  const m = Math.floor((diff % 3600000) / 60000);
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

interface FeedEvent {
  type: string;
  message: string;
  time: string;
}

export default function System() {
  const [events, setEvents] = useState<FeedEvent[]>([]);
  const [connected, setConnected] = useState(false);

  const { data: status } = useQuery<SystemStatus>({
    queryKey: ['system-status'],
    queryFn: () => api.get('/system/status').then((r) => r.data),
    refetchInterval: 30000,
  });

  const { data: diskSpaces = [] } = useQuery<DiskSpace[]>({
    queryKey: ['disk-space'],
    queryFn: () => api.get('/diskspace').then((r) => r.data),
    refetchInterval: 60000,
  });

  // SSE feed
  useEffect(() => {
    const es = new EventSource('/api/v1/feed');

    es.onopen = () => setConnected(true);
    es.onerror = () => setConnected(false);

    es.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data) as FeedEvent;
        setEvents((prev) => [{ ...data, time: new Date().toLocaleTimeString() }, ...prev].slice(0, 50));
      } catch {
        // ignore parse errors
      }
    };

    return () => es.close();
  }, []);

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-100">System Status</h1>
        <div className="flex items-center gap-2">
          {connected ? (
            <><Wifi size={14} className="text-green-400" /><span className="text-xs text-green-400">Live</span></>
          ) : (
            <><WifiOff size={14} className="text-gray-500" /><span className="text-xs text-gray-500">Disconnected</span></>
          )}
        </div>
      </div>

      {/* Status cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatusCard
          icon={<Server size={18} />}
          label="Version"
          value={status?.version ?? '—'}
          sub={status?.branch ?? ''}
        />
        <StatusCard
          icon={<Clock size={18} />}
          label="Uptime"
          value={status ? formatUptime(status.startTime) : '—'}
          sub={status ? `Started ${new Date(status.startTime).toLocaleString()}` : ''}
        />
        <StatusCard
          icon={<Cpu size={18} />}
          label="Runtime"
          value={status?.runtimeName ?? '—'}
          sub={status?.runtimeVersion ?? ''}
        />
        <StatusCard
          icon={<GitBranch size={18} />}
          label="Platform"
          value={status?.osName ?? '—'}
          sub={status?.osVersion ?? ''}
        />
      </div>

      {/* Info details */}
      {status && (
        <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
          <h3 className="text-sm font-semibold text-gray-300 mb-4 flex items-center gap-2">
            <Server size={14} /> Application Info
          </h3>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2.5">
            {[
              ['App Data', status.appData],
              ['Startup Path', status.startupPath],
              ['URL Base', status.urlBase || '/'],
              ['SQLite', status.sqliteVersion],
              ['Authentication', status.authentication],
            ].map(([k, v]) => (
              <div key={k} className="flex gap-2 text-sm">
                <dt className="text-gray-500 w-28 shrink-0">{k}</dt>
                <dd className="text-gray-300 truncate">{v}</dd>
              </div>
            ))}
          </dl>
        </div>
      )}

      {/* Disk space */}
      {diskSpaces.length > 0 && (
        <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
          <h3 className="text-sm font-semibold text-gray-300 mb-4 flex items-center gap-2">
            <HardDrive size={14} /> Disk Space
          </h3>
          <div className="space-y-4">
            {diskSpaces.map((disk) => {
              const usedPct = Math.round(((disk.totalSpace - disk.freeSpace) / disk.totalSpace) * 100);
              return (
                <div key={disk.path}>
                  <div className="flex justify-between text-sm mb-1.5">
                    <span className="text-gray-300 truncate">{disk.label || disk.path}</span>
                    <span className="text-gray-400 ml-4 shrink-0">
                      {formatBytes(disk.freeSpace)} free of {formatBytes(disk.totalSpace)}
                    </span>
                  </div>
                  <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${usedPct > 90 ? 'bg-red-500' : usedPct > 70 ? 'bg-yellow-400' : 'bg-green-500'}`}
                      style={{ width: `${usedPct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Event feed */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl p-5">
        <h3 className="text-sm font-semibold text-gray-300 mb-4 flex items-center gap-2">
          <Wifi size={14} /> Live Event Feed
          {connected && <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />}
        </h3>
        {events.length === 0 ? (
          <p className="text-sm text-gray-600 text-center py-4">Waiting for events…</p>
        ) : (
          <div className="space-y-1.5 max-h-64 overflow-y-auto font-mono text-xs">
            {events.map((e, i) => (
              <div key={i} className="flex gap-3 text-gray-400">
                <span className="text-gray-600 shrink-0">{e.time}</span>
                <span className="text-yellow-400 shrink-0">[{e.type}]</span>
                <span className="text-gray-300">{e.message}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function StatusCard({ icon, label, value, sub }: { icon: React.ReactNode; label: string; value: string; sub: string }) {
  return (
    <div className="bg-gray-900 border border-gray-800 rounded-xl p-4">
      <div className="flex items-center gap-2 text-yellow-400 mb-2">{icon}<span className="text-xs text-gray-400">{label}</span></div>
      <p className="text-lg font-semibold text-gray-100">{value}</p>
      {sub && <p className="text-xs text-gray-500 mt-0.5 truncate">{sub}</p>}
    </div>
  );
}
