import { useState } from 'react';
import {
  ArrowUpCircle, ArrowDownCircle, Wrench, Trash2, Flag, Play,
} from 'lucide-react';
import { api } from '../api';
import { Card, CardContent, CardHeader } from '../components/Card';
import { Button } from '../components/Button';

interface ActionLog {
  time: string;
  action: string;
  success: boolean;
  output: string;
}

export default function Actions() {
  const [loading, setLoading] = useState<string | null>(null);
  const [logs, setLogs] = useState<ActionLog[]>([]);

  const run = async (name: string, fn: () => Promise<{ output: string }>) => {
    setLoading(name);
    try {
      const result = await fn();
      setLogs((prev) => [{
        time: new Date().toLocaleTimeString(),
        action: name,
        success: true,
        output: result.output,
      }, ...prev]);
    } catch (err: any) {
      setLogs((prev) => [{
        time: new Date().toLocaleTimeString(),
        action: name,
        success: false,
        output: err.message,
      }, ...prev]);
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Actions</h2>
        <p className="text-sm text-slate-500 mt-1">Run migration operations</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <ActionCard
          icon={<ArrowUpCircle size={20} />}
          title="Migrate"
          description="Apply all pending migrations"
          color="text-green-400"
          loading={loading === 'migrate'}
          onClick={() => run('migrate', () => api.migrate())}
        />
        <ActionCard
          icon={<ArrowDownCircle size={20} />}
          title="Undo"
          description="Undo the last applied migration"
          color="text-yellow-400"
          loading={loading === 'undo'}
          onClick={() => run('undo', () => api.undo())}
        />
        <ActionCard
          icon={<Wrench size={20} />}
          title="Repair"
          description="Fix checksums and remove failed entries"
          color="text-blue-400"
          loading={loading === 'repair'}
          onClick={() => run('repair', () => api.repair())}
        />
        <ActionCard
          icon={<Flag size={20} />}
          title="Baseline"
          description="Set baseline at version 001"
          color="text-purple-400"
          loading={loading === 'baseline'}
          onClick={() => run('baseline', () => api.baseline('001'))}
        />
        <ActionCard
          icon={<Trash2 size={20} />}
          title="Clean"
          description="Drop the schema history table"
          color="text-red-400"
          loading={loading === 'clean'}
          onClick={() => {
            if (confirm('Are you sure? This will drop the history table.')) {
              run('clean', () => api.clean());
            }
          }}
        />
        <ActionCard
          icon={<Play size={20} />}
          title="Dry Run"
          description="Preview migrations without applying"
          color="text-cyan-400"
          loading={loading === 'dry-run'}
          onClick={() => run('dry-run', () => api.migrate(true))}
        />
      </div>

      {/* Action log */}
      {logs.length > 0 && (
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-white">Action Log</h3>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y divide-slate-800 max-h-96 overflow-y-auto">
              {logs.map((log, i) => (
                <div key={i} className="px-5 py-3">
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-slate-500 font-mono">{log.time}</span>
                    <span className={`text-sm font-medium ${log.success ? 'text-green-400' : 'text-red-400'}`}>
                      {log.action}
                    </span>
                    <span className={`text-xs ${log.success ? 'text-green-500' : 'text-red-500'}`}>
                      {log.success ? 'SUCCESS' : 'FAILED'}
                    </span>
                  </div>
                  {log.output && (
                    <pre className="mt-2 p-3 bg-slate-950 rounded text-xs text-slate-400 font-mono whitespace-pre-wrap overflow-x-auto">
                      {log.output}
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

function ActionCard({ icon, title, description, color, loading, onClick }: {
  icon: React.ReactNode;
  title: string;
  description: string;
  color: string;
  loading: boolean;
  onClick: () => void;
}) {
  return (
    <Card className="hover:border-slate-700 transition-colors">
      <CardContent>
        <div className="flex items-start justify-between mb-3">
          <div className={color}>{icon}</div>
        </div>
        <h3 className="font-semibold text-white mb-1">{title}</h3>
        <p className="text-xs text-slate-500 mb-3">{description}</p>
        <Button
          variant="secondary"
          onClick={onClick}
          loading={loading}
          className="w-full justify-center"
        >
          {loading ? 'Running...' : `Run ${title}`}
        </Button>
      </CardContent>
    </Card>
  );
}
