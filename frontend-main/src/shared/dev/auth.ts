// src/shared/dev/auth.ts
export type Role = 'user'

export type DevIdentity = {
  label: string
  userId: string
  role: Role
}

const KEY = 'itcube.dev.identity'

export const DEV_IDENTITIES: DevIdentity[] = [
  { label: 'Ученик', userId: '11111111-1111-1111-1111-111111111111', role: 'user' },
  { label: 'Преподаватель (назначенный)', userId: '22222222-2222-2222-2222-222222222222', role: 'user' },
]

export function getIdentity(): DevIdentity {
  if (typeof window === 'undefined') return DEV_IDENTITIES[0]
  const raw = localStorage.getItem(KEY)
  if (raw) {
    try { return JSON.parse(raw) } catch {}
  }
  return DEV_IDENTITIES[0]
}

export function setIdentity(id: DevIdentity) {
  if (typeof window === 'undefined') return
  localStorage.setItem(KEY, JSON.stringify(id))
}
