import { useState } from 'react'
import { useFormFields } from '../hooks/useFormFields'
import { requiredTrimmed } from '../lib/validators'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Modal } from './Modal'

type Props = {
  open: boolean
  onOpenChange: (v: boolean) => void
  onSubmit: (data: { name: string; description: string }) => Promise<void>
}

export function ProjectModal({ open, onOpenChange, onSubmit }: Props) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [loading, setLoading] = useState(false)
  const { fields, setFields, error, clearErrors, applyApiError } = useFormFields()

  function reset() {
    setName('')
    setDescription('')
    clearErrors()
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    clearErrors()
    const nameErr = requiredTrimmed(name, 'is required')
    if (nameErr) {
      setFields({ name: nameErr })
      return
    }
    setLoading(true)
    try {
      await onSubmit({ name: name.trim(), description: description.trim() })
      reset()
      onOpenChange(false)
    } catch (err) {
      applyApiError(err, 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      open={open}
      onOpenChange={(v) => {
        if (!v) reset()
        onOpenChange(v)
      }}
      title="New project"
      description="Give your project a name. You can add tasks next."
      footer={
        <>
          <Button type="button" variant="secondary" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="submit" form="project-form" disabled={loading}>
            {loading ? 'Creating…' : 'Create'}
          </Button>
        </>
      }
    >
      <form id="project-form" onSubmit={handleSubmit} className="space-y-4 text-fg">
        {error ? (
          <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-300">{error}</p>
        ) : null}
        <div>
          <Label htmlFor="p-name">Name</Label>
          <Input
            id="p-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="mt-1"
            autoFocus
          />
          {fields.name ? <p className="mt-1 text-sm text-red-400">{fields.name}</p> : null}
        </div>
        <div>
          <Label htmlFor="p-desc">Description (optional)</Label>
          <Input
            id="p-desc"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="mt-1"
          />
        </div>
      </form>
    </Modal>
  )
}
