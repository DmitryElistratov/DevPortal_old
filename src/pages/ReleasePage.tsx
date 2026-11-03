import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { api } from '../lib/api';
import {
  ArrowLeft, GitBranch, GitCommit, User, Calendar, Clock,
  CheckCircle2, XCircle, Rocket, RotateCcw, Tag, FileText, Loader2,
  Shield, FlaskConical, RefreshCw, AlertTriangle,
} from 'lucide-react';

interface Release {
  id: string;
  serviceId: string;
  version: string;
  status: string;
  description: string;
  branch: string;
  gitCommit?: string;
  createdBy: string;
  approvals: number;
  requiredApprovals: number;
  isPrimary?: boolean;
  trafficPercentage?: number;
  createdAt: string;
  updatedAt: string;
}

const statusStyles: Record<string, string> = {
  production: 'bg-green-100 text-green-700',
  testing: 'bg-yellow-100 text-yellow-700',
  building: 'bg-orange-100 text-orange-700',
  staging: 'bg-purple-100 text-purple-700',
  rolled_back: 'bg-red-100 text-red-700',
  closed: 'bg-gray-100 text-gray-600',
};

export default function ReleasePage() {
  const { releaseId } = useParams<{ releaseId: string }>();
  const [release, setRelease] = useState<Release | null>(null);
  const [serviceName, setServiceName] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [busy, setBusy] = useState<'deploy' | 'rollback' | 'approve' | null>(null);
  const [security, setSecurity] = useState<any | null>(null);
  const [tests, setTests] = useState<any | null>(null);
  const [qualityLoading, setQualityLoading] = useState(false);
  const [qualityError, setQualityError] = useState<string | null>(null);

  const loadQuality = async (id: string) => {
    setQualityLoading(true);
    setQualityError(null);
    try {
      const [sec, tst] = await Promise.allSettled([
        api.getReleaseSecurity(id),
        api.getReleaseTests(id),
      ]);
      setSecurity(sec.status === 'fulfilled' ? sec.value : null);
      setTests(tst.status === 'fulfilled' ? tst.value : null);
      if (sec.status === 'rejected' && tst.status === 'rejected') {
        setQualityError('GitLab integration is not available or the service has no git repository.');
      }
    } finally {
      setQualityLoading(false);
    }
  };

  const load = async () => {
    if (!releaseId) return;
    try {
      setLoading(true);
      const data = await api.getRelease(releaseId);
      setRelease(data);
      setError(null);
      if (data?.serviceId) {
        try {
          const svc = await api.getService(data.serviceId);
          setServiceName(svc?.name || '');
        } catch {
          /* service may be deleted — ignore */
        }
      }
    } catch (err) {
      console.error('Failed to load release:', err);
      setError(err instanceof Error && err.message.includes('404')
        ? 'Release not found'
        : 'Failed to load release');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [releaseId]);

  useEffect(() => {
    if (release?.id) loadQuality(release.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [release?.id]);

  const runAction = async (kind: 'deploy' | 'rollback' | 'approve') => {
    if (!release) return;
    setBusy(kind);
    setActionError(null);
    try {
      if (kind === 'deploy') {
        const environment = window.prompt('Deploy to environment (staging / production):', 'staging');
        if (!environment) { setBusy(null); return; }
        await api.deployRelease(release.id, environment);
      } else if (kind === 'rollback') {
        if (!window.confirm('Roll back this release?')) { setBusy(null); return; }
        await api.rollbackRelease(release.id);
      } else {
        await api.approveRelease(release.id);
      }
      await load();
    } catch (err) {
      console.error(`Failed to ${kind} release:`, err);
      setActionError(err instanceof Error ? err.message : `Failed to ${kind}`);
    } finally {
      setBusy(null);
    }
  };

  if (loading) {
    return (
      <div className="p-8 flex items-center gap-2 text-gray-500">
        <Loader2 className="animate-spin" size={20} /> Loading release...
      </div>
    );
  }

  if (error || !release) {
    return (
      <div className="p-8">
        <Link to="/releases" className="inline-flex items-center gap-1 text-sm text-orange-500 hover:text-orange-600 mb-4">
          <ArrowLeft size={16} /> Back to releases
        </Link>
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-4">
          {error || 'Release not found'}
        </div>
      </div>
    );
  }

  const canDeploy = ['building', 'testing', 'staging'].includes(release.status);
  const canRollback = ['production', 'staging'].includes(release.status);
  const canApprove = release.requiredApprovals > 0 && release.approvals < release.requiredApprovals;

  return (
    <div className="p-8 max-w-4xl mx-auto">
      <Link to="/releases" className="inline-flex items-center gap-1 text-sm text-orange-500 hover:text-orange-600 mb-4">
        <ArrowLeft size={16} /> Back to releases
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-2xl font-bold text-gray-900 font-mono">{release.version}</h1>
            <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${statusStyles[release.status] || 'bg-gray-100 text-gray-600'}`}>
              {release.status}
            </span>
            {release.isPrimary && (
              <span className="px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-700">primary</span>
            )}
          </div>
          <div className="text-sm text-gray-500">
            {release.serviceId ? (
              <Link to={`/service/${release.serviceId}`} className="text-orange-500 hover:text-orange-600 font-medium">
                {serviceName || 'View service'}
              </Link>
            ) : 'No service linked'}
          </div>
        </div>
        {(release.trafficPercentage ?? 0) > 0 && (
          <div className="text-right">
            <div className="text-2xl font-semibold text-gray-900">{release.trafficPercentage}%</div>
            <div className="text-xs text-gray-500">traffic</div>
          </div>
        )}
      </div>

      {actionError && (
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-3 mb-4 text-sm">
          {actionError}
        </div>
      )}

      {/* Actions */}
      <div className="flex flex-wrap gap-2 mb-6">
        {canDeploy && (
          <button
            onClick={() => runAction('deploy')}
            disabled={busy !== null}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded-lg disabled:opacity-50"
          >
            {busy === 'deploy' ? <Loader2 size={16} className="animate-spin" /> : <Rocket size={16} />}
            Deploy
          </button>
        )}
        {canApprove && (
          <button
            onClick={() => runAction('approve')}
            disabled={busy !== null}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg disabled:opacity-50"
          >
            {busy === 'approve' ? <Loader2 size={16} className="animate-spin" /> : <CheckCircle2 size={16} />}
            Approve ({release.approvals}/{release.requiredApprovals})
          </button>
        )}
        {canRollback && (
          <button
            onClick={() => runAction('rollback')}
            disabled={busy !== null}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-lg disabled:opacity-50"
          >
            {busy === 'rollback' ? <Loader2 size={16} className="animate-spin" /> : <RotateCcw size={16} />}
            Rollback
          </button>
        )}
        <Link
          to={`/release/${release.id}/canary`}
          className="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg"
        >
          Canary Deploy
        </Link>
        {!canDeploy && !canRollback && !canApprove && (
          <div className="text-sm text-gray-400 italic">No actions available for the current status.</div>
        )}
      </div>

      {/* Details */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <div className="bg-white border border-gray-200 rounded-lg p-4 space-y-3">
          <h2 className="text-sm font-semibold text-gray-900 uppercase tracking-wide">Details</h2>
          <div className="flex items-center gap-2 text-sm text-gray-700">
            <GitBranch size={16} className="text-gray-400 shrink-0" />
            <span className="font-mono">{release.branch || '—'}</span>
          </div>
          {release.gitCommit && (
            <div className="flex items-center gap-2 text-sm text-gray-700">
              <GitCommit size={16} className="text-gray-400 shrink-0" />
              <span className="font-mono">{release.gitCommit.slice(0, 12)}</span>
            </div>
          )}
          <div className="flex items-center gap-2 text-sm text-gray-700">
            <User size={16} className="text-gray-400 shrink-0" />
            {release.createdBy || '—'}
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-700">
            <Calendar size={16} className="text-gray-400 shrink-0" />
            Created {new Date(release.createdAt).toLocaleString()}
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-700">
            <Clock size={16} className="text-gray-400 shrink-0" />
            Updated {new Date(release.updatedAt || release.createdAt).toLocaleString()}
          </div>
        </div>

        <div className="bg-white border border-gray-200 rounded-lg p-4 space-y-3">
          <h2 className="text-sm font-semibold text-gray-900 uppercase tracking-wide">Approvals</h2>
          {release.requiredApprovals > 0 ? (
            <>
              <div className="flex items-center gap-2">
                <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-green-500 rounded-full transition-all"
                    style={{ width: `${Math.min(100, (release.approvals / release.requiredApprovals) * 100)}%` }}
                  />
                </div>
                <span className="text-sm font-medium text-gray-700">
                  {release.approvals}/{release.requiredApprovals}
                </span>
              </div>
              {release.approvals >= release.requiredApprovals ? (
                <div className="flex items-center gap-1.5 text-sm text-green-600">
                  <CheckCircle2 size={16} /> Fully approved
                </div>
              ) : (
                <div className="flex items-center gap-1.5 text-sm text-yellow-600">
                  <XCircle size={16} /> Awaiting approvals
                </div>
              )}
            </>
          ) : (
            <div className="text-sm text-gray-400">No approvals required</div>
          )}
        </div>
      </div>

      {/* Quality assessment: Security & Tests (from GitLab CI) */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-900 uppercase tracking-wide flex items-center gap-1.5">
              <Shield size={16} className="text-gray-400" /> Security Assessment
            </h2>
            <button
              onClick={() => release && loadQuality(release.id)}
              disabled={qualityLoading}
              title="Refresh from GitLab"
              className="p-1.5 text-gray-400 hover:text-gray-600 rounded disabled:opacity-50"
            >
              {qualityLoading ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
            </button>
          </div>
          {qualityError && !security && (
            <div className="text-xs text-gray-400 italic">{qualityError}</div>
          )}
          {security ? (() => {
            const rep = security.report || {};
            const gradeColors: Record<string, string> = {
              A: 'bg-green-100 text-green-700', B: 'bg-lime-100 text-lime-700',
              C: 'bg-yellow-100 text-yellow-700', D: 'bg-orange-100 text-orange-700',
              F: 'bg-red-100 text-red-700', 'not-scanned': 'bg-gray-100 text-gray-500',
              unknown: 'bg-gray-100 text-gray-500',
            };
            return (
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <span className={`px-2.5 py-1 rounded-full text-sm font-bold ${gradeColors[security.grade] || gradeColors.unknown}`}>
                    {security.grade === 'not-scanned' ? 'Not scanned' : `Grade ${security.grade}`}
                  </span>
                  {rep.available === false && (
                    <span className="text-xs text-gray-400">no security reports in GitLab for this ref</span>
                  )}
                </div>
                {(rep.total ?? 0) > 0 ? (
                  <div className="grid grid-cols-4 gap-2 text-center">
                    {[
                      ['Critical', rep.critical, 'text-red-600'],
                      ['High', rep.high, 'text-orange-600'],
                      ['Medium', rep.medium, 'text-yellow-600'],
                      ['Low', rep.low, 'text-blue-600'],
                    ].map(([label, val, cls]) => (
                      <div key={String(label)} className="bg-gray-50 rounded-md py-2">
                        <div className={`text-lg font-semibold ${cls}`}>{String(val)}</div>
                        <div className="text-[10px] uppercase text-gray-500">{String(label)}</div>
                      </div>
                    ))}
                  </div>
                ) : rep.available === true ? (
                  <div className="flex items-center gap-1.5 text-sm text-green-600">
                    <CheckCircle2 size={15} /> No vulnerabilities found
                  </div>
                ) : null}
                {Array.isArray(rep.sources) && rep.sources.length > 0 && (
                  <div className="text-xs text-gray-400 truncate">Scans: {rep.sources.join(', ')}</div>
                )}
              </div>
            );
          })() : !qualityLoading && !qualityError ? (
            <div className="text-xs text-gray-400 italic">No data</div>
          ) : null}
        </div>

        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-semibold text-gray-900 uppercase tracking-wide flex items-center gap-1.5">
              <FlaskConical size={16} className="text-gray-400" /> Test Results
            </h2>
            <button
              onClick={() => release && loadQuality(release.id)}
              disabled={qualityLoading}
              title="Refresh from GitLab CI"
              className="p-1.5 text-gray-400 hover:text-gray-600 rounded disabled:opacity-50"
            >
              {qualityLoading ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
            </button>
          </div>
          {tests ? (() => {
            const sum = tests.summary || {};
            const verdictStyles: Record<string, string> = {
              passed: 'bg-green-100 text-green-700', failed: 'bg-red-100 text-red-700',
              running: 'bg-blue-100 text-blue-700', skipped: 'bg-gray-100 text-gray-600',
              unknown: 'bg-gray-100 text-gray-500',
            };
            return (
              <div className="space-y-3">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className={`px-2.5 py-1 rounded-full text-sm font-semibold ${verdictStyles[tests.verdict] || verdictStyles.unknown}`}>
                    {tests.verdict}
                  </span>
                  {sum.ref && <span className="text-xs text-gray-500 font-mono">{sum.ref}</span>}
                  {sum.pipelineUrl && (
                    <a href={sum.pipelineUrl} target="_blank" rel="noreferrer" className="text-xs text-orange-500 hover:underline">
                      pipeline #{sum.pipelineId}
                    </a>
                  )}
                </div>
                {(sum.total ?? 0) > 0 ? (
                  <>
                    <div className="grid grid-cols-3 gap-2 text-center">
                      <div className="bg-green-50 rounded-md py-2">
                        <div className="text-lg font-semibold text-green-600">{sum.passed}</div>
                        <div className="text-[10px] uppercase text-gray-500">passed</div>
                      </div>
                      <div className="bg-red-50 rounded-md py-2">
                        <div className="text-lg font-semibold text-red-600">{sum.failed}</div>
                        <div className="text-[10px] uppercase text-gray-500">failed</div>
                      </div>
                      <div className="bg-gray-50 rounded-md py-2">
                        <div className="text-lg font-semibold text-gray-500">{sum.skipped}</div>
                        <div className="text-[10px] uppercase text-gray-500">skipped</div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-gray-700">
                      <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
                        <div className="h-full bg-green-500 rounded-full" style={{ width: `${Math.min(100, sum.passRate || 0)}%` }} />
                      </div>
                      <span className="text-xs font-medium">{(sum.passRate || 0).toFixed(0)}%</span>
                    </div>
                  </>
                ) : (
                  <div className="flex items-center gap-1.5 text-xs text-gray-400">
                    <AlertTriangle size={13} /> No test jobs found in the latest pipeline for this ref
                  </div>
                )}
                {(sum.coverage ?? 0) > 0 && (
                  <div className="text-xs text-gray-500">Coverage: <span className="font-semibold text-gray-700">{sum.coverage.toFixed(1)}%</span></div>
                )}
              </div>
            );
          })() : !qualityLoading ? (
            <div className="text-xs text-gray-400 italic">No data</div>
          ) : null}
        </div>
      </div>

      {/* Description */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h2 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-2 flex items-center gap-1.5">
          <FileText size={16} className="text-gray-400" /> Description
        </h2>
        {release.description ? (
          <p className="text-sm text-gray-700 whitespace-pre-wrap">{release.description}</p>
        ) : (
          <p className="text-sm text-gray-400 italic">No description provided</p>
        )}
      </div>

      <div className="mt-4 text-xs text-gray-400 flex items-center gap-1">
        <Tag size={12} /> Release ID: <span className="font-mono">{release.id}</span>
      </div>
    </div>
  );
}
