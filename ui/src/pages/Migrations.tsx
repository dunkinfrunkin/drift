import { api } from '../api';
import type { MigrationInfo } from '../api';
import { useApi } from '../hooks/useApi';
import { Card, CardContent, CardHeader } from '../components/Card';
import { Badge, stateBadgeVariant } from '../components/Badge';
import { Table, Thead, Th, Td } from '../components/Table';
import { FileText, Clock } from 'lucide-react';

export default function Migrations() {
  const { data: migrations, loading, error } = useApi<MigrationInfo[]>(() => api.info());

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Migrations</h2>
        <p className="text-sm text-slate-500 mt-1">All discovered and applied migrations</p>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/25 rounded-lg p-4 text-red-400 text-sm">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h3 className="font-semibold text-white">Migration History</h3>
            <span className="text-xs text-slate-500">{migrations?.length ?? 0} total</span>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <div className="p-8 text-center text-slate-500">Loading...</div>
          ) : (
            <Table>
              <Thead>
                <tr>
                  <Th>Version</Th>
                  <Th>Description</Th>
                  <Th>Type</Th>
                  <Th>State</Th>
                  <Th>Script</Th>
                  <Th>Installed On</Th>
                  <Th>Duration</Th>
                </tr>
              </Thead>
              <tbody>
                {migrations?.map((m, i) => (
                  <tr key={i} className="hover:bg-slate-800/50 transition-colors">
                    <Td>
                      <span className="font-mono text-blue-400">V{m.Version}</span>
                    </Td>
                    <Td>
                      <span className="text-white">{m.Description}</span>
                    </Td>
                    <Td>
                      <Badge variant="info">{m.Type}</Badge>
                    </Td>
                    <Td>
                      <Badge variant={stateBadgeVariant(m.State)}>{m.State}</Badge>
                    </Td>
                    <Td>
                      <div className="flex items-center gap-1.5 text-slate-400">
                        <FileText size={12} />
                        <span className="font-mono text-xs">{m.Script}</span>
                      </div>
                    </Td>
                    <Td>
                      {m.InstalledOn && m.InstalledOn !== '0001-01-01T00:00:00Z' ? (
                        <span className="text-slate-400 text-xs">
                          {new Date(m.InstalledOn).toLocaleString()}
                        </span>
                      ) : (
                        <span className="text-slate-600">-</span>
                      )}
                    </Td>
                    <Td>
                      {m.ExecTime > 0 ? (
                        <div className="flex items-center gap-1 text-slate-400 text-xs">
                          <Clock size={10} />
                          {m.ExecTime}ms
                        </div>
                      ) : (
                        <span className="text-slate-600">-</span>
                      )}
                    </Td>
                  </tr>
                ))}
                {!migrations?.length && (
                  <tr>
                    <Td colSpan={7}>
                      <div className="text-center py-8 text-slate-500">No migrations found</div>
                    </Td>
                  </tr>
                )}
              </tbody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Timeline */}
      {migrations && migrations.length > 0 && (
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-white">Timeline</h3>
          </CardHeader>
          <CardContent>
            <div className="relative">
              <div className="absolute left-3 top-0 bottom-0 w-px bg-slate-800" />
              <div className="space-y-4">
                {migrations.filter(m => m.State === 'Applied').reverse().map((m, i) => (
                  <div key={i} className="flex items-start gap-4 pl-8 relative">
                    <div className="absolute left-1.5 top-1.5 w-3 h-3 rounded-full bg-green-500 border-2 border-slate-900" />
                    <div>
                      <p className="text-sm text-white">
                        <span className="font-mono text-blue-400">V{m.Version}</span>
                        {' '}{m.Description}
                      </p>
                      {m.InstalledOn && m.InstalledOn !== '0001-01-01T00:00:00Z' && (
                        <p className="text-xs text-slate-500 mt-0.5">
                          {new Date(m.InstalledOn).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
