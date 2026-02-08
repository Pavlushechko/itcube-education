// src/pages/TeacherGroupManage.tsx

import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../lib/api'

export function TeacherGroupManage() {
  const { groupId } = useParams()
  const [students, setStudents] = useState<string[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!groupId) return
    setErr('')
    api.groupStudents(groupId)
      .then((res: any) => setStudents(res.students || []))
      .catch((e) => setErr(String(e.message || e)))
  }, [groupId])

  return (
    <div style={{ padding: 12 }}>
      <h2>Управление группой {groupId}</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}

      <h3>Студенты</h3>
      <ul>
        {students.map(s => <li key={s}>{s}</li>)}
      </ul>

      <div style={{ opacity: 0.75, marginTop: 10 }}>
        Следом сюда добавим формы: “добавить материал” и “добавить задание”.
      </div>
    </div>
  )
}
