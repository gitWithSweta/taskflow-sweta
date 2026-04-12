import * as Dialog from '@radix-ui/react-dialog'
import { useEffect, useState } from 'react'
import { api } from '../api/client'
import { useTaskForm } from '../hooks/useTaskForm'
import type { Task, User } from '../types'
import { ProjectMarkdownDescription } from './ProjectMarkdownDescription'
import { TaskDueDateField } from './task-form/TaskDueDateField'
import { TaskStatusPrioritySelects } from './task-form/TaskStatusPrioritySelects'
import { UserSearchPickerField } from './task-form/UserSearchPickerField'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'

const PANEL_WIDTH_STORAGE_KEY = 'taskflow-task-panel-width-px'
const PANEL_MIN_PX = 280

function clampTaskPanelWidth(w: number) {
  if (typeof window === 'undefined') return w
  const max = Math.max(PANEL_MIN_PX, window.innerWidth - 16)
  return Math.min(max, Math.max(PANEL_MIN_PX, Math.round(w)))
}

function readInitialPanelWidth(): number {
  if (typeof window === 'undefined') return 560
  try {
    const raw = localStorage.getItem(PANEL_WIDTH_STORAGE_KEY)
    if (raw) {
      const n = parseInt(raw, 10)
      if (Number.isFinite(n)) return clampTaskPanelWidth(n)
    }
  } catch {
    void 0
  }
  return clampTaskPanelWidth(window.innerWidth * 0.5)
}

type Props = {
  open: boolean
  onOpenChange: (v: boolean) => void
  projectId: string
  currentUserId: string
  task?: Task | null
  isProjectOwner?: boolean
  assigneeRoster?: User[]
  assigneeRosterLoading?: boolean
  onSubmit: (data: {
    title: string
    description: string
    status: Task['status']
    priority: Task['priority']
    assignee_id: string | null
    due_date: string
    creator_id?: string
  }) => Promise<void>
}

