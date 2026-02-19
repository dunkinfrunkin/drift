import { CheckCircle, XCircle, Clock, AlertTriangle, Database, FileText } from 'lucide-react';
import { api } from '../api';
import type { MigrationInfo, ValidationResult, LintResult, AppConfig } from '../api';
import { useApi } from '../hooks/useApi';
import { Card, CardContent } from '../components/Card';
import { Badge, stateBadgeVariant } from '../components/Badge';

export default function Dashboard() {
  const { data: migrations, loading: loadingM } = useApi<MigrationInfo[]>(() => api.info());
  const { data: validation } = useApi<ValidationResult>(() => api.validate());
  const { data: lintResult } = useApi<LintResult>(() => api.lint());
  const { data: config } = useApi<AppConfig>(() => api.config());

  const applied = migrations?.filter((m) => m.State === 'Applied').length ?? 0;
  const pending = migrations?.filter((m) => m.State === 'Pending').length ?? 0;
  const failed = migrations?.filter((m) => m.State === 'Failed').length ?? 0;
  const total = migrations?.length ?? 0;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Dashboard</h2>
        <p className="text-sm text-slate-500 mt-1">Migration status overview</p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={<Database size={20} />} label="Total Migrations" value={total} loading={loadingM} color="text-blue-400" />
        <StatCard icon={<CheckCircle size={20} />} label="Applied" value={applied} loading={loadingM} color="text-green-400" />
        <StatCard icon={<Clock size={20} />} label="Pending" value={pending} loading={loadingM} color="text-yellow-400" />
        <StatCard icon={<XCircle size={20} />} label="Failed" value={failed} loading={loadingM} color="text-red-400" />
      </div>

      {/* Status cards */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Validation */}
        <Card>
          <CardContent>
            <div className="flex items-center gap-3 mb-3">
              {validation?.valid ? (
                <CheckCircle className="text-green-400" size={20} />
              ) : (
                <AlertTriangle className="text-yellow-400" size={20} />
              )}
              <h3 className="font-semibold text-white">Validation</h3>
            </div>
            {validation?.valid ? (
              <p className="text-sm text-green-400">All migrations are valid</p>
            ) : (
              <div className="space-y-1">
                {validation?.errors?.map((e, i) => (
                  <p key={i} className="text-sm text-red-400">V{e.Version}: {e.Message}</p>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Lint */}
        <Card>
          <CardContent>
            <div className="flex items-center gap-3 mb-3">
              {lintResult?.clean ? (
                <CheckCircle className="text-green-400" size={20} />
              ) : (
                <AlertTriangle className="text-yellow-400" size={20} />
              )}
              <h3 className="font-semibold text-white">Lint</h3>
            </div>
            {lintResult?.clean ? (
              <p className="text-sm text-green-400">All files pass linting</p>
            ) : (
              <p className="text-sm text-yellow-400">
                {lintResult?.results?.length ?? 0} issue(s) found
              </p>
            )}
          </CardContent>
        </Card>

        {/* Config */}
        <Card>
          <CardContent>
            <div className="flex items-center gap-3 mb-3">
              <FileText className="text-slate-400" size={20} />
              <h3 className="font-semibold text-white">Configuration</h3>
            </div>
            <div className="text-sm space-y-1 text-slate-400">
              <p>Driver: <span className="text-white">{config?.driver ?? '...'}</span></p>
              <p>Table: <span className="text-white">{config?.table ?? '...'}</span></p>
              <p>Locations: <span className="text-white">{config?.locations?.join(', ') ?? '...'}</span></p>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent migrations */}
      <Card>
        <CardContent>
          <h3 className="font-semibold text-white mb-4">Recent Migrations</h3>
          {loadingM ? (
            <p className="text-sm text-slate-500">Loading...</p>
          ) : (
            <div className="space-y-2">
              {migrations?.slice(-5).reverse().map((m, i) => (
                <div key={i} className="flex items-center justify-between py-2 border-b border-slate-800 last:border-0">
                  <div className="flex items-center gap-3">
                    <span className="text-slate-500 font-mono text-xs w-10">V{m.Version}</span>
                    <span className="text-sm text-white">{m.Description}</span>
                  </div>
                  <Badge variant={stateBadgeVariant(m.State)}>{m.State}</Badge>
                </div>
              ))}
              {!migrations?.length && <p className="text-sm text-slate-500">No migrations found</p>}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({ icon, label, value, loading, color }: {
  icon: React.ReactNode; label: string; value: number; loading: boolean; color: string;
}) {
  return (
    <Card>
      <CardContent>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
            <p className="text-2xl font-bold text-white mt-1">
              {loading ? '...' : value}
            </p>
          </div>
          <div className={color}>{icon}</div>
        </div>
      </CardContent>
    </Card>
  );
}
