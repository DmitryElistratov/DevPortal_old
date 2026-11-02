import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../lib/api';
import { GitBranch, Clock, Users, ArrowUpRight, Plus } from 'lucide-react';

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
  dependsOn?: string[];
  links?: { title: string; url: string; icon: string }[];
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

const lifecycleColors: Record<string, string> = {
  production: 'bg-green-100 text-green-700',
  staging: 'bg-yellow-100 text-yellow-700',
  development: 'bg-orange-100 text-orange-700',
  deprecated: 'bg-red-100 text-red-700',
  experimental: 'bg-purple-100 text-purple-700',
};

export default function ComponentPage() {
  const { serviceId } = useParams();
  const [service, setService] = useState<Service | null>(null);
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'releases' | 'ci/cd' | 'api' | 'observability'>('overview');

  useEffect(() => {
    if (serviceId) {
      loadService();
    }
  }, [serviceId]);

  const loadService = async () => {
    try {
      setLoading(true);
      const serviceData = await api.getService(serviceId!);
      setService(serviceData);
      
      const releasesData = await api.listReleases(serviceId!);
      setReleases(releasesData || []);
      
      setError(null);
    } catch (err) {
      setError('Failed to load service');
      console.error('Failed to load service:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (error || !service) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error || 'Service not found'}</p>
          <Link to="/catalog" className="text-orange-500 hover:underline">← Back to catalog</Link>
        </div>
      </div>
    );
  }

  const latestRelease = releases[0];
  const tabs = [
    { id: 'overview' as const, label: 'Overview' },
    { id: 'releases' as const, label: 'Releases' },
    { id: 'ci/cd' as const, label: 'CI/CD' },
    { id: 'api' as const, label: 'API' },
    { id: 'observability' as const, label: 'Observability' },
  ];

  return (
    <div>
      {/* Header */}
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-6">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-2">
            <Link to="/catalog" className="hover:text-gray-300">Catalog</Link>
            <span className="mx-2">/</span>
            <span>{service.type}</span>
          </p>
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 rounded bg-white/10 flex items-center justify-center text-2xl font-medium">
              {service.name.charAt(0).toUpperCase()}
            </div>
            <div>
              <h1 className="text-2xl font-light">{service.name}</h1>
              <p className="text-gray-400 text-sm mt-0.5">{service.description}</p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        {/* Tabs */}
        <div className="bg-white rounded-t-lg shadow-sm border border-gray-200">
          <div className="flex border-b border-gray-200">
            {tabs.map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`px-6 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-orange-500 text-orange-500'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>

        {/* Tab content */}
        <div className="bg-white rounded-b-lg shadow-sm border border-gray-200 border-t-0 p-6">
          {activeTab === 'overview' && (
            <div className="grid grid-cols-3 gap-6">
              {/* Left - Info cards */}
              <div className="col-span-2 space-y-6">
                {/* About */}
                <div>
                  <h2 className="text-lg font-semibold text-gray-900 mb-3">About</h2>
                  <p className="text-sm text-gray-600 leading-relaxed">{service.description}</p>
                </div>

                {/* Dependencies */}
                {service.dependsOn && service.dependsOn.length > 0 && (
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-3">Dependencies</h2>
                    <div className="flex gap-2 flex-wrap">
                      {service.dependsOn.map((dep, i) => (
                        <span key={i} className="text-sm px-3 py-1 bg-gray-100 text-gray-700 rounded-full">
                          {dep}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Latest release */}
                {latestRelease && (
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900 mb-3">Latest Release</h2>
                    <div className="border border-gray-200 rounded-lg p-4">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm text-orange-500">v{latestRelease.version}</span>
                          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                            latestRelease.status === 'production' ? 'bg-green-100 text-green-700' :
                            latestRelease.status === 'testing' ? 'bg-yellow-100 text-yellow-700' :
                            latestRelease.status === 'building' ? 'bg-orange-100 text-orange-700' :
                            latestRelease.status === 'staging' ? 'bg-purple-100 text-purple-700' :
                            latestRelease.status === 'rolled_back' ? 'bg-red-100 text-red-700' :
                            'bg-gray-100 text-gray-700'
                          }`}>
                            {latestRelease.status}
                          </span>
                        </div>
                        <Link to={`/release/${latestRelease.id}`} className="text-sm text-orange-500 hover:text-orange-600 flex items-center gap-1">
                          View details <ArrowUpRight className="w-3 h-3" />
                        </Link>
                      </div>
                      {latestRelease.description && (
                        <p className="text-sm text-gray-600 mb-3">{latestRelease.description}</p>
                      )}
                      <div className="flex items-center gap-4 text-xs text-gray-500">
                        <span className="flex items-center gap-1"><GitBranch className="w-3 h-3" />{latestRelease.branch}</span>
                        <span className="flex items-center gap-1"><Users className="w-3 h-3" />{latestRelease.createdBy}</span>
                        <span className="flex items-center gap-1"><Clock className="w-3 h-3" />{new Date(latestRelease.createdAt).toLocaleDateString('ru-RU')}</span>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* Right sidebar - Metadata */}
              <div className="space-y-6">
                <div className="border border-gray-200 rounded-lg overflow-hidden">
                  <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
                    <h3 className="text-sm font-semibold text-gray-700">Metadata</h3>
                  </div>
                  <div className="divide-y divide-gray-100">
                    <div className="px-4 py-3 flex items-center justify-between">
                      <span className="text-xs text-gray-500">Type</span>
                      <span className="text-sm text-gray-900">{service.type}</span>
                    </div>
                    <div className="px-4 py-3 flex items-center justify-between">
                      <span className="text-xs text-gray-500">Lifecycle</span>
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${lifecycleColors[service.lifecycle]}`}>
                        {service.lifecycle}
                      </span>
                    </div>
                    <div className="px-4 py-3 flex items-center justify-between">
                      <span className="text-xs text-gray-500">Owner</span>
                      <span className="text-sm text-gray-900">{service.owner}</span>
                    </div>
                    <div className="px-4 py-3 flex items-center justify-between">
                      <span className="text-xs text-gray-500">Language</span>
                      <span className="text-sm text-gray-900">{service.language}</span>
                    </div>
                  </div>
                </div>

                {/* Links */}
                <div className="border border-gray-200 rounded-lg overflow-hidden">
                  <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
                    <h3 className="text-sm font-semibold text-gray-700">Links</h3>
                  </div>
                  <div className="divide-y divide-gray-100">
                    <a href={service.repository} target="_blank" rel="noopener noreferrer" className="px-4 py-3 flex items-center gap-2 text-sm text-orange-500 hover:bg-gray-50">
                      <GitBranch className="w-4 h-4" /> Repository
                    </a>
                    {service.links?.map((link, i) => (
                      <a key={i} href={link.url} target="_blank" rel="noopener noreferrer" className="px-4 py-3 flex items-center gap-2 text-sm text-orange-500 hover:bg-gray-50">
                        <ArrowUpRight className="w-4 h-4" /> {link.title}
                      </a>
                    ))}
                  </div>
                </div>

                {/* Tags */}
                {service.tags && service.tags.length > 0 && (
                  <div className="border border-gray-200 rounded-lg overflow-hidden">
                    <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
                      <h3 className="text-sm font-semibold text-gray-700">Labels</h3>
                    </div>
                    <div className="p-4 flex flex-wrap gap-1">
                      {service.tags.map((tag, i) => (
                        <span key={i} className="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded">
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === 'releases' && (
            <div>
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-semibold text-gray-900">Releases</h2>
                <Link
                  to={`/service/${service.id}/create-release`}
                  className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded text-sm font-medium"
                >
                  <Plus className="w-4 h-4" />
                  Create Release
                </Link>
              </div>
              {releases.length === 0 ? (
                <div className="text-center py-12 text-gray-500">No releases yet</div>
              ) : (
                <div className="border border-gray-200 rounded-lg overflow-hidden">
                  <table className="w-full">
                    <thead>
                      <tr className="bg-gray-50 border-b border-gray-200">
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Version</th>
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Status</th>
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Branch</th>
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Author</th>
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Created</th>
                        <th className="text-left px-6 py-3 text-xs font-semibold text-gray-500">Approvals</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {releases.map(rel => (
                        <tr key={rel.id} className="hover:bg-gray-50">
                          <td className="px-6 py-3">
                            <Link to={`/release/${rel.id}`} className="text-orange-500 hover:text-orange-600 font-mono text-sm">
                              v{rel.version}
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
                          <td className="px-6 py-3 text-sm text-gray-600 font-mono">{rel.branch}</td>
                          <td className="px-6 py-3 text-sm text-gray-600">{rel.createdBy}</td>
                          <td className="px-6 py-3 text-sm text-gray-500">{new Date(rel.createdAt).toLocaleDateString('ru-RU')}</td>
                          <td className="px-6 py-3">
                            <span className={`text-sm ${rel.approvals >= rel.requiredApprovals ? 'text-green-600' : 'text-yellow-600'}`}>
                              {rel.approvals}/{rel.requiredApprovals}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}

          {activeTab === 'observability' && (
            <div className="text-center py-12">
              <p className="text-gray-500 mb-4">View logs and metrics for this service</p>
              <Link
                to={`/service/${service.id}/observability`}
                className="inline-flex items-center gap-2 px-6 py-3 bg-orange-500 hover:bg-orange-600 text-white rounded-lg text-sm font-medium"
              >
                Open Observability Dashboard
              </Link>
            </div>
          )}

          {(activeTab === 'ci/cd' || activeTab === 'api') && (
            <div className="text-center py-12 text-gray-500">
              {activeTab === 'ci/cd' ? 'CI/CD pipeline data will be available from backend API' : 'API endpoints will be available from backend API'}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
