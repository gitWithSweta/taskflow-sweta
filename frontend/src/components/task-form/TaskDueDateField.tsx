import { Input } from '../ui/input'
import { Label } from '../ui/label'

type Props = {
  value: string
  onChange: (v: string) => void
  fieldError?: string
  min?: string
}

export function TaskDueDateField({ value, onChange, fieldError, min }: Props) {
  return (
    <div>
      <Label htmlFor="ts-due">Due date</Label>
      <Input
        id="ts-due"
        type="date"
        value={value}
        min={min}
        onChange={(e) => onChange(e.target.value)}
        className="mt-1"
      />
      {fieldError ? <p className="mt-1 text-sm text-[var(--tf-delete-fg)]">{fieldError}</p> : null}
    </div>
  )
}
