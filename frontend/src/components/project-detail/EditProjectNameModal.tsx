import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'
import { Modal } from '../Modal'

type Props = {
  open: boolean
  onOpenChange: (v: boolean) => void
  name: string
  onNameChange: (v: string) => void
  error: string | null
  saving: boolean
  onSubmit: (e: React.FormEvent) => void
}

export function EditProjectNameModal({
  open,
  onOpenChange,
  name,
  onNameChange,
  error,
  saving,
  onSubmit,
}: Props) {
  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Edit project name"
      description="Rename the project. Edit the description in the description box on this page."
      footer={
        <>
          <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" form="project-edit-form" disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
        </>
      }
    >
      <form id="project-edit-form" onSubmit={(e) => void onSubmit(e)} className="space-y-4">
        {error ? (
          <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-300">{error}</p>
        ) : null}
        <div>
          <Label htmlFor="pe-name">Name</Label>
          <Input
            id="pe-name"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            className="mt-1"
            autoFocus
          />
        </div>
      </form>
    </Modal>
  )
}
