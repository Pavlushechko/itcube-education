// src/ui/AdminCreateGroup.tsx
import { useMemo, useState } from 'react'
import { api } from '../lib/api'
import { DEV_IDENTITIES } from '../lib/auth'

export function AdminCreateGroup({ programId, onDone }: { programId: string; onDone: () => void }) {
  const [year, setYear] = useState<number>(2026)
  const [title, setTitle] = useState('Group A')
  const [capacity, setCapacity] = useState<number>(30)
  const [requiresInterview, setRequiresInterview] = useState(true)
  const [isOpen, setIsOpen] = useState(true)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  // кандидаты из DEV (обычные user)
  const candidates = useMemo(() => DEV_IDENTITIES.filter((x) => x.role === 'user'), [])
  const [selectedTeacher, setSelectedTeacher] = useState(candidates[0]?.userId ?? '')
  const [manualTeacher, setManualTeacher] = useState('')

  // teacher_user_id, который реально уйдёт в API
  const teacherUserId = (manualTeacher.trim() || selectedTeacher).trim()

  async function create() {
    setErr('')
    if (!teacherUserId) {
      setErr('teacher_user_id обязателен')
      return
    }

    setBusy(true)
    try {
      const c = await api.createCohort(programId, year)

      await api.createGroup({
        programId,
        cohortId: c.id,
        teacherUserId, // обязательно
        title,
        capacity,
        requiresInterview,
        isOpen,
      })

      alert('Группа создана')
      onDone()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div style={{ marginTop: 12, padding: 10, border: '1px solid #ddd', borderRadius: 6, maxWidth: 720 }}>
      <b>Создать группу (admin)</b>

      <div style={{ display: 'grid', gridTemplateColumns: '140px 1fr', gap: 8, marginTop: 10 }}>
        <div>Год (cohort)</div>
        <input value={year} onChange={(e) => setYear(Number(e.target.value))} disabled={busy} />

        <div>Преподаватель</div>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
          <select value={selectedTeacher} onChange={(e) => setSelectedTeacher(e.target.value)} disabled={busy}>
            {candidates.map((c) => (
              <option key={c.userId} value={c.userId}>
                {c.label}: {c.userId}
              </option>
            ))}
          </select>

          <span style={{ opacity: 0.75 }}>или UUID:</span>

          <input
            value={manualTeacher}
            onChange={(e) => setManualTeacher(e.target.value)}
            placeholder="teacher_user_id"
            style={{ padding: 6, minWidth: 320 }}
            disabled={busy}
          />
        </div>

        <div>Название группы</div>
        <input value={title} onChange={(e) => setTitle(e.target.value)} disabled={busy} />

        <div>Вместимость</div>
        <input value={capacity} onChange={(e) => setCapacity(Number(e.target.value))} disabled={busy} />

        <div>Интервью</div>
        <label>
          <input
            type="checkbox"
            checked={requiresInterview}
            onChange={(e) => setRequiresInterview(e.target.checked)}
            disabled={busy}
          />{' '}
          требуется
        </label>

        <div>Открыта</div>
        <label>
          <input type="checkbox" checked={isOpen} onChange={(e) => setIsOpen(e.target.checked)} disabled={busy} /> да
        </label>
      </div>

      {err && <div style={{ color: 'crimson', marginTop: 10 }}>{err}</div>}

      <div style={{ marginTop: 10 }}>
        <button onClick={create} disabled={busy || !teacherUserId}>
          {busy ? 'Создаю...' : 'Создать'}
        </button>
      </div>
    </div>
  )
}