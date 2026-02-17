// src/lib/api.ts
import { getIdentity } from './auth'

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

  return text ? JSON.parse(text) as T : (undefined as T)
}

export const api = {

  listPrograms: () => request<any[]>('/catalog/programs'),

  getProgram: (id: string) => request<any>(`/catalog/programs/${id}`),
  
  // staff: заявки (у тебя /applications для staff возвращает все)
  listApplications: (opts: { groupId?: string; programId?: string; status?: string; year?: number }) => {
    const qs = new URLSearchParams()
    if (opts.groupId) qs.set('group_id', opts.groupId)
    if (opts.programId) qs.set('program_id', opts.programId)
    if (opts.status) qs.set('status', opts.status)
    if (opts.year !== undefined) qs.set('year', String(opts.year))
    return request<any[]>(`/applications?${qs.toString()}`)
  },

  // admin programs/groups
  createProgram: (title: string, description: string) =>
    request<{ id: string }>('/admin/programs', {
      method: 'POST',
      body: JSON.stringify({ title, description }),
    }),

  publishProgram: (programId: string) =>
    request<void>(`/admin/programs/${programId}/publish`, { method: 'POST' }),

  listProgramsAdmin: () => request<any[]>('/admin/programs'),
  getProgramAdmin: (id: string) => request<any>(`/admin/programs/${id}`),

  updateProgram: (programId: string, patch: { title?: string; description?: string }) =>
    request<void>(`/admin/programs/${programId}`, {
      method: 'PATCH',
      body: JSON.stringify(patch),
    }),

  createCohort: (programId: string, year: number) =>
    request<{ id: string }>('/admin/cohorts', {
      method: 'POST',
      body: JSON.stringify({ program_id: programId, year }),
    }),

  createGroup: (args: {
    programId: string
    cohortId: string
    title: string
    capacity: number
    requiresInterview: boolean
    isOpen: boolean
  }) =>
    request<{ id: string }>('/admin/groups', {
      method: 'POST',
      body: JSON.stringify({
        program_id: args.programId,
        cohort_id: args.cohortId,
        title: args.title,
        capacity: args.capacity,
        requires_interview: args.requiresInterview,
        is_open: args.isOpen,
      }),
    }),

  updateGroup: (
    groupId: string,
    patch: { title?: string; capacity?: number; is_open?: boolean; requires_interview?: boolean }
  ) =>
    request<void>(`/admin/groups/${groupId}`, {
      method: 'PATCH',
      body: JSON.stringify(patch),
    }),

  assignTeacherToGroup: (groupId: string, teacherUserId: string) =>
    request<void>(`/admin/groups/${groupId}/teachers?teacher_user_id=${encodeURIComponent(teacherUserId)}`, {
      method: 'POST',
    }),

  listGroupTeachers: (groupId: string) =>
    request<{ group_id: string; teachers: string[] }>(`/admin/groups/${groupId}/teachers`),

  removeTeacherFromGroup: (groupId: string, teacherUserId: string) =>
    request<void>(`/admin/groups/${groupId}/teachers?teacher_user_id=${encodeURIComponent(teacherUserId)}`, {
      method: 'DELETE',
    }),

  changeApplicationStatus: (appId: string, status: string, reason: string) =>
    request<void>(`/admin/applications/${appId}/status`, {
      method: 'POST',
      body: JSON.stringify({ status, reason }),
    }),
  
  recordInterview: (appId: string, result: 'recommended' | 'not_recommended' | 'needs_more' | 'pending', comment: string) =>
    request<void>(`/teacher/applications/${appId}/interview`, {
      method: 'POST',
      body: JSON.stringify({ result, comment }),
    }),

}
