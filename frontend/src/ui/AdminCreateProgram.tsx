// src/ui/AdminCreateProgram.tsx

import { useState } from 'react'
import { api } from '../lib/api'

type Props = {
  onCreated?: (programId: string) => void
}

export function AdminCreateProgram(props: Props) {
  const [open, setOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')

  async function submit() {
    setErr('')
    if (!title.trim()) {
      setErr('Название обязательно')
      return
    }
    setBusy(true)
    try {
      const res = await api.createProgram(title.trim(), description)
      props.onCreated?.(res.id)
      alert('Программа создана: ' + res.id)

      // очистить форму
      setTitle('')
      setDescription('')
      setOpen(false)
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div style={{ margin: '12px 0' }}>
      {!open ? (
        <button onClick={() => setOpen(true)}>+ Создать программу</button>
      ) : (
        <div style={{ border: '1px solid #ddd', padding: 12, borderRadius: 6, maxWidth: 720 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <b>Новая программа</b>
            <button onClick={() => setOpen(false)} disabled={busy}>Закрыть</button>
          </div>

          <div style={{ marginTop: 10 }}>
            <div>Название</div>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              style={{ width: '100%', padding: 8 }}
              placeholder="Например: Go backend"
              disabled={busy}
            />
          </div>

          <div style={{ marginTop: 10 }}>
            <div>Описание (пока текст/ссылки, можно Markdown)</div>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              style={{ width: '100%', padding: 8, minHeight: 140 }}
              placeholder="Описание курса. Можно вставлять ссылки: https://..."
              disabled={busy}
            />
            <div style={{ fontSize: 12, opacity: 0.75, marginTop: 6 }}>
              Фото/гифки “как в редакторе” потребуют отдельного rich-text и загрузки файлов. Для MVP — текст/ссылки/Markdown.
            </div>
          </div>

          {err && <div style={{ color: 'crimson', marginTop: 10 }}>{err}</div>}

          <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
            <button onClick={submit} disabled={busy}>
              {busy ? 'Создаю...' : 'Создать'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
