// src/app/me/applications/page.tsx

'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/shared/api/client'

type Application = {
  ID: string
  UserID: string
  GroupID: string
  Status: string
  Comment: string
  CreatedAt: string
  UpdatedAt: string
}

export default function MyApplicationsPage() {
  const [items, setItems] = useState<Application[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    setErr('')
    api.listMyApplications()
      .then((x) => setItems((x ?? []) as Application[]))
      .catch((e: any) => setErr(String(e?.message || e)))
  }, [])

  return (
    <div style={{ padding: 12 }}>
      <div style={{ marginBottom: 8 }}>
        <Link href="/catalog">← Каталог</Link>
      </div>

      <h2>Мои заявки</h2>

      {err ? <div style={{ color: 'crimson' }}>{err}</div> : null}

      {items.length === 0 ? (
        <div style={{ opacity: 0.75 }}>Пока нет заявок</div>
      ) : (
        <ul>
          {items.map((a) => (
            <li key={a.ID}>
              group={a.GroupID} • status={a.Status} • {a.Comment}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
