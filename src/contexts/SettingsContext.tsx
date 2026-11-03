import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { api } from '../lib/api';

export interface GitIntegration {
  id: string;
  name: string;
  type: 'gitlab' | 'github' | 'bitbucket';
  url: string;
  token: string;
  enabled: boolean;
  lastSync?: string;
}

export interface JiraIntegration {
  id: string;
  name: string;
  url: string;
  username: string;
  apiToken: string;
  projectKey: string;
  enabled: boolean;
  lastSync?: string;
}

export interface NotificationIntegration {
  id: string;
  name: string;
  type: 'slack' | 'telegram' | 'email';
  webhookUrl: string;
  enabled: boolean;
  channels: string[];
  lastSync?: string;
}

export interface ArgoCDIntegration {
  id: string;
  name: string;
  url: string;
  token: string;
  enabled: boolean;
  lastSync?: string;
}

export interface KubernetesCluster {
  id: string;
  name: string;
  server: string;
  namespace: string;
  environment: 'production' | 'staging' | 'development';
  enabled: boolean;
}

export interface GrafanaIntegration {
  id: string;
  name: string;
  url: string;
  apiKey: string;
  enabled: boolean;
  lastSync?: string;
}

export interface GraylogIntegration {
  id: string;
  name: string;
  url: string;
  username: string;
  password: string;
  enabled: boolean;
  lastSync?: string;
}

export interface VaultIntegration {
  id: string;
  name: string;
  url: string;
  token: string;
  enabled: boolean;
  lastSync?: string;
}

export interface AuthIntegration {
  id: string;
  name: string;
  provider: 'keycloak' | 'github' | 'gitlab' | 'ldap';
  url: string;
  realm?: string;
  clientId?: string;
  clientSecret?: string;
  bindDn?: string;
  bindPassword?: string;
  baseDn?: string;
  enabled: boolean;
}

export interface Settings {
  gitIntegrations: GitIntegration[];
  jiraIntegrations: JiraIntegration[];
  notificationIntegrations: NotificationIntegration[];
  argoIntegrations: ArgoCDIntegration[];
  kubernetesClusters: KubernetesCluster[];
  grafanaIntegrations: GrafanaIntegration[];
  graylogIntegrations: GraylogIntegration[];
  vaultIntegrations: VaultIntegration[];
  keycloak?: {
    url: string;
    realm: string;
    clientId: string;
  };
  authIntegrations: AuthIntegration[];
  general: {
    companyName: string;
    defaultBranch: string;
    autoCreateRelease: boolean;
    requireApprovals: boolean;
    minApprovals: number;
  };
}

interface SettingsContextType {
  settings: Settings;
  loading: boolean;
  error: string | null;
  reloadIntegrations: () => Promise<void>;
  updateSettings: (data: Partial<Settings>) => void;
  updateGitIntegration: (id: string, data: Partial<GitIntegration>) => void;
  addGitIntegration: (data: Omit<GitIntegration, 'id'>) => Promise<boolean>;
  deleteGitIntegration: (id: string) => Promise<void>;
  updateJiraIntegration: (id: string, data: Partial<JiraIntegration>) => void;
  addJiraIntegration: (data: Omit<JiraIntegration, 'id'>) => Promise<boolean>;
  deleteJiraIntegration: (id: string) => Promise<void>;
  updateNotificationIntegration: (id: string, data: Partial<NotificationIntegration>) => void;
  addNotificationIntegration: (data: Omit<NotificationIntegration, 'id'>) => Promise<boolean>;
  deleteNotificationIntegration: (id: string) => Promise<void>;
  updateArgoIntegration: (id: string, data: Partial<ArgoCDIntegration>) => void;
  addArgoIntegration: (data: Omit<ArgoCDIntegration, 'id'>) => Promise<boolean>;
  deleteArgoIntegration: (id: string) => Promise<void>;
  updateKubernetesCluster: (id: string, data: Partial<KubernetesCluster>) => void;
  addKubernetesCluster: (data: Omit<KubernetesCluster, 'id'>) => Promise<boolean>;
  deleteKubernetesCluster: (id: string) => Promise<void>;
  updateGrafanaIntegration: (id: string, data: Partial<GrafanaIntegration>) => void;
  addGrafanaIntegration: (data: Omit<GrafanaIntegration, 'id'>) => Promise<boolean>;
  deleteGrafanaIntegration: (id: string) => Promise<void>;
  updateGraylogIntegration: (id: string, data: Partial<GraylogIntegration>) => void;
  addGraylogIntegration: (data: Omit<GraylogIntegration, 'id'>) => Promise<boolean>;
  deleteGraylogIntegration: (id: string) => Promise<void>;
  updateVaultIntegration: (id: string, data: Partial<VaultIntegration>) => void;
  addVaultIntegration: (data: Omit<VaultIntegration, 'id'>) => Promise<boolean>;
  deleteVaultIntegration: (id: string) => Promise<void>;
  addAuthIntegration: (data: Omit<AuthIntegration, 'id'>) => Promise<boolean>;
  updateAuthIntegration: (id: string, data: Partial<AuthIntegration>) => void;
  deleteAuthIntegration: (id: string) => Promise<void>;
  updateGeneralSettings: (data: Partial<Settings['general']>) => void;
  testConnection: (type: string, id: string) => Promise<boolean>;
  syncIntegration: (type: string, id: string) => Promise<boolean>;
}

