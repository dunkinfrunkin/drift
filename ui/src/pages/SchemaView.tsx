import { Table2, Key, Hash, Link } from 'lucide-react';
import { api } from '../api';
import type { SchemaSnapshot } from '../api';
import { useApi } from '../hooks/useApi';
import { Card, CardContent, CardHeader } from '../components/Card';
import { Badge } from '../components/Badge';
import { Table, Thead, Th, Td } from '../components/Table';

export default function SchemaView() {
  const { data: schema, loading, error } = useApi<SchemaSnapshot>(() => api.snapshot());

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Schema</h2>
        <p className="text-sm text-slate-500 mt-1">Current database schema snapshot</p>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500/25 rounded-lg p-4 text-red-400 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-slate-500">Loading schema...</div>
      ) : (
        <>
          {/* Tables */}
          <div className="space-y-4">
            {schema?.tables?.map((table) => (
              <Card key={table.name}>
                <CardHeader>
                  <div className="flex items-center gap-3">
                    <Table2 size={16} className="text-blue-400" />
                    <h3 className="font-semibold text-white font-mono">{table.name}</h3>
                    {table.schema && (
                      <span className="text-xs text-slate-500">{table.schema}</span>
                    )}
                    <Badge variant="info">{table.columns?.length ?? 0} columns</Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <Table>
                    <Thead>
                      <tr>
                        <Th>Column</Th>
                        <Th>Type</Th>
                        <Th>Nullable</Th>
                        <Th>Default</Th>
                        <Th>Constraints</Th>
                      </tr>
                    </Thead>
                    <tbody>
                      {table.columns?.map((col) => {
                        const isPK = table.primaryKey?.includes(col.name);
                        const fk = table.foreignKeys?.find(f => f.columns?.includes(col.name));
                        return (
                          <tr key={col.name} className="hover:bg-slate-800/50">
                            <Td>
                              <div className="flex items-center gap-2">
                                {isPK && <Key size={12} className="text-yellow-400" />}
                                {fk && <Link size={12} className="text-purple-400" />}
                                <span className="font-mono text-sm text-white">{col.name}</span>
                              </div>
                            </Td>
                            <Td>
                              <span className="font-mono text-xs text-blue-400">{col.dataType}</span>
                            </Td>
                            <Td>
                              {col.nullable ? (
                                <span className="text-slate-500 text-xs">NULL</span>
                              ) : (
                                <span className="text-orange-400 text-xs">NOT NULL</span>
                              )}
                            </Td>
                            <Td>
                              {col.defaultValue ? (
                                <span className="font-mono text-xs text-slate-400">{col.defaultValue}</span>
                              ) : (
                                <span className="text-slate-600">-</span>
                              )}
                            </Td>
                            <Td>
                              <div className="flex gap-1">
                                {isPK && <Badge variant="warning">PK</Badge>}
                                {fk && (
                                  <Badge variant="info">
                                    FK → {fk.referencedTable}
                                  </Badge>
                                )}
                              </div>
                            </Td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </Table>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Indexes */}
          {schema?.indexes && schema.indexes.length > 0 && (
            <Card>
              <CardHeader>
                <div className="flex items-center gap-3">
                  <Hash size={16} className="text-slate-400" />
                  <h3 className="font-semibold text-white">Indexes</h3>
                  <Badge variant="muted">{schema.indexes.length}</Badge>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <Table>
                  <Thead>
                    <tr>
                      <Th>Name</Th>
                      <Th>Table</Th>
                      <Th>Unique</Th>
                    </tr>
                  </Thead>
                  <tbody>
                    {schema.indexes.map((idx) => (
                      <tr key={idx.name} className="hover:bg-slate-800/50">
                        <Td>
                          <span className="font-mono text-sm text-white">{idx.name}</span>
                        </Td>
                        <Td>
                          <span className="font-mono text-xs text-slate-400">{idx.table}</span>
                        </Td>
                        <Td>
                          {idx.unique ? (
                            <Badge variant="warning">UNIQUE</Badge>
                          ) : (
                            <span className="text-slate-600">-</span>
                          )}
                        </Td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </CardContent>
            </Card>
          )}

          {!schema?.tables?.length && (
            <Card>
              <CardContent>
                <div className="text-center py-8 text-slate-500">
                  No tables found in the database
                </div>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  );
}
