import { useEffect, useState } from 'react';
import { api } from '../lib/api';

interface ArgoApp {
  name: string;
  namespace?: string;
  sync_status?: string;
  health_status?: string;
  revision?: string;
  repo_url?: string;
  path?: string;
  targetRevision?: string;
  dest_server?: string;
  dest_namespace?: string;
}

const SYNC_STYLES: Record<string, string> = {
  Synced: 'bg-green-100 text-green-700',
  OutOfSync: 'bg-yellow-100 text-yellow-700',
  Unknown: 'bg-gray-100 text-gray-600',
};

const HEALTH_STYLES: Record<string, string> = {
  Healthy: 'bg-green-100 text-green-700',
  Progressing: 'bg-blue-100 text-blue-700',
  Degraded: 'bg-red-100 text-red-700',
  Suspended: 'bg-purple-100 text-purple-700',
  Missing: 'bg-gray-100 text-gray-600',
};

function Badge({ value, styles }: { value: string; styles: Record<string, string> }) {
  const cls = styles[value] || 'bg-gray-100 text-gray-600';
  return (
    <span className={`inline-block px-2.5 py-0.5 rounded-full text-xs font-medium ${cls}`}>
      {value || 'Unknown'}
    </span>
  );
}

export default function DeploymentsPage() {
  const [apps, setApps] = useState<ArgoApp[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState('');
  const [syncing, setSyncing] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.listArgoApplications();
      setApps(Array.isArray(data) ? data : []);
    } catch (e: any) {
      setError(
        e?.message ||
          'Failed to load applications. Configure an ArgoCD integration in Settings (URL + token), press Test Connection and retry.'
      );
      setApps([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const sync = async (name: string) => {
    setSyncing(name);
    setNotice(null);
    try {
      await api.syncArgoApplication(name);
      setNotice(`Sync triggered for “${name}”. Refresh to see the updated state.`);
    } catch (e: any) {
      setNotice(`Sync failed for “${name}”: ${e?.message || 'unknown error'}`);
    } finally {
      setSyncing(null);
    }
  };

  const visible = apps.filter(
    (a) =>
      !filter ||
      a.name.toLowerCase().includes(filter.toLowerCase()) ||
      (a.dest_namespace || '').toLowerCase().includes(filter.toLowerCase())
  );

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-light">Deployments (ArgoCD)</h1>
            <p className="text-sm text-white/60 mt-1">
              GitOps source of truth — applications managed by ArgoCD across all destination clusters.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <input
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="filter by app or namespace"
              className="px-3 py-1.5 text-sm rounded bg-white/10 border border-white/20 text-white placeholder-white/50 focus:outline-none focus:border-white/40"
            />
            <button
              onClick={load}
              disabled={loading}
              className="px-4 py-1.5 text-sm bg-white text-black rounded hover:bg-gray-200 transition-colors disabled:opacity-50"
            >
              {loading ? 'Loading…' : 'Refresh'}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 py-8">
        {notice && (
          <div className="mb-4 bg-blue-50 border border-blue-200 text-blue-800 rounded-lg p-4 text-sm">
            {notice}
          </div>
        )}

        {loading && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center text-gray-500">
            Loading ArgoCD applications…
          </div>
        )}

        {!loading && error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-6 text-sm">{error}</div>
        )}

        {!loading && !error && visible.length === 0 && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center text-gray-500">
            No ArgoCD applications found{filter ? ` matching “${filter}”` : ''}.
          </div>
        )}

        {!loading && visible.length > 0 && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200 text-left text-gray-600">
                <tr>
                  <th className="px-6 py-3 font-medium">Application</th>
                  <th className="px-6 py-3 font-medium">Sync</th>
                  <th className="px-6 py-3 font-medium">Health</th>
                  <th className="px-6 py-3 font-medium">Target revision</th>
                  <th className="px-6 py-3 font-medium">Repository</th>
                  <th className="px-6 py-3 font-medium">Destination</th>
                  <th className="px-6 py-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {visible.map((a) => (
                  <tr key={a.name} className="hover:bg-gray-50 align-top">
                    <td className="px-6 py-3">
                      <div className="font-mono text-gray-900">{a.name}</div>
                      {a.path && <div className="text-xs text-gray-500 mt-0.5">{a.path}</div>}
                    </td>
                    <td className="px-6 py-3">
                      <Badge value={a.sync_status || 'Unknown'} styles={SYNC_STYLES} />
                    </td>
                    <td className="px-6 py-3">
                      <Badge value={a.health_status || 'Unknown'} styles={HEALTH_STYLES} />
                    </td>
                    <td className="px-6 py-3">
                      <div className="text-gray-900">{a.targetRevision || '—'}</div>
                      {a.revision && (
                        <div className="text-xs text-gray-500 font-mono mt-0.5" title={a.revision}>
                          {a.revision.slice(0, 8)}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-3 text-gray-600 max-w-[220px] truncate" title={a.repo_url}>
                      {a.repo_url || '—'}
                    </td>
                    <td className="px-6 py-3 text-gray-600">
                      <div>{a.dest_namespace || '—'}</div>
                      {a.dest_server && <div className="text-xs text-gray-400 mt-0.5">{a.dest_server}</div>}
                    </td>
                    <td className="px-6 py-3 text-right">
                      <button
                        onClick={() => sync(a.name)}
                        disabled={syncing === a.name}
                        className="px-3 py-1 text-xs rounded border border-gray-300 text-gray-700 hover:bg-gray-100 transition-colors disabled:opacity-50"
                      >
                        {syncing === a.name ? 'Syncing…' : 'Sync'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
