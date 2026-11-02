import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { GitBranch, Clock, AlertCircle, CheckCircle, TrendingUp, Activity } from 'lucide-react';

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
}

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

export default function HomePage() {
  const [services, setServices] = useState<Service[]>([]);
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [servicesData, releasesData] = await Promise.all([
        api.listServices(),
        api.listReleases(),
      ]);
      setServices(servicesData || []);
      setReleases(releasesData || []);
      setError(null);
    } catch (err) {
      setError('Failed to load data. Please check if backend is running.');
      console.error('Failed to load ', err);
    } finally {
      setLoading(false);
    }
  };

  const totalReleases = releases.length;
  const activeReleases = releases.filter(r => !['closed', 'rolled_back'].includes(r.status)).length;
  const failedReleases = releases.filter(r => r.status === 'rolled_back').length;

  const starredServices = services.slice(0, 3);
  const recentReleases = [...releases]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5);

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
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={loadData}
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
        <div className="max-w-7xl mx-auto px-8 py-10">
          <h1 className="text-3xl font-light">Welcome back!</h1>
          <p className="text-gray-400 mt-2">Your developer portal for release management</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-6 pb-8">
        <div className="grid grid-cols-4 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-orange-50 rounded-lg">
                <Activity className="w-5 h-5 text-orange-600" />
              </div>
              <span className="text-sm text-gray-500">Total Services</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{services.length}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-green-50 rounded-lg">
                <GitBranch className="w-5 h-5 text-green-600" />
              </div>
              <span className="text-sm text-gray-500">Active Releases</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{activeReleases}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-purple-50 rounded-lg">
                <TrendingUp className="w-5 h-5 text-purple-600" />
              </div>
              <span className="text-sm text-gray-500">Total Releases</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{totalReleases}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-red-50 rounded-lg">
                <AlertCircle className="w-5 h-5 text-red-600" />
              </div>
              <span className="text-sm text-gray-500">Rolled Back</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{failedReleases}</p>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-6">
          <div className="col-span-2">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200">
              <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h2 className="text-base font-semibold text-gray-900">Starred Components</h2>
                <Link to="/catalog" className="text-sm text-orange-500 hover:text-orange-600">View all</Link>
              </div>
              <div className="divide-y divide-gray-100">
                {starredServices.length === 0 ? (
                  <div className="px-6 py-8 text-center">
                    <p className="text-gray-500 mb-4">No services yet</p>
                    <button
                      onClick={() => {}}
                      className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-lg text-sm font-medium"
                    >
                      Load Demo Data
                    </button>
                  </div>
                ) : (
                  starredServices.map((svc) => (
                    <Link
                      key={svc.id}
                      to={`/service/${svc.id}`}
                      className="flex items-center justify-between px-6 py-4 hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center text-white text-sm font-medium">
                          {svc.name.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <h3 className="text-sm font-medium text-gray-900">{svc.name}</h3>
                          <p className="text-xs text-gray-500 mt-0.5">{svc.owner} • {svc.language}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700 font-medium">
                          {svc.lifecycle}
                        </span>
                      </div>
                    </Link>
                  ))
                )}
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 mt-6">
              <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h2 className="text-base font-semibold text-gray-900">Recent Releases</h2>
                <Link to="/releases" className="text-sm text-orange-500 hover:text-orange-600">View all</Link>
              </div>
              <div className="divide-y divide-gray-100">
                {recentReleases.length === 0 ? (
                  <div className="px-6 py-8 text-center text-gray-500">No releases yet</div>
                ) : (
                  recentReleases.map((rel) => {
                    const service = services.find(s => s.id === rel.serviceId);
                    return (
                      <Link
                        key={rel.id}
                        to={`/release/${rel.id}`}
                        className="flex items-center justify-between px-6 py-3 hover:bg-gray-50 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <GitBranch className="w-4 h-4 text-orange-500" />
                          <div>
                            <span className="text-sm text-gray-900 font-mono">{service?.name || 'Unknown'}</span>
                            <span className="text-sm text-gray-500 ml-2">v{rel.version}</span>
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
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
                          <span className="text-xs text-gray-400">
                            {new Date(rel.createdAt).toLocaleDateString('ru-RU')}
                          </span>
                        </div>
                      </Link>
                    );
                  })
                )}
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-base font-semibold text-gray-900">Quick Access</h2>
              </div>
              <div className="p-4 space-y-2">
                <Link to="/releases" className="flex items-center gap-3 px-3 py-2 rounded hover:bg-gray-50 text-sm text-gray-700">
                  <GitBranch className="w-4 h-4 text-gray-400" />
                  All Releases
                </Link>
                <Link to="/checklists" className="flex items-center gap-3 px-3 py-2 rounded hover:bg-gray-50 text-sm text-gray-700">
                  <CheckCircle className="w-4 h-4 text-gray-400" />
                  Checklists
                </Link>
                <Link to="/ttm" className="flex items-center gap-3 px-3 py-2 rounded hover:bg-gray-50 text-sm text-gray-700">
                  <Clock className="w-4 h-4 text-gray-400" />
                  Time to Market
                </Link>
                <Link to="/audit" className="flex items-center gap-3 px-3 py-2 rounded hover:bg-gray-50 text-sm text-gray-700">
                  <AlertCircle className="w-4 h-4 text-gray-400" />
                  Audit Log
                </Link>
                <Link to="/vault/secrets" className="flex items-center gap-3 px-3 py-2 rounded hover:bg-gray-50 text-sm text-gray-700">
                  <Activity className="w-4 h-4 text-gray-400" />
                  Vault Secrets
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
