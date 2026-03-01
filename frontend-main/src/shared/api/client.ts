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

  // assignments
  teacherListAssignments: (groupId: string) =>
    request<any[]>(`/teacher/groups/${groupId}/assignments`),

  learnerListAssignments: (groupId: string) =>
    request<any[]>(`/learn/groups/${groupId}/assignments`),

  listAssignmentsLearner: (groupId: string) =>
    request<any[]>(`/learn/groups/${groupId}/assignments`),

  // submissions
  submitAssignment: (assignmentId: string, contentType: string, content: string) =>
    request<{ id: string }>(`/learn/assignments/${assignmentId}/submissions`, {
      method: 'POST',
      body: JSON.stringify({ content_type: contentType, content }),
    }),

  mySubmission: (assignmentId: string) =>
    request<any>(`/learn/assignments/${assignmentId}/submissions/me`),

  teacherListSubmissions: (groupId: string, status?: string) => {
    const qs = new URLSearchParams()
    if (status) qs.set('status', status)
    const tail = qs.toString() ? `?${qs.toString()}` : ''
    return request<any[]>(`/teacher/groups/${groupId}/submissions${tail}`)
  },

  teacherCreateMaterial: (
    groupId: string,
    body: { type: string; title: string; content?: string; attachments?: string[] }
  ) =>
    request<{ id: string }>(`/teacher/groups/${groupId}/materials`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  
  teacherListMaterials: (groupId: string) =>
    request<any[]>(`/teacher/groups/${groupId}/materials`),

  teacherGetMaterial: (materialId: string) =>
    request<any>(`/teacher/materials/${materialId}`),

  teacherGetMaterialPage: (materialId: string) =>
    request<any>(`/teacher/materials/${materialId}`),

  teacherReviewSubmission: (submissionId: string, grade: number | null, comment: string) =>
    request<void>(`/teacher/submissions/${submissionId}/review`, {
      method: 'POST',
      body: JSON.stringify({ grade, comment }),
    }),
  
  myGroups: () => request<any[]>('/teacher/groups'),

  teacherCreateAssignment: (
    groupId: string,
    body: { title: string; description?: string; due_at?: string }
  ) =>
    request<{ id: string }>(`/teacher/groups/${groupId}/assignments`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  
  getLearnGroupInfo: (groupId: string) =>
    request<any>(`/learn/groups/${groupId}`),

  createUploadURL: (body: { original_name: string; mime_type: string; size_bytes: number }) =>
    request<{ file_id: string; put_url: string; object_key: string; bucket: string }>(`/files/upload-url`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  downloadURL: (fileId: string) =>
    request<{ url: string }>(`/files/${fileId}/download-url`),

  learnerGetMaterialPage: (materialId: string) =>
    request<any>(`/learn/materials/${materialId}`),

  teacherUpdateMaterial: (materialId: string, body: { type?: string; title?: string; content?: string; external_url?: string | null }) =>
    request<void>(`/teacher/materials/${materialId}/update`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  // upload file (multipart) -> {file_id}
  uploadFile: async (file: File) => {
    const ident = getIdentity()
    const fd = new FormData()
    fd.append('file', file)

    const res = await fetch('/api/files/upload', {
      method: 'POST',
      headers: {
        'X-User-Id': ident.userId,
        'X-Role': ident.role,
      },
      body: fd,
    })

    const text = await res.text()
    if (!res.ok) throw new ApiError(res.status, text || res.statusText)
    return JSON.parse(text) as { file_id: string }
  },

  teacherAddMaterialFile: (materialId: string, fileId: string) =>
    request<void>(`/teacher/materials/${materialId}/files/add`, {
      method: 'POST',
      body: JSON.stringify({ file_id: fileId }),
    }),

  teacherRemoveMaterialFile: (materialId: string, fileId: string) =>
    request<void>(`/teacher/materials/${materialId}/files/remove`, {
      method: 'POST',
      body: JSON.stringify({ file_id: fileId }),
    }),

  downloadFile: async (fileId: string) => {
    const ident = getIdentity()
    const res = await fetch(`/api/files/${fileId}/download`, {
      method: 'GET',
      headers: {
        'X-User-Id': ident.userId,
        'X-Role': ident.role,
      },
    })
    if (!res.ok) throw new ApiError(res.status, await res.text())
    const blob = await res.blob()

    const cd = res.headers.get('content-disposition') || ''
    let filename = 'file'
    const m = cd.match(/filename\*\=UTF-8''([^;]+)/i)
    if (m?.[1]) {
      try { filename = decodeURIComponent(m[1]) } catch {}
    }

    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = filename
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
  },
}
