import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { GitBranch, Search } from 'lucide-react';

interface Release {
  id: string;
  serviceId: string;
  version: string;
  status: string;
  description: string;
  branch: string;
  createdBy: string;
  approvals: number;
  requiredApprovals: number;
  createdAt: string;
}

interface Service {
  id: string;
  name: string;
}

export default function ReleasesPage() {
  const [releases, setReleases] = useState<Release[]>([]);
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [releasesData, servicesData] = await Promise.all([
        api.listReleases(),
        api.listServices(),
      ]);
      setReleases(releasesData || []);
      setServices(servicesData || []);
      setError(null);
    } catch (err) {
      setError('Failed to load releases');
      console.error('Failed to load releases:', err);
    } finally {
      setLoading(false);
    }
  };

  const filteredReleases = releases.filter(r => {
    const service = services.find(s => s.id === r.serviceId);
    if (statusFilter !== 'all' && r.status !== statusFilter) return false;
    if (searchQuery && !r.version.includes(searchQuery) && !(service?.name || '').includes(searchQuery)) return false;
    return true;
  });

  if (loading) {
    return <div className="flex items-center justify-center min-h-screen"><div className="text-gray-500">Loading...</div></div>;
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error}</p>
          <button onClick={loadData} className="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600">Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Releases</p>
          <h1 className="text-2xl font-light">Releases</h1>
          <p className="text-gray-400 mt-1 text-sm">All releases across services</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6 p-4 flex items-center gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by version or service..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
          <div className="flex gap-2">
            {['all', 'building', 'testing', 'staging', 'production', 'closed', 'rolled_back'].map(status => (
              <button
                key={status}
                onClick={() => setStatusFilter(status)}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  statusFilter === status ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                {status === 'all' ? 'All' : status}
              </button>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-3 border-b border-gray-200">
            <span className="text-sm text-gray-500">{filteredReleases.length} results</span>
          </div>
          {filteredReleases.length === 0 ? (
            <div className="px-6 py-12 text-center text-gray-400 text-sm">No releases found</div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200 bg-gray-50">
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Version</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Service</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Status</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Author</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Created</th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500 uppercase">Approvals</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filteredReleases.map(rel => {
                  const service = services.find(s => s.id === rel.serviceId);
                  return (
                    <tr key={rel.id} className="hover:bg-orange-50/30">
                      <td className="px-6 py-3">
                        <Link to={`/release/${rel.id}`} className="text-orange-500 hover:text-orange-600 font-mono text-sm">
                          v{rel.version}
                        </Link>
                      </td>
                      <td className="px-6 py-3">
                        <Link to={`/service/${rel.serviceId}`} className="text-sm text-gray-700 hover:text-orange-500">
                          {service?.name || 'Unknown'}
                        </Link>
                      </td>
                      <td className="px-6 py-3">
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                          rel.status === 'production' ? 'bg-green-100 text-green-700' :
                          rel.status === 'testing' ? 'bg-yellow-100 text-yellow-700' :
                          rel.status === 'building' ? 'bg-orange-100 text-orange-700' :
                          rel.status === 'staging' ? 'bg-purple-100 text-purple-700' :
                          rel.status === 'rolled_back' ? 'bg-red-100 text-red-700' :
                          'bg-gray-100 text-gray-700'
                        }`}>
                          {rel.status}
                        </span>
                      </td>
                      <td className="px-6 py-3 text-sm text-gray-600">{rel.createdBy}</td>
                      <td className="px-6 py-3 text-sm text-gray-500">{new Date(rel.createdAt).toLocaleDateString('ru-RU')}</td>
                      <td className="px-6 py-3">
                        <span className={`text-sm ${rel.approvals >= rel.requiredApprovals ? 'text-green-600' : 'text-yellow-600'}`}>
                          {rel.approvals}/{rel.requiredApprovals}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
