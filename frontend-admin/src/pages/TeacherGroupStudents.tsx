// src/pages/TeacherGroupStudents.tsx

import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import { useParams } from 'react-router-dom'

export function TeacherGroupStudents() {
  const { groupId } = useParams()
  const [students, setStudents] = useState<string[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!groupId) return
    api.groupStudents(groupId)
      .then((res: any) => setStudents(res.students || []))
      .catch(e => setErr(String(e.message || e)))
  }, [groupId])

  return (
    <div style={{ padding: 12 }}>
      <h2>Студенты группы {groupId}</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}
      <ul>
        {students.map(id => <li key={id}>{id}</li>)}
      </ul>
    </div>
  )
}
