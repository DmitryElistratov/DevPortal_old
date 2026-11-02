import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import {
  Home,
  Layers,
  GitBranch,
  ClipboardList,
  Clock,
  Shield,
  Search,
  Settings,
  ChevronLeft,
  ChevronRight,
  FileEdit,
  BookOpen,
  LogOut,
  Cloud,
  Terminal,
  Key,
} from 'lucide-react';

const navGroups = [
  {
    label: '',
    items: [
      { path: '/', label: 'Home', icon: Home },
      { path: '/catalog', label: 'Catalog', icon: Layers },
      { path: '/releases', label: 'Releases', icon: GitBranch },
      { path: '/deployments', label: 'Deployments', icon: Cloud },
      { path: '/kube-api', label: 'Kubernetes API', icon: Terminal },
    ],
  },
  {
    label: 'RELEASE MANAGEMENT',
    items: [
      { path: '/checklists', label: 'Checklists', icon: ClipboardList },
      { path: '/ttm', label: 'Time to Market', icon: Clock },
      { path: '/audit', label: 'Audit Log', icon: Shield },
    ],
  },
  {
    label: 'TOOLS',
    items: [
      { path: '/templates', label: 'Templates', icon: FileEdit },
      { path: '/plugins', label: 'Plugins', icon: FileEdit },
      { path: '/docs', label: 'TechDocs', icon: BookOpen },
      { path: '/vault/secrets', label: 'Vault Secrets', icon: Key },
      { path: '/settings', label: 'Settings', icon: Settings },
    ],
  },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const location = useLocation();
  const { user, logout } = useAuth();

  return (
    <div className="flex h-screen bg-[#f0f0f0] overflow-hidden">
      {/* Sidebar */}
      <aside
        className={`${collapsed ? 'w-[72px]' : 'w-[240px]'} transition-all duration-200 flex flex-col`}
        style={{ backgroundColor: '#171717' }}
      >
        <div className="flex items-center gap-3 px-4 h-[64px] border-b border-white/10">
          <div className="w-8 h-8 rounded bg-orange-500 flex items-center justify-center flex-shrink-0">
            <span className="text-white font-bold text-sm">R</span>
          </div>
          {!collapsed && (
            <div className="overflow-hidden">
              <span className="text-white font-semibold text-base whitespace-nowrap">Release Portal</span>
              <span className="text-gray-500 text-xs block whitespace-nowrap">Developer Portal</span>
            </div>
          )}
        </div>

        <nav className="flex-1 py-3 overflow-y-auto">
          {navGroups.map((group, gi) => (
            <div key={gi} className="mb-2">
              {group.label && !collapsed && (
                <div className="px-4 py-2 text-[10px] font-semibold tracking-wider text-gray-500 uppercase">
                  {group.label}
                </div>
              )}
              {group.items.map((item) => {
                const isActive = location.pathname === item.path ||
                  (item.path !== '/' && location.pathname.startsWith(item.path));
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`flex items-center gap-3 px-4 py-2.5 mx-2 rounded transition-colors relative group ${
                      isActive
                        ? 'text-white'
                        : 'text-gray-400 hover:text-white hover:bg-white/5'
                    }`}
                    title={collapsed ? item.label : undefined}
                  >
                    {isActive && (
                      <div className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-6 bg-orange-500 rounded-r" />
                    )}
                    <item.icon className="w-5 h-5 flex-shrink-0" />
                    {!collapsed && <span className="text-sm">{item.label}</span>}
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>

        <div className="border-t border-white/10 p-2">
          <button
            onClick={() => setCollapsed(!collapsed)}
            className="w-full flex items-center justify-center py-2 text-gray-500 hover:text-white transition-colors rounded"
          >
            {collapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-[56px] bg-white border-b border-gray-200 flex items-center justify-between px-6 flex-shrink-0">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSearchOpen(!searchOpen)}
              className="flex items-center gap-2 bg-gray-100 hover:bg-gray-200 rounded-full px-4 py-1.5 text-sm text-gray-500 transition-colors"
            >
              <Search className="w-4 h-4" />
              <span>Search</span>
              <kbd className="text-xs bg-gray-200 px-1.5 py-0.5 rounded">/</kbd>
            </button>
          </div>
          <div className="flex items-center gap-3">
            {user && (
              <>
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900">{user.fullName}</p>
                  <p className="text-xs text-gray-500 capitalize">{user.role}</p>
                </div>
                <div className="w-8 h-8 rounded-full bg-orange-500 flex items-center justify-center text-white text-sm font-medium">
                  {user.fullName.split(' ').map(n => n[0]).join('')}
                </div>
                <button
                  onClick={logout}
                  className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                  title="Выйти"
                >
                  <LogOut className="w-4 h-4" />
                </button>
              </>
            )}
          </div>
        </header>

        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>

      {searchOpen && (
        <div className="fixed inset-0 z-50 flex items-start justify-center pt-24" onClick={() => setSearchOpen(false)}>
          <div className="absolute inset-0 bg-black/40" />
          <div className="relative bg-white rounded-lg shadow-2xl w-[600px] overflow-hidden" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-200">
              <Search className="w-5 h-5 text-gray-400" />
              <input
                autoFocus
                type="text"
                placeholder="Search for components, APIs, templates..."
                className="flex-1 text-sm outline-none placeholder-gray-400"
              />
              <kbd className="text-xs bg-gray-100 px-2 py-0.5 rounded text-gray-500">ESC</kbd>
            </div>
            <div className="p-4">
              <p className="text-xs text-gray-400 uppercase tracking-wider mb-2">Recent</p>
              <div className="space-y-1">
                <div className="px-3 py-2 hover:bg-gray-50 rounded cursor-pointer text-sm text-gray-700">payment-service</div>
                <div className="px-3 py-2 hover:bg-gray-50 rounded cursor-pointer text-sm text-gray-700">order-service</div>
                <div className="px-3 py-2 hover:bg-gray-50 rounded cursor-pointer text-sm text-gray-700">user-service</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
