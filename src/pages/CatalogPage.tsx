import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { Search, Star, Plus, X } from 'lucide-react';

interface Service {
  id: string;
  name: string;
  team: string;
  type: string;
  lifecycle: string;
  description: string;
  language: string;
  repository: string;
  owner: string;
  tags: string[];
}

const lifecycleColors: Record<string, string> = {
  production: 'bg-green-100 text-green-700',
  staging: 'bg-yellow-100 text-yellow-700',
  development: 'bg-orange-100 text-orange-700',
  deprecated: 'bg-red-100 text-red-700',
  experimental: 'bg-purple-100 text-purple-700',
};

const typeIcons: Record<string, string> = {
  service: '⚙️',
  library: '📦',
  website: '🌐',
  api: '🔌',
};

export default function CatalogPage() {
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [lifecycleFilter, setLifecycleFilter] = useState<string>('all');
  const [starred, setStarred] = useState<Set<string>>(new Set());
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newService, setNewService] = useState({
    name: '',
    team: '',
    type: 'service',
    lifecycle: 'development',
    description: '',
    language: '',
    repository: '',
    owner: '',
    tags: [] as string[],
  });
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    loadServices();
  }, []);

  const loadServices = async () => {
    try {
      setLoading(true);
      const data = await api.listServices();
      setServices(data || []);
      setError(null);
    } catch (err) {
      setError('Failed to load services');
      console.error('Failed to load services:', err);
    } finally {
      setLoading(false);
    }
  };

  const filteredServices = services.filter(s => {
    if (searchQuery && !s.name.includes(searchQuery) && !s.description.includes(searchQuery)) return false;
    if (typeFilter !== 'all' && s.type !== typeFilter) return false;
    if (lifecycleFilter !== 'all' && s.lifecycle !== lifecycleFilter) return false;
    return true;
  });

  const toggleStar = (id: string) => {
    setStarred(prev => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleCreateService = async () => {
    if (!newService.name || !newService.team || !newService.owner) {
      alert('Please fill in all required fields');
      return;
    }

    setCreating(true);
    try {
      await api.createService(newService);
      await loadServices();
      setShowCreateModal(false);
      setNewService({
        name: '',
        team: '',
        type: 'service',
        lifecycle: 'development',
        description: '',
        language: '',
        repository: '',
        owner: '',
        tags: [],
      });
    } catch (err) {
      console.error('Failed to create service:', err);
      alert('Failed to create service');
    } finally {
      setCreating(false);
    }
  };

  const handleLoadDemoData = async () => {
    try {
      const demoServices = [
        {
          name: 'payment-service',
          team: 'Payments',
          type: 'service',
          lifecycle: 'production',
          description: 'Payment processing service',
          language: 'Go',
          repository: 'https://github.com/company/payment-service',
          owner: 'Payments Team',
          tags: ['go', 'payments', 'pci-dss'],
        },
        {
          name: 'catalog-service',
          team: 'Catalog',
          type: 'service',
          lifecycle: 'production',
          description: 'Product catalog service',
          language: 'Java',
          repository: 'https://github.com/company/catalog-service',
          owner: 'Catalog Team',
          tags: ['java', 'elasticsearch', 'catalog'],
        },
        {
          name: 'user-service',
          team: 'Identity',
          type: 'service',
          lifecycle: 'production',
          description: 'User authentication and authorization',
          language: 'Go',
          repository: 'https://github.com/company/user-service',
          owner: 'Identity Team',
          tags: ['go', 'auth', 'oauth2'],
        },
      ];

      for (const service of demoServices) {
        await api.createService(service);
      }

      await loadServices();
      alert('Demo data loaded successfully!');
    } catch (err) {
      console.error('Failed to load demo ', err);
      alert('Failed to load demo data');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={loadServices}
            className="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Catalog</p>
          <h1 className="text-2xl font-light">Catalog</h1>
          <p className="text-gray-400 mt-1 text-sm">A centralized view of all software components</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        <div className="flex gap-3 mb-6">
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-lg text-sm font-medium"
          >
            <Plus className="w-4 h-4" />
            Create Service
          </button>
          {services.length === 0 && (
            <button
              onClick={handleLoadDemoData}
              className="flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-sm font-medium"
            >
              Load Demo Data
            </button>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
          <div className="p-4 flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Filter by name or description..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:border-orange-500 focus:ring-1 focus:ring-orange-500"
              />
            </div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
            >
              <option value="all">All Types</option>
              <option value="service">Service</option>
              <option value="library">Library</option>
              <option value="website">Website</option>
              <option value="api">API</option>
            </select>
            <select
              value={lifecycleFilter}
              onChange={(e) => setLifecycleFilter(e.target.value)}
              className="border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
            >
              <option value="all">All Lifecycles</option>
              <option value="production">Production</option>
              <option value="staging">Staging</option>
              <option value="development">Development</option>
              <option value="experimental">Experimental</option>
              <option value="deprecated">Deprecated</option>
            </select>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-3 border-b border-gray-200 flex items-center justify-between">
            <span className="text-sm text-gray-500">{filteredServices.length} results</span>
          </div>
          {filteredServices.length === 0 ? (
            <div className="px-6 py-12 text-center text-gray-500">No services found</div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50">
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider w-8"></th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Name</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Owner</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Lifecycle</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Type</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wider">Tags</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filteredServices.map((svc) => (
                  <tr key={svc.id} className="hover:bg-orange-50/30 transition-colors">
                    <td className="px-6 py-3">
                      <button onClick={() => toggleStar(svc.id)} className="text-gray-300 hover:text-yellow-400 transition-colors">
                        <Star className={`w-4 h-4 ${starred.has(svc.id) ? 'fill-yellow-400 text-yellow-400' : ''}`} />
                      </button>
                    </td>
                    <td className="px-6 py-3">
                      <Link to={`/service/${svc.id}`} className="text-orange-500 hover:text-orange-600 text-sm font-medium">
                        {svc.name}
                      </Link>
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600">{svc.owner}</td>
                    <td className="px-6 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${lifecycleColors[svc.lifecycle] || 'bg-gray-100 text-gray-700'}`}>
                        {svc.lifecycle}
                      </span>
                    </td>
                    <td className="px-6 py-3 text-sm text-gray-600">
                      <span className="mr-1">{typeIcons[svc.type] || '⚙️'}</span>
                      {svc.type}
                    </td>
                    <td className="px-6 py-3">
                      <div className="flex gap-1 flex-wrap">
                        {(svc.tags || []).slice(0, 3).map((tag, i) => (
                          <span key={i} className="text-xs px-1.5 py-0.5 bg-gray-100 text-gray-600 rounded">
                            {tag}
                          </span>
                        ))}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={() => setShowCreateModal(false)} />
          <div className="relative bg-white rounded-lg shadow-2xl w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Create New Service</h2>
              <button onClick={() => setShowCreateModal(false)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Service Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={newService.name}
                  onChange={(e) => setNewService({ ...newService, name: e.target.value })}
                  placeholder="payment-service"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Team <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    value={newService.team}
                    onChange={(e) => setNewService({ ...newService, team: e.target.value })}
                    placeholder="Payments"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Owner <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    value={newService.owner}
                    onChange={(e) => setNewService({ ...newService, owner: e.target.value })}
                    placeholder="Payments Team"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
                  <select
                    value={newService.type}
                    onChange={(e) => setNewService({ ...newService, type: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  >
                    <option value="service">Service</option>
                    <option value="library">Library</option>
                    <option value="website">Website</option>
                    <option value="api">API</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Lifecycle</label>
                  <select
                    value={newService.lifecycle}
                    onChange={(e) => setNewService({ ...newService, lifecycle: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  >
                    <option value="development">Development</option>
                    <option value="staging">Staging</option>
                    <option value="production">Production</option>
                    <option value="deprecated">Deprecated</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <textarea
                  value={newService.description}
                  onChange={(e) => setNewService({ ...newService, description: e.target.value })}
                  placeholder="Service description..."
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Language</label>
                  <input
                    type="text"
                    value={newService.language}
                    onChange={(e) => setNewService({ ...newService, language: e.target.value })}
                    placeholder="Go"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Repository</label>
                  <input
                    type="text"
                    value={newService.repository}
                    onChange={(e) => setNewService({ ...newService, repository: e.target.value })}
                    placeholder="https://github.com/company/service"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                </div>
              </div>

              <div className="flex gap-3 pt-4 border-t border-gray-200">
                <button
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateService}
                  disabled={creating}
                  className="flex-1 px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white rounded-lg font-medium"
                >
                  {creating ? 'Creating...' : 'Create Service'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
