import { useState, useEffect, useRef } from 'react';
import { api } from '../lib/api';
import { useSettings } from '../contexts/SettingsContext';
import { 
  Key, 
  Plus, 
  Edit2, 
  Trash2, 
  RefreshCw, 
  AlertCircle, 
  Loader2,
  Eye,
  EyeOff,
  Copy,
  Check,
  X
} from 'lucide-react';

interface VaultSecret {
  path: string;
  name?: string;
  isDir?: boolean;
  data: Record<string, any>;
  metadata?: {
    created_time: string;
    updated_time: string;
    version: number;
  };
}

export default function VaultSecretsPage() {
  const { settings } = useSettings();
  const [secrets, setSecrets] = useState<VaultSecret[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingSecret, setEditingSecret] = useState<VaultSecret | null>(null);
  const [newSecret, setNewSecret] = useState({
    path: '',
    data: {} as Record<string, string>,
  });
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [visibleSecrets, setVisibleSecrets] = useState<Set<string>>(new Set());
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  // currentPath is what the user types; committedPath is what we actually query.
  // Splitting them prevents an API request on every keystroke — browsing happens
  // via the "Browse" button / Enter or after a short idle debounce.
  const [currentPath, setCurrentPath] = useState('secret');
  const [committedPath, setCommittedPath] = useState('secret');
  const debounceRef = useRef<number | null>(null);

  const hasVaultIntegration = settings.vaultIntegrations.length > 0 && settings.vaultIntegrations[0].enabled;

  useEffect(() => {
    if (hasVaultIntegration) {
      loadSecrets();
    } else {
      setLoading(false);
    }
  }, [committedPath, hasVaultIntegration]);

  const handlePathChange = (value: string) => {
    setCurrentPath(value);
    // Debounced commit: fire the request only after the user stops typing.
    if (debounceRef.current) window.clearTimeout(debounceRef.current);
    debounceRef.current = window.setTimeout(() => setCommittedPath(value.trim() || 'secret'), 500);
  };

  const browseTo = (path: string) => {
    if (debounceRef.current) window.clearTimeout(debounceRef.current);
    setCurrentPath(path);
    setCommittedPath(path);
  };

  const loadSecrets = async () => {
    try {
      setLoading(true);
      setError(null);
      const base = committedPath.trim().replace(/\/+$/, '') || 'secret';
      // Backend returns a list of key names at the path; fetch full data for each.
      const keys = await api.listVaultSecrets(base);
      const entries = await Promise.all(
        (keys || []).map(async (rawKey: string) => {
          const isDir = rawKey.endsWith('/');
          const key = rawKey.replace(/\/+$/, '');
          const fullPath = base === 'secret' ? key : `${base}/${key}`;
          if (isDir) {
            // Sub-directory: show it as a navigable entry without fetching data.
            return { path: fullPath, name: key, isDir, data: {} } as VaultSecret;
          }
          try {
            const secret = await api.getVaultSecret(fullPath);
            return {
              path: fullPath,
              name: key.split('/').pop() || key,
              isDir: false,
              data: secret?.data ?? secret ?? {},
              metadata: secret?.metadata,
            } as VaultSecret;
          } catch {
            return { path: fullPath, name: key, isDir: false, data: {} } as VaultSecret;
          }
        })
      );
      setSecrets(entries);
    } catch (err) {
      console.error('Failed to load Vault secrets:', err);
      setError('Failed to load secrets. Please check Vault integration.');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!newSecret.path || Object.keys(newSecret.data).length === 0) {
      alert('Please fill in path and at least one key-value pair');
      return;
    }

    try {
      await api.createVaultSecret(newSecret.path, newSecret.data);
      await loadSecrets();
      setShowCreateModal(false);
      setNewSecret({ path: '', data: {} });
      setNewKey('');
      setNewValue('');
    } catch (err) {
      console.error('Failed to create secret:', err);
      alert('Failed to create secret');
    }
  };

  const handleUpdate = async () => {
    if (!editingSecret) return;

    try {
      await api.updateVaultSecret(editingSecret.path, editingSecret.data);
      await loadSecrets();
      setEditingSecret(null);
    } catch (err) {
      console.error('Failed to update secret:', err);
      alert('Failed to update secret');
    }
  };

  const handleDelete = async (path: string) => {
    if (!confirm(`Are you sure you want to delete secret at path: ${path}?`)) return;

    try {
      await api.deleteVaultSecret(path);
      await loadSecrets();
    } catch (err) {
      console.error('Failed to delete secret:', err);
      alert('Failed to delete secret');
    }
  };

  const addKeyValue = () => {
    if (!newKey.trim()) return;
    setNewSecret(prev => ({
      ...prev,
      data: { ...prev.data, [newKey]: newValue },
    }));
    setNewKey('');
    setNewValue('');
  };

  const removeKeyValue = (key: string) => {
    setNewSecret(prev => {
      const newData = { ...prev.data };
      delete newData[key];
      return { ...prev, data: newData };
    });
  };

  const toggleSecretVisibility = (key: string) => {
    setVisibleSecrets(prev => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  const copyToClipboard = async (value: string, key: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="w-8 h-8 animate-spin text-orange-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <p className="text-red-600 mb-4">{error}</p>
          {!hasVaultIntegration && (
            <div className="bg-orange-50 border border-orange-200 rounded-lg p-4 mb-6 max-w-md">
              <p className="text-sm text-orange-900 mb-2">
                Vault integration is not configured. Please configure it in Settings.
              </p>
              <a href="#/settings" className="text-sm text-orange-600 hover:text-orange-700 font-medium">
                Go to Settings →
              </a>
            </div>
          )}
          <button
            onClick={loadSecrets}
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
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Vault</p>
          <h1 className="text-2xl font-light">Vault Secrets</h1>
          <p className="text-gray-400 mt-1 text-sm">Manage secrets stored in HashiCorp Vault</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        {/* Actions */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Current Path</label>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={currentPath}
                  onChange={(e) => handlePathChange(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') browseTo(currentPath.trim() || 'secret'); }}
                  className="px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:border-orange-500"
                  placeholder="secret"
                />
                <button
                  onClick={() => browseTo(currentPath.trim() || 'secret')}
                  className="px-3 py-2 text-sm border border-gray-300 rounded hover:bg-gray-50 text-gray-700"
                  title="Browse this path"
                >
                  Browse
                </button>
              </div>
            </div>
            <button
              onClick={loadSecrets}
              className="p-2 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
              title="Refresh"
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
          >
            <Plus className="w-4 h-4" />
            Create Secret
          </button>
        </div>

        {/* Secrets List */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-3 border-b border-gray-200">
            <span className="text-sm text-gray-500">{secrets.length} secrets</span>
          </div>
          
          {secrets.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <Key className="w-16 h-16 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500 mb-2">No secrets found</p>
              <p className="text-sm text-gray-400 mb-4">Create your first secret to get started</p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded text-sm font-medium"
              >
                Create Secret
              </button>
            </div>
          ) : (
            <div className="divide-y divide-gray-100">
              {secrets.map((secret) => (
                <div key={secret.path} className="px-6 py-4 hover:bg-gray-50">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3
                        className={`text-sm font-medium text-gray-900 font-mono ${secret.isDir ? 'cursor-pointer hover:text-orange-600' : ''}`}
                        onClick={() => { if (secret.isDir) browseTo(secret.path); }}
                        title={secret.isDir ? 'Open directory' : undefined}
                      >
                        {secret.name || secret.path}{secret.isDir ? '/' : ''}
                      </h3>
                      {!secret.isDir && secret.metadata && (
                        <div className="flex items-center gap-4 mt-1 text-xs text-gray-500">
                          <span>Version: {secret.metadata.version}</span>
                          <span>Updated: {new Date(secret.metadata.updated_time).toLocaleString()}</span>
                        </div>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      {secret.isDir ? (
                        <button
                          onClick={() => browseTo(secret.path)}
                          className="p-1.5 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
                          title="Open"
                        >
                          <RefreshCw className="w-4 h-4" />
                        </button>
                      ) : (
                        <>
                          <button
                            onClick={() => setEditingSecret(secret)}
                            className="p-1.5 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
                            title="Edit"
                          >
                            <Edit2 className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => handleDelete(secret.path)}
                            className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
                            title="Delete"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </>
                      )}
                    </div>
                  </div>
                  
                  <div className="space-y-2">
                    {Object.entries(secret.data).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2 bg-gray-50 rounded px-3 py-2">
                        <span className="text-xs font-medium text-gray-700 min-w-[120px]">{key}:</span>
                        <span className="text-xs text-gray-600 font-mono flex-1">
                          {visibleSecrets.has(`${secret.path}-${key}`) ? (
                            String(value)
                          ) : (
                            '••••••••••••'
                          )}
                        </span>
                        <button
                          onClick={() => toggleSecretVisibility(`${secret.path}-${key}`)}
                          className="p-1 text-gray-400 hover:text-gray-600"
                          title={visibleSecrets.has(`${secret.path}-${key}`) ? 'Hide' : 'Show'}
                        >
                          {visibleSecrets.has(`${secret.path}-${key}`) ? (
                            <EyeOff className="w-3 h-3" />
                          ) : (
                            <Eye className="w-3 h-3" />
                          )}
                        </button>
                        <button
                          onClick={() => copyToClipboard(String(value), `${secret.path}-${key}`)}
                          className="p-1 text-gray-400 hover:text-gray-600"
                          title="Copy"
                        >
                          {copiedKey === `${secret.path}-${key}` ? (
                            <Check className="w-3 h-3 text-green-500" />
                          ) : (
                            <Copy className="w-3 h-3" />
                          )}
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Create Secret Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={() => setShowCreateModal(false)} />
          <div className="relative bg-white rounded-lg shadow-2xl w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Create New Secret</h2>
              <button onClick={() => setShowCreateModal(false)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Path <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={newSecret.path}
                  onChange={(e) => setNewSecret({ ...newSecret, path: e.target.value })}
                  placeholder="secret/my-app/database"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Path where the secret will be stored in Vault
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Key-Value Pairs <span className="text-red-500">*</span>
                </label>
                
                {Object.keys(newSecret.data).length > 0 && (
                  <div className="space-y-2 mb-3">
                    {Object.entries(newSecret.data).map(([key, value]) => (
                      <div key={key} className="flex items-center gap-2 bg-gray-50 rounded px-3 py-2">
                        <span className="text-sm font-medium text-gray-700 min-w-[120px]">{key}:</span>
                        <span className="text-sm text-gray-600 flex-1">{String(value)}</span>
                        <button
                          onClick={() => removeKeyValue(key)}
                          className="p-1 text-gray-400 hover:text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newKey}
                    onChange={(e) => setNewKey(e.target.value)}
                    placeholder="Key"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                  <input
                    type="text"
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    placeholder="Value"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                  />
                  <button
                    onClick={addKeyValue}
                    disabled={!newKey.trim()}
                    className="px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white rounded-lg"
                  >
                    Add
                  </button>
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
                  onClick={handleCreate}
                  disabled={!newSecret.path || Object.keys(newSecret.data).length === 0}
                  className="flex-1 px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white rounded-lg font-medium"
                >
                  Create Secret
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Edit Secret Modal */}
      {editingSecret && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={() => setEditingSecret(null)} />
          <div className="relative bg-white rounded-lg shadow-2xl w-full max-w-2xl p-6 max-h-[90vh] overflow-y-auto">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-gray-900">Edit Secret</h2>
              <button onClick={() => setEditingSecret(null)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Secret Path
                </label>
                <input
                  type="text"
                  value={editingSecret.path}
                  disabled
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Key-Value Pairs
                </label>
                
                <div className="space-y-2 mb-3">
                  {Object.entries(editingSecret.data).map(([key, value]) => (
                    <div key={key} className="flex items-center gap-2">
                      <input
                        type="text"
                        value={key}
                        disabled
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
                      />
                      <input
                        type="text"
                        value={String(value)}
                        onChange={(e) => {
                          setEditingSecret(prev => {
                            if (!prev) return null;
                            return {
                              ...prev,
                              data: {
                                ...prev.data,
                                [key]: e.target.value,
                              },
                            };
                          });
                        }}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"
                      />
                      <button
                        onClick={() => {
                          setEditingSecret(prev => {
                            if (!prev) return null;
                            const newData = { ...prev.data };
                            delete newData[key];
                            return { ...prev, data: newData };
                          });
                        }}
                        className="p-2 text-gray-400 hover:text-red-600"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>

                <button
                  onClick={() => {
                    const newKey = prompt('Enter new key:');
                    if (newKey && editingSecret) {
                      setEditingSecret({
                        ...editingSecret,
                        data: { ...editingSecret.data, [newKey]: '' },
                      });
                    }
                  }}
                  className="w-full px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 text-sm"
                >
                  + Add New Key
                </button>
              </div>

              <div className="flex gap-3 pt-4 border-t border-gray-200">
                <button
                  onClick={() => setEditingSecret(null)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleUpdate}
                  className="flex-1 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-lg font-medium"
                >
                  Update Secret
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
