// src/pages/AdminCreateProgram.tsx

import { useState } from 'react'
import { api, ApiError } from '../lib/api'
import { getIdentity } from '../lib/auth'
import { Link } from 'react-router-dom'

export function AdminCreateProgram() {
  const ident = getIdentity()
  const [title, setTitle] = useState('Go backend')
  const [description, setDescription] = useState('описание')
  const [createdId, setCreatedId] = useState('')
  const [err, setErr] = useState('')

  if (ident.role !== 'admin') {
    return (
      <div style={{ padding: 12 }}>
        <div style={{ color: 'crimson' }}>Нет доступа. Нужно войти как admin.</div>
        <div style={{ marginTop: 8 }}><Link to="/">В каталог</Link></div>
      </div>
    )
  }

  async function createDraft() {
    setErr('')
    try {
      const res = await api.createProgram(title, description)
      setCreatedId(res.id)
      alert('Черновик создан: ' + res.id)
    } catch (e: any) {
      setErr(e instanceof ApiError ? e.message : String(e))
    }
  }

  async function publish() {
    if (!createdId) return
    setErr('')
    try {
      await api.publishProgram(createdId)
      alert('Опубликовано')
    } catch (e: any) {
      setErr(e instanceof ApiError ? e.message : String(e))
    }
  }

  return (
    <div style={{ padding: 12 }}>
      <h2>Создать курс (admin)</h2>
      {err && <div style={{ color: 'crimson' }}>{err}</div>}

      <div style={{ display: 'grid', gap: 10, maxWidth: 520 }}>
        <label>
          Название
          <input value={title} onChange={(e) => setTitle(e.target.value)} style={{ width: '100%' }} />
        </label>

        <label>
          Описание
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} style={{ width: '100%' }} />
        </label>

        <button onClick={createDraft}>Создать черновик</button>

        <div>program_id: <code>{createdId || '—'}</code></div>
        <button disabled={!createdId} onClick={publish}>Опубликовать</button>
      </div>
    </div>
  )
}
