export function requiredTrimmed(value: string, message = 'is required'): string | null {
  return value.trim() ? null : message
}

export function passwordMinLength(password: string, min = 8): string | null {
  if (!password) return 'is required'
  if (password.length < min) return `must be at least ${min} characters`
  return null
}
