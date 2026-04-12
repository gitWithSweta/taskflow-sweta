import clsx from 'clsx'
import type { Task, User } from '../types'

type Props = {
  sortedTasks: Task[]
  usersById: Record<string, User>
  currentUserId: string | undefined
  isProjectOwner: boolean
  statusFilter: string
  assigneeFilter: string
  onOpenTask: (task: Task) => void
  onBumpStatus: (task: Task, next: Task['status']) => void
  onRequestDeleteTask: (task: Task) => void
}

function formatCreatedShort(iso: string) {
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return '—'
    return d.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch {
    return '—'
  }
}

export function TaskListSection({
  sortedTasks,
  usersById,
  currentUserId,
  isProjectOwner,
  statusFilter,
  assigneeFilter,
  onOpenTask,
  onBumpStatus,
  onRequestDeleteTask,
}: Props) {
  if (sortedTasks.length === 0) {
    return (
      <p className="p-8 text-center text-sm text-fg-faint">
        {statusFilter || assigneeFilter
          ? 'No tasks match the current filters.'
          : 'No tasks yet. Use New task to add one.'}
      </p>
    )
  }

  return (
    <>
      <ul className="divide-y divide-border lg:hidden">
        {sortedTasks.map((t) => {
          const canDeleteTask = isProjectOwner || (currentUserId != null && currentUserId === t.creator_id)
          const createdShort = formatCreatedShort(t.created_at)
          return (
            <li key={t.id} className="space-y-3 px-4 py-4">
              <div className="min-w-0">
                <button
                  type="button"
                  className="block w-full text-left text-base font-medium text-fg hover:text-[var(--tf-link-hover)]"
                  onClick={() => onOpenTask(t)}
                >
                  {t.title}
                </button>
                {t.description ? (
                  <p className="mt-1 line-clamp-2 break-words text-xs text-fg-soft">{t.description}</p>
                ) : null}
              </div>
              <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                <div>
                  <span className="text-[10px] font-medium uppercase tracking-wide text-fg-soft">
                    Status
                  </span>
                  <select
                    className="mt-1 h-9 w-full rounded-lg border border-border bg-input px-2 text-sm text-fg focus:border-blue-500 focus:outline-none"
                    value={t.status}
                    aria-label={`Status for ${t.title}`}
                    onChange={(e) => void onBumpStatus(t, e.target.value as Task['status'])}
                  >
                    <option value="todo">To do</option>
                    <option value="in_progress">In progress</option>
                    <option value="done">Done</option>
                  </select>
                </div>
                <div>
                  <span className="text-[10px] font-medium uppercase tracking-wide text-fg-soft">
                    Priority
                  </span>
                  <div className="mt-1">
                    <span className="inline-flex rounded bg-[var(--tf-chip)] px-2 py-1 text-sm capitalize text-fg-dim">
                      {t.priority}
                    </span>
                  </div>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-[10px] font-medium uppercase tracking-wide text-fg-soft">
                    Assignee
                  </span>
                  <p className="mt-0.5 truncate text-fg-dim">
                    {t.assignee_id && usersById[t.assignee_id] ? (
                      <>
                        {usersById[t.assignee_id].name}
                        {t.assignee_id === currentUserId ? ' (you)' : ''}
                      </>
                    ) : t.assignee_id ? (
                      <span className="text-fg-soft">…</span>
                    ) : (
                      <span className="text-fg-faint">—</span>
                    )}
                  </p>
                </div>
                <div>
                  <span className="text-[10px] font-medium uppercase tracking-wide text-fg-soft">Due</span>
                  <p className="mt-0.5 text-fg-muted">{t.due_date ?? '—'}</p>
                </div>
                <div className="col-span-2">
                  <span className="text-[10px] font-medium uppercase tracking-wide text-fg-soft">
                    Created
                  </span>
                  <p className="mt-0.5 text-fg-muted">{createdShort}</p>
                </div>
              </div>
              <div className="flex justify-end border-t border-border pt-3">
                <span
                  title={
                    !canDeleteTask
                      ? 'Only the project owner or task creator can delete this task.'
                      : undefined
                  }
                  className={clsx('inline-flex rounded', !canDeleteTask && 'cursor-not-allowed')}
                >
                  <button
                    type="button"
                    className={clsx(
                      'text-sm font-medium',
                      canDeleteTask
                        ? 'text-[var(--tf-delete-fg)] hover:underline'
                        : 'pointer-events-none text-[var(--tf-delete-fg-muted)] opacity-60',
                    )}
                    onClick={() => {
                      if (!canDeleteTask) return
                      onRequestDeleteTask(t)
                    }}
                  >
                    Delete
                  </button>
                </span>
              </div>
            </li>
          )
        })}
      </ul>

      <div className="hidden lg:block">
        <div className="divide-y divide-border">
          <div className="grid grid-cols-[minmax(8rem,1.2fr)_minmax(6rem,0.55fr)_minmax(4.5rem,0.35fr)_minmax(7rem,0.75fr)_minmax(5.5rem,0.45fr)_minmax(6.5rem,0.55fr)_auto] gap-2 bg-[var(--tf-table-header)] px-3 py-2.5 text-[11px] font-semibold uppercase tracking-wide text-fg-soft">
            <div>Task</div>
            <div>Status</div>
            <div>Priority</div>
            <div>Assignee</div>
            <div>Due</div>
            <div>Created</div>
            <div className="text-right"> </div>
          </div>
          <ul className="divide-y divide-border">
            {sortedTasks.map((t) => {
              const canDeleteTask =
                isProjectOwner || (currentUserId != null && currentUserId === t.creator_id)
              const createdShort = formatCreatedShort(t.created_at)
              return (
                <li
                  key={t.id}
                  className="grid grid-cols-[minmax(8rem,1.2fr)_minmax(6rem,0.55fr)_minmax(4.5rem,0.35fr)_minmax(7rem,0.75fr)_minmax(5.5rem,0.45fr)_minmax(6.5rem,0.55fr)_auto] items-center gap-2 px-3 py-2.5"
                >
                  <div className="min-w-0">
                    <button
                      type="button"
                      className="block w-full truncate text-left text-sm font-medium text-fg hover:text-[var(--tf-link-hover)]"
                      title={t.title}
                      onClick={() => onOpenTask(t)}
                    >
                      {t.title}
                    </button>
                    {t.description ? (
                      <p
                        className="mt-0.5 line-clamp-1 break-all text-xs text-fg-soft"
                        title={t.description}
                      >
                        {t.description}
                      </p>
                    ) : null}
                  </div>
                  <div className="flex min-w-0 flex-col gap-1">
                    <select
                      className="h-9 w-full min-w-0 rounded-lg border border-border bg-input px-2 text-xs text-fg focus:border-blue-500 focus:outline-none"
                      value={t.status}
                      aria-label={`Status for ${t.title}`}
                      onChange={(e) => void onBumpStatus(t, e.target.value as Task['status'])}
                    >
                      <option value="todo">To do</option>
                      <option value="in_progress">In progress</option>
                      <option value="done">Done</option>
                    </select>
                  </div>
                  <div>
                    <span className="inline-flex rounded bg-[var(--tf-chip)] px-2 py-1 text-xs capitalize text-fg-dim">
                      {t.priority}
                    </span>
                  </div>
                  <div className="min-w-0 truncate text-xs text-fg-dim">
                    {t.assignee_id && usersById[t.assignee_id] ? (
                      <>
                        {usersById[t.assignee_id].name}
                        {t.assignee_id === currentUserId ? ' (you)' : ''}
                      </>
                    ) : t.assignee_id ? (
                      <span className="text-fg-soft">…</span>
                    ) : (
                      <span className="text-fg-faint">—</span>
                    )}
                  </div>
                  <div className="text-xs text-fg-muted">{t.due_date ?? '—'}</div>
                  <div className="text-xs text-fg-muted">{createdShort}</div>
                  <div className="flex justify-end">
                    <span
                      title={
                        !canDeleteTask
                          ? 'Only the project owner or task creator can delete this task.'
                          : undefined
                      }
                      className={clsx('inline-flex rounded', !canDeleteTask && 'cursor-not-allowed')}
                    >
                      <button
                        type="button"
                    className={clsx(
                      'text-xs font-medium',
                      canDeleteTask
                        ? 'text-[var(--tf-delete-fg)] hover:underline'
                        : 'pointer-events-none text-[var(--tf-delete-fg-muted)] opacity-60',
                    )}
                        onClick={() => {
                          if (!canDeleteTask) return
                          onRequestDeleteTask(t)
                        }}
                      >
                        Delete
                      </button>
                    </span>
                  </div>
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </>
  )
}
