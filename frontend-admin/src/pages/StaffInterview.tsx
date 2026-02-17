// src/pages/StaffInterview.tsx

import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'

const RESULTS = ['recommended', 'not_recommended', 'needs_more', 'pending'] as const
type Result = (typeof RESULTS)[number]

export function StaffInterview() {
  const { appId } = useParams()

  const [result, setResult] = useState<Result>('recommended')
  const [comment, setComment] = useState<string>('ok')
  const [err, setErr] = useState<string>('')
  const [saving, setSaving] = useState<boolean>(false)
  const [forbidden, setForbidden] = useState<boolean>(false)

  async function submit() {
    if (!appId) return
    try {
      setErr('')
      setForbidden(false)
      setSaving(true)

      await api.recordInterview(appId, result, comment)
      alert('Интервью сохранено')
    } catch (e: any) {
      const msg = e?.message || String(e)
      // если у тебя ApiError кидает status — используй его
      if (e?.status === 403 || msg.includes('forbidden')) {
        setForbidden(true)
        return
      }
      setErr(msg)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h2>Зафиксировать интервью</h2>

      <div>
        <Link to={-1 as any}>Назад</Link>
      </div>

      {!appId ? <div>Нет appId</div> : null}

      {forbidden ? (
        <div style={{ color: 'crimson' }}>Доступ запрещён</div>
      ) : null}

      <div>
        <div>Результат</div>
        <select value={result} onChange={(e) => setResult(e.target.value as Result)}>
          {RESULTS.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>
      </div>

      <div>
        <div>Комментарий</div>
        <textarea value={comment} onChange={(e) => setComment(e.target.value)} rows={4} />
      </div>

      <div>
        <button disabled={saving || !appId} onClick={submit}>
          {saving ? 'Сохраняю...' : 'Сохранить'}
        </button>
      </div>

      {err ? <div style={{ color: 'crimson' }}>{err}</div> : null}
    </div>
  )
}
