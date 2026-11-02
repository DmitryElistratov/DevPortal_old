import { useState } from 'react';
import { usePlugins, BackstagePlugin } from '../contexts/PluginContext';
import { Download, Trash2, ToggleLeft, ToggleRight, ExternalLink, Github, Search, Upload, Plus } from 'lucide-react';

export default function PluginsPage() {
  const { plugins, installPlugin, uninstallPlugin, togglePlugin, addCustomPlugin } = usePlugins();
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [showUploadForm, setShowUploadForm] = useState(false);
  const [newPlugin, setNewPlugin] = useState({
    name: '',
    version: '',
    description: '',
    author: '',
    category: 'ci-cd' as BackstagePlugin['category'],
    icon: '🔌',
    homepage: '',
    repository: '',
  });

  const filteredPlugins = plugins.filter(p => {
    if (searchQuery && !p.name.toLowerCase().includes(searchQuery.toLowerCase()) && !p.description.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (categoryFilter !== 'all' && p.category !== categoryFilter) return false;
    return true;
  });

  const installedCount = plugins.filter(p => p.installed).length;
  const enabledCount = plugins.filter(p => p.enabled).length;

  const categories = [
    { id: 'all', label: 'All', color: 'bg-gray-100 text-gray-700' },
    { id: 'ci-cd', label: 'CI/CD', color: 'bg-blue-100 text-blue-700' },
    { id: 'monitoring', label: 'Monitoring', color: 'bg-green-100 text-green-700' },
    { id: 'documentation', label: 'Documentation', color: 'bg-purple-100 text-purple-700' },
    { id: 'security', label: 'Security', color: 'bg-red-100 text-red-700' },
    { id: 'testing', label: 'Testing', color: 'bg-yellow-100 text-yellow-700' },
    { id: 'infrastructure', label: 'Infrastructure', color: 'bg-indigo-100 text-indigo-700' },
  ];

  const handleUploadPlugin = () => {
    if (!newPlugin.name || !newPlugin.version) return;
    
    const plugin: BackstagePlugin = {
      id: `plugin-custom-${Date.now()}`,
      name: newPlugin.name,
      version: newPlugin.version,
      description: newPlugin.description,
      author: newPlugin.author || 'Custom',
      category: newPlugin.category,
      installed: false,
      enabled: false,
      icon: newPlugin.icon,
      homepage: newPlugin.homepage,
      repository: newPlugin.repository,
      compatibility: '>=1.0.0',
      lastUpdated: new Date().toISOString().split('T')[0],
    };
    
    addCustomPlugin(plugin);
    alert(`Плагин "${plugin.name}" успешно загружен!`);
    setShowUploadForm(false);
    setNewPlugin({
      name: '',
      version: '',
      description: '',
      author: '',
      category: 'ci-cd',
      icon: '🔌',
      homepage: '',
      repository: '',
    });
  };

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Plugins</p>
          <h1 className="text-2xl font-light">Backstage Plugins</h1>
          <p className="text-gray-400 mt-1 text-sm">Extend functionality with Backstage plugins</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        {/* Stats */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-orange-50 rounded-lg">
                <Download className="w-5 h-5 text-orange-600" />
              </div>
              <span className="text-sm text-gray-500">Total Plugins</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{plugins.length}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-green-50 rounded-lg">
                <ToggleRight className="w-5 h-5 text-green-600" />
              </div>
              <span className="text-sm text-gray-500">Installed</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{installedCount}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-purple-50 rounded-lg">
                <ToggleRight className="w-5 h-5 text-purple-600" />
              </div>
              <span className="text-sm text-gray-500">Enabled</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{enabledCount}</p>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-5">
            <div className="flex items-center gap-3 mb-2">
              <div className="p-2 bg-orange-50 rounded-lg">
                <Upload className="w-5 h-5 text-orange-600" />
              </div>
              <span className="text-sm text-gray-500">Custom</span>
            </div>
            <p className="text-3xl font-light text-gray-900">{plugins.filter(p => p.id.startsWith('plugin-custom-')).length}</p>
          </div>
        </div>

        {/* Upload Plugin Form */}
        {showUploadForm && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Загрузить новый плагин</h2>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Название плагина *</label>
                  <input
                    type="text"
                    value={newPlugin.name}
                    onChange={(e) => setNewPlugin({ ...newPlugin, name: e.target.value })}
                    placeholder="@company/plugin-name"
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Версия *</label>
                  <input
                    type="text"
                    value={newPlugin.version}
                    onChange={(e) => setNewPlugin({ ...newPlugin, version: e.target.value })}
                    placeholder="1.0.0"
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                <textarea
                  value={newPlugin.description}
                  onChange={(e) => setNewPlugin({ ...newPlugin, description: e.target.value })}
                  placeholder="Описание функциональности плагина..."
                  rows={3}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500 resize-none"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Автор</label>
                  <input
                    type="text"
                    value={newPlugin.author}
                    onChange={(e) => setNewPlugin({ ...newPlugin, author: e.target.value })}
                    placeholder="Company Team"
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Категория</label>
                  <select
                    value={newPlugin.category}
                    onChange={(e) => setNewPlugin({ ...newPlugin, category: e.target.value as any })}
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  >
                    <option value="ci-cd">CI/CD</option>
                    <option value="monitoring">Monitoring</option>
                    <option value="documentation">Documentation</option>
                    <option value="security">Security</option>
                    <option value="testing">Testing</option>
                    <option value="infrastructure">Infrastructure</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Иконка (emoji)</label>
                  <input
                    type="text"
                    value={newPlugin.icon}
                    onChange={(e) => setNewPlugin({ ...newPlugin, icon: e.target.value })}
                    placeholder="🔌"
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Homepage URL</label>
                  <input
                    type="url"
                    value={newPlugin.homepage}
                    onChange={(e) => setNewPlugin({ ...newPlugin, homepage: e.target.value })}
                    placeholder="https://..."
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Repository URL</label>
                <input
                  type="url"
                  value={newPlugin.repository}
                  onChange={(e) => setNewPlugin({ ...newPlugin, repository: e.target.value })}
                  placeholder="https://github.com/..."
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleUploadPlugin}
                  disabled={!newPlugin.name || !newPlugin.version}
                  className="px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-gray-300 text-white text-sm font-medium rounded"
                >
                  Загрузить плагин
                </button>
                <button
                  onClick={() => setShowUploadForm(false)}
                  className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
                >
                  Отмена
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6 p-4 flex items-center gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search plugins..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:border-orange-500"
            />
          </div>
          <button
            onClick={() => setShowUploadForm(!showUploadForm)}
            className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
          >
            <Upload className="w-4 h-4" />
            Загрузить плагин
          </button>
          <div className="flex gap-2 flex-wrap">
            {categories.map(cat => (
              <button
                key={cat.id}
                onClick={() => setCategoryFilter(cat.id)}
                className={`px-3 py-1.5 rounded text-xs font-medium transition-colors ${
                  categoryFilter === cat.id
                    ? 'bg-orange-500 text-white'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                {cat.label}
              </button>
            ))}
          </div>
        </div>

        {/* Plugins grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredPlugins.map(plugin => (
            <div key={plugin.id} className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow">
              <div className="p-6">
                <div className="flex items-start gap-3 mb-3">
                  <div className="text-3xl">{plugin.icon}</div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-medium text-gray-900 text-sm truncate">{plugin.name}</h3>
                    <p className="text-xs text-gray-500">v{plugin.version} • {plugin.author}</p>
                  </div>
                </div>
                <p className="text-sm text-gray-600 mb-4 line-clamp-2">{plugin.description}</p>
                
                <div className="flex items-center gap-2 mb-4">
                  <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">{plugin.category}</span>
                  {plugin.installed && (
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      plugin.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'
                    }`}>
                      {plugin.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-2 mb-4">
                  {plugin.homepage && (
                    <a href={plugin.homepage} target="_blank" rel="noopener noreferrer" className="text-xs text-orange-500 hover:text-orange-600 flex items-center gap-1">
                      <ExternalLink className="w-3 h-3" /> Homepage
                    </a>
                  )}
                  {plugin.repository && (
                    <a href={plugin.repository} target="_blank" rel="noopener noreferrer" className="text-xs text-gray-600 hover:text-gray-800 flex items-center gap-1">
                      <Github className="w-3 h-3" /> Source
                    </a>
                  )}
                </div>

                <div className="flex items-center justify-between pt-4 border-t border-gray-100">
                  <span className="text-xs text-gray-400">Updated: {new Date(plugin.lastUpdated).toLocaleDateString('ru-RU')}</span>
                  <div className="flex items-center gap-2">
                    {!plugin.installed ? (
                      <button
                        onClick={() => installPlugin(plugin.id)}
                        className="px-3 py-1.5 bg-orange-500 hover:bg-orange-600 text-white text-xs font-medium rounded flex items-center gap-1"
                      >
                        <Download className="w-3 h-3" /> Install
                      </button>
                    ) : (
                      <>
                        <button
                          onClick={() => togglePlugin(plugin.id)}
                          className={`p-1.5 rounded ${plugin.enabled ? 'text-orange-500 hover:bg-orange-50' : 'text-gray-400 hover:bg-gray-50'}`}
                          title={plugin.enabled ? 'Disable' : 'Enable'}
                        >
                          {plugin.enabled ? <ToggleRight className="w-5 h-5" /> : <ToggleLeft className="w-5 h-5" />}
                        </button>
                        <button
                          onClick={() => uninstallPlugin(plugin.id)}
                          className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
                          title="Uninstall"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredPlugins.length === 0 && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
            <p className="text-gray-500">No plugins found</p>
          </div>
        )}
      </div>
    </div>
  );
}
