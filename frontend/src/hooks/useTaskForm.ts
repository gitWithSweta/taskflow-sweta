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
}

type Options = {
  open: boolean
  projectId: string
  task?: Task | null
  onSubmit: (data: TaskFormSubmitData) => Promise<void>
  onSuccessClose: () => void
}

function utcTodayYYYYMMDD(): string {
  return new Date().toISOString().slice(0, 10)
}

export function useTaskForm({
  open,
  projectId,
  task,
  onSubmit,
  onSuccessClose,
}: Options) {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [status, setStatus] = useState<Task['status']>('todo')
  const [priority, setPriority] = useState<Task['priority']>('medium')
  const [assigneeId, setAssigneeId] = useState('')
  const [dueDate, setDueDate] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fields, setFields] = useState<Record<string, string>>({})
  const [assigneeSearch, setAssigneeSearch] = useState('')
  const [assigneePickerOpen, setAssigneePickerOpen] = useState(false)
  const assigneeInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    if (task) {
      setTitle(task.title)
      setDescription(task.description ?? '')
      setStatus(task.status)
      setPriority(task.priority)
      setAssigneeId(task.assignee_id ?? '')
      setDueDate(task.due_date ?? '')
    } else {
      setTitle('')
      setDescription('')
      setStatus('todo')
      setPriority('medium')
      setAssigneeId('')
      setDueDate('')
    }
    setError(null)
    setFields({})
    setAssigneeSearch('')
    setAssigneePickerOpen(false)
  }, [open, task?.id, projectId])

  useEffect(() => {
    if (!assigneePickerOpen) return
    const id = requestAnimationFrame(() => assigneeInputRef.current?.focus())
    return () => cancelAnimationFrame(id)
  }, [assigneePickerOpen])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setFields({})
    if (!title.trim()) {
      setFields({ title: 'is required' })
      return
    }
    if (!task && dueDate && dueDate < utcTodayYYYYMMDD()) {
      setFields({ due_date: 'Due date cannot be in the past' })
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
    loading,
    error,
    fields,
    assigneeSearch,
    setAssigneeSearch,
    assigneePickerOpen,
    setAssigneePickerOpen,
    assigneeInputRef,
    handleSubmit,
  }
}
