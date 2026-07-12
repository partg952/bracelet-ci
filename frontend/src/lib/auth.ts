import type { User } from './types'

const KEY = 'bracelet_user'

export function saveSession(user: User): void {
  localStorage.setItem(KEY, JSON.stringify(user))
}

export function loadSession(): User | null {
  try {
    const raw = localStorage.getItem(KEY)
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

export function clearSession(): void {
  localStorage.removeItem(KEY)
}
