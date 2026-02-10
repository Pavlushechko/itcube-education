// src/pages/ProgramApplications.tsx

import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import { getIdentity } from '../lib/auth'

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

export function ProgramApplications() {
  const { id } = useParams() // это programId из /program/:id/applications
  const ident = getIdentity()

  const canChangeStatus = ident.role === 'admin' || ident.role === 'moderator'

  const [items, setItems] = useState<Row[]>([])
  const [err, setErr] = useState('')
  const [status, setStatus] = useState('')

  async function reload() {
    if (!id) return
    try {
      setErr('')
      const data = await api.listApplications({ programId: id, status: status || undefined })
      setItems(data as Row[])
    } catch (e: any) {
      setErr(e.message || String(e))
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, status])

  async function move(appId: string, to: string) {
    const reason = prompt(`Причина для ${to}:`, 'ok') ?? ''
    try {
      await api.changeApplicationStatus(appId, to, reason)
      await reload()
    } catch (e: any) {
      const msg = e.message || String(e)

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
    <div>
      <h2>Заявки на курс</h2>

      <div>
        <Link to={`/program/${id}`}>Назад к программе</Link>
      </div>

      <div>
        <div>Фильтр по статусу</div>
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

      {err && <div>{err}</div>}

      {items.length === 0 ? (
        <div>Пока нет заявок.</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th align="left">app</th>
              <th align="left">user</th>
              <th align="left">group</th>
              <th align="left">status</th>
              <th align="left">comment</th>
              <th align="left">actions</th>
              <th align="left">interview</th>
              <th align="left">interview_comment</th>
            </tr>
          </thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.ID}>
                <td>{a.ID}</td>
                <td>{a.UserID}</td>
                <td>{a.GroupTitle} ({a.GroupID})</td>
                <td>{a.ProgramTitle} ({a.ProgramID})</td>
                <td>{a.Status}</td>
                <td>{a.Comment}</td>
                <td>{a.InterviewResult ?? '-'}</td>
                <td>{a.InterviewComment ?? '-'}</td>
                <td>
                  {canChangeStatus
                    ? (NEXT[a.Status] || []).map((btn) => (
                        <button key={btn.status} onClick={() => move(a.ID, btn.status)}>
                          {btn.label}
                        </button>
                      ))
                    : null}

                  {a.Status === 'in_review' ? (
                    <Link to={`/teacher/applications/${a.ID}/interview`}>Интервью</Link>
                  ) : null}

                  {!canChangeStatus ? (
                    <div>
                      <span>Интервью фиксирует преподаватель</span>
                      <div>
                        {a.Status === 'in_review' ? (
                          <Link to={`/teacher/applications/${a.ID}/interview`}>Зафиксировать интервью</Link>
                        ) : (
                          <span>Интервью доступно только в in_review</span>
                        )}
                      </div>
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
