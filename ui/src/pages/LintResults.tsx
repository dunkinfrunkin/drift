import { Shield, AlertTriangle, XCircle, CheckCircle, RefreshCw } from 'lucide-react';
import { api } from '../api';
import type { LintResult } from '../api';
import { useApi } from '../hooks/useApi';
import { Card, CardContent, CardHeader } from '../components/Card';
import { Badge } from '../components/Badge';
import { Button } from '../components/Button';

export default function LintResults() {
  const { data: lint, loading, error, refetch } = useApi<LintResult>(() => api.lint());

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Lint Results</h2>
          <p className="text-sm text-slate-500 mt-1">SQL migration file analysis</p>
        </div>
        <Button variant="secondary" onClick={refetch}>
          <RefreshCw size={14} />
          Re-lint
        </Button>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/25 rounded-lg p-4 text-red-400 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-slate-500">Running lint...</div>
      ) : lint?.clean ? (
        <Card>
          <CardContent>
            <div className="flex items-center gap-3 py-8 justify-center">
              <CheckCircle className="text-green-400" size={24} />
              <p className="text-lg text-green-400">All migration files pass linting</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Shield size={16} className="text-yellow-400" />
                <h3 className="font-semibold text-white">Issues Found</h3>
              </div>
              <Badge variant="warning">{lint?.results?.length ?? 0} issue(s)</Badge>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-slate-800">
              {lint?.results?.map((r, i) => (
                <div key={i} className="px-5 py-3 hover:bg-slate-800/50 transition-colors">
                  <div className="flex items-center gap-3">
                    {r.Severity === 'ERROR' ? (
                      <XCircle size={14} className="text-red-400" />
                    ) : (
                      <AlertTriangle size={14} className="text-yellow-400" />
                    )}
                    <Badge variant={r.Severity === 'ERROR' ? 'error' : 'warning'}>
                      {r.Severity}
                    </Badge>
                    <Badge variant="info">{r.Rule}</Badge>
                    <span className="font-mono text-xs text-slate-400">{r.File}</span>
                  </div>
                  <p className="text-sm text-slate-300 mt-1 ml-7">{r.Message}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
