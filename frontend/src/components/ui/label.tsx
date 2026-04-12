import * as LabelPrimitive from '@radix-ui/react-label'
import { clsx } from 'clsx'
import type { ComponentProps } from 'react'

export function Label({ className, ...props }: ComponentProps<typeof LabelPrimitive.Root>) {
  return (
    <LabelPrimitive.Root
      className={clsx('text-sm font-medium text-fg-dim', className)}
      {...props}
    />
  )
}
