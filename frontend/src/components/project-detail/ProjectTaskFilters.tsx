import type { User } from '../../types'
import { Label } from '../ui/label'

export type TaskSortMode =
  | 'alpha_asc'
  | 'alpha_desc'
  | 'created_asc'
  | 'created_desc'
  | 'due_asc'
  | 'due_desc'

type Props = {
  taskSort: TaskSortMode
  onTaskSortChange: (v: TaskSortMode) => void
  statusFilter: string
  onStatusFilterChange: (v: string) => void
  assigneeFilter: string
  onAssigneeFilterChange: (v: string) => void
  assigneeUsers: User[]
  currentUserId: string | undefined
  tasksLoading: boolean
}

export function ProjectTaskFilters({
  taskSort,
  onTaskSortChange,
  statusFilter,
  onStatusFilterChange,
  assigneeFilter,
  onAssigneeFilterChange,
  assigneeUsers,
  currentUserId,
  tasksLoading,
}: Props) {
  const selectClass =
    'mt-1.5 h-10 w-full min-w-0 rounded-lg border border-border bg-input px-3 text-sm text-fg focus:border-blue-500 focus:outline-none'

  return (
    <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3 sm:items-end">
      <div className="min-w-0">
        <Label htmlFor="taskflow-task-sort" className="mb-0 block">
          Sort tasks
        </Label>
        <select
          id="taskflow-task-sort"
          className={selectClass}
          value={taskSort}
          onChange={(e) => onTaskSortChange(e.target.value as TaskSortMode)}
        >
          <option value="alpha_asc">Title (A–Z)</option>
          <option value="alpha_desc">Title (Z–A)</option>
          <option value="created_asc">Created date (oldest first)</option>
          <option value="created_desc">Created date (newest first)</option>
          <option value="due_asc">Due date (soonest first)</option>
          <option value="due_desc">Due date (latest first)</option>
        </select>
      </div>
      <div className="min-w-0">
        <Label htmlFor="taskflow-task-status-filter" className="mb-0 block">
          Status
        </Label>
        <select
          id="taskflow-task-status-filter"
          className={selectClass}
          value={statusFilter}
          onChange={(e) => onStatusFilterChange(e.target.value)}
        >
          <option value="">All</option>
          <option value="todo">To do</option>
          <option value="in_progress">In progress</option>
          <option value="done">Done</option>
        </select>
      </div>
      <div className="min-w-0">
        <Label htmlFor="taskflow-task-assignee-filter" className="mb-0 block">
          Assignee
        </Label>
        <select
          id="taskflow-task-assignee-filter"
          className={selectClass}
          value={assigneeFilter}
          onChange={(e) => onAssigneeFilterChange(e.target.value)}
        >
          <option value="">All</option>
          {currentUserId ? <option value={currentUserId}>Me</option> : null}
          {assigneeUsers
            .filter((u) => u.id !== currentUserId)
            .map((u) => (
              <option key={u.id} value={u.id}>
                {u.name} — {u.email}
              </option>
            ))}
        </select>
      </div>
      {tasksLoading ? (
        <div className="flex items-end sm:col-span-3">
          <span className="pb-1 text-sm text-fg-soft">Refreshing tasks…</span>
        </div>
      ) : null}
    </div>
  )
}
