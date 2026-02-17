'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/shared/api/client'

type Program = {
  ID: string
  Title: string
  Description: string
  Status?: string
  CreatedAt?: string
}

export default function CatalogPage() {
  const [publicItems, setPublicItems] = useState<Program[]>([])
  const [err, setErr] = useState<string>('')

  useEffect(() => {
    setErr('')
    api.listPrograms()
      .then(setPublicItems)
      .catch((e: any) => setErr(String(e?.message || e)))
  }, [])

  return (
    <div style={{ padding: 12 }}>
      <h2>Каталог программ</h2>

      {err ? <div style={{ color: 'crimson' }}>{err}</div> : null}

      <ul>
        {publicItems.map((p) => (
          <li key={p.ID}>
            <Link href={`/program/${p.ID}`}>{p.Title}</Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
