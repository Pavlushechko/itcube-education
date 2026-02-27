// src/app/learn/group/[groupId]/assignment/[assignmentId]/page.tsx

'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import { api } from '@/shared/api/client'

type Assignment = {
  ID: string
  GroupID: string
  Title: string
  Description: string
  DueAt?: string | null
}

export default function AssignmentPage() {
  const router = useRouter()
  const params = useParams<{ groupId: string; assignmentId: string }>()
  const groupId = params?.groupId
  const assignmentId = params?.assignmentId

  const [asg, setAsg] = useState<Assignment | null>(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)

  const [contentType, setContentType] = useState('text')
  const [content, setContent] = useState('')
  const [saving, setSaving] = useState(false)

  const [mySubmission, setMySubmission] = useState<any>(null)
  const [myReview, setMyReview] = useState<any>(null)

  async function reload() {
    if (!groupId || !assignmentId) return
    setErr('')
    setLoading(true)
    try {
      const list = await api.listAssignmentsLearner(groupId)
      const found = (list || []).find((x: any) => x.ID === assignmentId)
      if (!found) throw new Error('Задание не найдено')
      setAsg(found)

      const me = await api.mySubmission(assignmentId)
      setMySubmission(me?.submission ?? null)
      setMyReview(me?.review ?? null)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId, assignmentId])

  async function submit() {
    if (!assignmentId) return
    setSaving(true)
    setErr('')
    try {
      await api.submitAssignment(assignmentId, contentType, content)
      setContent('')
      await reload()
      alert('Отправлено')
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div>Загрузка...</div>

  return (
    <div>
      <div style={{ display: 'flex', gap: 12 }}>
        <button onClick={() => router.back()}>Назад</button>
        {groupId ? <Link href={`/learn/group/${groupId}`}>К группе</Link> : null}
      </div>

      {err ? <div>{err}</div> : null}

      {!asg ? <div>Нет данных</div> : (
        <>
          <h2>{asg.Title}</h2>
          {asg.DueAt ? <div>Дедлайн: {asg.DueAt}</div> : null}

          <h3>Описание</h3>
          <div style={{ whiteSpace: 'pre-wrap' }}>{asg.Description}</div>

          <h3>Моя сдача</h3>
          {mySubmission ? (
            <div>
              <div>Статус: {mySubmission.Status}</div>
              <div>{mySubmission.ContentType}: {mySubmission.Content}</div>
            </div>
          ) : (
            <div>Пока не сдавал</div>
          )}

          <h3>Отправить решение</h3>
          <div>
            <div>Тип</div>
            <select value={contentType} onChange={(e) => setContentType(e.target.value)}>
              <option value="text">text</option>
              <option value="link">link</option>
              <option value="file_link">file_link</option>
            </select>
          </div>
          <div>
            <div>Содержимое</div>
            <textarea rows={5} value={content} onChange={(e) => setContent(e.target.value)} />
          </div>
          <div>
            <button disabled={saving} onClick={submit}>
              {saving ? 'Отправляю...' : 'Сдать'}
            </button>
          </div>

          <h3>Проверка преподавателя</h3>
          {myReview ? (
            <div>
              <div>Оценка: {myReview.Grade ?? '-'}</div>
              <div>Комментарий: {myReview.Comment ?? ''}</div>
            </div>
          ) : (
            <div>Пока нет проверки</div>
          )}
        </>
      )}
    </div>
  )
}