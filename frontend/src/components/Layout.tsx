import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';

export default function Layout() {
  return (
    <div className="flex min-h-screen bg-gray-950">
      <Sidebar />
      <main className="flex-1 ml-16 md:ml-56 min-h-screen">
        <Outlet />
      </main>
    </div>
  );
}
