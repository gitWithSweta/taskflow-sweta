import clsx from 'clsx'
import { Link } from 'react-router-dom'
import { ProjectMarkdownDescription } from '../ProjectMarkdownDescription'
import { Button } from '../ui/button'

type Props = {
  projectName: string
  projectDescription: string | null | undefined
  isProjectOwner: boolean
  onEditProjectName: () => void
  onNewTask: () => void
  onTransferOwnership: () => void
  onDeleteProject: () => void
  onSaveDescription: (markdown: string) => Promise<void>
}

export function ProjectDetailHeader({
  projectName,
  projectDescription,
  isProjectOwner,
  onEditProjectName,
  onNewTask,
  onTransferOwnership,
  onDeleteProject,
  onSaveDescription,
}: Props) {
  return (
    <div className="border-b border-border pb-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <Link
            to="/project"
            className="text-sm font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)] hover:underline"
          >
            ← All projects
          </Link>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight text-fg">{projectName}</h1>
          {isProjectOwner ? (
            <Button
              type="button"
              variant="secondary"
              size="sm"
              className="mt-2"
              onClick={onEditProjectName}
            >
              Edit project name
            </Button>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button type="button" size="md" onClick={onNewTask}>
            New task
          </Button>
          {isProjectOwner ? (
            <Button type="button" variant="secondary" size="md" onClick={onTransferOwnership}>
              Transfer ownership
            </Button>
          ) : null}
          <span
            title={!isProjectOwner ? 'Only the project owner can delete this project.' : undefined}
            className={clsx('inline-flex rounded-lg', !isProjectOwner && 'cursor-not-allowed')}
          >
            <Button
              type="button"
              variant="danger"
              size="md"
              className={clsx(!isProjectOwner && 'pointer-events-none opacity-40')}
              onClick={() => {
                if (!isProjectOwner) return
                onDeleteProject()
              }}
            >
              Delete project
            </Button>
          </span>
        </div>
      </div>
      <ProjectMarkdownDescription
        description={projectDescription}
        canEdit={isProjectOwner}
        onSave={onSaveDescription}
      />
    </div>
  )
}
