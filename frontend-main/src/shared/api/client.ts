// src/shared/api/client.ts
import { getIdentity } from '@/shared/dev/auth'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const ident = getIdentity()

  const res = await fetch(BASE + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      'X-User-Id': ident.userId,
      'X-Role': ident.role,
      ...(init?.headers || {}),
    },
  })

  if (res.status === 204) return undefined as T

  const text = await res.text()
  if (!res.ok) throw new ApiError(res.status, text || res.statusText)

  return text ? (JSON.parse(text) as T) : (undefined as T)
}

export const api = {
  // public catalog
  listPrograms: () => request<any[]>('/catalog/programs'),
  getProgram: (id: string) => request<any>(`/catalog/programs/${id}`),

  // teacher access check
  teacherProgramAccess: (programId: string) =>
    request<{ ok: boolean }>(`/teacher/programs/${programId}/access`),

  listMyApplications: () => request<any[]>('/enrollments/me/applications'),
  createApplication: (groupId: string, comment: string) =>
    request<{ id: string }>('/enrollments/applications', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupId, comment }),
    }),

  listMaterials: (groupId: string) => request<any[]>(`/learn/groups/${groupId}/materials`),

  // teacher interview (для страницы интервью в main)
  recordInterview: (
    appId: string,
    result: 'recommended' | 'not_recommended' | 'needs_more' | 'pending',
    comment: string
  ) =>
    request<void>(`/teacher/applications/${appId}/interview`, {
      method: 'POST',
      body: JSON.stringify({ result, comment }),
    }),
  
  // applications (teacher can read; student uses me/*)
  listApplications: (opts: { groupId?: string; programId?: string; status?: string }) => {
    const qs = new URLSearchParams()
    if (opts.groupId) qs.set('group_id', opts.groupId)
    if (opts.programId) qs.set('program_id', opts.programId)
    if (opts.status) qs.set('status', opts.status)
    return request<any[]>(`/applications?${qs.toString()}`)
  },

  changeApplicationStatus: (appId: string, status: string, reason: string) =>
    request<void>(`/admin/applications/${appId}/status`, {
      method: 'POST',
      body: JSON.stringify({ status, reason }),
    }),
  
  cancelMyApplication: (appId: string) =>
    request<void>(`/enrollments/applications/${appId}/cancel`, { method: 'POST' }),

}
