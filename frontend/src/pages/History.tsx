import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronLeft, ChevronRight, Download, Search } from 'lucide-react';
import api from '../lib/api';
import type { HistoryResponse } from '../lib/types';

const PAGE_SIZE = 25;

function eventBadge(type: string) {
  const map: Record<string, string> = {
    grabbed: 'bg-blue-500/20 text-blue-400',
    downloadFolderImported: 'bg-green-500/20 text-green-400',
    downloadFailed: 'bg-red-500/20 text-red-400',
    movieFileDeleted: 'bg-red-500/20 text-red-400',
    movieFolderImported: 'bg-green-500/20 text-green-400',
  };
  return map[type] ?? 'bg-gray-700/40 text-gray-400';
}

export default function History() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');

  const { data, isLoading } = useQuery<HistoryResponse>({
    queryKey: ['history', page],
    queryFn: () =>
      api.get(`/history?page=${page}&pageSize=${PAGE_SIZE}&sortKey=date&sortDirection=descending`).then((r) => r.data),
  });

  const records = data?.records ?? [];
  const total = data?.totalRecords ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  const filtered = search
    ? records.filter((r) =>
        r.sourceTitle.toLowerCase().includes(search.toLowerCase()) ||
        r.movie?.title?.toLowerCase().includes(search.toLowerCase())
      )
    : records;

  return (
    <div className="p-4 md:p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-100">History</h1>
          <p className="text-sm text-gray-400 mt-0.5">{total} records</p>
        </div>
      </div>

      {/* Search */}
      <div className="relative mb-4 max-w-md">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
        <input
          type="text"
          placeholder="Filter results..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full bg-gray-800 border border-gray-700 rounded-lg pl-9 pr-4 py-2 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-yellow-400/50 transition-colors"
        />
      </div>

      {/* Table */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800 text-left">
                <th className="px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider">Movie</th>
                <th className="px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider">Source</th>
                <th className="px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider">Quality</th>
                <th className="px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider">Event</th>
                <th className="px-4 py-3 text-xs font-medium text-gray-400 uppercase tracking-wider">Date</th>
              </tr>
            </thead>
            <tbody>
              {isLoading &&
                Array.from({ length: 10 }).map((_, i) => (
                  <tr key={i} className="border-b border-gray-800/50 animate-pulse">
                    {Array.from({ length: 5 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <div className="h-3 bg-gray-800 rounded w-3/4" />
                      </td>
                    ))}
                  </tr>
                ))}
              {!isLoading &&
                filtered.map((record) => (
                  <tr key={record.id} className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-200">
                      {record.movie?.title ?? `Movie #${record.movieId}`}
                    </td>
                    <td className="px-4 py-3 text-gray-400 max-w-xs truncate" title={record.sourceTitle}>
                      <div className="flex items-center gap-1.5">
                        <Download size={12} className="shrink-0 text-gray-500" />
                        <span className="truncate">{record.sourceTitle}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-400">
                      {record.quality?.quality?.name ?? '—'}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${eventBadge(record.eventType)}`}>
                        {record.eventType}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-gray-400 whitespace-nowrap">
                      {new Date(record.date).toLocaleString()}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>

        {!isLoading && filtered.length === 0 && (
          <div className="text-center py-12 text-gray-500">No history records found</div>
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <span className="text-sm text-gray-400">
            Page {page} of {totalPages}
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed text-gray-300 rounded-lg text-sm transition-colors"
            >
              <ChevronLeft size={14} /> Prev
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed text-gray-300 rounded-lg text-sm transition-colors"
            >
              Next <ChevronRight size={14} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
