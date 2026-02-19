import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom';
import {
  LayoutDashboard, List, GitCompare, Shield, Eye, Wrench,
} from 'lucide-react';
import Dashboard from './pages/Dashboard';
import Migrations from './pages/Migrations';
import DiffViewer from './pages/DiffViewer';
import SchemaView from './pages/SchemaView';
import LintResults from './pages/LintResults';
import Actions from './pages/Actions';

const nav = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/migrations', icon: List, label: 'Migrations' },
  { to: '/diff', icon: GitCompare, label: 'Diff' },
  { to: '/schema', icon: Eye, label: 'Schema' },
  { to: '/lint', icon: Shield, label: 'Lint' },
  { to: '/actions', icon: Wrench, label: 'Actions' },
];

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen">
        {/* Sidebar */}
        <aside className="w-56 bg-slate-900 border-r border-slate-700 flex flex-col">
          <div className="p-4 border-b border-slate-700">
            <h1 className="text-xl font-bold tracking-tight text-white">drift</h1>
            <p className="text-xs text-slate-500 mt-0.5">Database Migrations</p>
          </div>
          <nav className="flex-1 p-2 space-y-0.5">
            {nav.map(({ to, icon: Icon, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
                    isActive
                      ? 'bg-blue-600/20 text-blue-400'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
                  }`
                }
              >
                <Icon size={16} />
                {label}
              </NavLink>
            ))}
          </nav>
          <div className="p-3 border-t border-slate-700 text-xs text-slate-600">
            drift v0.1.0
          </div>
        </aside>

        {/* Main content */}
        <main className="flex-1 overflow-auto bg-slate-950 p-6">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/migrations" element={<Migrations />} />
            <Route path="/diff" element={<DiffViewer />} />
            <Route path="/schema" element={<SchemaView />} />
            <Route path="/lint" element={<LintResults />} />
            <Route path="/actions" element={<Actions />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}
