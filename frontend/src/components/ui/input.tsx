import { clsx } from 'clsx'
import { forwardRef, type InputHTMLAttributes } from 'react'

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  function Input({ className, ...props }, ref) {
    return (
      <input
        ref={ref}
        className={clsx(
          'h-10 w-full rounded-lg border border-border bg-input px-3 text-sm text-fg placeholder:text-fg-soft focus:border-blue-500 focus:outline-none',
          className,
        )}
        {...props}
      />
    )
  },
)
