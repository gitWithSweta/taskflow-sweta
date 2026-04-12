import { ApiError } from '../api/client'

export function applyApiFormError(
  err: unknown,
  setFields: (f: Record<string, string>) => void,
  setError: (s: string | null) => void,
  fallback: string,
  mapMessage?: (msg: string) => string,
) {
  if (err instanceof ApiError && err.fields) {
    setFields(err.fields)
    return
  }
  if (err instanceof ApiError) {
    setError(mapMessage ? mapMessage(err.message) : err.message)
    return
  }
  setError(fallback)
}
