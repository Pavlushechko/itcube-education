import type { Role } from '@/shared/types/education'

export type DevIdentity = {
  label: string
  userId: string
  role: Role
}

export const DEV_IDENTITIES: DevIdentity[] = [
  { label: 'Ученик', userId: '11111111-1111-1111-1111-111111111111', role: 'user' },
  { label: 'Модератор', userId: '33333333-3333-3333-3333-333333333333', role: 'moderator' },
  { label: 'Админ', userId: 'aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa', role: 'admin' },
  { label: 'Преподаватель (назначенный)', userId: '22222222-2222-2222-2222-222222222222', role: 'user' },
]

const COOKIE_NAME = 'itcube.dev.identity'

export function getDefaultIdentity(): DevIdentity {
  return DEV_IDENTITIES[0]
}

// клиент: читаем cookie
export function getIdentityClient(): DevIdentity {
  if (typeof document === 'undefined') return getDefaultIdentity()

  const raw = document.cookie
    .split('; ')
    .find((x) => x.startsWith(COOKIE_NAME + '='))
    ?.split('=')[1]

  if (!raw) return getDefaultIdentity()

  try {
    const json = decodeURIComponent(raw)
    return JSON.parse(json) as DevIdentity
  } catch {
    return getDefaultIdentity()
  }
}

// сервер: читаем cookie через next/headers
export function getIdentityServer(cookieValue?: string | null): DevIdentity {
  if (!cookieValue) return getDefaultIdentity()
  try {
    return JSON.parse(cookieValue) as DevIdentity
  } catch {
    return getDefaultIdentity()
  }
}

export function setIdentityClient(id: DevIdentity) {
  const json = encodeURIComponent(JSON.stringify(id))
  document.cookie = `${COOKIE_NAME}=${json}; Path=/; SameSite=Lax`
}

export function clearIdentityClient() {
  document.cookie = `${COOKIE_NAME}=; Max-Age=0; Path=/`
}

export { COOKIE_NAME }
