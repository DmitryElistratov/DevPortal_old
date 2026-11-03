import { useState } from 'react';
import { useSettings } from '../contexts/SettingsContext';
import { useAuth } from '../contexts/AuthContext';
import {
  GitBranch,
  ExternalLink,
  Bell,
  Settings as SettingsIcon,
  Plus,
  Trash2,
  Check,
  X,
  Loader2,
  TestTube,
  Cloud,
  Server,
  Activity,
  FileText,
  Lock,
  Users,
  RefreshCw,
  AlertCircle,
  Edit2,
  Key,
} from 'lucide-react';

export default function SettingsPage() {
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState<'general' | 'users' | 'auth' | 'git' | 'jira' | 'notifications' | 'argo' | 'kubernetes' | 'grafana' | 'graylog' | 'vault'>('general');

  if (user?.role !== 'admin') {
    return (
      <div className="p-8 text-center">
        <p className="text-gray-500">Доступ только для администраторов</p>
      </div>
    );
  }

  const menuGroups = [
    {
      label: 'Основные',
      items: [
        { id: 'general' as const, label: 'Общие', icon: SettingsIcon },
        { id: 'users' as const, label: 'Пользователи', icon: Users },
        { id: 'auth' as const, label: 'Авторизация', icon: Lock },
      ],
    },
    {
      label: 'Интеграции',
      items: [
        { id: 'git' as const, label: 'Git', icon: GitBranch },
        { id: 'jira' as const, label: 'Jira', icon: ExternalLink },
        { id: 'notifications' as const, label: 'Уведомления', icon: Bell },
      ],
    },
    {
      label: 'Инфраструктура',
      items: [
        { id: 'argo' as const, label: 'ArgoCD', icon: Cloud },
        { id: 'kubernetes' as const, label: 'Kubernetes', icon: Server },
        { id: 'grafana' as const, label: 'Grafana', icon: Activity },
        { id: 'graylog' as const, label: 'Graylog', icon: FileText },
        { id: 'vault' as const, label: 'Vault', icon: Key },
      ],
    },
  ];

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      <div className="bg-gradient-to-r from-[#1a237e] to-[#283593] text-white flex-shrink-0">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-white/60 uppercase tracking-wider mb-1">Home / Settings</p>
          <h1 className="text-2xl font-light">Настройки</h1>
          <p className="text-white/70 mt-1 text-sm">Управление интеграциями и конфигурацией</p>
        </div>
      </div>

      <div className="flex-1 overflow-hidden">
        <div className="max-w-7xl mx-auto h-full flex gap-6 px-8 py-6">
          <div className="w-64 flex-shrink-0 overflow-y-auto">
            <nav className="space-y-6">
              {menuGroups.map((group, groupIndex) => (
                <div key={groupIndex}>
                  <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 px-3">
                    {group.label}
                  </h3>
                  <div className="space-y-1">
                    {group.items.map(item => (
                      <button
                        key={item.id}
                        onClick={() => setActiveTab(item.id)}
                        className={`w-full flex items-center gap-3 px-3 py-2 text-sm rounded-lg transition-colors ${
                          activeTab === item.id
                            ? 'bg-orange-500 text-white'
                            : 'text-gray-700 hover:bg-gray-100'
                        }`}
                      >
                        <item.icon className="w-4 h-4" />
                        <span>{item.label}</span>
                      </button>
                    ))}
                  </div>
                </div>
              ))}
            </nav>
          </div>

          <div className="flex-1 bg-white rounded-lg shadow-sm border border-gray-200 overflow-y-auto">
            <div className="p-6">
              {activeTab === 'general' && <GeneralSettings />}
              {activeTab === 'users' && <UsersSettings />}
              {activeTab === 'auth' && <AuthSettings />}
              {activeTab === 'git' && <GitSettings />}
              {activeTab === 'jira' && <JiraSettings />}
              {activeTab === 'notifications' && <NotificationSettings />}
              {activeTab === 'argo' && <ArgoSettings />}
              {activeTab === 'kubernetes' && <KubernetesSettings />}
              {activeTab === 'grafana' && <GrafanaSettings />}
              {activeTab === 'graylog' && <GraylogSettings />}
              {activeTab === 'vault' && <VaultSettings />}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function GeneralSettings() {
  const { settings, updateGeneralSettings } = useSettings();
  const [localSettings, setLocalSettings] = useState(settings.general);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  const handleChange = (field: string, value: any) => {
    setLocalSettings(prev => ({ ...prev, [field]: value }));
    setHasChanges(true);
    setSaved(false);
  };

  const handleSave = async () => {
    setSaving(true);
    updateGeneralSettings(localSettings);
    await new Promise(resolve => setTimeout(resolve, 500));
    setSaving(false);
    setSaved(true);
    setHasChanges(false);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-base font-semibold text-gray-900 mb-4">Основные настройки</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Название компании</label>
            <input
              type="text"
              value={localSettings.companyName}
              onChange={(e) => handleChange('companyName', e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Ветка по умолчанию</label>
            <input
              type="text"
              value={localSettings.defaultBranch}
              onChange={(e) => handleChange('defaultBranch', e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
            />
          </div>

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-700">Автоматическое создание релиза</p>
              <p className="text-xs text-gray-500">Создавать релиз при мерже в основную ветку</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={localSettings.autoCreateRelease}
                onChange={(e) => handleChange('autoCreateRelease', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-orange-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-500"></div>
            </label>
          </div>

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-700">Требовать одобрения</p>
              <p className="text-xs text-gray-500">Релиз должен быть одобрен перед деплоем</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={localSettings.requireApprovals}
                onChange={(e) => handleChange('requireApprovals', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-orange-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-orange-500"></div>
            </label>
          </div>

          {localSettings.requireApprovals && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Минимальное количество одобрений</label>
              <input
                type="number"
                min="1"
                value={localSettings.minApprovals}
                onChange={(e) => handleChange('minApprovals', parseInt(e.target.value) || 1)}
                className="w-32 border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
          )}
        </div>
      </div>

      <div className="pt-4 border-t border-gray-200 flex items-center gap-3">
        <button
          onClick={handleSave}
          disabled={saving || !hasChanges}
          className={`px-4 py-2 text-white text-sm font-medium rounded ${
            saving || !hasChanges
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-orange-500 hover:bg-orange-600'
          }`}
        >
          {saving ? (
            <span className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              Сохранение...
            </span>
          ) : (
            'Сохранить изменения'
          )}
        </button>
        {saved && (
          <span className="text-sm text-green-600 flex items-center gap-1 animate-pulse">
            <Check className="w-4 h-4" />
            Настройки сохранены
          </span>
        )}
        {hasChanges && !saved && (
          <span className="text-sm text-orange-600 flex items-center gap-1">
            <AlertCircle className="w-4 h-4" />
            Есть несохранённые изменения
          </span>
        )}
      </div>
    </div>
  );
}

function UsersSettings() {
  const [users] = useState([
    { id: '1', username: 'admin', email: 'admin@company.ru', fullName: 'Admin User', role: 'admin', authProvider: 'local', lastLogin: '2024-02-27 14:30' },
    { id: '2', username: 'developer', email: 'dev@company.ru', fullName: 'Developer User', role: 'developer', authProvider: 'local', lastLogin: '2024-02-27 10:15' },
  ]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Пользователи</h3>
          <p className="text-sm text-gray-500 mt-1">Управление пользователями портала</p>
        </div>
      </div>

      <div className="border border-gray-200 rounded-lg overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Пользователь</th>
              <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Email</th>
              <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Роль</th>
              <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Источник</th>
              <th className="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Последний вход</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {users.map(user => (
              <tr key={user.id} className="hover:bg-gray-50">
                <td className="px-4 py-3">
                  <div>
                    <p className="text-sm font-medium text-gray-900">{user.fullName}</p>
                    <p className="text-xs text-gray-500">@{user.username}</p>
                  </div>
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">{user.email}</td>
                <td className="px-4 py-3">
                  <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-700 rounded">{user.role}</span>
                </td>
                <td className="px-4 py-3">
                  <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-700 rounded">{user.authProvider}</span>
                </td>
                <td className="px-4 py-3 text-xs text-gray-500">{user.lastLogin}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function AuthSettings() {
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-base font-semibold text-gray-900 mb-2">Провайдеры аутентификации</h3>
        <p className="text-sm text-gray-500 mb-6">Настройте внешние провайдеры для входа пользователей</p>
      </div>

      <div className="border border-gray-200 rounded-lg p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h4 className="text-sm font-medium text-gray-900">Keycloak (OIDC)</h4>
            <p className="text-xs text-gray-500 mt-1">OpenID Connect провайдер</p>
          </div>
          <span className="text-xs px-2 py-0.5 bg-green-100 text-green-700 rounded">Доступен</span>
        </div>
      </div>

      <div className="border border-gray-200 rounded-lg p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h4 className="text-sm font-medium text-gray-900">LDAP</h4>
            <p className="text-xs text-gray-500 mt-1">Корпоративный каталог</p>
          </div>
          <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">Не настроен</span>
        </div>
      </div>

      <div className="border border-gray-200 rounded-lg p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h4 className="text-sm font-medium text-gray-900">GitHub OAuth</h4>
            <p className="text-xs text-gray-500 mt-1">Вход через GitHub аккаунт</p>
          </div>
          <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">Не настроен</span>
        </div>
      </div>

      <div className="border border-gray-200 rounded-lg p-4">
        <div className="flex items-start justify-between mb-3">
          <div>
            <h4 className="text-sm font-medium text-gray-900">GitLab OAuth</h4>
            <p className="text-xs text-gray-500 mt-1">Вход через GitLab аккаунт</p>
          </div>
          <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">Не настроен</span>
        </div>
      </div>
    </div>
  );
}

function GitSettings() {
  const { settings, addGitIntegration, updateGitIntegration, deleteGitIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    type: 'gitlab' as const,
    url: '',
    token: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    type: 'gitlab' as const,
    url: '',
    token: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.token) return;
    const ok = await addGitIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', type: 'gitlab', url: '', token: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('git', id);
    setTestingId(null);
    if (success) {
      alert('✅ Соединение успешно!');
    } else {
      alert('❌ Ошибка соединения!');
    }
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('git', id);
    setSyncingId(null);
    if (success) {
      alert('✅ Синхронизация завершена!');
    } else {
      alert('❌ Ошибка синхронизации!');
    }
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      type: integration.type,
      url: integration.url,
      token: integration.token || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateGitIntegration(editingId, editIntegration);
    setEditingId(null);
    setEditIntegration({ name: '', type: 'gitlab', url: '', token: '', enabled: true });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Git интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Подключите GitLab, GitHub или Bitbucket</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main GitLab"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Тип</label>
              <select
                value={newIntegration.type}
                onChange={(e) => setNewIntegration({ ...newIntegration, type: e.target.value as any })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              >
                <option value="gitlab">GitLab</option>
                <option value="github">GitHub</option>
                <option value="bitbucket">Bitbucket</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="https://gitlab.company.ru"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Access Token <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.token}
                onChange={(e) => setNewIntegration({ ...newIntegration, token: e.target.value })}
                placeholder="glpat-xxxxxxxxxxxx"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Access Token</label>
              <input
                type="password"
                value={editIntegration.token}
                onChange={(e) => setEditIntegration({ ...editIntegration, token: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => {
                  setEditingId(null);
                  setEditIntegration({ name: '', type: 'gitlab', url: '', token: '', enabled: true });
                }}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.gitIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-orange-100 text-orange-700 rounded uppercase">
                    {integration.type}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateGitIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteGitIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Jira Settings
function JiraSettings() {
  const { settings, addJiraIntegration, updateJiraIntegration, deleteJiraIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    url: '',
    username: '',
    apiToken: '',
    projectKey: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    url: '',
    username: '',
    apiToken: '',
    projectKey: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.username || !newIntegration.apiToken) return;
    const ok = await addJiraIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', url: '', username: '', apiToken: '', projectKey: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('jira', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('jira', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      url: integration.url,
      username: integration.username || '',
      apiToken: integration.apiToken || '',
      projectKey: integration.projectKey || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateJiraIntegration(editingId, editIntegration);
    setEditingId(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Jira интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Подключите Jira для управления задачами</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main Jira"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="https://jira.company.ru"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Username <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.username}
                onChange={(e) => setNewIntegration({ ...newIntegration, username: e.target.value })}
                placeholder="jira-bot"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Token <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.apiToken}
                onChange={(e) => setNewIntegration({ ...newIntegration, apiToken: e.target.value })}
                placeholder="jira-api-token"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Project Key</label>
              <input
                type="text"
                value={newIntegration.projectKey}
                onChange={(e) => setNewIntegration({ ...newIntegration, projectKey: e.target.value })}
                placeholder="PAY,CAT,USR"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Username</label>
              <input
                type="text"
                value={editIntegration.username}
                onChange={(e) => setEditIntegration({ ...editIntegration, username: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Token</label>
              <input
                type="password"
                value={editIntegration.apiToken}
                onChange={(e) => setEditIntegration({ ...editIntegration, apiToken: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Project Key</label>
              <input
                type="text"
                value={editIntegration.projectKey}
                onChange={(e) => setEditIntegration({ ...editIntegration, projectKey: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => setEditingId(null)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.jiraIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded uppercase">
                    jira
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                <p className="text-xs text-gray-500 mb-1">Username: {integration.username}</p>
                {integration.projectKey && (
                  <p className="text-xs text-gray-500 mb-1">Projects: {integration.projectKey}</p>
                )}
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateJiraIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteJiraIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Notification Settings
function NotificationSettings() {
  const { settings, addNotificationIntegration, updateNotificationIntegration, deleteNotificationIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    type: 'slack' as const,
    webhookUrl: '',
    enabled: true,
    channels: [] as string[],
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.webhookUrl) return;
    const ok = await addNotificationIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', type: 'slack', webhookUrl: '', enabled: true, channels: [] });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('notification', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('notification', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Уведомления</h3>
          <p className="text-sm text-gray-500 mt-1">Настройте каналы для уведомлений о релизах</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новое уведомление</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Release Notifications"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Тип</label>
              <select
                value={newIntegration.type}
                onChange={(e) => setNewIntegration({ ...newIntegration, type: e.target.value as any })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              >
                <option value="slack">Slack</option>
                <option value="telegram">Telegram</option>
                <option value="email">Email</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Webhook URL</label>
              <input
                type="url"
                value={newIntegration.webhookUrl}
                onChange={(e) => setNewIntegration({ ...newIntegration, webhookUrl: e.target.value })}
                placeholder="https://hooks.slack.com/..."
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.notificationIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активно' : 'Отключено'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded uppercase">
                    {integration.type}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">Webhook: {integration.webhookUrl}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateNotificationIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteNotificationIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ArgoCD Settings
function ArgoSettings() {
  const { settings, addArgoIntegration, updateArgoIntegration, deleteArgoIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    url: '',
    token: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    url: '',
    token: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.token) return;
    const ok = await addArgoIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', url: '', token: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('argo', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('argo', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      url: integration.url,
      token: integration.token || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateArgoIntegration(editingId, editIntegration);
    setEditingId(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">ArgoCD интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Подключите ArgoCD для управления деплоями в Kubernetes</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main ArgoCD"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="https://argocd.company.ru"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Token <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.token}
                onChange={(e) => setNewIntegration({ ...newIntegration, token: e.target.value })}
                placeholder="argocd-token"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Token</label>
              <input
                type="password"
                value={editIntegration.token}
                onChange={(e) => setEditIntegration({ ...editIntegration, token: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => setEditingId(null)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.argoIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-purple-100 text-purple-700 rounded uppercase">
                    argocd
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateArgoIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteArgoIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Kubernetes Settings
function KubernetesSettings() {
  const { settings, addKubernetesCluster, updateKubernetesCluster, deleteKubernetesCluster, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newCluster, setNewCluster] = useState({
    name: '',
    server: '',
    namespace: 'default',
    environment: 'production' as const,
    enabled: true,
  });
  const [editCluster, setEditCluster] = useState({
    name: '',
    server: '',
    namespace: 'default',
    environment: 'production' as const,
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newCluster.name || !newCluster.server) return;
    const ok = await addKubernetesCluster(newCluster);
    if (!ok) return;
    setNewCluster({ name: '', server: '', namespace: 'default', environment: 'production', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('kubernetes', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('kubernetes', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  const handleEdit = (cluster: any) => {
    setEditingId(cluster.id);
    setEditCluster({
      name: cluster.name,
      server: cluster.server,
      namespace: cluster.namespace || 'default',
      environment: cluster.environment || 'production',
      enabled: cluster.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateKubernetesCluster(editingId, editCluster);
    setEditingId(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Kubernetes кластеры</h3>
          <p className="text-sm text-gray-500 mt-1">Управление кластерами для деплоя приложений</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новый кластер</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newCluster.name}
                onChange={(e) => setNewCluster({ ...newCluster, name: e.target.value })}
                placeholder="production-cluster"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Server URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newCluster.server}
                onChange={(e) => setNewCluster({ ...newCluster, server: e.target.value })}
                placeholder="https://k8s-prod.company.ru:6443"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Namespace</label>
              <input
                type="text"
                value={newCluster.namespace}
                onChange={(e) => setNewCluster({ ...newCluster, namespace: e.target.value })}
                placeholder="default"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Окружение</label>
              <select
                value={newCluster.environment}
                onChange={(e) => setNewCluster({ ...newCluster, environment: e.target.value as any })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              >
                <option value="production">Production</option>
                <option value="staging">Staging</option>
                <option value="development">Development</option>
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование кластера</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editCluster.name}
                onChange={(e) => setEditCluster({ ...editCluster, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Server URL</label>
              <input
                type="url"
                value={editCluster.server}
                onChange={(e) => setEditCluster({ ...editCluster, server: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Namespace</label>
              <input
                type="text"
                value={editCluster.namespace}
                onChange={(e) => setEditCluster({ ...editCluster, namespace: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Окружение</label>
              <select
                value={editCluster.environment}
                onChange={(e) => setEditCluster({ ...editCluster, environment: e.target.value as any })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              >
                <option value="production">Production</option>
                <option value="staging">Staging</option>
                <option value="development">Development</option>
              </select>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => setEditingId(null)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.kubernetesClusters.map(cluster => (
          <div key={cluster.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{cluster.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded font-medium ${
                    cluster.environment === 'production'
                      ? 'bg-green-100 text-green-700'
                      : cluster.environment === 'staging'
                      ? 'bg-yellow-100 text-yellow-700'
                      : 'bg-blue-100 text-blue-700'
                  }`}>
                    {cluster.environment}
                  </span>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    cluster.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {cluster.enabled ? 'Активен' : 'Отключен'}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">Server: {cluster.server}</p>
                <p className="text-xs text-gray-500">Namespace: {cluster.namespace}</p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(cluster.id)}
                  disabled={syncingId === cluster.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === cluster.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(cluster)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(cluster.id)}
                  disabled={testingId === cluster.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === cluster.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateKubernetesCluster(cluster.id, { enabled: !cluster.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={cluster.enabled ? 'Отключить' : 'Включить'}
                >
                  {cluster.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteKubernetesCluster(cluster.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Grafana Settings
function GrafanaSettings() {
  const { settings, addGrafanaIntegration, updateGrafanaIntegration, deleteGrafanaIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    url: '',
    apiKey: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    url: '',
    apiKey: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.apiKey) return;
    const ok = await addGrafanaIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', url: '', apiKey: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('grafana', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('grafana', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      url: integration.url,
      apiKey: integration.apiKey || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateGrafanaIntegration(editingId, editIntegration);
    setEditingId(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Grafana интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Подключите Grafana для просмотра метрик</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main Grafana"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="https://grafana.company.ru"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Key <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.apiKey}
                onChange={(e) => setNewIntegration({ ...newIntegration, apiKey: e.target.value })}
                placeholder="grafana-api-key"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">API Key</label>
              <input
                type="password"
                value={editIntegration.apiKey}
                onChange={(e) => setEditIntegration({ ...editIntegration, apiKey: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => setEditingId(null)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.grafanaIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-orange-100 text-orange-700 rounded uppercase">
                    grafana
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateGrafanaIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteGrafanaIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Graylog Settings
function GraylogSettings() {
  const { settings, addGraylogIntegration, updateGraylogIntegration, deleteGraylogIntegration, testConnection, syncIntegration } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [syncingId, setSyncingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    url: '',
    username: '',
    password: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    url: '',
    username: '',
    password: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.username || !newIntegration.password) return;
    const ok = await addGraylogIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', url: '', username: '', password: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    const success = await testConnection('graylog', id);
    setTestingId(null);
    alert(success ? '✅ Соединение успешно!' : '❌ Ошибка соединения!');
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    const success = await syncIntegration('graylog', id);
    setSyncingId(null);
    alert(success ? '✅ Синхронизация завершена!' : '❌ Ошибка синхронизации!');
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      url: integration.url,
      username: integration.username || '',
      password: integration.password || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateGraylogIntegration(editingId, editIntegration);
    setEditingId(null);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Graylog интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Подключите Graylog для просмотра и анализа логов</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main Graylog"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="https://graylog.company.ru"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Username <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.username}
                onChange={(e) => setNewIntegration({ ...newIntegration, username: e.target.value })}
                placeholder="admin"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Password <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.password}
                onChange={(e) => setNewIntegration({ ...newIntegration, password: e.target.value })}
                placeholder="••••••••"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Username</label>
              <input
                type="text"
                value={editIntegration.username}
                onChange={(e) => setEditIntegration({ ...editIntegration, username: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Password</label>
              <input
                type="password"
                value={editIntegration.password}
                onChange={(e) => setEditIntegration({ ...editIntegration, password: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => setEditingId(null)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.graylogIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                  <span className="text-xs px-2 py-0.5 bg-green-100 text-green-700 rounded uppercase">
                    graylog
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                <p className="text-xs text-gray-500 mb-1">Username: {integration.username}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleSync(integration.id)}
                  disabled={syncingId === integration.id}
                  className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
                  title="Синхронизировать"
                >
                  {syncingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <RefreshCw className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateGraylogIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteGraylogIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function VaultSettings() {
  const { settings, addVaultIntegration, updateVaultIntegration, deleteVaultIntegration, testConnection } = useSettings();
  const [showForm, setShowForm] = useState(false);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newIntegration, setNewIntegration] = useState({
    name: '',
    url: '',
    token: '',
    enabled: true,
  });
  const [editIntegration, setEditIntegration] = useState({
    name: '',
    url: '',
    token: '',
    enabled: true,
  });

  const handleAdd = async () => {
    if (!newIntegration.name || !newIntegration.url || !newIntegration.token) return;
    const ok = await addVaultIntegration(newIntegration);
    if (!ok) return;
    setNewIntegration({ name: '', url: '', token: '', enabled: true });
    setShowForm(false);
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    await testConnection('vault', id);
    setTestingId(null);
  };

  const handleEdit = (integration: any) => {
    setEditingId(integration.id);
    setEditIntegration({
      name: integration.name,
      url: integration.url,
      token: integration.token || '',
      enabled: integration.enabled,
    });
  };

  const handleUpdate = () => {
    if (!editingId) return;
    updateVaultIntegration(editingId, editIntegration);
    setEditingId(null);
    setEditIntegration({ name: '', url: '', token: '', enabled: true });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-semibold text-gray-900">Vault интеграции</h3>
          <p className="text-sm text-gray-500 mt-1">Управление секретами через HashiCorp Vault</p>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
        >
          <Plus className="w-4 h-4" />
          Добавить
        </button>
      </div>

      {showForm && (
        <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Новая интеграция</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название <span className="text-red-500">*</span></label>
              <input
                type="text"
                value={newIntegration.name}
                onChange={(e) => setNewIntegration({ ...newIntegration, name: e.target.value })}
                placeholder="Main Vault"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL <span className="text-red-500">*</span></label>
              <input
                type="url"
                value={newIntegration.url}
                onChange={(e) => setNewIntegration({ ...newIntegration, url: e.target.value })}
                placeholder="http://vault:8200"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Token <span className="text-red-500">*</span></label>
              <input
                type="password"
                value={newIntegration.token}
                onChange={(e) => setNewIntegration({ ...newIntegration, token: e.target.value })}
                placeholder="hvs.xxxxxxxxxxxx"
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleAdd}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Сохранить
              </button>
              <button
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      {editingId && (
        <div className="border border-orange-200 rounded-lg p-4 bg-orange-50">
          <h4 className="text-sm font-medium text-gray-900 mb-4">Редактирование интеграции</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Название</label>
              <input
                type="text"
                value={editIntegration.name}
                onChange={(e) => setEditIntegration({ ...editIntegration, name: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">URL</label>
              <input
                type="url"
                value={editIntegration.url}
                onChange={(e) => setEditIntegration({ ...editIntegration, url: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-700 mb-1">Token</label>
              <input
                type="password"
                value={editIntegration.token}
                onChange={(e) => setEditIntegration({ ...editIntegration, token: e.target.value })}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleUpdate}
                className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
              >
                Обновить
              </button>
              <button
                onClick={() => {
                  setEditingId(null);
                  setEditIntegration({ name: '', url: '', token: '', enabled: true });
                }}
                className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {settings.vaultIntegrations.map(integration => (
          <div key={integration.id} className="border border-gray-200 rounded-lg p-4">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="text-sm font-medium text-gray-900">{integration.name}</h4>
                  <span className={`text-xs px-2 py-0.5 rounded ${
                    integration.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {integration.enabled ? 'Активна' : 'Отключена'}
                  </span>
                </div>
                <p className="text-xs text-gray-500 mb-1">URL: {integration.url}</p>
                {integration.lastSync && (
                  <p className="text-xs text-gray-400">
                    Последняя синхронизация: {new Date(integration.lastSync).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleEdit(integration)}
                  className="p-2 text-gray-400 hover:text-orange-600"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleTest(integration.id)}
                  disabled={testingId === integration.id}
                  className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50"
                  title="Проверить соединение"
                >
                  {testingId === integration.id ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <TestTube className="w-4 h-4" />
                  )}
                </button>
                <button
                  onClick={() => updateVaultIntegration(integration.id, { enabled: !integration.enabled })}
                  className="p-2 text-gray-400 hover:text-green-600"
                  title={integration.enabled ? 'Отключить' : 'Включить'}
                >
                  {integration.enabled ? <X className="w-4 h-4" /> : <Check className="w-4 h-4" />}
                </button>
                <button
                  onClick={() => deleteVaultIntegration(integration.id)}
                  className="p-2 text-gray-400 hover:text-red-600"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
        <h4 className="text-sm font-medium text-orange-900 mb-2">Как использовать Vault</h4>
        <ul className="text-xs text-orange-800 space-y-1">
          <li>• Секреты интеграций хранятся в Vault KV engine</li>
          <li>• Путь: <code className="bg-orange-100 px-1 rounded">secret/integrations/&lt;type&gt;/&lt;id&gt;</code></li>
          <li>• Backend автоматически получает credentials при вызове интеграций</li>
          <li>• Перейдите на страницу <a href="#/vault/secrets" className="text-orange-600 hover:text-orange-700 underline">Vault Secrets</a> для управления секретами</li>
        </ul>
      </div>
    </div>
  );
}
