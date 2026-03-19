import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronLeft, ChevronRight, Film } from 'lucide-react';
import api from '../lib/api';
import type { CalendarMovie } from '../lib/types';
import { clsx } from 'clsx';

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
const MONTHS = [
  'January','February','March','April','May','June',
  'July','August','September','October','November','December'
];

function isoDateStr(d: Date) {
  return d.toISOString().split('T')[0];
}

export default function Calendar() {
  const [current, setCurrent] = useState(new Date());

  const year = current.getFullYear();
  const month = current.getMonth();

  const startOfMonth = new Date(year, month, 1);
  const endOfMonth = new Date(year, month + 1, 0);

  const { data: movies = [] } = useQuery<CalendarMovie[]>({
    queryKey: ['calendar', year, month],
    queryFn: () =>
      api
        .get(`/calendar?start=${isoDateStr(startOfMonth)}&end=${isoDateStr(endOfMonth)}`)
        .then((r) => r.data),
  });

  // Build day -> movies map
  const moviesByDate = new Map<string, CalendarMovie[]>();
  movies.forEach((m) => {
    [m.inCinemas, m.physicalRelease, m.digitalRelease].forEach((d) => {
      if (!d) return;
      const day = d.split('T')[0];
      if (!moviesByDate.has(day)) moviesByDate.set(day, []);
      moviesByDate.get(day)!.push(m);
    });
  });

  // Calendar grid: pad to start on Sunday
  const startDay = startOfMonth.getDay();
  const daysInMonth = endOfMonth.getDate();
  const cells: (number | null)[] = [
    ...Array(startDay).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => i + 1),
  ];
  while (cells.length % 7 !== 0) cells.push(null);

  const todayStr = isoDateStr(new Date());

  function prevMonth() { setCurrent(new Date(year, month - 1, 1)); }
  function nextMonth() { setCurrent(new Date(year, month + 1, 1)); }

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-100">Calendar</h1>
        <div className="flex items-center gap-3">
          <button onClick={prevMonth} className="p-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-gray-200 transition-colors">
            <ChevronLeft size={16} />
          </button>
          <span className="text-sm font-medium text-gray-200 min-w-32 text-center">
            {MONTHS[month]} {year}
          </span>
          <button onClick={nextMonth} className="p-2 rounded-lg bg-gray-800 hover:bg-gray-700 text-gray-400 hover:text-gray-200 transition-colors">
            <ChevronRight size={16} />
          </button>
        </div>
      </div>

      {/* Grid */}
      <div className="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        {/* Day headers */}
        <div className="grid grid-cols-7 border-b border-gray-800">
          {DAYS.map((d) => (
            <div key={d} className="py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
              {d}
            </div>
          ))}
        </div>

        {/* Cells */}
        <div className="grid grid-cols-7">
          {cells.map((day, idx) => {
            const dateStr = day ? `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}` : '';
            const dayMovies = day ? (moviesByDate.get(dateStr) ?? []) : [];
            const isToday = dateStr === todayStr;

            return (
              <div
                key={idx}
                className={clsx(
                  'min-h-20 p-1.5 border-r border-b border-gray-800 last:border-r-0',
                  !day && 'bg-gray-950/50',
                  idx % 7 === 6 && 'border-r-0'
                )}
              >
                {day && (
                  <>
                    <span className={clsx(
                      'inline-flex items-center justify-center w-6 h-6 text-xs rounded-full mb-1',
                      isToday ? 'bg-yellow-400 text-gray-900 font-bold' : 'text-gray-400'
                    )}>
                      {day}
                    </span>
                    <div className="space-y-0.5">
                      {dayMovies.slice(0, 3).map((m) => (
                        <div
                          key={m.id}
                          className="flex items-center gap-1 bg-yellow-400/10 text-yellow-300 text-xs px-1.5 py-0.5 rounded truncate"
                          title={m.title}
                        >
                          <Film size={9} className="shrink-0" />
                          <span className="truncate">{m.title}</span>
                        </div>
                      ))}
                      {dayMovies.length > 3 && (
                        <div className="text-xs text-gray-500 px-1.5">+{dayMovies.length - 3} more</div>
                      )}
                    </div>
                  </>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
