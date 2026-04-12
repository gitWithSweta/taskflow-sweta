export type User = {
  id: string
  name: string
  email: string
}

export type Project = {
  id: string
  name: string
  description?: string | null
  owner_id: string
  created_at: string
}

export type Task = {
  id: string
  title: string
  description?: string | null
  status: 'todo' | 'in_progress' | 'done'
  priority: 'low' | 'medium' | 'high'
  project_id?: string
  assignee_id?: string | null
  creator_id?: string
  due_date?: string | null
  created_at: string
  updated_at: string
}

export type ProjectDetail = Project & { tasks: Task[] }
