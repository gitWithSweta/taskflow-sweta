import * as Select from '@radix-ui/react-select'
import type { Task } from '../../types'
import { Label } from '../ui/label'

const statuses: Task['status'][] = ['todo', 'in_progress', 'done']
const priorities: Task['priority'][] = ['low', 'medium', 'high']

function Chevron() {
  return <span className="text-fg-muted">▾</span>
}

type Props = {
  status: Task['status']
  priority: Task['priority']
  onStatusChange: (v: Task['status']) => void
  onPriorityChange: (v: Task['priority']) => void
}

export function TaskStatusPrioritySelects({
  status,
  priority,
  onStatusChange,
  onPriorityChange,
}: Props) {
  return (
    <div className="grid grid-cols-1 gap-4">
      <div>
        <Label>Status</Label>
        <Select.Root value={status} onValueChange={(v) => onStatusChange(v as Task['status'])}>
          <Select.Trigger className="mt-1 flex h-10 w-full items-center justify-between rounded-lg border border-border bg-input px-3 text-sm text-fg">
            <Select.Value />
            <Select.Icon>
              <Chevron />
            </Select.Icon>
          </Select.Trigger>
          <Select.Portal>
            <Select.Content className="z-[200] overflow-hidden rounded-lg border border-border bg-panel shadow-xl">
              <Select.Viewport className="p-1">
                {statuses.map((s) => (
                  <Select.Item
                    key={s}
                    value={s}
                    className="cursor-pointer rounded-md px-3 py-2 text-sm text-fg outline-none data-[highlighted]:bg-[var(--tf-chip)]"
                  >
                    <Select.ItemText>{s.replace('_', ' ')}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>
      </div>
      <div>
        <Label>Priority</Label>
        <Select.Root value={priority} onValueChange={(v) => onPriorityChange(v as Task['priority'])}>
          <Select.Trigger className="mt-1 flex h-10 w-full items-center justify-between rounded-lg border border-border bg-input px-3 text-sm text-fg">
            <Select.Value />
            <Select.Icon>
              <Chevron />
            </Select.Icon>
          </Select.Trigger>
          <Select.Portal>
            <Select.Content className="z-[200] overflow-hidden rounded-lg border border-border bg-panel shadow-xl">
              <Select.Viewport className="p-1">
                {priorities.map((p) => (
                  <Select.Item
                    key={p}
                    value={p}
                    className="cursor-pointer rounded-md px-3 py-2 text-sm text-fg outline-none data-[highlighted]:bg-[var(--tf-chip)]"
                  >
                    <Select.ItemText>{p}</Select.ItemText>
                  </Select.Item>
                ))}
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>
      </div>
    </div>
  )
}