export function TaskSidePanel({
  open,
  onOpenChange,
  projectId,
  currentUserId,
  task,
  isProjectOwner = false,
  assigneeRoster,
  assigneeRosterLoading: assigneeRosterLoadingProp,
  onSubmit,
}: Props) {
  const [internalUsers, setInternalUsers] = useState<User[]>([])
  const [internalUsersLoading, setInternalUsersLoading] = useState(false)
  const [usersError, setUsersError] = useState<string | null>(null)
  const [panelWidthPx, setPanelWidthPx] = useState(() => readInitialPanelWidth())

  const useParentRoster = assigneeRoster !== undefined
  const allUsers = useParentRoster ? assigneeRoster : internalUsers
  const usersLoading = useParentRoster ? (assigneeRosterLoadingProp ?? false) : internalUsersLoading

  const form = useTaskForm({
    open,
    projectId,
    task,
    isProjectOwner,
    onSubmit,
    onSuccessClose: () => onOpenChange(false),
  })

  useEffect(() => {
    const onResize = () => setPanelWidthPx((w) => clampTaskPanelWidth(w))
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  useEffect(() => {
    if (!open || assigneeRoster !== undefined) return
    let cancelled = false
    void (async () => {
      await Promise.resolve()
      if (cancelled) return
      setInternalUsersLoading(true)
      setUsersError(null)
      try {
        const r = await api.listUsers()
        if (!cancelled) setInternalUsers(r.users ?? [])
      } catch {
        if (!cancelled) {
          setInternalUsers([])
          setUsersError('Could not load users for assignment')
        }
      } finally {
        if (!cancelled) setInternalUsersLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [open, assigneeRoster])

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="task-panel-overlay fixed inset-0 z-50 bg-black/50 backdrop-blur-[2px]" />
        <Dialog.Content
          style={{ width: panelWidthPx }}
          className="task-panel-content fixed inset-y-0 right-0 z-50 flex min-w-[280px] max-w-[calc(100vw-8px)] flex-col border-l border-border bg-panel shadow-[var(--tf-panel-shadow)] outline-none"
        >
          <div
            role="separator"
            aria-orientation="vertical"
            aria-label="Resize task panel"
            tabIndex={0}
            onKeyDown={(e) => {
              const step = e.shiftKey ? 40 : 16
              if (e.key === 'ArrowLeft') {
                e.preventDefault()
                setPanelWidthPx((w) => {
                  const n = clampTaskPanelWidth(w + step)
                  try {
                    localStorage.setItem(PANEL_WIDTH_STORAGE_KEY, String(n))
                  } catch {
                    void 0
                  }
                  return n
                })
              }
              if (e.key === 'ArrowRight') {
                e.preventDefault()
                setPanelWidthPx((w) => {
                  const n = clampTaskPanelWidth(w - step)
                  try {
                    localStorage.setItem(PANEL_WIDTH_STORAGE_KEY, String(n))
                  } catch {
                    void 0
                  }
                  return n
                })
              }
            }}
            onMouseDown={(e) => {
              e.preventDefault()
              const startX = e.clientX
              const startW = panelWidthPx
              let lastW = startW
              const move = (ev: MouseEvent) => {
                lastW = clampTaskPanelWidth(startW + (startX - ev.clientX))
                setPanelWidthPx(lastW)
              }
              const up = () => {
                window.removeEventListener('mousemove', move)
                window.removeEventListener('mouseup', up)
                try {
                  localStorage.setItem(PANEL_WIDTH_STORAGE_KEY, String(lastW))
                } catch {
                  void 0
                }
              }
              window.addEventListener('mousemove', move)
              window.addEventListener('mouseup', up)
            }}
            className="group absolute left-0 top-0 z-[60] flex h-full w-3 -translate-x-1/2 cursor-col-resize items-stretch justify-center bg-transparent"
          >
            <span className="pointer-events-none my-2 w-1 shrink-0 rounded-full bg-fg/10 transition-colors group-hover:bg-[var(--tf-link)]/50" />
          </div>
          <div className="flex shrink-0 items-start justify-between gap-3 border-b border-border px-5 py-4">
            <div className="min-w-0 pr-2">
              <Dialog.Title className="text-lg font-semibold text-fg">
                {task ? 'Edit task' : 'New task'}
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-sm text-fg-muted">
                {task
                  ? 'Update title, Markdown description, status, priority, assignee, and due date.'
                  : 'Add a task to this project.'}
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <Button type="button" variant="ghost" size="sm" className="shrink-0" aria-label="Close panel">
                ✕
              </Button>
            </Dialog.Close>
          </div>

          <form
            id="task-side-form"
            onSubmit={(e) => void form.handleSubmit(e)}
            className="flex min-h-0 flex-1 flex-col"
          >
            <div className="min-h-0 flex-1 overflow-y-auto px-5 py-4">
              {form.error ? (
                <p className="mb-4 rounded-lg border border-red-500/20 bg-red-500/10 px-3 py-2 text-sm text-[var(--tf-delete-fg)]">
                  {form.error}
                </p>
              ) : null}
              <div className="space-y-4">
                <div>
                  <Label htmlFor="ts-title">Title</Label>
                  <Input
                    id="ts-title"
                    value={form.title}
                    onChange={(e) => form.setTitle(e.target.value)}
                    className="mt-1"
                    autoFocus
                  />
                  {form.fields.title ? (
                    <p className="mt-1 text-sm text-[var(--tf-delete-fg)]">{form.fields.title}</p>
                  ) : null}
                </div>
                <ProjectMarkdownDescription
                  className="mt-0 shadow-none"
                  description={form.description}
                  canEdit
                  editorRows={10}
                  onSave={async (md) => {
                    form.setDescription(md)
                  }}
                />
                <TaskStatusPrioritySelects
                  status={form.status}
                  priority={form.priority}
                  onStatusChange={form.setStatus}
                  onPriorityChange={form.setPriority}
                />
                <UserSearchPickerField
                  label="Assignee"
                  users={allUsers}
                  usersLoading={usersLoading}
                  usersError={usersError}
                  currentUserId={currentUserId}
                  value={form.assigneeId}
                  onChange={form.setAssigneeId}
                  searchOpen={form.assigneePickerOpen}
                  onSearchOpenChange={form.setAssigneePickerOpen}
                  search={form.assigneeSearch}
                  onSearchChange={form.setAssigneeSearch}
                  inputRef={form.assigneeInputRef}
                  inputId="ts-assignee-search"
                  listboxId="ts-assignee-list"
                  ariaListLabel="Choose assignee"
                  allowUnassigned
                  fieldError={form.fields.assignee_id}
                />
                {task && isProjectOwner ? (
                  <UserSearchPickerField
                    label="Task creator (ownership)"
                    hint="Only the project owner can transfer who is recorded as the task creator (for delete permissions)."
                    users={allUsers}
                    usersLoading={usersLoading}
                    usersError={null}
                    currentUserId={currentUserId}
                    value={form.creatorId}
                    onChange={form.setCreatorId}
                    searchOpen={form.creatorPickerOpen}
                    onSearchOpenChange={form.setCreatorPickerOpen}
                    search={form.creatorSearch}
                    onSearchChange={form.setCreatorSearch}
                    inputRef={form.creatorInputRef}
                    inputId="ts-creator-search"
                    listboxId="ts-creator-list"
                    ariaListLabel="Choose task creator"
                    fieldError={form.fields.creator_id}
                    editButtonLabel="Transfer"
                  />
                ) : null}
                <TaskDueDateField
                  value={form.dueDate}
                  onChange={form.setDueDate}
                  fieldError={form.fields.due_date}
                />
              </div>
            </div>

            <div className="shrink-0 border-t border-border bg-surface px-5 py-4">
              <div className="flex flex-wrap justify-end gap-2">
                <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
                  Cancel
                </Button>
                <Button type="submit" form="task-side-form" disabled={form.loading}>
                  {form.loading ? 'Saving…' : task ? 'Save' : 'Create'}
                </Button>
              </div>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
