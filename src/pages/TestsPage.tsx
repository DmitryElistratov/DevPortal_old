import { useEffect, useMemo, useState } from 'react';
import { api } from '../lib/api';

export interface TestRun {
  id: string;
  serviceId?: string;
  releaseId?: string;
  environment?: string;
  suite?: string;
  status: string; // passed | failed | running | skipped
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  durationMs?: number;
  reportUrl?: string;
  pipelineRef?: string;
  triggeredBy?: string;
  createdAt?: string;
}

const statusStyles: Record<string, string> = {
  passed: 'bg-green-100 text-green-700',
  failed: 'bg-red-100 text-red-700',
  running: 'bg-blue-100 text-blue-700',
  skipped: 'bg-gray-100 text-gray-600',
};

interface Pipeline { id: number; sha: string; ref: string; status: string; web_url: string }
interface MR { iid: number; title: string; state: string; source_branch: string; target_branch: string; author?: { username: string }; web_url: string }

function PipelinesTab() {
  const [project, setProject] = useState(localStorage.getItem('rp.ci.project') || '');
  const [ref, setRef] = useState('');
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [mrs, setMrs] = useState<MR[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [triggering, setTriggering] = useState(false);

  const load = async () => {
    if (!project.trim()) return;
    localStorage.setItem('rp.ci.project', project.trim());
    setLoading(true);
    setError(null);
    try {
      const [pl, mr] = await Promise.all([
        api.listCIPipelines(project.trim(), ref || undefined),
        api.listCIMergeRequests(project.trim()).catch(() => []),
      ]);
      setPipelines(pl || []);
      setMrs(mr || []);
    } catch (e: any) {
      setError(e?.message || 'Failed to load pipelines. Check GitLab integration in Settings.');
    } finally {
      setLoading(false);
    }
  };

  const trigger = async () => {
    if (!project.trim() || !ref.trim()) { setError('Project and ref are required to trigger a pipeline'); return; }
    setTriggering(true);
    try {
      await api.triggerCIPipeline(project.trim(), ref.trim());
      await load();
    } catch (e: any) {
      setError(e?.message || 'Failed to trigger pipeline');
    } finally {
      setTriggering(false);
    }
  };

  useEffect(() => { if (project) load(); /* eslint-disable-next-line */ }, []);

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-2 items-end">
        <div className="flex-1 min-w-[220px]">
          <label className="block text-xs text-gray-500 mb-1">GitLab project (ID or group/path)</label>
          <input value={project} onChange={(e) => setProject(e.target.value)} placeholder="e.g. 42 or backend/api"
            className="w-full border border-gray-300 rounded px-3 py-2 text-sm" />
        </div>
        <div className="w-48">
          <label className="block text-xs text-gray-500 mb-1">Branch (ref)</label>
          <input value={ref} onChange={(e) => setRef(e.target.value)} placeholder="main"
            className="w-full border border-gray-300 rounded px-3 py-2 text-sm" />
        </div>
        <button onClick={load} disabled={loading} className="px-4 py-2 bg-black text-white rounded text-sm disabled:opacity-50">
          {loading ? 'Loading…' : 'Load'}
        </button>
        <button onClick={trigger} disabled={triggering} className="px-4 py-2 border border-gray-300 rounded text-sm hover:bg-gray-50 disabled:opacity-50">
          {triggering ? 'Triggering…' : '▶ Trigger pipeline'}
        </button>
      </div>
      {error && <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-3">{error}</div>}

      <div className="grid md:grid-cols-2 gap-6">
        <div>
          <h3 className="font-medium text-sm mb-2">Pipelines</h3>
          <div className="border border-gray-200 rounded-lg divide-y">
            {pipelines.length === 0 && <div className="p-4 text-sm text-gray-400">No data — press Load</div>}
            {pipelines.map((p) => (
              <a key={p.id} href={p.web_url} target="_blank" rel="noreferrer"
                className="flex items-center justify-between px-4 py-2.5 hover:bg-gray-50 text-sm">
                <div>
                  <span className="font-mono text-xs text-gray-500">#{p.id}</span>{' '}
                  <span className="text-gray-800">{p.ref}</span>
                  <div className="text-xs text-gray-400 font-mono">{p.sha?.slice(0, 8)}</div>
                </div>
                <span className={`px-2 py-0.5 rounded text-xs ${statusStyles[p.status] || 'bg-yellow-100 text-yellow-700'}`}>{p.status}</span>
              </a>
            ))}
          </div>
        </div>
        <div>
          <h3 className="font-medium text-sm mb-2">Open Merge Requests</h3>
          <div className="border border-gray-200 rounded-lg divide-y">
            {mrs.length === 0 && <div className="p-4 text-sm text-gray-400">No open MRs</div>}
            {mrs.map((mr) => (
              <a key={mr.iid} href={mr.web_url} target="_blank" rel="noreferrer"
                className="block px-4 py-2.5 hover:bg-gray-50 text-sm">
                <div className="text-gray-800">{mr.title}</div>
                <div className="text-xs text-gray-400">!{mr.iid} · {mr.source_branch} → {mr.target_branch}{mr.author?.username ? ` · ${mr.author.username}` : ''}</div>
              </a>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function TestsPage() {
  const [tab, setTab] = useState<'runs' | 'pipelines'>('runs');
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ serviceId: '', releaseId: '', environment: 'staging', suite: 'smoke', status: 'passed', total: 0, passed: 0, failed: 0, skipped: 0, reportUrl: '' });

  const load = async () => {
    setLoading(true);
    try {
      const data = await api.listTestRuns();
      setRuns(data || []);
      setError(null);
    } catch (e: any) {
      setError(e?.message || 'Failed to load test runs');
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);

  const stats = useMemo(() => ({
    total: runs.length,
    passed: runs.filter(r => r.status === 'passed').length,
    failed: runs.filter(r => r.status === 'failed').length,
    passRate: runs.reduce((a, r) => a + r.total, 0) > 0
      ? Math.round((runs.reduce((a, r) => a + r.passed, 0) / runs.reduce((a, r) => a + r.total, 0)) * 100) : 0,
  }), [runs]);

  const submit = async () => {
    try {
      await api.createTestRun(form);
      setShowForm(false);
      await load();
    } catch (e: any) {
      alert(e?.message || 'Failed to create test run');
    }
  };

  const remove = async (id: string) => {
    if (!confirm('Пометить тест-ран как удалённый (soft-delete, запись сохранится в БД)?')) return;
    try {
      await api.deleteTestRun(id);
      await load();
    } catch (e: any) {
      alert(e?.message || 'Failed');
    }
  };

  const num = (v: string) => parseInt(v || '0', 10) || 0;

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <h1 className="text-2xl font-light">Tests / TestOps</h1>
          <p className="text-gray-400 mt-1 text-sm">Test runs, QA reports and CI/CD pipelines</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 py-8">
        <div className="flex gap-2 mb-6">
          {(['runs', 'pipelines'] as const).map((t) => (
            <button key={t} onClick={() => setTab(t)}
              className={`px-4 py-2 rounded text-sm ${tab === t ? 'bg-black text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}`}>
              {t === 'runs' ? 'Test Runs' : 'CI/CD Pipelines'}
            </button>
          ))}
        </div>

        {tab === 'runs' && (
          <>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              {[
                { label: 'Runs', value: stats.total },
                { label: 'Passed runs', value: stats.passed },
                { label: 'Failed runs', value: stats.failed },
                { label: 'Pass rate', value: `${stats.passRate}%` },
              ].map((s) => (
                <div key={s.label} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="text-2xl font-light">{s.value}</div>
                  <div className="text-xs text-gray-500 mt-1">{s.label}</div>
                </div>
              ))}
            </div>

            <div className="flex justify-between items-center mb-4">
              <h2 className="font-medium">Test runs</h2>
              <div className="flex gap-2">
                <button onClick={load} className="px-3 py-1.5 border border-gray-300 rounded text-sm hover:bg-gray-50">Refresh</button>
                <button onClick={() => setShowForm(true)} className="px-3 py-1.5 bg-black text-white rounded text-sm">+ New test run</button>
              </div>
            </div>

            {error && <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded p-3 mb-4">{error}</div>}

            {showForm && (
              <div className="bg-white border border-gray-200 rounded-lg p-5 mb-5 grid md:grid-cols-3 gap-3">
                {[
                  ['serviceId', 'Service ID'], ['releaseId', 'Release ID'], ['environment', 'Environment'],
                  ['suite', 'Suite'], ['total', 'Total'], ['passed', 'Passed'], ['failed', 'Failed'], ['skipped', 'Skipped'], ['reportUrl', 'Report URL'],
                ].map(([key, label]) => (
                  <div key={key}>
                    <label className="block text-xs text-gray-500 mb-1">{label}</label>
                    <input
                      value={(form as any)[key]}
                      onChange={(e) => setForm({ ...form, [key]: ['total', 'passed', 'failed', 'skipped'].includes(key) ? num(e.target.value) : e.target.value })}
                      className="w-full border border-gray-300 rounded px-2 py-1.5 text-sm" />
                  </div>
                ))}
                <div>
                  <label className="block text-xs text-gray-500 mb-1">Status</label>
                  <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}
                    className="w-full border border-gray-300 rounded px-2 py-1.5 text-sm">
                    {['passed', 'failed', 'running', 'skipped'].map((s) => <option key={s}>{s}</option>)}
                  </select>
                </div>
                <div className="md:col-span-3 flex gap-2 justify-end">
                  <button onClick={() => setShowForm(false)} className="px-4 py-2 border border-gray-300 rounded text-sm">Cancel</button>
                  <button onClick={submit} className="px-4 py-2 bg-black text-white rounded text-sm">Save</button>
                </div>
              </div>
            )}

            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left text-xs text-gray-500">
                  <tr>
                    <th className="px-4 py-2.5">Suite / Env</th>
                    <th className="px-4 py-2.5">Status</th>
                    <th className="px-4 py-2.5">Results</th>
                    <th className="px-4 py-2.5">Duration</th>
                    <th className="px-4 py-2.5">Triggered by</th>
                    <th className="px-4 py-2.5"></th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {loading && <tr><td colSpan={6} className="px-4 py-6 text-center text-gray-400">Loading…</td></tr>}
                  {!loading && runs.length === 0 && <tr><td colSpan={6} className="px-4 py-6 text-center text-gray-400">No test runs yet</td></tr>}
                  {runs.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50">
                      <td className="px-4 py-2.5">
                        <div className="text-gray-800">{r.suite || '—'}</div>
                        <div className="text-xs text-gray-400">{r.environment}{r.releaseId ? ` · ${r.releaseId}` : ''}</div>
                      </td>
                      <td className="px-4 py-2.5"><span className={`px-2 py-0.5 rounded text-xs ${statusStyles[r.status] || 'bg-gray-100'}`}>{r.status}</span></td>
                      <td className="px-4 py-2.5">
                        <span className="text-green-600">{r.passed}✓</span>{' '}
                        <span className="text-red-600">{r.failed}✗</span>{' '}
                        <span className="text-gray-400">{r.skipped}⊘</span>{' '}
                        <span className="text-gray-500">/ {r.total}</span>
                      </td>
                      <td className="px-4 py-2.5 text-gray-500">{r.durationMs ? `${Math.round(r.durationMs / 1000)}s` : '—'}</td>
                      <td className="px-4 py-2.5 text-gray-500">{r.triggeredBy || '—'}</td>
                      <td className="px-4 py-2.5 text-right whitespace-nowrap">
                        {r.reportUrl && <a href={r.reportUrl} target="_blank" rel="noreferrer" className="text-blue-600 hover:underline mr-3">Report</a>}
                        <button onClick={() => remove(r.id)} className="text-red-500 hover:underline">Delete</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}

        {tab === 'pipelines' && <PipelinesTab />}
      </div>
    </div>
  );
}
