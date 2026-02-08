// src/pages/LearnGroup.tsx

import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import { useParams } from 'react-router-dom'
import type { Material } from '../lib/types'

export function LearnGroup() {
  const { groupId } = useParams()
  const [items, setItems] = useState<Material[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!groupId) return
    api.listMaterials(groupId)
      .then(setItems)
      .catch(e => setErr(String(e.message || e)))
  }, [groupId])

  return (
    <div style={{ padding: 12 }}>
      <h2>Обучение — группа {groupId}</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}

      <h3>Материалы</h3>
      <ul>
        {(items ?? []).map(m => (
          <li key={m.ID}>
            <b>{m.Title}</b> — {m.Type} — {m.Content}
          </li>
        ))}
      </ul>
    </div>
  )
}
