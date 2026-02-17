// src/shared/api/http.ts

import type { DevIdentity } from '@/shared/lib/devIdentity.ts'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const BASE = () => {
  const url = process.env.NEXT_PUBLIC_API_URL
  if (!url) throw new Error('NEXT_PUBLIC_API_URL is not set')
  return url
}

export async function request<T>(
  path: string,
  init?: RequestInit & { identity?: DevIdentity }
): Promise<T> {
  const ident = init?.identity

  const res = await fetch(BASE() + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(ident
        ? { 'X-User-Id': ident.userId, 'X-Role': ident.role }
        : {}),
      ...(init?.headers || {}),
    },
  })

  if (res.status === 204) return undefined as T

  const text = await res.text()
  if (!res.ok) {
    throw new ApiError(res.status, text || res.statusText)
  }

  return text ? (JSON.parse(text) as T) : (undefined as T)
}
