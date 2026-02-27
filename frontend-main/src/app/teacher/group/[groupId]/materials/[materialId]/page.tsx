// src/app/teacher/group/[groupId]/materials/[materialId]/page.tsx
'use client'

import { useEffect, useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { api } from '@/shared/api/client'

type FileItem = {
  id: string
  original_name: string
  mime_type: string
  size_bytes: number
}

type MaterialPage = {
  id: string
  group_id: string
  type: string
  title: string
  content: string
  external_url?: string | null
  created_at: string
  files: FileItem[]
}

function fmtSize(bytes: number) {
  if (!Number.isFinite(bytes)) return ''
  const kb = bytes / 1024
  if (kb < 1024) return `${Math.round(kb)} KB`
  const mb = kb / 1024
  return `${mb.toFixed(1)} MB`
}

export default function TeacherMaterialPage() {
  const router = useRouter()
  const params = useParams<{ groupId: string; materialId: string }>()
  const groupId = params?.groupId
  const materialId = params?.materialId

  const [mat, setMat] = useState<MaterialPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState('')

  const [edit, setEdit] = useState(false)
  const [saving, setSaving] = useState(false)

  // edit fields
  const [eType, setEType] = useState('text')
  const [eTitle, setETitle] = useState('')
  const [eContent, setEContent] = useState('')
  const [eURL, setEURL] = useState<string>('')

  const [fileBusy, setFileBusy] = useState(false)
  const [newFiles, setNewFiles] = useState<FileList | null>(null)

  const canEdit = useMemo(() => !!mat, [mat])

  async function load() {
    if (!materialId) return
    setErr('')
    setLoading(true)
    try {
      const data = await api.teacherGetMaterialPage(materialId)
      setMat(data)
      setEType(data?.type || 'text')
      setETitle(data?.title || '')
      setEContent(data?.content || '')
      setEURL(data?.external_url || '')
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [materialId])

  async function addFiles() {
    if (!materialId) return
    if (!newFiles || newFiles.length === 0) return

    setFileBusy(true)
    setErr('')
    try {
      for (const f of Array.from(newFiles)) {
        const { file_id } = await api.uploadFile(f)
        await api.teacherAddMaterialFile(materialId, file_id)
      }
      setNewFiles(null)
      await load()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setFileBusy(false)
    }
  }

  async function removeFile(fileId: string) {
    if (!materialId) return
    const ok = confirm('Удалить файл из материала?')
    if (!ok) return

    setFileBusy(true)
    setErr('')
    try {
      await api.teacherRemoveMaterialFile(materialId, fileId)
      await load()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setFileBusy(false)
    }
  }

  async function save() {
    if (!materialId) return
    setSaving(true)
    setErr('')
    try {
      await api.teacherUpdateMaterial(materialId, {
        type: eType,
        title: eTitle,
        content: eContent,
        external_url: eURL.trim() === '' ? null : eURL.trim(),
      })
      setEdit(false)
      await load()
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setSaving(false)
    }
  }

  if (!groupId || !materialId) return <div>Загрузка...</div>
  if (loading) return <div>Загрузка...</div>

  return (
    <div>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
        <button onClick={() => router.back()}>Назад</button>
        <h2 style={{ margin: 0 }}>Материал (преподаватель)</h2>
      </div>

      {err ? <div style={{ marginTop: 12, color: 'crimson' }}>{err}</div> : null}

      {!mat ? (
        <div style={{ marginTop: 12 }}>Материал не найден</div>
      ) : (
        <div style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <div>
              <b>ID:</b> {mat.id}
            </div>
            <div>
              <b>Тип:</b> {mat.type}
            </div>
          </div>

          {!edit ? (
            <>
              <h3 style={{ marginTop: 12 }}>{mat.title}</h3>

              {mat.external_url ? (
                <div style={{ marginBottom: 8 }}>
                  <b>Ссылка:</b>{' '}
                  <a href={mat.external_url} target="_blank" rel="noreferrer">
                    {mat.external_url}
                  </a>
                </div>
              ) : null}

              <div style={{ whiteSpace: 'pre-wrap' }}>{mat.content}</div>

              <div style={{ marginTop: 16 }}>
                <h3>Файлы</h3>

                {mat.files?.length ? (
                  <ul style={{ paddingLeft: 18 }}>
                    {mat.files.map((f) => (
                      <li key={f.id} style={{ marginBottom: 10 }}>
                        <div>
                          <b>{f.original_name}</b>{' '}
                          {f.mime_type ? `(${f.mime_type})` : ''}
                          {typeof f.size_bytes === 'number' ? ` — ${fmtSize(f.size_bytes)}` : ''}
                        </div>

                        <div style={{ display: 'flex', gap: 8, marginTop: 6 }}>
                          <button disabled={fileBusy} onClick={() => api.downloadFile(f.id)}>
                            Скачать
                          </button>
                          <button disabled={fileBusy} onClick={() => removeFile(f.id)}>
                            Удалить
                          </button>
                        </div>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <div>Файлов нет</div>
                )}

                <div style={{ marginTop: 12 }}>
                  <div>Добавить файлы</div>
                  <input
                    type="file"
                    multiple
                    onChange={(e) => setNewFiles(e.target.files)}
                    disabled={fileBusy}
                  />
                  <div style={{ marginTop: 8 }}>
                    <button
                      disabled={fileBusy || !newFiles || newFiles.length === 0}
                      onClick={addFiles}
                    >
                      {fileBusy ? 'Обрабатываю...' : 'Добавить'}
                    </button>
                  </div>
                </div>
              </div>

              <div style={{ marginTop: 16 }}>
                <button disabled={!canEdit} onClick={() => setEdit(true)}>
                  Редактировать
                </button>
              </div>
            </>
          ) : (
            <>
              <h3 style={{ marginTop: 12 }}>Редактирование</h3>

              <div style={{ marginTop: 8 }}>
                <div>Тип</div>
                <select value={eType} onChange={(e) => setEType(e.target.value)}>
                  <option value="text">text</option>
                  <option value="file">file</option>
                  <option value="link">link</option>
                  <option value="video">video</option>
                </select>
              </div>

              <div style={{ marginTop: 8 }}>
                <div>Название</div>
                <input value={eTitle} onChange={(e) => setETitle(e.target.value)} />
              </div>

              <div style={{ marginTop: 8 }}>
                <div>Ссылка (опционально)</div>
                <input value={eURL} onChange={(e) => setEURL(e.target.value)} />
              </div>

              <div style={{ marginTop: 8 }}>
                <div>Контент / описание</div>
                <textarea
                  value={eContent}
                  onChange={(e) => setEContent(e.target.value)}
                  rows={8}
                  style={{ width: '100%' }}
                />
              </div>

              <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
                <button disabled={saving} onClick={save}>
                  {saving ? 'Сохраняю...' : 'Сохранить'}
                </button>
                <button disabled={saving} onClick={() => setEdit(false)}>
                  Отмена
                </button>
              </div>

              <div style={{ marginTop: 16, opacity: 0.8 }}>
                Примечание: управление файлами (добавить/удалить) можно сделать следующим шагом.
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}