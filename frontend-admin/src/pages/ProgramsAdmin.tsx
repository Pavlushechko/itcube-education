// src/pages/ProgramsAdmin.tsx
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import { getIdentity } from '../lib/auth'
import { AdminCreateProgram } from '../ui/AdminCreateProgram'

type Program = {
  ID: string
  Title: string
  Description: string
  Status?: string
}

export function ProgramsAdmin() {
  const ident = getIdentity()
  const isAdmin = ident.role === 'admin'

  const [items, setItems] = useState<Program[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    setErr('')
    api.listProgramsAdmin()
      .then(setItems)
      .catch((e: any) => setErr(e?.message || String(e)))
  }, [])

  const drafts = useMemo(() => items.filter(p => p.Status === 'draft'), [items])
  const published = useMemo(() => items.filter(p => p.Status === 'published'), [items])

  return (
    <div style={{ padding: 12 }}>
      <h2>Программы (staff)</h2>

      {isAdmin ? (
        <AdminCreateProgram onCreated={() => window.location.reload()} />
      ) : null}

      {err ? <div style={{ color: 'crimson' }}>{err}</div> : null}

      <h3>Черновики</h3>
      {drafts.length === 0 ? <div>Нет</div> : (
        <ul>
          {drafts.map(p => (
            <li key={p.ID}>
              {p.Title} <span style={{ opacity: 0.75 }}>({p.ID})</span>
            </li>
          ))}
        </ul>
      )}

      <h3>Опубликованные</h3>
      {published.length === 0 ? <div>Нет</div> : (
        <ul>
          {published.map(p => (
            <li key={p.ID}>
              {p.Title} <span style={{ opacity: 0.75 }}>({p.ID})</span>
            </li>
          ))}
        </ul>
      )}

      <div style={{ marginTop: 12, opacity: 0.75 }}>
        Страницы “каталог/программа/ученик/препод” перенесены в frontend-main (Next.js).
      </div>
    </div>
  )
}
