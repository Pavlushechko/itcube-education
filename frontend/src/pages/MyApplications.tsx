// src/pages/MyApplications.tsx

import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { Application } from '../lib/types'

export function MyApplications() {
  const [items, setItems] = useState<Application[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    api.listMyApplications()
      .then(setItems)
      .catch(e => setErr(String(e.message || e)))
  }, [])

  return (
    <div style={{ padding: 12 }}>
      <h2>Мои заявки</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}
      <ul>
        {items.map(a => (
          <li key={a.ID}>
            group={a.GroupID} • status={a.Status} • {a.Comment}
          </li>
        ))}
      </ul>
    </div>
  )
}
