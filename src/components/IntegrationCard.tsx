import { useState } from 'react';
import { RefreshCw, TestTube, Loader2, CheckCircle, XCircle } from 'lucide-react';

interface IntegrationCardProps {
  name: string;
  type: string;
  url: string;
  enabled: boolean;
  lastSync?: string;
  onSync: () => Promise<void>;
  onTest: () => Promise<boolean>;
  onEdit: () => void;
  onDelete: () => void;
  onToggle: () => void;
}

export function IntegrationCard({
  name,
  type,
  url,
  enabled,
  lastSync,
  onSync,
  onTest,
  onEdit,
  onDelete,
  onToggle,
}: IntegrationCardProps) {
  const [syncing, setSyncing] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<'success' | 'error' | null>(null);

  const handleSync = async () => {
    setSyncing(true);
    try {
      await onSync();
    } finally {
      setSyncing(false);
    }
  };

  const handleTest = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const success = await onTest();
      setTestResult(success ? 'success' : 'error');
    } catch {
      setTestResult('error');
    } finally {
      setTesting(false);
    }
  };

  const getTypeColor = () => {
    const colors: Record<string, string> = {
      gitlab: 'bg-orange-100 text-orange-700',
      github: 'bg-gray-100 text-gray-700',
      jira: 'bg-blue-100 text-blue-700',
      argocd: 'bg-purple-100 text-purple-700',
      kubernetes: 'bg-blue-100 text-blue-700',
      grafana: 'bg-orange-100 text-orange-700',
      graylog: 'bg-green-100 text-green-700',
      vault: 'bg-yellow-100 text-yellow-700',
    };
    return colors[type] || 'bg-gray-100 text-gray-700';
  };

  return (
    <div className="border border-gray-200 rounded-lg p-4">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h4 className="text-sm font-medium text-gray-900">{name}</h4>
            <span className={`text-xs px-2 py-0.5 rounded ${enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
              {enabled ? 'Активна' : 'Отключена'}
            </span>
            <span className={`text-xs px-2 py-0.5 rounded uppercase ${getTypeColor()}`}>
              {type}
            </span>
          </div>
          <p className="text-xs text-gray-500 mb-1">URL: {url}</p>
          {lastSync && (
            <p className="text-xs text-gray-400">
              Последняя синхронизация: {new Date(lastSync).toLocaleString('ru-RU')}
            </p>
          )}
          {testResult && (
            <div className={`mt-2 flex items-center gap-1 text-xs ${testResult === 'success' ? 'text-green-600' : 'text-red-600'}`}>
              {testResult === 'success' ? (
                <>
                  <CheckCircle className="w-3 h-3" />
                  <span>Соединение успешно</span>
                </>
              ) : (
                <>
                  <XCircle className="w-3 h-3" />
                  <span>Ошибка соединения</span>
                </>
              )}
            </div>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleSync}
            disabled={syncing || !enabled}
            className="p-2 text-gray-400 hover:text-orange-600 disabled:opacity-50 disabled:cursor-not-allowed"
            title="Синхронизировать"
          >
            {syncing ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <RefreshCw className="w-4 h-4" />
            )}
          </button>
          <button
            onClick={handleTest}
            disabled={testing}
            className="p-2 text-gray-400 hover:text-blue-600 disabled:opacity-50"
            title="Проверить соединение"
          >
            {testing ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <TestTube className="w-4 h-4" />
            )}
          </button>
          <button
            onClick={onEdit}
            className="p-2 text-gray-400 hover:text-orange-600"
            title="Редактировать"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
          </button>
          <button
            onClick={onToggle}
            className="p-2 text-gray-400 hover:text-green-600"
            title={enabled ? 'Отключить' : 'Включить'}
          >
            {enabled ? <XCircle className="w-4 h-4" /> : <CheckCircle className="w-4 h-4" />}
          </button>
          <button
            onClick={onDelete}
            className="p-2 text-gray-400 hover:text-red-600"
            title="Удалить"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
