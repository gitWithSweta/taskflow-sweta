import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { Layout } from '../components/Layout'
import { TaskListSection } from '../components/TaskListSection'
import { TaskSidePanel } from '../components/TaskSidePanel'
import { EditProjectNameModal } from '../components/project-detail/EditProjectNameModal'
import { ProjectDetailHeader } from '../components/project-detail/ProjectDetailHeader'
import { ProjectStatusStats } from '../components/project-detail/ProjectStatusStats'
import {
  ProjectTaskFilters,
  type TaskSortMode,
} from '../components/project-detail/ProjectTaskFilters'
import { TransferOwnershipModal } from '../components/project-detail/TransferOwnershipModal'
import { Button } from '../components/ui/button'
import { useAuth } from '../contexts/useAuth'
import type { ProjectDetail, Task, User } from '../types'

type DeleteTarget = { kind: 'none' } | { kind: 'task'; task: Task } | { kind: 'project' }

export function ProjectDetailPage() {
  const { id = '' } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const [project, setProject] = useState<ProjectDetail | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [stats, setStats] = useState<{ by_status: Record<string, number> } | null>(null)
  const [loading, setLoading] = useState(true)
  const [tasksLoading, setTasksLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [assigneeFilter, setAssigneeFilter] = useState<string>('')
  const [taskPanelOpen, setTaskPanelOpen] = useState(false)
  const [editing, setEditing] = useState<Task | null>(null)
  const [statusFlash, setStatusFlash] = useState<string | null>(null)
  const [usersById, setUsersById] = useState<Record<string, User>>({})
  const [assigneeDirectoryLoading, setAssigneeDirectoryLoading] = useState(false)
  const [projectEditOpen, setProjectEditOpen] = useState(false)
  const [editProjectName, setEditProjectName] = useState('')
  const [editProjectSaving, setEditProjectSaving] = useState(false)
  const [taskSort, setTaskSort] = useState<TaskSortMode>('created_desc')
  const [editProjectError, setEditProjectError] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget>({ kind: 'none' })
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [transferOwnerOpen, setTransferOwnerOpen] = useState(false)
  const [transferOwnerId, setTransferOwnerId] = useState('')
  const [transferOwnerSaving, setTransferOwnerSaving] = useState(false)
  const [transferOwnerError, setTransferOwnerError] = useState<string | null>(null)

  const loadProject = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const p = await api.getProject(id)
      setProject(p)
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        setError('not found')
      } else {
        setError(e instanceof ApiError ? e.message : 'Failed to load project')
      }
    } finally {
      setLoading(false)
    }
  }, [id])

  const refreshStats = useCallback(async () => {
    if (!id) return
    try {
      const r = await api.stats(id)
      setStats({ by_status: r.by_status })
    } catch {
      setStats(null)
    }
  }, [id])

  const loadTasks = useCallback(async () => {
    setTasksLoading(true)
    try {
      const res = await api.listTasks(id, {
        status: statusFilter || undefined,
        assignee: assigneeFilter || undefined,
        limit: 200,
      })
      setTasks(res.tasks ?? [])
    } catch {
      setTasks([])
    } finally {
      setTasksLoading(false)
    }
    await refreshStats()
  }, [id, statusFilter, assigneeFilter, refreshStats])

  useEffect(() => {
    void loadProject()
  }, [loadProject])

  useEffect(() => {
    if (!id) return
    void loadTasks()
  }, [id, loadTasks])

  useEffect(() => {
    if (!id) return
    setAssigneeDirectoryLoading(true)
    void api
      .listUsers()
      .then((r) => {
        const m: Record<string, User> = {}
        for (const u of r.users ?? []) m[u.id] = u
        setUsersById(m)
      })
      .catch(() => setUsersById({}))
      .finally(() => setAssigneeDirectoryLoading(false))
  }, [id])

  const sortedTasks = useMemo(() => {
    const list = [...tasks]
    const mul = taskSort.endsWith('_desc') ? -1 : 1
    const mode = taskSort.replace(/_asc|_desc$/, '') as 'alpha' | 'created' | 'due'
    const titleTie = (a: Task, b: Task) =>
      a.title.localeCompare(b.title, undefined, { sensitivity: 'base' })

    list.sort((a, b) => {
      let primary = 0
      if (mode === 'alpha') {
        primary = a.title.localeCompare(b.title, undefined, { sensitivity: 'base' })
        return primary * mul
      }
      if (mode === 'created') {
        primary = a.created_at.localeCompare(b.created_at)
        if (primary !== 0) return primary * mul
        return titleTie(a, b)
      }
      if (!a.due_date && !b.due_date) {
        return titleTie(a, b)
      }
      if (!a.due_date) primary = 1
      else if (!b.due_date) primary = -1
      else primary = a.due_date.localeCompare(b.due_date)
      if (primary !== 0) return primary * mul
      return titleTie(a, b)
    })
    return list
  }, [tasks, taskSort])

  const assigneeFilterUsers = useMemo(() => {
    return Object.values(usersById).sort((a, b) =>
      a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }),
    )
  }, [usersById])

  const isProjectOwner = Boolean(user && project && user.id === project.owner_id)

  useEffect(() => {
    if (!projectEditOpen || !project) return
    setEditProjectName(project.name)
    setEditProjectError(null)
  }, [projectEditOpen, project])

  useEffect(() => {
    if (!transferOwnerOpen) return
    setTransferOwnerId('')
    setTransferOwnerError(null)
  }, [transferOwnerOpen])

  async function bumpStatus(task: Task, next: Task['status']) {
    if (task.status === next) return
    const snapshot = tasks.map((x) => ({ ...x }))
    setTasks((cur) => cur.map((x) => (x.id === task.id ? { ...x, status: next } : x)))
    setStatusFlash(null)
    try {
      const updated = await api.patchTask(task.id, { status: next })
      setTasks((cur) => cur.map((x) => (x.id === task.id ? updated : x)))
      void refreshStats()
    } catch {
      setTasks(snapshot)
      setStatusFlash('Could not update status — reverted.')
    }
  }

  async function confirmDelete() {
    if (deleteTarget.kind === 'none') return
    if (deleteTarget.kind === 'project') {
      if (!project || !isProjectOwner) {
        setDeleteTarget({ kind: 'none' })
        return
      }
    }
    if (deleteTarget.kind === 'task') {
      const t = deleteTarget.task
      if (!isProjectOwner && user?.id !== t.creator_id) {
        setDeleteTarget({ kind: 'none' })
        return
      }
    }
    setDeletePending(true)
    setDeleteError(null)
    try {
      if (deleteTarget.kind === 'task') {
        await api.deleteTask(deleteTarget.task.id)
        setTasks((t) => t.filter((x) => x.id !== deleteTarget.task.id))
        void refreshStats()
        setDeleteTarget({ kind: 'none' })
      } else {
        await api.deleteProject(project!.id)
        setDeleteTarget({ kind: 'none' })
        navigate('/project')
      }
    } catch (e) {
      setDeleteError(e instanceof ApiError ? e.message : 'Delete failed')
    } finally {
      setDeletePending(false)
    }
  }

  async function saveTransferOwner(e: React.FormEvent) {
    e.preventDefault()
    if (!project || !isProjectOwner) return
    if (!transferOwnerId.trim()) {
      setTransferOwnerError('Choose a user to transfer ownership to.')
      return
    }
    setTransferOwnerSaving(true)
    setTransferOwnerError(null)
    try {
      const updated = await api.patchProject(project.id, { owner_id: transferOwnerId.trim() })
      setProject((cur) => (cur ? { ...cur, owner_id: updated.owner_id } : null))
      setTransferOwnerOpen(false)
    } catch (err) {
      setTransferOwnerError(err instanceof ApiError ? err.message : 'Transfer failed')
    } finally {
      setTransferOwnerSaving(false)
    }
  }

  async function saveProjectDetails(e: React.FormEvent) {
    e.preventDefault()
    if (!project || !isProjectOwner) return
    setEditProjectError(null)
    const name = editProjectName.trim()
    if (!name) {
      setEditProjectError('Name is required')
      return
    }
    setEditProjectSaving(true)
    try {
      const updated = await api.patchProject(project.id, { name })
      setProject((cur) =>
        cur
          ? {
              ...cur,
              name: updated.name,
            }
          : null,
      )
      setProjectEditOpen(false)
    } catch (err) {
      setEditProjectError(err instanceof ApiError ? err.message : 'Could not save project')
    } finally {
      setEditProjectSaving(false)
    }
  }

  async function saveProjectMarkdownDescription(markdown: string) {
    if (!project || !isProjectOwner) return
    const updated = await api.patchProject(project.id, { description: markdown })
    setProject((cur) =>
      cur
        ? {
            ...cur,
            description: updated.description,
          }
        : null,
    )
  }

  if (loading) {
    return (
      <Layout>
        <p className="text-fg-muted">Loading project…</p>
      </Layout>
    )
  }

  if (error === 'not found' || !project) {
    return (
      <Layout>
        <p className="text-fg-muted">Project not found.</p>
        <Link
          to="/project"
          className="mt-4 inline-block text-sm font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)] hover:underline"
        >
          Back to projects
        </Link>
      </Layout>
    )
  }

  if (error) {
    return (
      <Layout>
        <p className="text-[var(--tf-delete-fg)]">{error}</p>
        <Button type="button" variant="secondary" className="mt-4" onClick={() => void loadProject()}>
          Retry
        </Button>
      </Layout>
    )
  }

  return (
    <Layout>
      <ProjectDetailHeader
        projectName={project.name}
        projectDescription={project.description}
        isProjectOwner={isProjectOwner}
        onEditProjectName={() => setProjectEditOpen(true)}
        onNewTask={() => {
          setEditing(null)
          setTaskPanelOpen(true)
        }}
        onTransferOwnership={() => setTransferOwnerOpen(true)}
        onDeleteProject={() => {
          setDeleteError(null)
          setDeleteTarget({ kind: 'project' })
        }}
        onSaveDescription={saveProjectMarkdownDescription}
      />

      {isProjectOwner ? (
        <TransferOwnershipModal
          open={transferOwnerOpen}
          onOpenChange={setTransferOwnerOpen}
          users={assigneeFilterUsers}
          transferOwnerId={transferOwnerId}
          onTransferOwnerIdChange={setTransferOwnerId}
          error={transferOwnerError}
          saving={transferOwnerSaving}
          onSubmit={saveTransferOwner}
        />
      ) : null}

      {isProjectOwner ? (
        <EditProjectNameModal
          open={projectEditOpen}
          onOpenChange={setProjectEditOpen}
          name={editProjectName}
          onNameChange={setEditProjectName}
          error={editProjectError}
          saving={editProjectSaving}
          onSubmit={saveProjectDetails}
        />
      ) : null}

      <ProjectStatusStats byStatus={stats?.by_status} />

      <ProjectTaskFilters
        taskSort={taskSort}
        onTaskSortChange={setTaskSort}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        assigneeFilter={assigneeFilter}
        onAssigneeFilterChange={setAssigneeFilter}
        assigneeUsers={assigneeFilterUsers}
        currentUserId={user?.id}
        tasksLoading={tasksLoading}
      />

      {statusFlash ? (
        <p className="mt-4 rounded-lg border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-sm text-fg">
          {statusFlash}
        </p>
      ) : null}

      <div className="mt-8 overflow-hidden rounded-xl border border-border bg-input shadow-[var(--tf-card-shadow)]">
        <TaskListSection
          sortedTasks={sortedTasks}
          usersById={usersById}
          currentUserId={user?.id}
          isProjectOwner={isProjectOwner}
          statusFilter={statusFilter}
          assigneeFilter={assigneeFilter}
          onOpenTask={(t) => {
            setEditing(t)
            setTaskPanelOpen(true)
          }}
          onBumpStatus={bumpStatus}
          onRequestDeleteTask={(t) => {
            setDeleteError(null)
            setDeleteTarget({ kind: 'task', task: t })
          }}
        />
      </div>

      <ConfirmDialog
        open={deleteTarget.kind !== 'none'}
        onOpenChange={(open) => {
          if (!open) {
            setDeleteTarget({ kind: 'none' })
            setDeleteError(null)
          }
        }}
        title={
          deleteTarget.kind === 'task'
            ? 'Delete task?'
            : deleteTarget.kind === 'project'
              ? 'Delete project?'
              : ''
        }
        description={
          deleteTarget.kind === 'task'
            ? `This will permanently delete “${deleteTarget.task.title}”. This cannot be undone.`
            : deleteTarget.kind === 'project' && project
              ? `This will permanently delete “${project.name}” and all of its tasks. This cannot be undone.`
              : ''
        }
        confirmLabel={
          deleteTarget.kind === 'task'
            ? 'Delete task'
            : deleteTarget.kind === 'project'
              ? 'Delete project'
              : 'Delete'
        }
        pending={deletePending}
        error={deleteError}
        onConfirm={confirmDelete}
      />

      {user ? (
        <TaskSidePanel
          open={taskPanelOpen}
          onOpenChange={(v) => {
            setTaskPanelOpen(v)
            if (!v) setEditing(null)
          }}
          projectId={id}
          currentUserId={user.id}
          task={editing}
          isProjectOwner={isProjectOwner}
          assigneeRoster={assigneeFilterUsers}
          assigneeRosterLoading={assigneeDirectoryLoading}
          onSubmit={async (data) => {
            const payload: Record<string, unknown> = {
              title: data.title,
              description: data.description || null,
              status: data.status,
              priority: data.priority,
            }
            if (data.assignee_id) {
              payload.assignee_id = data.assignee_id
            } else {
              payload.assignee_id = null
            }
            if (data.due_date) {
              payload.due_date = data.due_date
            } else {
              payload.due_date = null
            }
            if (editing && isProjectOwner && data.creator_id !== undefined) {
              payload.creator_id = data.creator_id
            }
            if (editing) {
              await api.patchTask(editing.id, payload)
            } else {
              await api.createTask(id, {
                title: data.title,
                description: data.description || undefined,
                status: data.status,
                priority: data.priority,
                assignee_id: data.assignee_id ?? undefined,
                due_date: data.due_date || undefined,
              })
            }
            await loadTasks()
          }}
        />
      ) : null}
    </Layout>
  )
}
