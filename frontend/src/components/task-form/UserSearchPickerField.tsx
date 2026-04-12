import type { RefObject } from 'react'
import { useMemo } from 'react'
import type { User } from '../../types'
import { Button } from '../ui/button'
import { Input } from '../ui/input'
import { Label } from '../ui/label'

type Props = {
  label: string
  hint?: string
  users: User[]
  usersLoading: boolean
  usersError: string | null
  currentUserId: string
  value: string
  onChange: (userId: string) => void
  searchOpen: boolean
  onSearchOpenChange: (open: boolean) => void
  search: string
  onSearchChange: (q: string) => void
  inputRef: RefObject<HTMLInputElement | null>
  inputId: string
  listboxId: string
  ariaListLabel: string
  allowUnassigned?: boolean
  fieldError?: string
  editButtonLabel?: string
}

export function UserSearchPickerField({
  label,
  hint,
  users,
  usersLoading,
  usersError,
  currentUserId,
  value,
  onChange,
  searchOpen,
  onSearchOpenChange,
  search,
  onSearchChange,
  inputRef,
  inputId,
  listboxId,
  ariaListLabel,
  allowUnassigned = false,
  fieldError,
  editButtonLabel = 'Edit',
}: Props) {
  const selectedUser = useMemo(
    () => (value ? users.find((u) => u.id === value) : undefined),
    [users, value],
  )

  const filteredUsers = useMemo(() => {
    const q = search.trim().toLowerCase()
    let list = users
    if (q) {
      list = users.filter(
        (u) => u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q),
      )
    }
    const sel = value ? users.find((u) => u.id === value) : undefined
    if (sel && !list.some((u) => u.id === value)) {
      return [sel, ...list]
    }
    return list
  }, [users, search, value])

  return (
    <div>
      <div className="flex items-center justify-between gap-2">
        <Label className="mb-0">{label}</Label>
        {searchOpen ? (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-8 shrink-0 text-xs text-fg-muted hover:text-fg"
            onClick={() => {
              onSearchOpenChange(false)
              onSearchChange('')
            }}
          >
            Cancel
          </Button>
        ) : (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-8 shrink-0 text-xs font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)]"
            disabled={usersLoading}
            onClick={() => onSearchOpenChange(true)}
          >
            {editButtonLabel}
          </Button>
        )}
      </div>
      {hint ? <p className="mt-1 text-xs text-fg-soft">{hint}</p> : null}
      {!searchOpen ? (
        <div className="mt-1 flex min-h-10 items-center rounded-lg border border-border bg-input px-3 text-sm">
          {value && selectedUser ? (
            <span className="truncate text-fg">
              {selectedUser.name}
              {value === currentUserId ? <span className="ml-1 text-xs text-fg-muted">(you)</span> : null}
              <span className="text-fg-muted"> — {selectedUser.email}</span>
            </span>
          ) : value && !selectedUser ? (
            <span className="text-amber-200/90">{allowUnassigned ? 'Loading assignee…' : 'Loading user…'}</span>
          ) : (
            <span className="text-fg-soft">{allowUnassigned ? 'Unassigned' : 'Unknown'}</span>
          )}
        </div>
      ) : (
        <>
          <Input
            ref={inputRef}
            id={inputId}
            type="text"
            autoComplete="off"
            inputMode="search"
            enterKeyHint="search"
            placeholder="Search by name or email…"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') e.preventDefault()
            }}
            className="mt-1"
            disabled={usersLoading}
            aria-controls={listboxId}
          />
          <div
            id={listboxId}
            role="listbox"
            aria-label={ariaListLabel}
            className="mt-2 max-h-44 overflow-y-auto rounded-lg border border-border bg-input p-1"
          >
            {allowUnassigned ? (
              <button
                type="button"
                role="option"
                aria-selected={value === ''}
                disabled={usersLoading}
                onClick={() => {
                  onChange('')
                  onSearchChange('')
                  onSearchOpenChange(false)
                }}
                className="flex w-full rounded-md px-3 py-2 text-left text-sm text-fg outline-none hover:bg-[var(--tf-chip)] disabled:opacity-50 aria-selected:bg-[var(--tf-chip)]"
              >
                {usersLoading ? 'Loading users…' : 'Unassigned'}
              </button>
            ) : null}
            {filteredUsers.map((u) => (
              <button
                key={u.id}
                type="button"
                role="option"
                aria-selected={value === u.id}
                onClick={() => {
                  onChange(u.id)
                  onSearchChange('')
                  onSearchOpenChange(false)
                }}
                className="flex w-full flex-col gap-0.5 rounded-md px-3 py-2 text-left text-sm outline-none hover:bg-[var(--tf-chip)] aria-selected:bg-[var(--tf-chip)]"
              >
                <span className="text-fg">
                  {u.name}
                  {u.id === currentUserId ? <span className="ml-1 text-xs text-fg-muted">(you)</span> : null}
                </span>
                <span className="text-xs text-fg-muted">{u.email}</span>
              </button>
            ))}
            {!usersLoading && users.length > 0 && filteredUsers.length === 0 ? (
              <p className="px-3 py-2 text-xs text-fg-soft">No users match your search.</p>
            ) : null}
          </div>
          {usersError ? (
            <p className="mt-1 text-xs text-amber-300">{usersError}</p>
          ) : allowUnassigned ? (
            <p className="mt-1 text-xs text-fg-soft">
              Search by name or email, then pick someone. The list closes after you choose.
            </p>
          ) : null}
        </>
      )}
      {fieldError ? <p className="mt-1 text-sm text-[var(--tf-delete-fg)]">{fieldError}</p> : null}
    </div>
  )
}
