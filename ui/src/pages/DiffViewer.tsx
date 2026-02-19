import { Plus, Minus, RefreshCw } from 'lucide-react';
import { api } from '../api';
import type { SchemaDiff } from '../api';
import { useApi } from '../hooks/useApi';
import { Card, CardContent, CardHeader } from '../components/Card';
import { Badge } from '../components/Badge';
import { Button } from '../components/Button';

export default function DiffViewer() {
  const { data: diff, loading, error, refetch } = useApi<SchemaDiff>(() => api.diff());

  const actionColor = (action: string) => {
    switch (action) {
      case 'ADD': return 'success';
      case 'DROP': return 'error';
      case 'MODIFY': return 'warning';
      default: return 'muted';
    }
  };

  const ActionIcon = ({ action }: { action: string }) => {
    switch (action) {
      case 'ADD': return <Plus size={14} className="text-green-400" />;
      case 'DROP': return <Minus size={14} className="text-red-400" />;
      default: return <RefreshCw size={14} className="text-yellow-400" />;
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Schema Diff</h2>
          <p className="text-sm text-slate-500 mt-1">Compare current database state against empty baseline</p>
        </div>
        <Button variant="secondary" onClick={refetch}>
          <RefreshCw size={14} />
          Refresh
        </Button>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/25 rounded-lg p-4 text-red-400 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-slate-500">Loading schema diff...</div>
      ) : !diff?.changes?.length ? (
        <Card>
          <CardContent>
            <div className="text-center py-8 text-slate-500">
              No schema changes detected
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <h3 className="font-semibold text-white">Changes</h3>
              <span className="text-xs text-slate-500">{diff.changes.length} change(s)</span>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-slate-800">
              {diff.changes.map((change, i) => (
                <div key={i} className="px-5 py-3 hover:bg-slate-800/50 transition-colors">
                  <div className="flex items-center gap-3">
                    <ActionIcon action={change.action} />
                    <Badge variant={actionColor(change.action) as any}>{change.action}</Badge>
                    <Badge variant="info">{change.objectType}</Badge>
                    <span className="font-mono text-sm text-white">{change.name}</span>
                    {change.details && (
                      <span className="text-xs text-slate-500">({change.details})</span>
                    )}
                  </div>
                  {change.sql && (
                    <pre className="mt-2 ml-8 p-3 bg-slate-950 rounded text-xs text-slate-300 font-mono overflow-x-auto">
                      {change.sql}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
