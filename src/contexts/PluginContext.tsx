import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

export interface BackstagePlugin {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  category: 'ci-cd' | 'monitoring' | 'documentation' | 'security' | 'testing' | 'infrastructure';
  installed: boolean;
  enabled: boolean;
  icon: string;
  homepage?: string;
  repository?: string;
  compatibility: string;
  lastUpdated: string;
}

interface PluginContextType {
  plugins: BackstagePlugin[];
  installPlugin: (id: string) => void;
  uninstallPlugin: (id: string) => void;
  togglePlugin: (id: string) => void;
  addCustomPlugin: (plugin: BackstagePlugin) => void;
}

const defaultPlugins: BackstagePlugin[] = [
  {
    id: 'plugin-argocd',
    name: '@backstage/plugin-argocd',
    version: '2.0.5',
    description: 'ArgoCD integration for monitoring application deployments',
    author: 'Backstage Community',
    category: 'ci-cd',
    installed: true,
    enabled: true,
    icon: '🚀',
    compatibility: '>=1.0.0',
    lastUpdated: '2024-02-20',
  },
  {
    id: 'plugin-kubernetes',
    name: '@backstage/plugin-kubernetes',
    version: '1.5.2',
    description: 'Kubernetes cluster integration',
    author: 'Backstage Community',
    category: 'infrastructure',
    installed: true,
    enabled: true,
    icon: '☸️',
    compatibility: '>=1.0.0',
    lastUpdated: '2024-02-18',
  },
];

const PluginContext = createContext<PluginContextType | undefined>(undefined);

export function PluginProvider({ children }: { children: ReactNode }) {
  const [plugins, setPlugins] = useState<BackstagePlugin[]>(() => {
    const saved = localStorage.getItem('release_portal_plugins');
    return saved ? JSON.parse(saved) : defaultPlugins;
  });

  useEffect(() => {
    localStorage.setItem('release_portal_plugins', JSON.stringify(plugins));
  }, [plugins]);

  const installPlugin = (id: string) => {
    setPlugins(prev => prev.map(p => p.id === id ? { ...p, installed: true, enabled: true } : p));
  };

  const uninstallPlugin = (id: string) => {
    setPlugins(prev => prev.map(p => p.id === id ? { ...p, installed: false, enabled: false } : p));
  };

  const togglePlugin = (id: string) => {
    setPlugins(prev => prev.map(p => p.id === id && p.installed ? { ...p, enabled: !p.enabled } : p));
  };

  const addCustomPlugin = (plugin: BackstagePlugin) => {
    setPlugins(prev => [...prev, plugin]);
  };

  return (
    <PluginContext.Provider value={{ plugins, installPlugin, uninstallPlugin, togglePlugin, addCustomPlugin }}>
      {children}
    </PluginContext.Provider>
  );
}

export function usePlugins() {
  const context = useContext(PluginContext);
  if (context === undefined) {
    throw new Error('usePlugins must be used within a PluginProvider');
  }
  return context;
}
