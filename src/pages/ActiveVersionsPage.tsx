import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../lib/api';
import { GitBranch, RefreshCw, AlertCircle, Loader2, CheckCircle, XCircle } from 'lucide-react';

export default function ActiveVersionsPage() {
  const { serviceId } = useParams();
  const [versions, setVersions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [environment, setEnvironment] = useState('production');

  useEffect(() => {
    if (serviceId) {
      api.getActiveVersions(serviceId, environment).then(setVersions).finally(() => setLoading(false));
    }
  }, [serviceId, environment]);

  if (loading) {
    return <div className="flex items-center justify-center min-h-screen"><Loader2 className="w-8 h-8 animate-spin text-orange-500" /></div>;
  }

  return (
    <div className="max-w-7xl mx-auto px-8 py-8">
      <div className="mb-8">
        <div className="flex items-center gap-2 text-sm text-gray-500 mb-2">
          <Link to="/" className="hover:text-gray-700">Home</Link>
          <span>/</span>
          <Link to={`/service/${serviceId}`} className="hover:text-gray-700">Service</Link>
          <span>/</span>
          <span className="text-gray-900">Active Versions</span>
        </div>
        <h1 className="text-3xl font-light text-gray-900">Active Versions</h1>
        <p className="text-gray-600 mt-2">Manage multiple deployed versions and traffic distribution</p>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
        <div className="flex items-center gap-4">
          <span className="text-sm font-medium text-gray-700">Environment:</span>
          <div className="flex gap-2">
            {['production', 'staging', 'development'].map(env => (
              <button
                key={env}
                onClick={() => setEnvironment(env)}
                className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
                  environment === env ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {env.charAt(0).toUpperCase() + env.slice(1)}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Deployed Versions</h2>
        </div>

        {versions.length === 0 ? (
          <div className="px-6 py-12 text-center">
            <GitBranch className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <p className="text-gray-500 mb-2">No active versions</p>
            <p className="text-sm text-gray-400">Deploy a new version to get started</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {versions.map(version => (
              <div key={version.id} className="px-6 py-4 hover:bg-gray-50">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="text-lg font-semibold text-gray-900">v{version.version}</h3>
                      <span className={`px-2 py-1 rounded text-xs font-medium ${
                        version.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'
                      }`}>
                        {version.status}
                      </span>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4 mb-3">
                      <div>
                        <p className="text-xs text-gray-500 mb-1">Branch</p>
                        <div className="flex items-center gap-2">
                          <GitBranch className="w-4 h-4 text-gray-400" />
                          <span className="text-sm font-mono text-gray-900">{version.branch}</span>
                        </div>
                      </div>
                      <div>
                        <p className="text-xs text-gray-500 mb-1">Traffic</p>
                        <span className="text-sm font-semibold text-gray-900">{version.trafficPercentage}%</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-4">
                      <span className="text-sm font-medium text-gray-700">Traffic:</span>
                      <input
                        type="range"
                        min="0"
                        max="100"
                        value={version.trafficPercentage}
                        onChange={(e) => api.updateVersionTraffic(version.id, parseInt(e.target.value)).then(() => {
                          setVersions(prev => prev.map(v => v.id === version.id ? { ...v, trafficPercentage: parseInt(e.target.value) } : v));
                        })}
                        className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                      />
                      <span className="text-sm font-semibold text-gray-900 w-12 text-right">{version.trafficPercentage}%</span>
                    </div>
                  </div>

                  <div className="flex flex-col gap-2 ml-4">
                    <button
                      onClick={() => api.promoteVersion(version.id).then(() => api.getActiveVersions(serviceId!, environment).then(setVersions))}
                      disabled={version.trafficPercentage === 100}
                      className="flex items-center gap-2 px-3 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:bg-gray-300 text-sm"
                    >
                      <CheckCircle className="w-4 h-4" />
                      Promote
                    </button>
                    <button
                      onClick={() => api.rollbackVersion(version.id).then(() => api.getActiveVersions(serviceId!, environment).then(setVersions))}
                      className="flex items-center gap-2 px-3 py-2 bg-red-500 text-white rounded hover:bg-red-600 text-sm"
                    >
                      <XCircle className="w-4 h-4" />
                      Rollback
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
