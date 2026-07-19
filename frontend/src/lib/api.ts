import type {
  RawEvent,
  User,
  UserDetails,
  Project,
  ProjectDetails,
  Job,
  JobLog,
} from './types'

const BASE = '/api'

// ─── Core transport ───────────────────────────────────────────────────────────

async function event<T>(payload: RawEvent): Promise<T> {
  const res = await fetch(`${BASE}/event`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error ?? `HTTP ${res.status}`)
  }

  const text = await res.text()
  return text ? (JSON.parse(text) as T) : (null as unknown as T)
}

// ─── User ─────────────────────────────────────────────────────────────────────

export async function signIn(email: string, password: string): Promise<User> {
  return event<User>({
    method: 'query',
    entity_name: 'user',
    operation: 'sign_in',
    entity_data: { email, password },
  })
}

export async function register(
  userId: string,
  username: string,
  email: string,
  password: string,
): Promise<void> {
  return event<void>({
    method: 'create',
    entity_name: 'user',
    entity_data: { user_id: userId, username, email, password },
  })
}

export async function getUserDetails(userId: string): Promise<UserDetails> {
  return event<UserDetails>({
    method: 'query',
    entity_name: 'user',
    operation: 'get_details',
    entity_data: { user_id: userId },
  })
}

export async function checkEmailExists(email: string): Promise<boolean> {
  const res = await event<{ exists: boolean }>({
    method: 'query',
    entity_name: 'user',
    operation: 'exists_by_email',
    entity_data: { email },
  })
  return res.exists
}

// ─── Project ──────────────────────────────────────────────────────────────────

export async function createProject(
  projectId: string,
  userId: string,
  name: string,
  repoUrl: string,
): Promise<void> {
  return event<void>({
    method: 'create',
    entity_name: 'project',
    entity_data: { project_id: projectId, user_id: userId, name, repo_url: repoUrl },
  })
}

export async function getProjectsByUser(userId: string): Promise<Project[]> {
  return event<Project[]>({
    method: 'query',
    entity_name: 'project',
    operation: 'get_by_user_id',
    entity_data: { user_id: userId },
  })
}

export async function getProjectDetails(projectId: string): Promise<ProjectDetails> {
  return event<ProjectDetails>({
    method: 'query',
    entity_name: 'project',
    operation: 'get_details',
    entity_data: { project_id: projectId },
  })
}

// ─── Job ──────────────────────────────────────────────────────────────────────

export async function getJobsByProject(projectId: string): Promise<Job[]> {
  return event<Job[]>({
    method: 'query',
    entity_name: 'job',
    operation: 'get_by_project_id',
    entity_data: { project_id: projectId },
  })
}

export async function getJob(jobId: string): Promise<Job> {
  return event<Job>({
    method: 'read',
    entity_name: 'job',
    entity_data: { job_id: jobId },
  })
}

// ─── Job logs ─────────────────────────────────────────────────────────────────

export async function getLogsByJob(jobId: string): Promise<JobLog[]> {
  return event<JobLog[]>({
    method: 'query',
    entity_name: 'job_log',
    operation: 'get_by_job_id',
    entity_data: { job_id: jobId },
  })
}

// ─── SSE URL helper ───────────────────────────────────────────────────────────

export function sseUrl(projectId: string): string {
  return `${BASE}/stream?project_id=${encodeURIComponent(projectId)}`
}
