import { useEffect, useState } from 'react';
import { api } from '../lib/api';

interface Deployment {
  metadata: { name: string; namespace: string };
  spec: { replicas: number };
  status: { readyReplicas: number };
}

export default function DeploymentsPage() {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [namespace, setNamespace] = useState('default');

  const load = async (ns: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.listKubernetesDeployments(ns);
      setDeployments(Array.isArray(data) ? data : []);
    } catch (e: any) {
      setError(e?.message || 'Failed to load deployments. Check that a Kubernetes integration is configured and enabled in Settings.');
      setDeployments([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(namespace);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8 flex items-center justify-between">
          <h1 className="text-2xl font-light">Kubernetes Deployments</h1>
          <div className="flex items-center gap-3">
            <input
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              placeholder="namespace"
              className="px-3 py-1.5 text-sm rounded bg-white/10 border border-white/20 text-white placeholder-white/50 focus:outline-none focus:border-white/40"
            />
            <button
              onClick={() => load(namespace)}
              className="px-4 py-1.5 text-sm bg-white text-black rounded hover:bg-gray-200 transition-colors"
            >
              Refresh
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 py-8">
        {loading && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center text-gray-500">
            Loading deployments…
          </div>
        )}

        {!loading && error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-6 text-sm">
            {error}
          </div>
        )}

        {!loading && !error && deployments.length === 0 && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center text-gray-500">
            No deployments found in namespace “{namespace}”.
          </div>
        )}

        {!loading && deployments.length > 0 && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200 text-left text-gray-600">
                <tr>
                  <th className="px-6 py-3 font-medium">Name</th>
                  <th className="px-6 py-3 font-medium">Namespace</th>
                  <th className="px-6 py-3 font-medium">Ready</th>
                  <th className="px-6 py-3 font-medium">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {deployments.map((d) => {
                  const ready = d.status?.readyReplicas ?? 0;
                  const desired = d.spec?.replicas ?? 0;
                  const healthy = ready === desired && desired > 0;
                  return (
                    <tr key={`${d.metadata.namespace}/${d.metadata.name}`} className="hover:bg-gray-50">
                      <td className="px-6 py-3 font-mono text-gray-900">{d.metadata.name}</td>
                      <td className="px-6 py-3 text-gray-600">{d.metadata.namespace}</td>
                      <td className="px-6 py-3 text-gray-900">{ready}/{desired}</td>
                      <td className="px-6 py-3">
                        <span
                          className={`inline-block px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            healthy
                              ? 'bg-green-100 text-green-700'
                              : desired === 0
                              ? 'bg-gray-100 text-gray-600'
                              : 'bg-yellow-100 text-yellow-700'
                          }`}
                        >
                          {healthy ? 'Healthy' : desired === 0 ? 'Scaled to zero' : 'Progressing'}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
