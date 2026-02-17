// src/pages/GroupApplicationsStaff.tsx

import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'

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

export function GroupApplicationsStaff() {
  const { groupId } = useParams()
  const [items, setItems] = useState<Row[]>([])
  const [err, setErr] = useState('')
  const [filter, setFilter] = useState<string>('')

  async function reload() {
    if (!groupId) return
    try {
      setErr('')
      const data = await api.listApplications({ groupId, status: filter || undefined })
      setItems(data as Row[])
    } catch (e: any) {
      setErr(e?.message || String(e))
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId, filter])

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
      <h2>Заявки группы {groupId}</h2>

      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <span>Фильтр:</span>
        <select value={filter} onChange={(e) => setFilter(e.target.value)}>
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

      <table style={{ width: '100%', marginTop: 12, borderCollapse: 'collapse' }}>
        <thead>
          <tr>
            <th align="left">app</th>
            <th align="left">user</th>
            <th align="left">program</th>
            <th align="left">group</th>
            <th align="left">status</th>
            <th align="left">comment</th>
            <th align="left">actions</th>
          </tr>
        </thead>
        <tbody>
          {items.map((a) => (
            <tr key={a.ID} style={{ borderTop: '1px solid #eee' }}>
              <td>{a.ID}</td>
              <td>{a.UserID}</td>
              <td>
                {a.ProgramTitle} ({a.ProgramID})
              </td>
              <td>
                {a.GroupTitle} ({a.GroupID})
              </td>
              <td>
                <b>{a.Status}</b>
              </td>
              <td>{a.Comment}</td>
              <td>
                {(NEXT[a.Status] || []).map((btn) => (
                  <button key={btn.status} onClick={() => move(a.ID, btn.status)} style={{ marginRight: 8 }}>
                    {btn.label}
                  </button>
                ))}
                {/* В админке ссылку на интервью можно оставить как read-only переход в main (или убрать) */}
                {a.Status === 'in_review' ? (
                  <span style={{ marginLeft: 8, opacity: 0.75 }}>Интервью фиксирует преподаватель</span>
                ) : null}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
