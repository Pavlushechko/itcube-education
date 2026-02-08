// src/ui/GroupTeacherAdminPanel.tsx

import { useEffect, useMemo, useState } from 'react'
import { api } from '../lib/api'
import { DEV_IDENTITIES } from '../lib/auth'

type Props = {
  groupId: string
  // показывать ли блок назначения (только admin)
  canAssign: boolean
}

export function GroupTeacherAdminPanel({ groupId, canAssign }: Props) {
  const [teachers, setTeachers] = useState<string[]>([])
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  // кандидаты из DEV (можно выбрать id)
  const candidates = useMemo(() => {
    // берём всех, кто не admin/moderator (по факту “обычные юзеры”)
    return DEV_IDENTITIES.filter(x => x.role === 'user')
  }, [])

  const [selected, setSelected] = useState(candidates[0]?.userId ?? '')
  const [manual, setManual] = useState('')

  async function reload() {
    setErr('')
    try {
      const res = await api.listGroupTeachers(groupId)
      setTeachers(res.teachers || [])
    } catch (e: any) {
      setErr(e?.message || String(e))
    }
  }

  useEffect(() => {
    reload()
  }, [groupId])

  async function assign() {
    const teacherUserId = (manual.trim() || selected).trim()
    if (!teacherUserId) return

    setBusy(true)
    setErr('')
    try {
      await api.assignTeacherToGroup(groupId, teacherUserId)
      await reload()
      setManual('')
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

return (
  <div style={{ marginTop: 10, padding: 10, border: '1px solid #ddd', borderRadius: 6 }}>
    <div style={{ fontWeight: 600, marginBottom: 6 }}>Назначенные преподаватели:</div>

    {teachers.length === 0 ? (
      <div style={{ opacity: 0.75 }}>пока нет</div>
    ) : (
      <ul style={{ marginTop: 6 }}>
        {teachers.map(t => (
          <li key={t} style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <span>{t}</span>

            {canAssign && (
              <button
                onClick={async () => {
                  if (!confirm('Убрать преподавателя из группы?')) return
                  setBusy(true)
                  try {
                    await api.removeTeacherFromGroup(groupId, t)
                    await reload()
                  } finally {
                    setBusy(false)
                  }
                }}
                disabled={busy}
              >
                Убрать
              </button>
            )}
          </li>
        ))}
      </ul>
    )}

    {canAssign && (
      <>
        <div style={{ marginTop: 10, fontWeight: 600 }}>Назначить преподавателя</div>

        <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginTop: 6, flexWrap: 'wrap' }}>
          <select value={selected} onChange={(e) => setSelected(e.target.value)} disabled={busy}>
            {candidates.map(c => (
              <option key={c.userId} value={c.userId}>
                {c.label}: {c.userId}
              </option>
            ))}
          </select>

          <span style={{ opacity: 0.75 }}>или UUID:</span>

          <input
            value={manual}
            onChange={(e) => setManual(e.target.value)}
            placeholder="teacher_user_id"
            style={{ padding: 6, minWidth: 320 }}
            disabled={busy}
          />

          <button onClick={assign} disabled={busy}>
            {busy ? 'Назначаю...' : 'Назначить'}
          </button>

          <button onClick={reload} disabled={busy}>
            Обновить
          </button>
        </div>

        {err && <div style={{ color: 'crimson', marginTop: 8 }}>{err}</div>}
      </>
    )}
  </div>
)

}
