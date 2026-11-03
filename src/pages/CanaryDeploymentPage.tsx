import { useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../lib/api';

const STEPS = [5, 10, 25, 50, 100];

type StatusBadge = { label: string; cls: string };

function statusBadge(status: string): StatusBadge {
  switch (status) {
    case 'running': return { label: 'В процессе', cls: 'bg-blue-100 text-blue-700' };
    case 'pending': return { label: 'Ожидание', cls: 'bg-yellow-100 text-yellow-700' };
    case 'promoted': return { label: 'Продвинут', cls: 'bg-green-100 text-green-700' };
    case 'rolled_back': return { label: 'Откачен', cls: 'bg-red-100 text-red-700' };
    default: return { label: status || '—', cls: 'bg-gray-100 text-gray-600' };
  }
}

interface ReleaseOption {
  id: string;
  version: string;
  serviceId: string;
  serviceName?: string;
  status: string;
}

export default function CanaryDeploymentPage() {
  const { releaseId: routeReleaseId } = useParams<{ releaseId: string }>();

  const [canaries, setCanaries] = useState<any[]>([]);
  const [releases, setReleases] = useState<ReleaseOption[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  // creation form
  const [showForm, setShowForm] = useState(false);
  const [formReleaseId, setFormReleaseId] = useState('');
  const [formPercentage, setFormPercentage] = useState(5);
  const [formArgoApp, setFormArgoApp] = useState('');
  const [creating, setCreating] = useState(false);

  // live analysis state per canary id
  const [analysis, setAnalysis] = useState<Record<string, any>>({});

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [c, r, s] = await Promise.all([
        api.listAllCanaries().catch(() => [] as any[]),
        api.listReleases().catch(() => [] as any[]),
        api.listServices().catch(() => [] as any[]),
      ]);
      setCanaries(Array.isArray(c) ? c : []);
      const svcMap = new Map<string, string>((Array.isArray(s) ? s : []).map((sv: any) => [sv.id, sv.name]));
      setServices(Array.isArray(s) ? s : []);
      setReleases(
        (Array.isArray(r) ? r : []).map((rel: any) => ({
          id: rel.id,
          version: rel.version,
          serviceId: rel.serviceId,
          serviceName: svcMap.get(rel.serviceId) || rel.serviceName,
          status: rel.status,
        })),
      );
    } catch (e: any) {
      setError(e?.message || 'Не удалось загрузить данные');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  // prefill form when arriving from a release page
  useEffect(() => {
    if (routeReleaseId) {
      setFormReleaseId(routeReleaseId);
      setShowForm(true);
    }
  }, [routeReleaseId]);

  const activeCanaries = useMemo(
    () => canaries.filter((c) => c.status === 'running' || c.status === 'pending'),
    [canaries],
  );
  const finishedCanaries = useMemo(
    () => canaries.filter((c) => c.status !== 'running' && c.status !== 'pending'),
    [canaries],
  );

  const releaseInfo = (id: string) => releases.find((r) => r.id === id);

  const doCreate = async () => {
    if (!formReleaseId) return;
    setCreating(true);
    setError(null);
    try {
      await api.createCanary(formReleaseId, formPercentage, formArgoApp.trim());
      setShowForm(false);
      setFormReleaseId('');
      setFormArgoApp('');
      setFormPercentage(5);
      await load();
    } catch (e: any) {
      setError(e?.message || 'Не удалось запустить canary');
    } finally {
      setCreating(false);
    }
  };

  const doAction = async (id: string, fn: (id: string) => Promise<any>) => {
    setBusyId(id);
    setError(null);
    try {
      await fn(id);
      await load();
    } catch (e: any) {
      setError(e?.message || 'Действие не выполнено');
    } finally {
      setBusyId(null);
    }
  };

  const doAnalyze = async (id: string) => {
    setBusyId(id);
    setError(null);
    try {
      const res = await api.analyzeCanary(id);
      setAnalysis((prev) => ({ ...prev, [id]: res }));
    } catch (e: any) {
      setError(e?.message || 'Не удалось получить статус из ArgoCD');
    } finally {
      setBusyId(null);
    }
  };

  const renderCard = (canary: any) => {
    const badge = statusBadge(canary.status);
    const isActive = canary.status === 'running' || canary.status === 'pending';
    const rel = releaseInfo(canary.releaseId);
    const an = analysis[canary.id];
    const pct = Math.max(0, Math.min(100, canary.percentage ?? 0));

    return (
      <div key={canary.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h3 className="text-lg font-medium text-gray-900">
                {canary.argoApp || rel?.serviceName || 'Сервис'}
              </h3>
              <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${badge.cls}`}>{badge.label}</span>
            </div>
            <p className="text-sm text-gray-500 mt-1">
              {rel?.serviceName ? `${rel.serviceName} · ` : ''}
              <Link to={`/release/${canary.releaseId}`} className="text-indigo-600 hover:underline">
                {canary.canaryVersion || rel?.version || canary.releaseId}
              </Link>
              {' ← '}
              <span className="font-mono">{canary.stableVersion || 'stable'}</span>
            </p>
            <p className="text-xs text-gray-400 mt-1">
              Started: {canary.startedAt ? new Date(canary.startedAt).toLocaleString() : '—'}
              {canary.gitTag ? ` · tag ${canary.gitTag}` : ''}
            </p>
          </div>
          <div className="text-right shrink-0">
            <div className="text-3xl font-light text-gray-900">{pct}%</div>
            <div className="text-xs text-gray-500">трафика на canary</div>
          </div>
        </div>

        {/* traffic bar */}
        <div className="mt-4">
          <div className="h-2 w-full bg-gray-200 rounded overflow-hidden flex">
            <div className="bg-indigo-500 h-full" style={{ width: `${pct}%` }} />
            <div className="bg-emerald-400 h-full" style={{ width: `${100 - pct}%` }} />
          </div>
          <div className="flex justify-between text-xs text-gray-500 mt-1">
            <span>canary {pct}%</span>
            <span>stable {100 - pct}%</span>
          </div>
        </div>

        {/* steps */}
        {isActive && (
          <div className="mt-4 flex items-center gap-2 flex-wrap">
            {STEPS.map((s) => (
              <button
                key={s}
                disabled={busyId === canary.id || pct >= s}
                onClick={() => doAction(canary.id, (id) => api.updateCanaryPercentage(id, s))}
                className={`px-3 py-1 rounded text-sm border ${
                  pct >= s
                    ? 'bg-indigo-600 text-white border-indigo-600 cursor-default'
                    : 'border-gray-300 text-gray-700 hover:border-indigo-400 disabled:opacity-40'
                }`}
              >
                {s}%
              </button>
            ))}
          </div>
        )}

        {/* analysis result */}
        {an && (
          <div className="mt-4 bg-gray-50 border border-gray-200 rounded p-3 text-sm space-y-1">
            <div className="flex gap-4 flex-wrap">
              <span>ArgoCD: <b className={an.argoConnected ? 'text-green-700' : 'text-red-600'}>{an.argoConnected ? 'подключен' : 'нет связи'}</b></span>
              <span>Sync: <b>{an.syncStatus || '—'}</b></span>
              <span>Health: <b className={an.healthStatus === 'Healthy' ? 'text-green-700' : 'text-amber-600'}>{an.healthStatus || '—'}</b></span>
              {an.liveTraffic != null && <span>Traffic: <b>{an.liveTraffic}%</b></span>}
            </div>
            {an.currentRevision && <div className="text-xs text-gray-500 font-mono">rev: {String(an.currentRevision).slice(0, 12)}</div>}
            {an.error && <div className="text-xs text-red-600">{an.error}</div>}
          </div>
        )}

        {/* actions */}
        <div className="mt-5 flex gap-2 flex-wrap">
          <button
            onClick={() => doAnalyze(canary.id)}
            disabled={busyId === canary.id}
            className="px-4 py-1.5 rounded border border-gray-300 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            Проверить ArgoCD
          </button>
          {isActive && (
            <>
              <button
                onClick={() => doAction(canary.id, (id) => api.promoteCanary(id))}
                disabled={busyId === canary.id}
                className="px-4 py-1.5 rounded bg-green-600 text-white text-sm hover:bg-green-700 disabled:opacity-50"
              >
                Продвинуть на 100% (Promote)
              </button>
              <button
                onClick={() => {
                  if (confirm('Отменить canary-деплой и вернуть весь трафик на stable?')) {
                    doAction(canary.id, (id) => api.rollbackCanary(id));
                  }
                }}
                disabled={busyId === canary.id}
                className="px-4 py-1.5 rounded bg-red-600 text-white text-sm hover:bg-red-700 disabled:opacity-50"
              >
                Откатить (Rollback)
              </button>
            </>
          )}
        </div>
      </div>
    );
  };

  const eligibleReleases = releases.filter((r) => !['rolled_back', 'failed'].includes(r.status));

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8 flex items-center justify-between gap-4 flex-wrap">
          <div>
            <h1 className="text-2xl font-light">Canary Deployment</h1>
            <p className="text-sm text-gray-400 mt-1">
              Поэтапная выдача трафика через ArgoCD Rollouts: 5 → 10 → 25 → 50 → 100% с возможностью отката
            </p>
          </div>
          <div className="flex gap-2">
            <button onClick={load} className="px-4 py-2 rounded border border-gray-600 text-sm hover:bg-gray-800">
              Обновить
            </button>
            <button
              onClick={() => setShowForm((v) => !v)}
              className="px-4 py-2 rounded bg-indigo-600 text-sm hover:bg-indigo-500"
            >
              {showForm ? 'Скрыть форму' : '+ Запустить Canary'}
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-8 py-8 space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded p-3">{error}</div>
        )}

        {showForm && (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-medium mb-4">Новый canary-деплой</h2>
            <div className="grid md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Релиз</label>
                <select
                  value={formReleaseId}
                  onChange={(e) => setFormReleaseId(e.target.value)}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm bg-white"
                >
                  <option value="">— выберите релиз —</option>
                  {eligibleReleases.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.serviceName || 'сервис'} · {r.version} ({r.status})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Начальный процент трафика</label>
                <select
                  value={formPercentage}
                  onChange={(e) => setFormPercentage(Number(e.target.value))}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm bg-white"
                >
                  {STEPS.map((s) => (
                    <option key={s} value={s}>{s}%</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Имя приложения ArgoCD (опционально)</label>
                <input
                  value={formArgoApp}
                  onChange={(e) => setFormArgoApp(e.target.value)}
                  placeholder={services.length ? `напр. ${services[0].name}` : 'по умолчанию = имя сервиса'}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
                />
              </div>
            </div>
            <div className="mt-4 flex gap-2">
              <button
                onClick={doCreate}
                disabled={!formReleaseId || creating}
                className="px-5 py-2 rounded bg-indigo-600 text-white text-sm hover:bg-indigo-700 disabled:opacity-50"
              >
                {creating ? 'Запуск…' : 'Запустить Canary'}
              </button>
              <p className="text-xs text-gray-500 self-center">
                Бэкенд обновит приложение в ArgoCD, выставит traffic split и запишет Git-тег при доступном GitLab.
              </p>
            </div>
          </div>
        )}

        {loading ? (
          <div className="text-gray-500 text-sm">Загрузка…</div>
        ) : (
          <>
            <section>
              <h2 className="text-lg font-medium text-gray-900 mb-3">
                Активные canary {activeCanaries.length > 0 && <span className="text-sm text-gray-500">({activeCanaries.length})</span>}
              </h2>
              {activeCanaries.length === 0 ? (
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center text-gray-500 text-sm">
                  Нет активных canary-деплоев. Запустите новый для одного из релизов.
                </div>
              ) : (
                <div className="space-y-4">{activeCanaries.map(renderCard)}</div>
              )}
            </section>

            {finishedCanaries.length > 0 && (
              <section>
                <h2 className="text-lg font-medium text-gray-900 mb-3">История</h2>
                <div className="space-y-4">{finishedCanaries.slice(0, 10).map(renderCard)}</div>
              </section>
            )}
          </>
        )}
      </div>
    </div>
  );
}
