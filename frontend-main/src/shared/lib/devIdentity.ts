// src/shared/lib/devIdentity.ts

export type Role = 'user' | 'moderator' | 'admin'

export type DevIdentity = {
  label: string
  userId: string
  role: Role
}

const KEY = 'itcube.dev.identity'

export const DEV_IDENTITIES: DevIdentity[] = [
  { label: 'Ученик', userId: '11111111-1111-1111-1111-111111111111', role: 'user' },
  { label: 'Модератор', userId: '33333333-3333-3333-3333-333333333333', role: 'moderator' },
  { label: 'Админ', userId: 'aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa', role: 'admin' },
  { label: 'Преподаватель (назначенный)', userId: '22222222-2222-2222-2222-222222222222', role: 'user' },
]

export function getDevIdentity(): DevIdentity {
  if (typeof window === 'undefined') return DEV_IDENTITIES[0]
  const raw = localStorage.getItem(KEY)
  if (raw) {
    try { return JSON.parse(raw) } catch {}
  }
  return DEV_IDENTITIES[0]
}

export function setDevIdentity(id: DevIdentity) {
  localStorage.setItem(KEY, JSON.stringify(id))
}

export function clearDevIdentity() {
  localStorage.removeItem(KEY)
}
