import { useEffect, useRef, useState } from 'react'
import { ApiError } from '../api/client'
import type { Task } from '../types'

export type TaskFormSubmitData = {
  title: string
  description: string
  status: Task['status']
  priority: Task['priority']
  assignee_id: string | null
  due_date: string
  creator_id?: string
}

type Options = {
  open: boolean
  projectId: string
  task?: Task | null
  isProjectOwner: boolean
  onSubmit: (data: TaskFormSubmitData) => Promise<void>
  onSuccessClose: () => void
}

export function useTaskForm({
  open,
  projectId,
  task,
  isProjectOwner,
  onSubmit,
  onSuccessClose,
}: Options) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<Task['status']>('todo')
  const [priority, setPriority] = useState<Task['priority']>('medium')
  const [assigneeId, setAssigneeId] = useState('')
  const [dueDate, setDueDate] = useState('')
  const [creatorId, setCreatorId] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const [assigneeSearch, setAssigneeSearch] = useState('')
  const [assigneePickerOpen, setAssigneePickerOpen] = useState(false)
  const [creatorSearch, setCreatorSearch] = useState('')
  const [creatorPickerOpen, setCreatorPickerOpen] = useState(false)
  const assigneeInputRef = useRef<HTMLInputElement>(null)
  const creatorInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    if (task) {
      setTitle(task.title)
      setDescription(task.description ?? '')
      setStatus(task.status)
      setPriority(task.priority)
      setAssigneeId(task.assignee_id ?? '')
      setDueDate(task.due_date ?? '')
      setCreatorId(task.creator_id ?? '')
    } else {
      setTitle('')
      setDescription('')
      setStatus('todo')
      setPriority('medium')
      setAssigneeId('')
      setDueDate('')
      setCreatorId('')
    }
    setError(null)
    setFields({})
    setAssigneeSearch('')
    setAssigneePickerOpen(false)
    setCreatorSearch('')
    setCreatorPickerOpen(false)
  }, [open, task?.id, projectId])

  useEffect(() => {
    if (!assigneePickerOpen) return
    const id = requestAnimationFrame(() => assigneeInputRef.current?.focus())
    return () => cancelAnimationFrame(id)
  }, [assigneePickerOpen])

  useEffect(() => {
    if (!creatorPickerOpen) return
    const id = requestAnimationFrame(() => creatorInputRef.current?.focus())
    return () => cancelAnimationFrame(id)
  }, [creatorPickerOpen])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setFields({})
    if (!title.trim()) {
      setFields({ title: 'is required' })
      return
    }
    if (task && isProjectOwner && !creatorId) {
      setFields({ creator_id: 'Task creator is required' })
      return
    }
    setLoading(true)
    try {
      await onSubmit({
        title: title.trim(),
        description: description.trim(),
        status,
        priority,
        assignee_id: assigneeId || null,
        due_date: dueDate,
        ...(task && isProjectOwner ? { creator_id: creatorId } : {}),
      })
      onSuccessClose()
    } catch (err) {
      if (err instanceof ApiError && err.fields) {
        setFields(err.fields)
      } else if (err instanceof ApiError) {
        setError(err.message)
      } else {
        setError('Something went wrong')
      }
    } finally {
      setLoading(false)
    }
  }

  return {
    title,
    setTitle,
    description,
    setDescription,
    status,
    setStatus,
    priority,
    setPriority,
    assigneeId,
    setAssigneeId,
    dueDate,
    setDueDate,
    creatorId,
    setCreatorId,
    loading,
    error,
    fields,
    assigneeSearch,
    setAssigneeSearch,
    assigneePickerOpen,
    setAssigneePickerOpen,
    creatorSearch,
    setCreatorSearch,
    creatorPickerOpen,
    setCreatorPickerOpen,
    assigneeInputRef,
    creatorInputRef,
    handleSubmit,
  }
}
