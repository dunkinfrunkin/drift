const BASE = '/api/v1';

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || res.statusText);
  }
  return res.json();
}

export interface MigrationInfo {
  Version: string;
  Description: string;
  Type: string;
  Script: string;
  State: 'Applied' | 'Pending' | 'Failed' | 'Missing' | 'Undone';
  InstalledOn: string;
  ExecTime: number;
}

export interface ValidationResult {
  valid: boolean;
  errors: { Version: string; Message: string }[];
}

export interface LintResult {
  results: { File: string; Rule: string; Severity: string; Message: string; Line: number }[];
  clean: boolean;
}

export interface SchemaSnapshot {
  tables: {
    name: string;
    schema?: string;
    columns: { name: string; dataType: string; nullable: boolean; defaultValue?: string }[];
    primaryKey?: string[];
    foreignKeys?: { name: string; columns: string[]; referencedTable: string; referencedColumns: string[] }[];
  }[];
  indexes: { name: string; table: string; columns: string[]; unique: boolean }[];
}

export interface SchemaDiff {
  changes: { action: string; objectType: string; name: string; details?: string; sql?: string }[];
}

export interface ActionResult {
  output: string;
  results?: unknown[];
}

export interface AppConfig {
  driver: string;
  locations: string[];
  table: string;
  schemas: string[];
}

export const api = {
  info: () => request<MigrationInfo[]>('/info'),
  validate: () => request<ValidationResult>('/validate'),
  lint: () => request<LintResult>('/lint'),
  snapshot: () => request<SchemaSnapshot>('/snapshot'),
  diff: () => request<SchemaDiff>('/diff'),
  config: () => request<AppConfig>('/config'),
  migrate: (dryRun = false) => request<ActionResult>(`/migrate?dryRun=${dryRun}`, { method: 'POST' }),
  undo: (count = 1) => request<ActionResult>(`/undo?count=${count}`, { method: 'POST' }),
  repair: () => request<ActionResult>('/repair', { method: 'POST' }),
  clean: () => request<ActionResult>('/clean', { method: 'POST' }),
  baseline: (version: string) => request<ActionResult>(`/baseline?version=${version}`, { method: 'POST' }),
};
