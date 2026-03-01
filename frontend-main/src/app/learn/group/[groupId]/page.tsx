// src/app/learn/group/[groupId]/page.tsx

'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { api } from '@/shared/api/client'
import Link from 'next/link'

type GroupInfo = {
  ProgramID: string
  ProgramTitle: string
  GroupID: string
  GroupTitle: string
}

type Material = {
  ID: string
  GroupID: string
  Type: string
  Title: string
  Content: string
  CreatedBy: string
  CreatedAt: string
}

type Assignment = {
  ID: string
  GroupID: string
  Title: string
  Description: string
  DueAt?: string | null
  CreatedAt?: string
  MyStatus?: 'not_done' | 'in_review' | 'reviewed'
}

function humanStatus(s?: string) {
  if (s === 'reviewed') return 'Проверено'
  if (s === 'in_review') return 'На проверке'
  return 'Не сделано'
}

export default function LearnGroupPage() {
  const router = useRouter()
  const params = useParams<{ groupId: string }>()
  const groupId = params?.groupId

  const [info, setInfo] = useState<GroupInfo | null>(null)
  const [materials, setMaterials] = useState<Material[]>([])
  const [assignments, setAssignments] = useState<Assignment[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!groupId) return

    setLoading(true)
    setErr('')

    Promise.all([
      api.getLearnGroupInfo(groupId),
      api.listMaterials(groupId),
      api.listAssignmentsLearner(groupId),
    ])
      .then(([gi, ms, as]) => {
        setInfo(gi)
        setMaterials(ms || [])
        setAssignments(as || [])
      })
      .catch((e: any) => setErr(e?.message || String(e)))
      .finally(() => setLoading(false))
  }, [groupId])

  if (!groupId) return <div>Загрузка...</div>
  if (loading) return <div>Загрузка...</div>
  if (err) return <div>{err}</div>

  return (
    <div>
      <div>
        <button onClick={() => router.back()}>Назад</button>{' '}
        <Link href="/">В каталог</Link>
      </div>

      <h2>Обучение</h2>

      <div>
        <div><b>Курс:</b> {info?.ProgramTitle ?? '-'}</div>
        <div><b>Группа:</b> {info?.GroupTitle ?? '-'}</div>
      </div>

      <h3>Материалы</h3>
      {materials.length === 0 ? (
        <div>Пока нет материалов</div>
      ) : (
        <ul>
          {materials.map((m) => (
            <li key={m.ID}>
              <Link href={`/learn/group/${groupId}/materials/${m.ID}`}>
                {m.Title}
              </Link>{' '}
              ({m.Type})
            </li>
          ))}
        </ul>
      )}

      <h3>Задания</h3>
      {assignments.length === 0 ? (
        <div>Пока нет заданий</div>
      ) : (
        <ul>
          {assignments.map((a) => (
            <li key={a.ID}>
              <Link href={`/learn/group/${groupId}/assignment/${a.ID}`}>{a.Title}</Link>
              {' '}— <span>{humanStatus(a.MyStatus)}</span>
              {a.DueAt ? <span> (до {a.DueAt})</span> : null}
            </li>
          ))}
        </ul>
      )}

      <h3>Расписание</h3>
      <div>Пока не реализовано</div>
    </div>
  )
}