// src/pages/TeacherInterview.tsx

import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'

const RESULTS = [
  'recommended',
  'not_recommended',
  'needs_more',
  'pending',
] as const

type Result = typeof RESULTS[number]

export function TeacherInterview() {
  const { appId } = useParams()
  const [result, setResult] = useState<Result>('recommended')
  const [comment, setComment] = useState<string>('ok')
  const [err, setErr] = useState<string>('')
  const [saving, setSaving] = useState<boolean>(false)
  const [allowed, setAllowed] = useState<boolean | null>(null)

  useEffect(() => {
    api.myGroups()
      .then((gs: any) => {
        if (Array.isArray(gs) && gs.length > 0) setAllowed(true)
        else setAllowed(false)
      })
      .catch((e: any) => {
        if (e?.status === 403) setAllowed(false)
        else { setErr(e?.message || String(e)); setAllowed(false) }
      })
  }, [])

  if (allowed === false) {
    return (
      <div>
        <h2>Зафиксировать интервью</h2>
        <div><Link to={-1 as any}>Назад</Link></div>
        <div>Доступ запрещён</div>
        {err ? <div>{err}</div> : null}
      </div>
    )
  }

  if (allowed === null) {
    return <div>Загрузка...</div>
  }

  async function submit() {
    if (!appId) return
    try {
      setErr('')
      setSaving(true)
      await api.recordInterview(appId, result, comment)
      alert('Интервью сохранено')
    } catch (e: any) {
      setErr(e?.message || String(e))
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

      <div>
        <div>Результат</div>
        <select value={result} onChange={(e) => setResult(e.target.value as Result)}>
          {RESULTS.map((r) => (
            <option key={r} value={r}>{r}</option>
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

      {err ? <div>{err}</div> : null}
    </div>
  )
}
