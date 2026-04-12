import type { Project, ProjectDetail, Task, User } from '../types'

const API_PREFIX = '/api'

const TOKEN_KEY = 'taskflow_token'
const USER_KEY = 'taskflow_user'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setSession(token: string, user: User) {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function getStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}

export class ApiError extends Error {
  status: number
  fields?: Record<string, string>
  constructor(message: string, status: number, fields?: Record<string, string>) {
    super(message)
    this.status = status
    this.fields = fields
  }
}

async function parseBody(res: Response): Promise<unknown> {
  const text = await res.text()
  if (!text) return null
  try {
    return JSON.parse(text)
  } catch {
    return { error: text }
  }
}

async function request<T>(
  path: string,
  init: RequestInit & { token?: string | null } = {},
): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const token = init.token !== undefined ? init.token : getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const sentAuth = headers.has('Authorization')
  const url = path.startsWith('http') ? path : `${API_PREFIX}${path}`
  const res = await fetch(url, { ...init, headers })
  if (res.status === 204) {
    return undefined as T
  }
  const body = (await parseBody(res)) as Record<string, unknown> | null
  if (!res.ok) {
    if (res.status === 401 && sentAuth) {
      clearSession()
      window.location.replace('/login')
      throw new ApiError('unauthorized', 401)
    }
    const err = (body?.error as string) || res.statusText
    const fields = body?.fields as Record<string, string> | undefined
    throw new ApiError(err, res.status, fields)
  }
  return body as T
}

export const api = {
  async register(data: { name: string; email: string; password: string }) {
    return request<{ token: string; user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
      token: null,
    })
  },

  async login(data: { email: string; password: string }) {
    return request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
      token: null,
    })
  },

  async me() {
    return request<{ user: User }>('/auth/me')
  },

  async listUsers() {
    return request<{ users: User[] }>('/users')
  },

  async listProjects(page = 1, limit = 50) {
    const q = new URLSearchParams({ page: String(page), limit: String(limit) })
    return request<{ projects: Project[]; total?: number }>(`/projects?${q}`)
  },

  async createProject(body: { name: string; description?: string }) {
    return request<Project>('/projects', { method: 'POST', body: JSON.stringify(body) })
  },

  async getProject(id: string) {
    return request<ProjectDetail>(`/projects/${id}`)
  },

  async patchProject(id: string, body: { name?: string; description?: string; owner_id?: string }) {
    return request<Project>(`/projects/${id}`, { method: 'PATCH', body: JSON.stringify(body) })
  },

  async deleteProject(id: string) {
    await request<unknown>(`/projects/${id}`, { method: 'DELETE' })
  },

  async listTasks(
    projectId: string,
    filters: { status?: string; assignee?: string; page?: number; limit?: number } = {},
  ) {
    const q = new URLSearchParams()
    if (filters.status) q.set('status', filters.status)
    if (filters.assignee) q.set('assignee', filters.assignee)
    q.set('page', String(filters.page ?? 1))
    q.set('limit', String(filters.limit ?? 100))
    return request<{ tasks: Task[]; total?: number }>(`/projects/${projectId}/tasks?${q}`)
  },

  async createTask(
    projectId: string,
    body: {
      title: string
      description?: string
      status?: string
      priority?: string
      assignee_id?: string
      due_date?: string
    },
  ) {
    return request<Task>(`/projects/${projectId}/tasks`, {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  async patchTask(id: string, body: Record<string, unknown>) {
    return request<Task>(`/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(body) })
  },

  async deleteTask(id: string) {
    await request<unknown>(`/tasks/${id}`, { method: 'DELETE' })
  },

  async collaborators(projectId: string) {
    return request<{ users: User[] }>(`/projects/${projectId}/collaborators`)
  },

  async stats(projectId: string) {
    return request<{ by_status: Record<string, number>; by_assignee: Record<string, number> }>(
      `/projects/${projectId}/stats`,
    )
  },
}
