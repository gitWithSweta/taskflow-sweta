import type { User } from '../../types'
import { Button } from '../ui/button'
import { Label } from '../ui/label'
import { Modal } from '../Modal'

type Props = {
  open: boolean
  onOpenChange: (v: boolean) => void
  users: User[]
  transferOwnerId: string
  onTransferOwnerIdChange: (id: string) => void
  error: string | null
  saving: boolean
  onSubmit: (e: React.FormEvent) => void
}

export function TransferOwnershipModal({
  open,
  onOpenChange,
  users,
  transferOwnerId,
  onTransferOwnerIdChange,
  error,
  saving,
  onSubmit,
}: Props) {
  return (
    <Modal
      open={open}
      onOpenChange={onOpenChange}
      title="Transfer project ownership"
      description="The new owner will be able to delete the project, transfer ownership, and edit the description. You will keep access if you still have tasks in this project."
      footer={
        <>
          <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" form="transfer-owner-form" disabled={saving}>
            {saving ? 'Saving…' : 'Transfer'}
          </Button>
        </>
      }
    >
      <form id="transfer-owner-form" onSubmit={(e) => void onSubmit(e)} className="space-y-4">
        {error ? (
          <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-300">{error}</p>
        ) : null}
        <div>
          <Label htmlFor="transfer-owner-select">New owner</Label>
          <select
            id="transfer-owner-select"
            className="mt-1 h-10 w-full rounded-lg border border-border bg-input px-3 text-sm text-fg"
            value={transferOwnerId}
            onChange={(e) => onTransferOwnerIdChange(e.target.value)}
            required
          >
            <option value="">Select a user…</option>
            {users.map((u) => (
              <option key={u.id} value={u.id}>
                {u.name} — {u.email}
              </option>
            ))}
          </select>
        </div>
      </form>
    </Modal>
  )
}
