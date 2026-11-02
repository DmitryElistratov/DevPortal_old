import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
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
  updateSettings: (data: Partial<Settings>) => void;
  updateGitIntegration: (id: string, data: Partial<GitIntegration>) => void;
  addGitIntegration: (data: Omit<GitIntegration, 'id'>) => void;
  deleteGitIntegration: (id: string) => void;
  updateJiraIntegration: (id: string, data: Partial<JiraIntegration>) => void;
  addJiraIntegration: (data: Omit<JiraIntegration, 'id'>) => void;
  deleteJiraIntegration: (id: string) => void;
  updateNotificationIntegration: (id: string, data: Partial<NotificationIntegration>) => void;
  addNotificationIntegration: (data: Omit<NotificationIntegration, 'id'>) => void;
  deleteNotificationIntegration: (id: string) => void;
  updateArgoIntegration: (id: string, data: Partial<ArgoCDIntegration>) => void;
  addArgoIntegration: (data: Omit<ArgoCDIntegration, 'id'>) => void;
  deleteArgoIntegration: (id: string) => void;
  updateKubernetesCluster: (id: string, data: Partial<KubernetesCluster>) => void;
  addKubernetesCluster: (data: Omit<KubernetesCluster, 'id'>) => void;
  deleteKubernetesCluster: (id: string) => void;
  updateGrafanaIntegration: (id: string, data: Partial<GrafanaIntegration>) => void;
  addGrafanaIntegration: (data: Omit<GrafanaIntegration, 'id'>) => void;
  deleteGrafanaIntegration: (id: string) => void;
  updateGraylogIntegration: (id: string, data: Partial<GraylogIntegration>) => void;
  addGraylogIntegration: (data: Omit<GraylogIntegration, 'id'>) => void;
  deleteGraylogIntegration: (id: string) => void;
  updateVaultIntegration: (id: string, data: Partial<VaultIntegration>) => void;
  addVaultIntegration: (data: Omit<VaultIntegration, 'id'>) => void;
  deleteVaultIntegration: (id: string) => void;
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
  general: {
    companyName: 'Company Inc.',
    defaultBranch: 'main',
    autoCreateRelease: false,
    requireApprovals: true,
    minApprovals: 2,
  },
};

