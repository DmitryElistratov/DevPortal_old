import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { Shield, GitBranch, Users, Clock } from 'lucide-react';

export default function AuditPage() {
  const [entries, setEntries] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.listAuditLogs().then(setEntries).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div>
      <div className="bg-black text-white">
        <div className="max-w-7xl mx-auto px-8 py-8">
          <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Home / Audit</p>
          <h1 className="text-2xl font-light">Audit Log</h1>
          <p className="text-gray-400 mt-1 text-sm">All manual actions are logged for transparency</p>
        </div>
      </div>
      <div className="max-w-7xl mx-auto px-8 -mt-4 pb-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="px-6 py-3 border-b border-gray-200">
            <span className="text-sm text-gray-500">{entries.length} entries</span>
          </div>
          {entries.length === 0 ? (
            <div className="px-6 py-12 text-center text-gray-500">No audit entries found</div>
          ) : (
            <div className="divide-y divide-gray-100">
              {entries.map(entry => (
                <div key={entry.id} className="px-6 py-4 hover:bg-gray-50">
                  <div className="flex items-start gap-4">
                    <div className="p-1.5 rounded bg-orange-50">
                      <Shield className="w-4 h-4 text-orange-600" />
                    </div>
                    <div className="flex-1">
                      <span className="text-sm font-medium text-gray-900">{entry.action}</span>
                      <p className="text-sm text-gray-600 mt-1">{entry.details}</p>
                      <div className="flex items-center gap-4 mt-2">
                        <span className="flex items-center gap-1 text-xs text-gray-500">
                          <Users className="w-3 h-3" /> {entry.userName}
                        </span>
                        <span className="flex items-center gap-1 text-xs text-gray-500">
                          <Clock className="w-3 h-3" /> {new Date(entry.timestamp).toLocaleString('ru-RU')}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
