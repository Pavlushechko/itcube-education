// src/ui/AdminPublishProgram.tsx

import { useState } from 'react'
import { api } from '../lib/api'

export function AdminPublishProgram({ programId }: { programId: string }) {
  const [busy, setBusy] = useState(false)

  async function publish() {
    setBusy(true)
    try {
      await api.publishProgram(programId)
      alert('Опубликовано')
      window.location.reload()
    } catch (e: any) {
      alert('Ошибка: ' + (e?.message || String(e)))
    } finally {
      setBusy(false)
    }
  }

  return (
    <button onClick={publish} disabled={busy}>
      {busy ? 'Публикую...' : 'Опубликовать'}
    </button>
  )
}
