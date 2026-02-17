'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useParams } from 'next/navigation'
import { api } from '@/shared/api/client'
import { getIdentity } from '@/shared/dev/auth'

type Row = {
  ID: string
  UserID: string
  GroupID: string

  ProgramID: string
  ProgramTitle: string
  GroupTitle: string

  Status: string
  Comment: string
  CreatedAt: string

  InterviewResult?: string | null
  InterviewComment?: string | null
  InterviewByRole?: string | null
  InterviewAt?: string | null
}

const NEXT: Record<string, { label: string; status: string }[]> = {
  submitted: [{ label: 'В работу', status: 'in_review' }],
  in_review: [
    { label: 'Одобрить', status: 'approved' },
    { label: 'Отклонить', status: 'rejected' },
  ],
  approved: [],
  rejected: [],
  cancelled: [],
}

export default function ProgramApplicationsPage() {
  const params = useParams<{ id: string }>()
  const programId = params?.id || ''

  const ident = getIdentity()
  const isStaff = ident.role === 'admin' || ident.role === 'moderator'
  const canChangeStatus = isStaff

  const [items, setItems] = useState<Row[]>([])
  const [err, setErr] = useState('')
  const [status, setStatus] = useState('')

  async function reload() {
    if (!programId) return
    try {
      setErr('')
      const data = await api.listApplications({ programId, status: status || undefined })
      setItems(data as Row[])
    } catch (e: any) {
      setErr(e?.message || String(e))
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [programId, status])

  async function move(appId: string, to: string) {
    const reason = prompt(`Причина для ${to}:`, 'ok') ?? ''
    try {
      await api.changeApplicationStatus(appId, to, reason)
      await reload()
    } catch (e: any) {
      const msg = e?.message || String(e)

      if (msg.includes('interview result is required') || msg.includes('ErrInterviewRequired')) {
        alert('Нельзя одобрить: нужно зафиксировать интервью (recommended) по этой заявке.')
        return
      }
      if (msg.includes('interview is not recommended') || msg.includes('ErrInterviewFailed')) {
        alert('Нельзя одобрить: интервью НЕ рекомендовано.')
        return
      }

      alert(msg)
    }
  }

  return (
    <div style={{ padding: 12 }}>
      <h2>Заявки на курс</h2>

      <div style={{ marginBottom: 8 }}>
        <Link href={`/program/${programId}`}>Назад к программе</Link>
      </div>

      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <span>Фильтр по статусу</span>
        <select value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">все</option>
          <option value="submitted">submitted</option>
          <option value="in_review">in_review</option>
          <option value="approved">approved</option>
          <option value="rejected">rejected</option>
          <option value="cancelled">cancelled</option>
        </select>
        <button onClick={reload}>Обновить</button>
      </div>

      {err ? <div style={{ color: 'crimson', marginTop: 8 }}>{err}</div> : null}

      {items.length === 0 ? (
        <div style={{ marginTop: 12 }}>Пока нет заявок.</div>
      ) : (
        <table style={{ width: '100%', marginTop: 12 }}>
          <thead>
            <tr>
              <th align="left">app</th>
              <th align="left">user</th>
              <th align="left">program</th>
              <th align="left">group</th>
              <th align="left">status</th>
              <th align="left">comment</th>
              <th align="left">interview</th>
              <th align="left">interview_comment</th>
              <th align="left">actions</th>
            </tr>
          </thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.ID}>
                <td>{a.ID}</td>
                <td>{a.UserID}</td>
                <td>{a.ProgramTitle} ({a.ProgramID})</td>
                <td>{a.GroupTitle} ({a.GroupID})</td>
                <td><b>{a.Status}</b></td>
                <td>{a.Comment}</td>
                <td>{a.InterviewResult ?? '-'}</td>
                <td>{a.InterviewComment ?? '-'}</td>
                <td>
                  {canChangeStatus
                    ? (NEXT[a.Status] || []).map((btn) => (
                        <button key={btn.status} onClick={() => move(a.ID, btn.status)} style={{ marginRight: 8 }}>
                          {btn.label}
                        </button>
                      ))
                    : null}

                  {/* teacher/user: интервью фиксируется тут */}
                  {!canChangeStatus ? (
                    <div style={{ marginTop: 6 }}>
                      <div style={{ opacity: 0.75 }}>Интервью фиксирует преподаватель</div>
                      {a.Status === 'in_review' ? (
                        <Link href={`/teacher/applications/${a.ID}/interview`}>Зафиксировать интервью</Link>
                      ) : (
                        <span style={{ opacity: 0.75 }}>Интервью доступно только в in_review</span>
                      )}
                    </div>
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
