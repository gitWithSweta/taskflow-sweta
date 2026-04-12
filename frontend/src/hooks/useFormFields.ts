import { useCallback, useState } from 'react'
import { applyApiFormError } from '../lib/applyApiFormError'

export function useFormFields() {
  const [fields, setFields] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)

  const clearErrors = useCallback(() => {
    setFields({})
    setError(null)
  }, [])

  const applyApiError = useCallback(
    (err: unknown, fallback: string, mapMessage?: (msg: string) => string) => {
      applyApiFormError(err, setFields, setError, fallback, mapMessage)
    },
    [],
  )

  return { fields, setFields, error, setError, clearErrors, applyApiError }
}
