import { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { api } from '../lib/api';
import {
  ArrowLeft, GitBranch, GitCommit, User, Calendar, Clock,
  CheckCircle2, XCircle, Rocket, RotateCcw, Tag, FileText, Loader2,
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
