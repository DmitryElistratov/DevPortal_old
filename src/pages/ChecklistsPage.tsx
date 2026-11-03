import { useEffect, useState } from 'react';
import { api } from '../lib/api';

interface ChecklistStep { title: string }
interface Checklist {
  id: string;
  name: string;
  description?: string;
  steps: ChecklistStep[];
  isDefault?: boolean;
  createdBy?: string;
}

export default function ChecklistsPage() {
  const [items, setItems] = useState<Checklist[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<{ name: string; description: string; isDefault: boolean; steps: string[] }>({
    name: '', description: '', isDefault: false, steps: [''],
  });

  const load = async () => {
    setLoading(true);
    try {
      const data = await api.listChecklists();
      setItems(data || []);
      setError(null);
    } catch (e: any) {
      setError(e?.message || 'Failed to load checklists');
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);

  const openCreate = () => {
    setEditingId(null);
    setForm({ name: '', description: '', isDefault: false, steps: [''] });
    setShowForm(true);
  };

  const openEdit = (cl: Checklist) => {
    setEditingId(cl.id);
    setForm({
      name: cl.name,
      description: cl.description || '',
      isDefault: !!cl.isDefault,
      steps: cl.steps?.length ? cl.steps.map(s => s.title) : [''],
    });
    setShowForm(true);
  };

  const save = async () => {
    const steps = form.steps.map(t => t.trim()).filter(Boolean).map(title => ({ title }));
    if (!form.name.trim() || steps.length === 0) {
      alert('Name and at least one step are required');
      return;
    }
    const payload = { name: form.name.trim(), description: form.description, isDefault: form.isDefault, steps };
    try {
      if (editingId) await api.updateChecklist(editingId, payload);
      else await api.createChecklist(payload);
      setShowForm(false);
      await load();
    } catch (e: any) {
      alert(e?.message || 'Failed to save checklist');
    }
  };

  // Soft-delete только с подтверждением — запись помечается удалённой, не удаляется физически.
  const remove = async (id: string) => {
    if (!confirm('Пометить чек-лист как удалённый? Запись сохранится в БД (soft-delete), восстановление — через администратора.')) return;
    try {
      await api.deleteChecklist(id);
      await load();
    } catch (e: any) {
      alert(e?.message || 'Failed to delete checklist');
    }
  };

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <h1 className="text-2xl font-light">QA Checklists</h1>
          <p className="text-gray-400 mt-1 text-sm">Release step templates for manual QA gates</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 py-8">
        <div className="flex justify-between items-center mb-5">
          <div className="flex gap-2">
            <button onClick={load} className="px-3 py-1.5 border border-gray-300 rounded text-sm hover:bg-gray-50">Refresh</button>
          </div>
          <button onClick={openCreate} className="px-4 py-2 bg-black text-white rounded text-sm">+ New checklist</button>
        </div>

        {error && <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-3 mb-4">{error}</div>}

        {showForm && (
          <div className="bg-white border border-gray-200 rounded-lg p-5 mb-6 space-y-3">
            <h3 className="font-medium">{editingId ? 'Edit checklist' : 'New checklist'}</h3>
            <div className="grid md:grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-gray-500 mb-1">Name</label>
                <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="w-full border border-gray-300 rounded px-2 py-1.5 text-sm" placeholder="Pre-release QA" />
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">Description</label>
                <input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })}
                  className="w-full border border-gray-300 rounded px-2 py-1.5 text-sm" />
              </div>
            </div>
            <div>
              <label className="block text-xs text-gray-500 mb-1">Steps</label>
              <div className="space-y-2">
                {form.steps.map((s, i) => (
                  <div key={i} className="flex gap-2">
                    <input value={s} placeholder={`Step ${i + 1}`}
                      onChange={(e) => { const next = [...form.steps]; next[i] = e.target.value; setForm({ ...form, steps: next }); }}
                      className="flex-1 border border-gray-300 rounded px-2 py-1.5 text-sm" />
                    {form.steps.length > 1 && (
                      <button onClick={() => setForm({ ...form, steps: form.steps.filter((_, j) => j !== i) })}
                        className="text-red-500 text-sm px-2">✕</button>
                    )}
                  </div>
                ))}
                <button onClick={() => setForm({ ...form, steps: [...form.steps, ''] })}
                  className="text-sm text-blue-600 hover:underline">+ Add step</button>
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm text-gray-600">
              <input type="checkbox" checked={form.isDefault} onChange={(e) => setForm({ ...form, isDefault: e.target.checked })} />
              Default checklist (auto-suggested for new releases)
            </label>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowForm(false)} className="px-4 py-2 border border-gray-300 rounded text-sm">Cancel</button>
              <button onClick={save} className="px-4 py-2 bg-black text-white rounded text-sm">Save</button>
            </div>
          </div>
        )}

        <div className="grid md:grid-cols-2 gap-4">
          {loading && <div className="text-sm text-gray-400 col-span-2">Loading…</div>}
          {!loading && items.length === 0 && <div className="text-sm text-gray-400 col-span-2">No checklists yet — create one.</div>}
          {items.map((cl) => (
            <div key={cl.id} className="bg-white border border-gray-200 rounded-lg p-5">
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="font-medium text-gray-900">
                    {cl.name}
                    {cl.isDefault && <span className="ml-2 px-2 py-0.5 bg-blue-100 text-blue-700 rounded text-xs">default</span>}
                  </h3>
                  {cl.description && <p className="text-sm text-gray-500 mt-0.5">{cl.description}</p>}
                </div>
                <div className="flex gap-2 shrink-0">
                  <button onClick={() => openEdit(cl)} className="text-sm text-blue-600 hover:underline">Edit</button>
                  <button onClick={() => remove(cl.id)} className="text-sm text-red-500 hover:underline">Delete</button>
                </div>
              </div>
              <ol className="mt-3 space-y-1 list-decimal list-inside text-sm text-gray-700">
                {(cl.steps || []).map((s, i) => <li key={i}>{s.title}</li>)}
              </ol>
              {cl.createdBy && <div className="text-xs text-gray-400 mt-3">by {cl.createdBy}</div>}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
