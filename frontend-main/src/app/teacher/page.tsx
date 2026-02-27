// src/app/teacher/page.tsx

'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/shared/api/client'

type Group = {
  ID: string
  Title: string
  ProgramID: string
  CohortID: string
}

export default function TeacherHomePage() {
  const [items, setItems] = useState<Group[]>([])
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setErr('')
    setLoading(true)
    api.myGroups()
      .then((gs: any) => setItems(Array.isArray(gs) ? gs : []))
      .catch((e: any) => setErr(e?.message || String(e)))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div>Загрузка...</div>
  if (err) return <div>{err}</div>

  return (
    <div>
      <h2>Преподаватель</h2>

      {items.length === 0 ? (
        <div>Нет назначенных групп</div>
      ) : (
        <ul>
          {items.map((g) => (
            <li key={g.ID}>
              <Link href={`/teacher/group/${g.ID}`}>{g.Title}</Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}