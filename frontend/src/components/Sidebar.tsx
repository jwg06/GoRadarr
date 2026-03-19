import { NavLink } from 'react-router-dom';
import { Film, PlusCircle, Calendar, Clock, Settings, Server } from 'lucide-react';
import { clsx } from 'clsx';

const links = [
  { to: '/', icon: Film, label: 'Movies', exact: true },
  { to: '/add', icon: PlusCircle, label: 'Add Movie', exact: false },
  { to: '/calendar', icon: Calendar, label: 'Calendar', exact: false },
  { to: '/history', icon: Clock, label: 'History', exact: false },
  { to: '/settings', icon: Settings, label: 'Settings', exact: false },
  { to: '/system/status', icon: Server, label: 'System', exact: false },
];

export default function Sidebar() {
  return (
    <aside className="fixed inset-y-0 left-0 w-16 md:w-56 bg-gray-900 border-r border-gray-800 flex flex-col z-40">
      {/* Logo */}
      <div className="h-14 flex items-center justify-center md:justify-start px-4 border-b border-gray-800 shrink-0">
        <Film className="text-yellow-400 shrink-0" size={24} />
        <span className="hidden md:block ml-3 text-yellow-400 font-bold text-lg tracking-wide">
          GoRadarr
        </span>
      </div>

      <nav className="flex-1 py-4 overflow-y-auto">
        {links.map(({ to, icon: Icon, label, exact }) => (
          <NavLink
            key={to}
            to={to}
            end={exact}
            className={({ isActive }) =>
              clsx(
                'flex items-center justify-center md:justify-start gap-3 px-4 py-3 mx-2 rounded-lg transition-colors text-sm font-medium',
                isActive
                  ? 'bg-yellow-400/10 text-yellow-400'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
              )
            }
          >
            <Icon size={18} className="shrink-0" />
            <span className="hidden md:block">{label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-gray-800">
        <div className="hidden md:flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
          <span className="text-xs text-gray-500">Online</span>
        </div>
        <div className="md:hidden flex justify-center">
          <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
        </div>
      </div>
    </aside>
  );
}
