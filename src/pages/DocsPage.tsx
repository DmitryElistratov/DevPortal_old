import { useState } from 'react';
import { BookOpen, FileText, Plus, Edit2, Trash2, Search, Eye, X } from 'lucide-react';

interface Doc {
  id: string;
  title: string;
  description: string;
  category: string;
  content: string;
  lastUpdated: string;
  author: string;
}

const initialDocs: Doc[] = [
  {
    id: 'd1',
    title: 'Release Process Guide',
    description: 'How to create and manage releases in the portal',
    category: 'Getting Started',
    content: '# Release Process Guide\n\nThis guide explains how to create and manage releases using the Release Portal.\n\n## Creating a Release\n\n1. Navigate to the service page\n2. Click "Create Release"\n3. Select tasks from Jira\n4. Configure the release branch\n5. Submit for approval\n\n## Deploying a Release\n\n1. Ensure all approvals are received\n2. Click "Deploy to Staging"\n3. Run integration tests\n4. Click "Deploy to Production"\n5. Monitor via Grafana',
    lastUpdated: '2024-02-27',
    author: 'Admin User',
  },
  {
    id: 'd2',
    title: 'CI/CD Pipeline Configuration',
    description: 'Setting up build, test, and deploy stages',
    category: 'CI/CD',
    content: '# CI/CD Pipeline Configuration\n\nConfigure your CI/CD pipelines for automated deployments.\n\n## Pipeline Stages\n\n- **Build**: Compile and package the application\n- **Test**: Run unit and integration tests\n- **Security**: SAST and dependency scanning\n- **Deploy**: Deploy to target environment\n\n## Configuration\n\nAdd `.gitlab-ci.yml` to your repository root.',
    lastUpdated: '2024-02-25',
    author: 'DevOps Team',
  },
  {
    id: 'd3',
    title: 'Canary Deployment Guide',
    description: 'How to use canary deployments for safe releases',
    category: 'Operations',
    content: '# Canary Deployment Guide\n\nCanary deployments allow you to gradually release changes to a small subset of users.\n\n## Steps\n\n1. Create a release\n2. Click "Deploy as Canary"\n3. Set initial traffic percentage (1-50%)\n4. Monitor metrics\n5. Promote or rollback based on analysis',
    lastUpdated: '2024-02-20',
    author: 'Platform Team',
  },
];

