// src/app/teacher/applications/[appId]/interview/page.tsx

'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import { api, ApiError } from '@/shared/api/client'

const RESULTS = ['recommended', 'not_recommended', 'needs_more', 'pending'] as const
type Result = typeof RESULTS[number]

export default function TeacherInterviewPage() {
  const params = useParams<{ appId: string }>()
  const router = useRouter()
  const appId = params?.appId || ''

  const [result, setResult] = useState<Result>('recommended')
  const [comment, setComment] = useState('ok')
  const [err, setErr] = useState('')
  const [saving, setSaving] = useState(false)

  async function submit() {
    if (!appId) return
    try {
      setErr('')
      setSaving(true)
      await api.recordInterview(appId, result, comment)
      alert('Интервью сохранено')
      router.back()
    } catch (e: any) {
      if (e instanceof ApiError && e.status === 403) {
        setErr('Доступ запрещён (вы не назначены преподавателем на эту группу).')
      } else {
        setErr(e?.message || String(e))
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div style={{ padding: 12 }}>
      <h2>Зафиксировать интервью</h2>

      <div style={{ marginBottom: 8 }}>
        <button onClick={() => router.back()}>Назад</button>{' '}
        <span style={{ opacity: 0.75, marginLeft: 8 }}>
          appId: {appId || '(нет)'}
        </span>
      </div>

      {!appId ? <div style={{ color: 'crimson' }}>Нет appId</div> : null}

      <div style={{ marginTop: 10 }}>
        <div>Результат</div>
        <select value={result} onChange={(e) => setResult(e.target.value as Result)}>
          {RESULTS.map((r) => (
            <option key={r} value={r}>
              {r}
            </option>
          ))}
        </select>
      </div>

      <div style={{ marginTop: 10 }}>
        <div>Комментарий</div>
        <textarea rows={4} value={comment} onChange={(e) => setComment(e.target.value)} />
      </div>

      <div style={{ marginTop: 10 }}>
        <button disabled={saving || !appId} onClick={submit}>
          {saving ? 'Сохраняю...' : 'Сохранить'}
        </button>
      </div>

      {err ? <div style={{ color: 'crimson', marginTop: 10 }}>{err}</div> : null}

      <div style={{ marginTop: 14, opacity: 0.75 }}>
        <Link href="/catalog">В каталог</Link>
      </div>
    </div>
  )
}
