type Props = {
  byStatus: Record<string, number> | undefined
}

export function ProjectStatusStats({ byStatus }: Props) {
  if (!byStatus) return null
  return (
    <div className="mt-4 flex flex-wrap gap-2 text-xs font-medium text-fg-muted">
      {Object.entries(byStatus).map(([k, v]) => (
        <span
          key={k}
          className="rounded-full border border-border/70 bg-[var(--tf-chip)] px-2.5 py-1"
        >
          {k.replace('_', ' ')}: {v}
        </span>
      ))}
    </div>
  )
}