const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<Settings>(() => {
    const saved = localStorage.getItem('release_portal_settings');
    if (saved) {
      const parsed = JSON.parse(saved);
      return {
        ...defaultSettings,
        ...parsed,
        gitIntegrations: parsed.gitIntegrations || [],
        jiraIntegrations: parsed.jiraIntegrations || [],
        notificationIntegrations: parsed.notificationIntegrations || [],
        argoIntegrations: parsed.argoIntegrations || [],
        kubernetesClusters: parsed.kubernetesClusters || [],
        grafanaIntegrations: parsed.grafanaIntegrations || [],
        graylogIntegrations: parsed.graylogIntegrations || [],
        vaultIntegrations: parsed.vaultIntegrations || [],
        general: { ...defaultSettings.general, ...(parsed.general || {}) },
      };
    }
    return defaultSettings;
  });

  useEffect(() => {
    localStorage.setItem('release_portal_settings', JSON.stringify(settings));
  }, [settings]);

  const updateSettings = (data: Partial<Settings>) => {
    setSettings(prev => ({ ...prev, ...data }));
  };

  const updateGitIntegration = (id: string, data: Partial<GitIntegration>) => {
    setSettings(prev => ({
      ...prev,
      gitIntegrations: prev.gitIntegrations.map(g => g.id === id ? { ...g, ...data } : g),
    }));
  };

  const addGitIntegration = (data: Omit<GitIntegration, 'id'>) => {
    const newId = `git-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      gitIntegrations: [...prev.gitIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteGitIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      gitIntegrations: prev.gitIntegrations.filter(g => g.id !== id),
    }));
  };

  const updateJiraIntegration = (id: string, data: Partial<JiraIntegration>) => {
    setSettings(prev => ({
      ...prev,
      jiraIntegrations: prev.jiraIntegrations.map(j => j.id === id ? { ...j, ...data } : j),
    }));
  };

  const addJiraIntegration = (data: Omit<JiraIntegration, 'id'>) => {
    const newId = `jira-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      jiraIntegrations: [...prev.jiraIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteJiraIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      jiraIntegrations: prev.jiraIntegrations.filter(j => j.id !== id),
    }));
  };

  const updateNotificationIntegration = (id: string, data: Partial<NotificationIntegration>) => {
    setSettings(prev => ({
      ...prev,
      notificationIntegrations: prev.notificationIntegrations.map(n => n.id === id ? { ...n, ...data } : n),
    }));
  };

  const addNotificationIntegration = (data: Omit<NotificationIntegration, 'id'>) => {
    const newId = `notif-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      notificationIntegrations: [...prev.notificationIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteNotificationIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      notificationIntegrations: prev.notificationIntegrations.filter(n => n.id !== id),
    }));
  };

  const updateArgoIntegration = (id: string, data: Partial<ArgoCDIntegration>) => {
    setSettings(prev => ({
      ...prev,
      argoIntegrations: prev.argoIntegrations.map(a => a.id === id ? { ...a, ...data } : a),
    }));
  };

  const addArgoIntegration = (data: Omit<ArgoCDIntegration, 'id'>) => {
    const newId = `argo-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      argoIntegrations: [...prev.argoIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteArgoIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      argoIntegrations: prev.argoIntegrations.filter(a => a.id !== id),
    }));
  };

  const updateKubernetesCluster = (id: string, data: Partial<KubernetesCluster>) => {
    setSettings(prev => ({
      ...prev,
      kubernetesClusters: prev.kubernetesClusters.map(k => k.id === id ? { ...k, ...data } : k),
    }));
  };

  const addKubernetesCluster = (data: Omit<KubernetesCluster, 'id'>) => {
    const newId = `k8s-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      kubernetesClusters: [...prev.kubernetesClusters, { ...data, id: newId }],
    }));
  };

  const deleteKubernetesCluster = (id: string) => {
    setSettings(prev => ({
      ...prev,
      kubernetesClusters: prev.kubernetesClusters.filter(k => k.id !== id),
    }));
  };

  const updateGrafanaIntegration = (id: string, data: Partial<GrafanaIntegration>) => {
    setSettings(prev => ({
      ...prev,
      grafanaIntegrations: prev.grafanaIntegrations.map(g => g.id === id ? { ...g, ...data } : g),
    }));
  };

  const addGrafanaIntegration = (data: Omit<GrafanaIntegration, 'id'>) => {
    const newId = `grafana-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      grafanaIntegrations: [...prev.grafanaIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteGrafanaIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      grafanaIntegrations: prev.grafanaIntegrations.filter(g => g.id !== id),
    }));
  };

  const updateGraylogIntegration = (id: string, data: Partial<GraylogIntegration>) => {
    setSettings(prev => ({
      ...prev,
      graylogIntegrations: prev.graylogIntegrations.map(g => g.id === id ? { ...g, ...data } : g),
    }));
  };

  const addGraylogIntegration = (data: Omit<GraylogIntegration, 'id'>) => {
    const newId = `graylog-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      graylogIntegrations: [...prev.graylogIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteGraylogIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      graylogIntegrations: prev.graylogIntegrations.filter(g => g.id !== id),
    }));
  };

  const updateVaultIntegration = (id: string, data: Partial<VaultIntegration>) => {
    setSettings(prev => ({
      ...prev,
      vaultIntegrations: prev.vaultIntegrations.map(v => v.id === id ? { ...v, ...data } : v),
    }));
  };

  const addVaultIntegration = (data: Omit<VaultIntegration, 'id'>) => {
    const newId = `vault-${Date.now()}`;
    setSettings(prev => ({
      ...prev,
      vaultIntegrations: [...prev.vaultIntegrations, { ...data, id: newId }],
    }));
  };

  const deleteVaultIntegration = (id: string) => {
    setSettings(prev => ({
      ...prev,
      vaultIntegrations: prev.vaultIntegrations.filter(v => v.id !== id),
    }));
  };

  const updateGeneralSettings = (data: Partial<Settings['general']>) => {
    setSettings(prev => ({
      ...prev,
      general: { ...prev.general, ...data },
    }));
  };

  const testConnection = async (type: string, id: string): Promise<boolean> => {
    try {
      await api.testIntegration(id);
      return true;
    } catch (error) {
      console.error('Connection test failed:', error);
      return false;
    }
  };

  const syncIntegration = async (type: string, id: string): Promise<boolean> => {
    try {
      await api.syncIntegration(id);
      return true;
    } catch (error) {
      console.error('Integration sync failed:', error);
      return false;
    }
  };

  return (
    <SettingsContext.Provider value={{
      settings,
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
