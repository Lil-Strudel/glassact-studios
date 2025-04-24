export interface AuditLog {
  id: number;
  table_name: string;
  record_id: number;
  action: string;
  changed_by?: number;
  changed_at: string;
  old_data: unknown;
  new_data: unknown;
}
