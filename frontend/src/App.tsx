import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Layout from './components/Layout';
import Movies from './pages/Movies';
import AddMovie from './pages/AddMovie';
import Calendar from './pages/Calendar';
import History from './pages/History';
import Settings from './pages/Settings';
import System from './pages/System';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 30,
      retry: 1,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Movies />} />
            <Route path="/add" element={<AddMovie />} />
            <Route path="/calendar" element={<Calendar />} />
            <Route path="/history" element={<History />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/system/status" element={<System />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
