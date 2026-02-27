// src/app/learn/groups/[groupId]/materials/[materialId]/page.tsx

'use client'

import { useEffect, useState } from 'react'
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

export default function LearnerMaterialPage() {
  const router = useRouter()
  const params = useParams<{ groupId: string; materialId: string }>()
  const groupId = params?.groupId
  const materialId = params?.materialId

  const [mat, setMat] = useState<MaterialPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState('')

  async function load() {
    if (!materialId) return
    setErr('')
    setLoading(true)
    try {
      const data = await api.learnerGetMaterialPage(materialId)
      setMat(data)
      // отметим “прочитано” (не блокируем отображение)
      api.markMaterialRead(materialId).catch(() => {})
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

  async function download(fileId: string) {
    try {
      const { url } = await api.downloadURL(fileId)
      window.open(url, '_blank', 'noopener,noreferrer')
    } catch (e: any) {
      alert(e?.message || String(e))
    }
  }

  if (!groupId || !materialId) return <div>Загрузка...</div>
  if (loading) return <div>Загрузка...</div>

  return (
    <div>
      <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
        <button onClick={() => router.back()}>Назад</button>
        <h2 style={{ margin: 0 }}>Материал</h2>
      </div>

      {err ? <div style={{ marginTop: 12, color: 'crimson' }}>{err}</div> : null}

      {!mat ? (
        <div style={{ marginTop: 12 }}>Материал не найден</div>
      ) : (
        <div style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
            <div><b>Тип:</b> {mat.type}</div>
          </div>

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
              <ul>
                {mat.files.map((f) => (
                  <li key={f.id}>
                    <b>{f.original_name}</b> {f.mime_type ? `(${f.mime_type})` : ''}{' '}
                    {f.size_bytes ? `— ${fmtSize(f.size_bytes)}` : ''}
                    <div>
                      <button onClick={() => download(f.id)}>Скачать</button>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <div>Файлов нет</div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}