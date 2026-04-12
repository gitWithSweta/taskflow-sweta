import { useState } from 'react'

type Props = {
  text: string
  threshold?: number
  className?: string
  collapsedClass?: string
  expandedMaxHeightClass?: string
  buttonClassName?: string
}

export function ClampableText({
  text,
  threshold = 160,
  className = '',
  collapsedClass = 'line-clamp-3',
  expandedMaxHeightClass,
  buttonClassName =
    'mt-1 text-xs font-medium text-[var(--tf-link)] hover:text-[var(--tf-link-hover)]',
}: Props) {
  const [expanded, setExpanded] = useState(false)
  const needsToggle = text.length > threshold

  if (!needsToggle) {
    return <p className={className}>{text}</p>
  }

  return (
    <div>
      <p
        className={
          expanded
            ? [className, expandedMaxHeightClass].filter(Boolean).join(' ')
            : [className, collapsedClass].filter(Boolean).join(' ')
        }
      >
        {text}
      </p>
      <button
        type="button"
        className={buttonClassName}
        onClick={(e) => {
          e.preventDefault()
          e.stopPropagation()
          setExpanded((v) => !v)
        }}
      >
        {expanded ? 'Show less' : 'Show more'}
      </button>
    </div>
  )
}
