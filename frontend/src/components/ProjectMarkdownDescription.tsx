import { useCallback, useEffect, useState } from 'react'
import clsx from 'clsx'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import type { Components } from 'react-markdown'
import { Button } from './ui/button'
import { Label } from './ui/label'

const SHOW_MORE_THRESHOLD = 400

const mdComponents: Components = {
  h1: ({ children }) => <h1 className="mb-2 text-xl font-semibold text-fg">{children}</h1>,
  h2: ({ children }) => <h2 className="mb-2 text-lg font-semibold text-fg">{children}</h2>,
  h3: ({ children }) => <h3 className="mb-1.5 text-base font-semibold text-fg">{children}</h3>,
  p: ({ children }) => <p className="mb-2 text-sm leading-relaxed text-fg-dim last:mb-0">{children}</p>,
  ul: ({ children }) => (
    <ul className="mb-2 list-inside list-disc space-y-1 text-sm text-fg-dim">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="mb-2 list-inside list-decimal space-y-1 text-sm text-fg-dim">{children}</ol>
  ),
  li: ({ children }) => <li className="ms-1">{children}</li>,
  code: (props) => {
    const { className, children, ...rest } = props
    const inline = !className
    if (inline) {
      return (
        <code
          className="rounded bg-[var(--tf-chip)] px-1 py-0.5 font-mono text-[0.8em] text-fg"
          {...rest}
        >
          {children}
        </code>
      )
    }
    return (
      <code
        className={`block overflow-x-auto rounded-md bg-[var(--tf-code-bg)] p-2 font-mono text-xs text-[var(--tf-code-fg)] ${className ?? ''}`}
        {...rest}
      >
        {children}
      </code>
    )
  },
  pre: ({ children }) => (
    <pre className="mb-2 overflow-x-auto rounded-lg bg-[var(--tf-code-bg)] p-3 text-xs text-[var(--tf-code-fg)]">
      {children}
    </pre>
  ),
  a: ({ href, children }) => (
    <a href={href} className="underline" target="_blank" rel="noreferrer">
      {children}
    </a>
  ),
  strong: ({ children }) => <strong className="font-semibold text-fg">{children}</strong>,
  em: ({ children }) => <em className="italic text-[var(--tf-em)]">{children}</em>,
  blockquote: ({ children }) => (
    <blockquote className="mb-2 border-l-2 border-[var(--tf-blockquote-border)] pl-3 text-sm italic text-fg-muted">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-3 border-border" />,
  table: ({ children }) => (
    <div className="mb-2 overflow-x-auto">
      <table className="min-w-full border-collapse text-left text-xs text-fg-dim">{children}</table>
    </div>
  ),
  th: ({ children }) => (
    <th className="border border-border bg-[var(--tf-table-header)] px-2 py-1 font-medium text-fg">
      {children}
    </th>
  ),
  td: ({ children }) => <td className="border border-border px-2 py-1">{children}</td>,
}

type Props = {
  description: string | null | undefined
  canEdit: boolean
  onSave: (markdown: string) => Promise<void>
  className?: string
  editorRows?: number
}

export function ProjectMarkdownDescription({
  description,
  canEdit,
  onSave,
  className,
  editorRows = 12,
}: Props) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState('')
  const [saving, setSaving] = useState(false)
  const [expanded, setExpanded] = useState(false)

  useEffect(() => {
    if (!editing) {
      setDraft(description ?? '')
    }
  }, [description, editing])

  const md = (description ?? '').trim()
  const needsLongToggle = md.length > SHOW_MORE_THRESHOLD

  const enterEdit = useCallback(() => {
    if (!canEdit) return
    setDraft(description ?? '')
    setEditing(true)
  }, [canEdit, description])

  async function save() {
    setSaving(true)
    try {
      await onSave(draft.trim())
      setEditing(false)
    } finally {
      setSaving(false)
    }
  }

  function cancelEdit() {
    setEditing(false)
    setDraft(description ?? '')
  }

  return (
    <div
      className={clsx(
        'mt-3 rounded-xl border border-border bg-input p-4 shadow-[var(--tf-card-shadow)]',
        className,
      )}
    >
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <Label className="mb-0">Description</Label>
        {canEdit && editing ? (
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" size="sm" onClick={cancelEdit}>
              Cancel
            </Button>
            <Button type="button" size="sm" disabled={saving} onClick={() => void save()}>
              {saving ? 'Saving…' : 'Save'}
            </Button>
          </div>
        ) : canEdit ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-8 shrink-0 text-xs font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)]"
            onClick={enterEdit}
          >
            Edit
          </Button>
        ) : null}
      </div>

      {editing ? (
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          rows={editorRows}
          className="w-full resize-y rounded-lg border border-border bg-[var(--tf-editor-bg)] px-3 py-2 font-mono text-sm leading-relaxed text-fg placeholder:text-fg-soft focus:border-blue-500 focus:outline-none"
          placeholder="Markdown supported: **bold**, lists, `code`, ## headings…"
          spellCheck={false}
        />
      ) : md ? (
        <div>
          <div
            role={canEdit ? 'button' : undefined}
            tabIndex={canEdit ? 0 : undefined}
            onClick={() => canEdit && enterEdit()}
            onKeyDown={(e) => {
              if (!canEdit) return
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                enterEdit()
              }
            }}
            className={
              canEdit
                ? 'cursor-pointer rounded-lg outline-none ring-offset-2 ring-offset-input focus-visible:ring-2 focus-visible:ring-blue-500'
                : ''
            }
          >
            <div
              className={
                needsLongToggle && !expanded
                  ? 'max-h-48 overflow-hidden'
                  : needsLongToggle && expanded
                    ? 'max-h-none'
                    : ''
              }
            >
              <div className="markdown-preview">
                <ReactMarkdown remarkPlugins={[remarkGfm]} components={mdComponents}>
                  {md}
                </ReactMarkdown>
              </div>
            </div>
          </div>
          {needsLongToggle ? (
            <button
              type="button"
              className="mt-2 text-xs font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)]"
              onClick={(e) => {
                e.stopPropagation()
                setExpanded((v) => !v)
              }}
            >
              {expanded ? 'Show less' : 'Show more'}
            </button>
          ) : null}
        </div>
      ) : (
        <button
          type="button"
          disabled={!canEdit}
          onClick={() => canEdit && enterEdit()}
          className={
            canEdit
              ? 'w-full rounded-lg border border-dashed border-border bg-[var(--tf-description-shell)] px-3 py-8 text-left text-sm italic text-fg-faint transition-colors hover:border-blue-500/40 hover:text-fg-muted'
              : 'w-full rounded-lg border border-dashed border-border/60 px-3 py-6 text-left text-sm italic text-fg-faint'
          }
        >
          {canEdit ? 'No description yet — click to add (Markdown supported).' : 'No description.'}
        </button>
      )}
    </div>
  )
}