const defaultSettings: Settings = {
  gitIntegrations: [],
  jiraIntegrations: [],
  notificationIntegrations: [],
  argoIntegrations: [],
  kubernetesClusters: [],
  grafanaIntegrations: [],
  graylogIntegrations: [],
  vaultIntegrations: [],
  authIntegrations: [],
  general: {
    companyName: 'Company Inc.',
    defaultBranch: 'main',
    autoCreateRelease: false,
    requireApprovals: true,
    minApprovals: 2,
  },
};

const SETTINGS_STORAGE_KEY = 'release_portal_settings';

// ---- Backend model.Integration <-> frontend mapping ------------------------

interface BackendIntegration {
  id: string;
  type: string;
  name: string;
  url: string;
  enabled: boolean;
  config: Record<string, any>;
  createdAt?: string;
  updatedAt?: string;
}

function toBackendPayload(frontendType: string, data: any): any {
  const { id, lastSync, ...rest } = data;
  const base = {
    id: id || undefined, // backend generates UUID if empty
    type: frontendType,
    name: rest.name || '',
    url: rest.url || rest.server || '',
    enabled: rest.enabled !== false,
  };

  switch (frontendType) {
    case 'gitlab':
    case 'github':
      return { ...base, config: { token: rest.token || '' } };
    case 'bitbucket':
      return { ...base, config: { username: rest.username || '', password: rest.password || rest.token || '' } };
    case 'jira':
      return {
        ...base,
        config: {
          username: rest.username || '',
          token: rest.apiToken || rest.token || '',
          projectKey: rest.projectKey || '',
        },
      };
    case 'argocd':
      return { ...base, config: { token: rest.token || '' } };
    case 'kubernetes':
      return {
        ...base,
        config: {
          token: rest.token || '',
          namespace: rest.namespace || '',
          environment: rest.environment || '',
        },
      };
    case 'grafana':
      return { ...base, config: { apiKey: rest.apiKey || rest.token || '' } };
    case 'graylog':
      return { ...base, config: { username: rest.username || '', password: rest.password || '' } };
    case 'vault':
      return { ...base, config: { token: rest.token || '' } };
    case 'auth': {
      const provider = rest.provider || 'keycloak';
      const defaultUrl =
        provider === 'github' ? 'https://github.com' :
        provider === 'gitlab' ? 'https://gitlab.com' : '';
      return {
        ...base,
        url: rest.url || defaultUrl,
        type: 'auth',
        config: {
          provider,
          realm: rest.realm || '',
          clientId: rest.clientId || '',
          clientSecret: rest.clientSecret || '',
          bindDn: rest.bindDn || '',
          bindPassword: rest.bindPassword || '',
          baseDn: rest.baseDn || '',
        },
      };
    }
    case 'notification':
      return {
        ...base,
        config: {
          subtype: rest.type || 'slack',
          webhookUrl: rest.webhookUrl || '',
          channels: rest.channels || [],
        },
      };
    default:
      return { ...base, config: {} };
  }
}

