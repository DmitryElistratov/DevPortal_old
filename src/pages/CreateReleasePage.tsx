import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useSettings } from '../contexts/SettingsContext';
import { GitBranch, GripVertical, AlertCircle, Loader2 } from 'lucide-react';

export default function CreateReleasePage() {
  const { serviceId } = useParams();
  const navigate = useNavigate();
  const { settings } = useSettings();
  const [service, setService] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [version, setVersion] = useState('');
  const [branch, setBranch] = useState('');
  const [availableBranches, setAvailableBranches] = useState<string[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(false);
  const [branchesError, setBranchesError] = useState('');
  const [creating, setCreating] = useState(false);

  const hasGitIntegration = settings.gitIntegrations.length > 0 && settings.gitIntegrations[0].enabled;

  const loadBranches = () => {
    if (!serviceId || !hasGitIntegration) return;
    setBranchesLoading(true);
    setBranchesError('');
    // Ветки тянутся из GitLab по репозиторию сервиса (GET /api/ci/branches).
    api.listGitBranches(undefined, serviceId)
      .then(branches => {
        const names = branches.map((b: any) => b.name);
        setAvailableBranches(names);
        if (names.length > 0 && !branch) {
          setBranch(names.find((n: string) => n === 'main' || n === 'master') || names[0]);
        }
      })
      .catch((err) => setBranchesError(err instanceof Error ? err.message : 'Failed to load branches'))
      .finally(() => setBranchesLoading(false));
  };

  useEffect(() => {
    if (serviceId) {
      api.getService(serviceId).then(setService).finally(() => setLoading(false));
      loadBranches();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [serviceId, hasGitIntegration]);

  const handleCreate = async () => {
    if (!version || !branch) return;
    setCreating(true);
    try {
      await api.createRelease({
        serviceId,
        version,
        branch,
        status: 'building',
        createdBy: 'Current User',
        approvals: 0,
        requiredApprovals: 2,
      });
      navigate(`/service/${serviceId}`);
    } catch (err) {
      alert('Failed to create release');
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <div className="p-8">Loading...</div>;
  if (!service) return <div className="p-8">Service not found</div>;

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-6">
          <h1 className="text-2xl font-light">Create Release</h1>
        </div>
      </div>
      <div className="max-w-4xl mx-auto px-8 py-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Version *</label>
              <input
                type="text"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder="e.g., 2.15.0"
                className="w-full border border-gray-300 rounded px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Git Branch *</label>
              {hasGitIntegration && (availableBranches.length > 0 || branchesLoading) ? (
                <>
                  <select
                    value={branch}
                    onChange={(e) => setBranch(e.target.value)}
                    disabled={branchesLoading}
                    className="w-full border border-gray-300 rounded px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                  >
                    <option value="">{branchesLoading ? 'Loading branches from GitLab…' : 'Select a branch...'}</option>
                    {availableBranches.map(b => <option key={b} value={b}>{b}</option>)}
                  </select>
                  <button
                    type="button"
                    onClick={loadBranches}
                    disabled={branchesLoading}
                    className="mt-1 text-xs text-orange-600 hover:underline disabled:opacity-50"
                  >
                    ⟳ Обновить список веток из GitLab
                  </button>
                </>
              ) : (
                <>
                  <input
                    type="text"
                    value={branch}
                    onChange={(e) => setBranch(e.target.value)}
                    placeholder="release/2.15.0"
                    className="w-full border border-gray-300 rounded px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                  {branchesError && (
                    <p className="mt-1 text-xs text-red-600 flex items-center gap-1">
                      <AlertCircle size={12} /> Ветки из GitLab не загрузились: {branchesError}
                    </p>
                  )}
                  {!hasGitIntegration && (
                    <p className="mt-1 text-xs text-gray-500">Git-интеграция не настроена — введите ветку вручную.</p>
                  )}
                </>
              )}
            </div>
            <div className="flex gap-3 pt-4">
              <button onClick={() => navigate(`/service/${serviceId}`)} className="px-4 py-2 border border-gray-300 rounded text-sm">Cancel</button>
              <button
                onClick={handleCreate}
                disabled={!version || !branch || creating}
                className="px-6 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white rounded text-sm font-medium"
              >
                {creating ? 'Creating...' : 'Create Release'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
