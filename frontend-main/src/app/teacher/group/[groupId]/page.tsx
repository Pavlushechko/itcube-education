// src/app/teacher/group/[groupId]/page.tsx

'use client'

import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { api } from '@/shared/api/client'

type Assignment = {
  ID: string
  GroupID: string
  Title: string
  Description: string
  DueAt?: string | null
}

type Submission = {
  ID: string
  AssignmentID: string
  GroupID: string
  StudentUserID: string
  ContentType: string
  Content: string
  Status: string
  CreatedAt: string
}

type Material = {
  ID: string
  GroupID: string
  Type: string
  Title: string
  Content: string
  CreatedAt: string
}

export default function TeacherGroupPage() {
  const router = useRouter()
  const params = useParams<{ groupId: string }>()
  const groupId = params?.groupId

  const [asg, setAsg] = useState<Assignment[]>([])
  const [subs, setSubs] = useState<Submission[]>([])
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)

  const [mTitle, setMTitle] = useState('Материал 1')
  const [mContent, setMContent] = useState('')
  const [mFiles, setMFiles] = useState<FileList | null>(null)
  const [mBusy, setMBusy] = useState(false)
  const [mats, setMats] = useState<Material[]>([])

  // create assignment form
  const [title, setTitle] = useState('Задание 1')
  const [desc, setDesc] = useState('Сделать что-то')
  const [dueAt, setDueAt] = useState('') // RFC3339 optional
  const [busy, setBusy] = useState(false)

  async function uploadOne(file: File): Promise<string> {
    const { file_id } = await api.uploadFile(file)
    return file_id
  }
  
  async function createMaterialWithFiles() {
    if (!groupId) return
    setMBusy(true)
    setErr('')
    try {
      const ids: string[] = []
      if (mFiles && mFiles.length > 0) {
        for (const file of Array.from(mFiles)) {
          const id = await uploadOne(file)
          ids.push(id)
        }
      }

      await api.teacherCreateMaterial(groupId, {
        type: 'file',
        title: mTitle,
        content: mContent,
        attachments: ids,
      })

      setMTitle('')
      setMContent('')
      setMFiles(null)
      await reload()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setMBusy(false)
    }
  }

  async function reload() {
    if (!groupId) return
    setErr('')
    setLoading(true)
    try {
      const [assignments, submissions, materials] = await Promise.all([
        api.teacherListAssignments(groupId),
        api.teacherListSubmissions(groupId, undefined),
        api.teacherListMaterials(groupId),
      ])
      setAsg(assignments || [])
      setSubs(submissions || [])
      setMats(materials || [])
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId])

  async function createAssignment() {
    if (!groupId) return
    setBusy(true)
    setErr('')
    try {
      await api.teacherCreateAssignment(groupId, {
        title,
        description: desc,
        due_at: dueAt || undefined,
      })
      setTitle('')
      setDesc('')
      setDueAt('')
      await reload()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

  async function review(submissionId: string) {
    const comment = prompt('Комментарий:', 'ok') ?? ''
    const gradeRaw = prompt('Оценка (число или пусто):', '') ?? ''
    const gradeNum = gradeRaw.trim() === '' ? null : Number(gradeRaw)

    try {
      await api.teacherReviewSubmission(
        submissionId,
        Number.isFinite(gradeNum as any) ? gradeNum : null,
        comment
      )
      await reload()
    } catch (e: any) {
      alert(e?.message || String(e))
    }
  }

  if (!groupId) return <div>Загрузка...</div>
  if (loading) return <div>Загрузка...</div>

  return (
    <div>
      <div>
        <button onClick={() => router.back()}>Назад</button>
      </div>

      <h2>Группа (преподаватель)</h2>

      {err ? <div>{err}</div> : null}

      <h3>Создать материал</h3>

      <div>
        <div>Название</div>
        <input value={mTitle} onChange={(e) => setMTitle(e.target.value)} />
      </div>

      <div>
        <div>Контент / описание</div>
        <textarea value={mContent} onChange={(e) => setMContent(e.target.value)} rows={4} />
      </div>

      <div>
        <div>Файлы</div>
        <input type="file" multiple onChange={(e) => setMFiles(e.target.files)} />
      </div>

      <div>
        <button disabled={mBusy} onClick={createMaterialWithFiles}>
          {mBusy ? 'Загружаю...' : 'Создать материал'}
        </button>
      </div>

      <h3>Создать задание</h3>
      <div>
        <div>Название</div>
        <input value={title} onChange={(e) => setTitle(e.target.value)} />
      </div>
      <div>
        <div>Описание</div>
        <textarea value={desc} onChange={(e) => setDesc(e.target.value)} rows={4} />
      </div>
      <div>
        <div>Дедлайн (RFC3339, опционально)</div>
        <input value={dueAt} onChange={(e) => setDueAt(e.target.value)} />
      </div>
      <div>
        <button disabled={busy} onClick={createAssignment}>
          {busy ? 'Создаю...' : 'Создать'}
        </button>
      </div>

      <h3>Задания</h3>
      {asg.length === 0 ? (
        <div>Пока нет заданий</div>
      ) : (
        <ul>
          {asg.map((a) => (
            <li key={a.ID}>
              <b>{a.Title}</b> {a.DueAt ? `(до ${a.DueAt})` : ''}
              <div>{a.Description}</div>
            </li>
          ))}
        </ul>
      )}

      <h3>Материалы</h3>
      {mats.length === 0 ? (
        <div>Пока нет материалов</div>
      ) : (
        <ul>
          {mats.map((m) => (
            <li key={m.ID}>
              <b>{m.Title}</b> {m.Type ? `(${m.Type})` : ''}
              <div style={{ whiteSpace: 'pre-wrap' }}>{m.Content}</div>

              {/* ссылка на отдельную страницу материала */}
              <div>
                <a href={`/teacher/group/${groupId}/materials/${m.ID}`}>Открыть</a>
              </div>
            </li>
          ))}
        </ul>
      )}

      <h3>Сданные работы</h3>
      {subs.length === 0 ? (
        <div>Пока нет сдач</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th align="left">submission</th>
              <th align="left">student</th>
              <th align="left">status</th>
              <th align="left">content</th>
              <th align="left">actions</th>
            </tr>
          </thead>
          <tbody>
            {subs.map((s) => (
              <tr key={s.ID}>
                <td>{s.ID}</td>
                <td>{s.StudentUserID}</td>
                <td>{s.Status}</td>
                <td>{s.ContentType}: {s.Content}</td>
                <td>
                  <button onClick={() => review(s.ID)}>Проверить</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}