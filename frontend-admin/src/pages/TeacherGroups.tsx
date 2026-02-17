// src/pages/TeacherGroups.tsx

import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { Group } from '../lib/types'
import { Link } from 'react-router-dom'

export function TeacherGroups() {
  const [items, setItems] = useState<Group[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    api.myGroups()
      .then(setItems)
      .catch(e => setErr(String(e.message || e)))
  }, [])

  return (
    <div style={{ padding: 12 }}>
      <h2>Мои группы (назначение преподавателем)</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}
      <ul>
        {items.map(g => (
          <li key={g.ID}>
            <b>{g.Title}</b> — <Link to={`/teacher/groups/${g.ID}/students`}>Студенты</Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