export default function DocsPage() {
  const [docs, setDocs] = useState<Doc[]>(initialDocs);
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('all');
  const [selectedDoc, setSelectedDoc] = useState<Doc | null>(null);
  const [editingDoc, setEditingDoc] = useState<Doc | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newDoc, setNewDoc] = useState({ title: '', description: '', category: '', content: '' });

  const categories = ['all', ...Array.from(new Set(docs.map(d => d.category)))];

  const filteredDocs = docs.filter(doc => {
    if (searchQuery && !doc.title.toLowerCase().includes(searchQuery.toLowerCase()) && !doc.description.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (categoryFilter !== 'all' && doc.category !== categoryFilter) return false;
    return true;
  });

  const handleCreate = () => {
    if (!newDoc.title || !newDoc.content) return;
    const doc: Doc = {
      id: `d${Date.now()}`,
      ...newDoc,
      lastUpdated: new Date().toISOString().split('T')[0],
      author: 'Admin User',
    };
    setDocs([...docs, doc]);
    setNewDoc({ title: '', description: '', category: '', content: '' });
    setShowCreateForm(false);
  };

  const handleUpdate = () => {
    if (!editingDoc) return;
    setDocs(docs.map(d => d.id === editingDoc.id ? { ...editingDoc, lastUpdated: new Date().toISOString().split('T')[0] } : d));
    setEditingDoc(null);
  };

  const handleDelete = (id: string) => {
    if (confirm('Удалить документ?')) {
      setDocs(docs.filter(d => d.id !== id));
      if (selectedDoc?.id === id) setSelectedDoc(null);
      if (editingDoc?.id === id) setEditingDoc(null);
    }
  };

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / TechDocs</p>
          <h1 className="text-2xl font-light">TechDocs</h1>
          <p className="text-gray-400 mt-1 text-sm">Documentation as Code</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        {/* Actions */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6 flex items-center justify-between">
          <div className="flex items-center gap-4 flex-1">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Поиск документации..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:border-orange-500"
              />
            </div>
            <select
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
              className="border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
            >
              {categories.map(cat => (
                <option key={cat} value={cat}>{cat === 'all' ? 'Все категории' : cat}</option>
              ))}
            </select>
          </div>
          <button
            onClick={() => setShowCreateForm(true)}
            className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
          >
            <Plus className="w-4 h-4" />
            Создать документ
          </button>
        </div>

        {/* Create Form */}
        {showCreateForm && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Новый документ</h2>
              <button onClick={() => setShowCreateForm(false)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Заголовок *</label>
                <input
                  type="text"
                  value={newDoc.title}
                  onChange={(e) => setNewDoc({ ...newDoc, title: e.target.value })}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  placeholder="Название документа"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                <input
                  type="text"
                  value={newDoc.description}
                  onChange={(e) => setNewDoc({ ...newDoc, description: e.target.value })}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  placeholder="Краткое описание"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Категория</label>
                <input
                  type="text"
                  value={newDoc.category}
                  onChange={(e) => setNewDoc({ ...newDoc, category: e.target.value })}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                  placeholder="Getting Started, CI/CD, API, etc."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Содержимое (Markdown) *</label>
                <textarea
                  value={newDoc.content}
                  onChange={(e) => setNewDoc({ ...newDoc, content: e.target.value })}
                  rows={12}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono focus:outline-none focus:border-orange-500"
                  placeholder="# Заголовок&#10;&#10;Содержимое документа в формате Markdown..."
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleCreate}
                  className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
                >
                  Создать
                </button>
                <button
                  onClick={() => setShowCreateForm(false)}
                  className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
                >
                  Отмена
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Edit Form */}
        {editingDoc && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Редактирование: {editingDoc.title}</h2>
              <button onClick={() => setEditingDoc(null)} className="text-gray-400 hover:text-gray-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Заголовок</label>
                <input
                  type="text"
                  value={editingDoc.title}
                  onChange={(e) => setEditingDoc({ ...editingDoc, title: e.target.value })}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Описание</label>
                <input
                  type="text"
                  value={editingDoc.description}
                  onChange={(e) => setEditingDoc({ ...editingDoc, description: e.target.value })}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm focus:outline-none focus:border-orange-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Содержимое (Markdown)</label>
                <textarea
                  value={editingDoc.content}
                  onChange={(e) => setEditingDoc({ ...editingDoc, content: e.target.value })}
                  rows={12}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm font-mono focus:outline-none focus:border-orange-500"
                />
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handleUpdate}
                  className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded"
                >
                  Сохранить
                </button>
                <button
                  onClick={() => setEditingDoc(null)}
                  className="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded hover:bg-gray-50"
                >
                  Отмена
                </button>
              </div>
            </div>
          </div>
        )}

        {/* View Document */}
        {selectedDoc && !editingDoc && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
            <div className="flex items-start justify-between mb-4">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">{selectedDoc.title}</h2>
                <p className="text-sm text-gray-500 mt-1">{selectedDoc.description}</p>
                <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                  <span>Категория: {selectedDoc.category}</span>
                  <span>Автор: {selectedDoc.author}</span>
                  <span>Обновлено: {selectedDoc.lastUpdated}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setEditingDoc(selectedDoc)}
                  className="p-2 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
                  title="Редактировать"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleDelete(selectedDoc.id)}
                  className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
                  title="Удалить"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setSelectedDoc(null)}
                  className="px-3 py-1.5 border border-gray-300 text-gray-700 text-sm rounded hover:bg-gray-50"
                >
                  Закрыть
                </button>
              </div>
            </div>
            <div className="border-t border-gray-200 pt-4">
              <div className="prose prose-sm max-w-none">
                <pre className="whitespace-pre-wrap text-sm text-gray-700 bg-gray-50 p-4 rounded">
                  {selectedDoc.content}
                </pre>
              </div>
            </div>
          </div>
        )}

        {/* Documents List */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-3 border-b border-gray-200">
            <span className="text-sm text-gray-500">{filteredDocs.length} документов</span>
          </div>
          <div className="divide-y divide-gray-100">
            {filteredDocs.map(doc => (
              <div key={doc.id} className="px-6 py-4 hover:bg-gray-50 transition-colors flex items-center gap-4">
                <div className="p-2 bg-orange-50 rounded-lg">
                  <FileText className="w-5 h-5 text-orange-600" />
                </div>
                <div className="flex-1">
                  <h3 className="text-sm font-medium text-gray-900">{doc.title}</h3>
                  <p className="text-xs text-gray-500 mt-0.5">{doc.description}</p>
                  <div className="flex items-center gap-3 mt-1 text-xs text-gray-400">
                    <span>Автор: {doc.author}</span>
                    <span>Обновлено: {doc.lastUpdated}</span>
                  </div>
                </div>
                <span className="text-xs px-2 py-0.5 bg-gray-100 text-gray-600 rounded">{doc.category}</span>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => setSelectedDoc(doc)}
                    className="p-1.5 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
                    title="Просмотр"
                  >
                    <Eye className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => setEditingDoc(doc)}
                    className="p-1.5 text-gray-400 hover:text-orange-500 hover:bg-orange-50 rounded"
                    title="Редактировать"
                  >
                    <Edit2 className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(doc.id)}
                    className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded"
                    title="Удалить"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
          {filteredDocs.length === 0 && (
            <div className="px-6 py-12 text-center text-gray-400 text-sm">
              Документы не найдены
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
