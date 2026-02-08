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
      'X-User-Id': ident.userId, // потом заменить заголовки X-* на Authorization
      'X-Role': ident.role,
      ...(init?.headers || {}),
    },
  })

  if (res.status === 204) return undefined as T

  const text = await res.text()
  if (!res.ok) {
    // backend пока шлёт plain text — под это и подстроимся
    throw new ApiError(res.status, text || res.statusText)
  }

  return text ? JSON.parse(text) as T : (undefined as T)
}

export const api = {

  listApplications: (opts: { groupId?: string; programId?: string; status?: string }) => {
  const qs = new URLSearchParams()
  if (opts.groupId) qs.set('group_id', opts.groupId)
  if (opts.programId) qs.set('program_id', opts.programId)
  if (opts.status) qs.set('status', opts.status)
  return request<any[]>(`/applications?${qs.toString()}`)
  },
  // catalog
  listPrograms: () => request<any[]>('/catalog/programs'),
  getProgram: (id: string) => request<any>(`/catalog/programs/${id}`),

  // applications
  createApplication: (groupId: string, comment: string) =>
    request<{ id: string }>('/enrollments/applications', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupId, comment }),
    }),
  listMyApplications: () => request<any[]>('/enrollments/me/applications'),

  // learn
  listMaterials: (groupId: string) => request<any[]>(`/learn/groups/${groupId}/materials`),

  // teacher
  myGroups: () => request<any[]>('/teacher/groups'),
  groupStudents: (groupId: string) => request<any>(`/teacher/groups/${groupId}/students`),

  recordInterview: (appId: string, result: 'recommended' | 'not_recommended' | 'needs_more' | 'pending', comment: string) =>
    request<void>(`/teacher/applications/${appId}/interview`, {
      method: 'POST',
      body: JSON.stringify({ result, comment }),
    }),

  // helper
  me: () => getIdentity(),

  // admin
  createProgram: (title: string, description: string) =>
    request<{ id: string }>('/admin/programs', {
      method: 'POST',
      body: JSON.stringify({ title, description }),
    }),

  publishProgram: (programId: string) =>
    request<void>(`/admin/programs/${programId}/publish`, { method: 'POST' }),

  listProgramsAdmin: () => request<any[]>('/admin/programs'),

  getProgramAdmin: (id: string) => request<any>(`/admin/programs/${id}`),

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

  updateGroup: (
    groupId: string,
    patch: { title?: string; capacity?: number; is_open?: boolean; requires_interview?: boolean }
  ) =>
    request<void>(`/admin/groups/${groupId}`, {
      method: 'PATCH',
      body: JSON.stringify(patch),
    }),

  changeApplicationStatus: (appId: string, status: string, reason: string) =>
    request<void>(`/admin/applications/${appId}/status`, {
      method: 'POST',
      body: JSON.stringify({ status, reason }),
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

}
