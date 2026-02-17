// src/pages/ApplicationsAll.tsx

import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { getIdentity } from '../lib/auth'

type Row = {
  ID: string
  UserID: string
  GroupID: string

  ProgramID: string
  ProgramTitle: string
  GroupTitle: string

  // NEW: приходит с бэка как CohortYear (может быть null)
  CohortYear?: number | null

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

export function ApplicationsAll() {
  const ident = getIdentity()
  const canChangeStatus = ident.role === 'admin' || ident.role === 'moderator'

  const [rawItems, setRawItems] = useState<Row[]>([])
  const [err, setErr] = useState('')

  // фильтры (серверные)
  const [status, setStatus] = useState('')
  const [programId, setProgramId] = useState('')
  const [groupId, setGroupId] = useState('')
  const [year, setYear] = useState<string>('') // храним строкой из select

  // фильтр (локальный)
  const [userQ, setUserQ] = useState('')

  async function reload() {
    try {
      setErr('')
      const data = await api.listApplications({
        status: status || undefined,
        programId: programId || undefined,
        groupId: groupId || undefined,
        year: year ? Number(year) : undefined,
      })
      setRawItems(data as Row[])
    } catch (e: any) {
      setErr(e.message || String(e))
    }
  }

  // дергаем при изменении фильтров (можешь оставить только на кнопку — но так удобнее)
  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, programId, groupId, year])

  const programs = useMemo(() => {
    const m = new Map<string, string>()
    for (const a of rawItems) m.set(a.ProgramID, a.ProgramTitle)
    return Array.from(m.entries())
      .map(([id, title]) => ({ id, title }))
      .sort((x, y) => x.title.localeCompare(y.title))
  }, [rawItems])

  const groups = useMemo(() => {
    const m = new Map<string, { title: string; programId: string }>()
    for (const a of rawItems) {
      m.set(a.GroupID, { title: a.GroupTitle, programId: a.ProgramID })
    }

    // Если выбран programId — показываем только группы этой программы
    const arr = Array.from(m.entries()).map(([id, v]) => ({ id, title: v.title, programId: v.programId }))
    const filtered = programId ? arr.filter(g => g.programId === programId) : arr
    return filtered.sort((x, y) => x.title.localeCompare(y.title))
  }, [rawItems, programId])

  const years = useMemo(() => {
    const s = new Set<number>()
    for (const a of rawItems) {
      if (typeof a.CohortYear === 'number') s.add(a.CohortYear)
    }
    return Array.from(s.values()).sort((a, b) => b - a) // новые сверху
  }, [rawItems])

  const items = useMemo(() => {
    const q = userQ.trim().toLowerCase()
    if (!q) return rawItems
    return rawItems.filter(a => a.UserID.toLowerCase().includes(q))
  }, [rawItems, userQ])

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

  function reset() {
    setStatus('')
    setProgramId('')
    setGroupId('')
    setYear('')
    setUserQ('')
    // reload() вызовется сам из useEffect
  }

  return (
    <div style={{ padding: 12 }}>
      <h2>Все заявки</h2>

      <div style={{ marginBottom: 12 }}>
        <Link to="/">Каталог</Link>
      </div>

      <div style={{ display: 'grid', gap: 8, maxWidth: 520, marginBottom: 12 }}>
        <label style={{ display: 'grid', gap: 4 }}>
          <div>Статус</div>
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="">все</option>
            <option value="submitted">submitted</option>
            <option value="in_review">in_review</option>
            <option value="approved">approved</option>
            <option value="rejected">rejected</option>
            <option value="cancelled">cancelled</option>
          </select>
        </label>

        <label style={{ display: 'grid', gap: 4 }}>
          <div>ProgramID</div>
          <select
            value={programId}
            onChange={(e) => {
              const next = e.target.value
              setProgramId(next)
              // если программа сменилась — сбросим группу, чтобы не было "несовместимых" фильтров
              setGroupId('')
            }}
          >
            <option value="">все</option>
            {programs.map(p => (
              <option key={p.id} value={p.id}>
                {p.title} ({p.id})
              </option>
            ))}
          </select>
        </label>

        <label style={{ display: 'grid', gap: 4 }}>
          <div>GroupID</div>
          <select value={groupId} onChange={(e) => setGroupId(e.target.value)}>
            <option value="">все</option>
            {groups.map(g => (
              <option key={g.id} value={g.id}>
                {g.title} ({g.id})
              </option>
            ))}
          </select>
        </label>

        <label style={{ display: 'grid', gap: 4 }}>
          <div>Набор (год)</div>
          <select value={year} onChange={(e) => setYear(e.target.value)}>
            <option value="">все</option>
            {years.map(y => (
              <option key={y} value={String(y)}>
                {y}
              </option>
            ))}
          </select>
        </label>

        <label style={{ display: 'grid', gap: 4 }}>
          <div>User (поиск, локально)</div>
          <input
            value={userQ}
            onChange={(e) => setUserQ(e.target.value)}
            placeholder="часть UserID..."
          />
        </label>

        <div style={{ display: 'flex', gap: 8 }}>
          <button onClick={reload}>Обновить</button>
          <button onClick={reset}>Сброс</button>
        </div>
      </div>

      {err && <div style={{ color: 'crimson' }}>{err}</div>}

      {items.length === 0 ? (
        <div>Нет заявок.</div>
      ) : (
        <table cellPadding={6} style={{ borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th align="left">app</th>
              <th align="left">user</th>
              <th align="left">program</th>
              <th align="left">group</th>
              <th align="left">year</th>
              <th align="left">status</th>
              <th align="left">comment</th>
              <th align="left">interview</th>
              <th align="left">interview_comment</th>
              <th align="left">actions</th>
            </tr>
          </thead>
          <tbody>
            {items.map((a) => (
              <tr key={a.ID} style={{ borderTop: '1px solid #ddd' }}>
                <td>{a.ID}</td>
                <td>{a.UserID}</td>
                <td>
                  {a.ProgramTitle} ({a.ProgramID})
                </td>
                <td>
                  {a.GroupTitle} ({a.GroupID})
                </td>
                <td>{a.CohortYear ?? '-'}</td>
                <td>{a.Status}</td>
                <td>{a.Comment}</td>
                <td>{a.InterviewResult ?? '-'}</td>
                <td>{a.InterviewComment ?? '-'}</td>
                <td style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  {canChangeStatus
                    ? (NEXT[a.Status] || []).map((btn) => (
                        <button key={btn.status} onClick={() => move(a.ID, btn.status)}>
                          {btn.label}
                        </button>
                      ))
                    : null}

                  {a.Status === 'in_review' ? (
                    <Link to={`/applications/${a.ID}/interview`}>Интервью</Link>
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
