// src/ui/GroupTeacherAdminPanel.tsx
import { useEffect, useMemo, useState } from 'react'
import { api } from '../lib/api'
import { DEV_IDENTITIES } from '../lib/auth'

type Props = {
  groupId: string
  canAssign: boolean
}

export function GroupTeacherAdminPanel({ groupId, canAssign }: Props) {
  const [teacherId, setTeacherId] = useState<string | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  const candidates = useMemo(() => {
    return DEV_IDENTITIES.filter(x => x.role === 'user')
  }, [])

  const [selected, setSelected] = useState(candidates[0]?.userId ?? '')
  const [manual, setManual] = useState('')

  async function reload() {
    setErr('')
    try {
      const res = await api.getGroupTeacher(groupId)
      if (res.teacher_id !== undefined) {
        setTeacherId(res.teacher_id) // string|null
      } else {
        const t = (res.teachers && res.teachers.length > 0) ? res.teachers[0] : null
        setTeacherId(t)
      }
    } catch (e: any) {
      setErr(e?.message || String(e))
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId])

  async function changeTeacher() {
    const tid = (manual.trim() || selected).trim()
    if (!tid) return

    // если уже такой же — не дергаем бек
    if (teacherId && teacherId === tid) return

    if (teacherId && !confirm('Сменить преподавателя для этой группы?')) return

    setBusy(true)
    setErr('')
    try {
      await api.setGroupTeacher(groupId, tid)
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
      <div style={{ fontWeight: 600, marginBottom: 6 }}>Преподаватель группы:</div>

      {teacherId ? (
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
          <span>{teacherId}</span>
        </div>
      ) : (
        <div style={{ opacity: 0.75 }}>
          не назначен (группа некорректна — назначь преподавателя)
        </div>
      )}

      {canAssign && (
        <>
          <div style={{ marginTop: 10, fontWeight: 600 }}>
            {teacherId ? 'Сменить преподавателя' : 'Назначить преподавателя'}
          </div>

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

            <button onClick={changeTeacher} disabled={busy}>
              {busy ? 'Сохраняю...' : (teacherId ? 'Сменить' : 'Назначить')}
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