function mapFromBackend(backendType: string, b: BackendIntegration): any {
  const cfg = b.config || {};
  const common = {
    id: b.id,
    name: b.name,
    enabled: b.enabled,
    lastSync: b.updatedAt,
  };
  switch (backendType) {
    case 'gitlab':
    case 'github':
    case 'bitbucket':
      return { ...common, type: backendType, url: b.url, token: cfg.token || cfg.password || '' };
    case 'jira':
      return {
        ...common,
        url: b.url,
        username: cfg.username || '',
        apiToken: cfg.token || '',
        projectKey: cfg.projectKey || '',
      };
    case 'argocd':
      return { ...common, url: b.url, token: cfg.token || '' };
    case 'kubernetes':
      return {
        ...common,
        server: b.url,
        namespace: cfg.namespace || 'default',
        environment: cfg.environment || 'production',
      };
    case 'grafana':
      return { ...common, url: b.url, apiKey: cfg.apiKey || '' };
    case 'graylog':
      return { ...common, url: b.url, username: cfg.username || '', password: cfg.password || '' };
    case 'vault':
      return { ...common, url: b.url, token: cfg.token || '' };
    case 'auth': {
      const provider = cfg.provider || 'keycloak';
      return {
        ...common,
        provider,
        url: b.url,
        realm: cfg.realm || '',
        clientId: cfg.clientId || '',
        clientSecret: cfg.clientSecret || '',
        bindDn: cfg.bindDn || '',
        bindPassword: cfg.bindPassword || '',
        baseDn: cfg.baseDn || '',
      };
    }
    case 'notification':
      return {
        ...common,
        type: cfg.subtype || 'slack',
        webhookUrl: cfg.webhookUrl || '',
        channels: cfg.channels || [],
      };
    default:
      return { ...common, url: b.url };
  }
}

const GIT_TYPES = ['gitlab', 'github', 'bitbucket'];

// -----------------------------------------------------------------------------

