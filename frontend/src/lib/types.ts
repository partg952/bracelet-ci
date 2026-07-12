// ─── Domain models ───────────────────────────────────────────────────────────

export interface User {
  user_id: string
  username: string
  email: string
}

export interface UserDetails extends User {
  project_count: number
  job_count: number
}

export interface Project {
  project_id: string
  user_id: string
  repo_url: string
  name: string
}

export interface ProjectDetails extends Project {
  owner_username: string
  owner_email: string
  job_count: number
}

export interface Job {
  job_id: string
  project_id: string | null
  repo_url: string | null
  commit_sha: string
  created_at: string
  finished_at: string | null
  status: string | null
}

// ─── API wire types ───────────────────────────────────────────────────────────

export type Method = 'create' | 'read' | 'update' | 'delete' | 'query'
export type EntityName = 'user' | 'project' | 'job'

export interface RawEvent {
  method: Method
  entity_name: EntityName
  operation?: string
  entity_data: Record<string, unknown>
}

// ─── SSE event payload ────────────────────────────────────────────────────────

export interface SSEDatabaseEvent {
  method: Method
  entity_name: EntityName
  operation: string
  entity_data: Record<string, unknown>
  result: unknown
}
