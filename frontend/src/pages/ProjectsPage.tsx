import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from '../api/client'
import { useAuth } from '../contexts/useAuth'
import type { Project } from '../types'
import { Layout } from '../components/Layout'
import { ProjectModal } from '../components/ProjectModal'
import { Button } from '../components/ui/button'

export function ProjectsPage() {
  const { user } = useAuth()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)

  const load = useCallback(async (opts?: { silent?: boolean }) => {
    const silent = opts?.silent === true
    if (!silent) {
      setLoading(true)
      setError(null)
    }
    try {
      const res = await api.listProjects()
      setProjects(res.projects ?? [])
      if (silent) setError(null)
    } catch (e) {
      if (!silent) {
        setError(e instanceof ApiError ? e.message : 'Failed to load projects')
      }
    } finally {
      if (!silent) setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <Layout>
      <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-fg">Projects</h1>
          <p className="mt-1 text-sm text-fg-muted">
            <span className="font-medium text-fg-dim">
              {user?.name?.trim() || user?.email || 'You'}
            </span>
            , here are the workspaces you own or contribute to.
          </p>
        </div>
        <Button type="button" onClick={() => setModalOpen(true)}>
          New project
        </Button>
      </div>

      <ProjectModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        onSubmit={async (data) => {
          const created = await api.createProject({
            name: data.name,
            description: data.description || undefined,
          })
          setProjects((prev) => {
            if (prev.some((p) => p.id === created.id)) return prev
            return [created, ...prev]
          })
          await load({ silent: true })
        }}
      />

      {loading ? (
        <p className="mt-10 text-center text-fg-muted">Loading projects…</p>
      ) : error ? (
        <div className="mt-10 rounded-xl border border-red-500/30 bg-red-500/10 p-4 text-center">
          <p className="text-red-300">{error}</p>
          <Button type="button" variant="secondary" className="mt-3" onClick={() => void load()}>
            Retry
          </Button>
        </div>
      ) : projects.length === 0 ? (
        <div className="mt-10 rounded-xl border border-dashed border-border bg-surface-2/40 p-10 text-center">
          <p className="text-fg-dim">No projects yet.</p>
          <p className="mt-2 text-sm text-fg-soft">Create one to start tracking tasks with your team.</p>
          <Button type="button" className="mt-6" onClick={() => setModalOpen(true)}>
            Create your first project
          </Button>
        </div>
      ) : (
        <ul className="mt-8 flex w-full min-w-0 flex-col gap-3">
          {projects.map((p) => (
            <li key={p.id}>
              <Link
                to={`/project/${p.id}`}
                className="flex h-[6.75rem] min-h-[6.75rem] max-h-[6.75rem] flex-col overflow-hidden rounded-xl border border-border bg-surface-2/50 px-4 py-3 transition-colors hover:border-blue-500/40 hover:bg-surface-2"
              >
                <h2 className="shrink-0 truncate text-base font-medium text-fg">{p.name}</h2>
                <p
                  className={`mt-2 line-clamp-2 text-sm leading-snug text-fg-muted break-words ${
                    !p.description ? 'italic text-fg-faint' : ''
                  }`}
                >
                  {p.description?.trim() ? p.description : 'No description'}
                </p>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </Layout>
  )
}