const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<Settings>(() => {
    // General/local-only settings still live in localStorage; integrations are
    // loaded from the backend API below.
    try {
      const saved = localStorage.getItem(SETTINGS_STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        return {
          ...defaultSettings,
          ...parsed,
          general: { ...defaultSettings.general, ...(parsed.general || {}) },
        };
      }
    } catch {
      /* corrupted storage – fall back to defaults */
    }
    return defaultSettings;
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Persist only non-integration parts (general settings) locally.
  useEffect(() => {
    const { gitIntegrations, jiraIntegrations, notificationIntegrations, argoIntegrations,
      kubernetesClusters, grafanaIntegrations, graylogIntegrations, vaultIntegrations,
      authIntegrations,
      ...persistable } = settings;
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(persistable));
  }, [settings]);

  const reloadIntegrations = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const list: BackendIntegration[] = await api.listIntegrations();
      const next: Settings = { ...settings };
      next.gitIntegrations = [];
      next.jiraIntegrations = [];
      next.notificationIntegrations = [];
      next.argoIntegrations = [];
      next.kubernetesClusters = [];
      next.grafanaIntegrations = [];
      next.graylogIntegrations = [];
      next.vaultIntegrations = [];
      next.authIntegrations = [];
      for (const item of list || []) {
        const mapped = mapFromBackend(item.type, item);
        if (GIT_TYPES.includes(item.type)) next.gitIntegrations.push(mapped);
        else if (item.type === 'jira') next.jiraIntegrations.push(mapped);
        else if (item.type === 'notification') next.notificationIntegrations.push(mapped);
        else if (item.type === 'argocd') next.argoIntegrations.push(mapped);
        else if (item.type === 'kubernetes') next.kubernetesClusters.push(mapped);
        else if (item.type === 'grafana') next.grafanaIntegrations.push(mapped);
        else if (item.type === 'graylog') next.graylogIntegrations.push(mapped);
        else if (item.type === 'vault') next.vaultIntegrations.push(mapped);
        else if (item.type === 'auth') next.authIntegrations.push(mapped);
      }
      setSettings(next);
    } catch (e: any) {
      console.error('Failed to load integrations from backend:', e);
      setError(e?.message || 'Failed to load integrations');
    } finally {
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    reloadIntegrations();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const updateSettings = (data: Partial<Settings>) => {
    setSettings(prev => ({ ...prev, ...data }));
  };

  // Generic CRUD helpers backed by the REST API -------------------------------

  const apiAdd = useCallback(
    async (frontendType: string, listKey: keyof Settings, data: any): Promise<boolean> => {
      try {
        const payload = toBackendPayload(frontendType, data);
        const created: BackendIntegration = await api.createIntegration(payload);
        const mapped = mapFromBackend(created.type || frontendType, created);
        setSettings(prev => ({
          ...prev,
          [listKey]: [...(prev[listKey] as any[]), mapped],
        }) as Settings);
        return true;
      } catch (e: any) {
        console.error('Failed to create integration:', e);
        alert(`❌ Ошибка сохранения интеграции в бэкенд: ${e?.message || e}`);
        return false;
      }
    },
    []
  );

  const apiUpdate = useCallback(
    async (frontendType: string, listKey: keyof Settings, id: string, data: any) => {
      // Optimistic local update so UI stays responsive.
      setSettings(prev => ({
        ...prev,
        [listKey]: (prev[listKey] as any[]).map(item =>
          item.id === id ? { ...item, ...data } : item
        ),
      }) as Settings);
      try {
        const existing = await api.getIntegration(id).catch(() => null);
        const merged = { ...(existing || {}), ...data, id };
        const payload = toBackendPayload(frontendType, merged);
        await api.updateIntegration(id, payload);
      } catch (e: any) {
        console.error('Failed to update integration:', e);
        alert(`❌ Ошибка обновления интеграции: ${e?.message || e}`);
      }
    },
    []
  );

  const apiDelete = useCallback(
    async (listKey: keyof Settings, id: string) => {
      try {
        await api.deleteIntegration(id);
        setSettings(prev => ({
          ...prev,
          [listKey]: (prev[listKey] as any[]).filter(item => item.id !== id),
        }) as Settings);
      } catch (e: any) {
        console.error('Failed to delete integration:', e);
        alert(`❌ Ошибка удаления интеграции: ${e?.message || e}`);
      }
    },
    []
  );

  // Git -----------------------------------------------------------------------

  const resolveGitType = (id: string, data?: any): string => {
    if (data?.type) return data.type;
    const found = settings.gitIntegrations.find(g => g.id === id);
    return found?.type || 'gitlab';
  };

  const addGitIntegration = async (data: Omit<GitIntegration, 'id'>) => {
    return apiAdd(data.type || 'gitlab', 'gitIntegrations', data);
  };

  const updateGitIntegration = (id: string, data: Partial<GitIntegration>) => {
    apiUpdate(resolveGitType(id, data), 'gitIntegrations', id, data);
  };

  const deleteGitIntegration = async (id: string) => {
    await apiDelete('gitIntegrations', id);
  };

  // Jira ----------------------------------------------------------------------

  const addJiraIntegration = async (data: Omit<JiraIntegration, 'id'>) =>
    apiAdd('jira', 'jiraIntegrations', data);

  const updateJiraIntegration = (id: string, data: Partial<JiraIntegration>) =>
    apiUpdate('jira', 'jiraIntegrations', id, data);

  const deleteJiraIntegration = async (id: string) => apiDelete('jiraIntegrations', id);

  // Notifications ---------------------------------------------------------------

  const addNotificationIntegration = async (data: Omit<NotificationIntegration, 'id'>) =>
    apiAdd('notification', 'notificationIntegrations', data);

  const updateNotificationIntegration = (id: string, data: Partial<NotificationIntegration>) =>
    apiUpdate('notification', 'notificationIntegrations', id, data);

  const deleteNotificationIntegration = async (id: string) =>
    apiDelete('notificationIntegrations', id);

  // ArgoCD ----------------------------------------------------------------------

  const addArgoIntegration = async (data: Omit<ArgoCDIntegration, 'id'>) =>
    apiAdd('argocd', 'argoIntegrations', data);

  const updateArgoIntegration = (id: string, data: Partial<ArgoCDIntegration>) =>
    apiUpdate('argocd', 'argoIntegrations', id, data);

  const deleteArgoIntegration = async (id: string) => apiDelete('argoIntegrations', id);

  // Kubernetes ------------------------------------------------------------------

  const addKubernetesCluster = async (data: Omit<KubernetesCluster, 'id'>) =>
    apiAdd('kubernetes', 'kubernetesClusters', data);

  const updateKubernetesCluster = (id: string, data: Partial<KubernetesCluster>) =>
    apiUpdate('kubernetes', 'kubernetesClusters', id, data);

  const deleteKubernetesCluster = async (id: string) => apiDelete('kubernetesClusters', id);

  // Grafana ---------------------------------------------------------------------

  const addGrafanaIntegration = async (data: Omit<GrafanaIntegration, 'id'>) =>
    apiAdd('grafana', 'grafanaIntegrations', data);

  const updateGrafanaIntegration = (id: string, data: Partial<GrafanaIntegration>) =>
    apiUpdate('grafana', 'grafanaIntegrations', id, data);

  const deleteGrafanaIntegration = async (id: string) => apiDelete('grafanaIntegrations', id);

  // Graylog ---------------------------------------------------------------------

  const addGraylogIntegration = async (data: Omit<GraylogIntegration, 'id'>) =>
    apiAdd('graylog', 'graylogIntegrations', data);

  const updateGraylogIntegration = (id: string, data: Partial<GraylogIntegration>) =>
    apiUpdate('graylog', 'graylogIntegrations', id, data);

  const deleteGraylogIntegration = async (id: string) => apiDelete('graylogIntegrations', id);

  // Vault -------------------------------------------------------------------------

  const addVaultIntegration = async (data: Omit<VaultIntegration, 'id'>) =>
    apiAdd('vault', 'vaultIntegrations', data);

  const updateVaultIntegration = (id: string, data: Partial<VaultIntegration>) =>
    apiUpdate('vault', 'vaultIntegrations', id, data);

  const deleteVaultIntegration = async (id: string) => apiDelete('vaultIntegrations', id);

  // Auth providers --------------------------------------------------------------

  const addAuthIntegration = async (data: Omit<AuthIntegration, 'id'>) =>
    apiAdd('auth', 'authIntegrations', data);

  const updateAuthIntegration = (id: string, data: Partial<AuthIntegration>) =>
    apiUpdate('auth', 'authIntegrations', id, data);

  const deleteAuthIntegration = async (id: string) => apiDelete('authIntegrations', id);

  // General -----------------------------------------------------------------------

  const updateGeneralSettings = (data: Partial<Settings['general']>) => {
    setSettings(prev => ({
      ...prev,
      general: { ...prev.general, ...data },
    }));
  };

  // Test / Sync via backend ---------------------------------------------------------

  const testConnection = async (_type: string, id: string): Promise<boolean> => {
    try {
      const res: any = await api.testIntegration(id);
      // Backend returns {success, message} with HTTP 200 even on failed tests.
      if (res && res.success === false) {
        console.warn('Connection test failed:', res.message);
        return false;
      }
      return true;
    } catch (error) {
      console.error('Connection test failed:', error);
      return false;
    }
  };

  const syncIntegration = async (_type: string, id: string): Promise<boolean> => {
    try {
      const res: any = await api.syncIntegration(id);
      if (res && res.success === false) {
        console.warn('Integration sync failed:', res.message);
        return false;
      }
      // Refresh lastSync timestamps from the backend.
      await reloadIntegrations();
      return true;
    } catch (error) {
      console.error('Integration sync failed:', error);
      return false;
    }
  };

  return (
    <SettingsContext.Provider value={{
      settings,
      loading,
      error,
      reloadIntegrations,
      updateSettings,
      updateGitIntegration,
      addGitIntegration,
      deleteGitIntegration,
      updateJiraIntegration,
      addJiraIntegration,
      deleteJiraIntegration,
      updateNotificationIntegration,
      addNotificationIntegration,
      deleteNotificationIntegration,
      updateArgoIntegration,
      addArgoIntegration,
      deleteArgoIntegration,
      updateKubernetesCluster,
      addKubernetesCluster,
      deleteKubernetesCluster,
      updateGrafanaIntegration,
      addGrafanaIntegration,
      deleteGrafanaIntegration,
      updateGraylogIntegration,
      addGraylogIntegration,
      deleteGraylogIntegration,
      updateVaultIntegration,
      addVaultIntegration,
      deleteVaultIntegration,
      addAuthIntegration,
      updateAuthIntegration,
      deleteAuthIntegration,
      updateGeneralSettings,
      testConnection,
      syncIntegration,
    }}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider');
  }
  return context;
}
