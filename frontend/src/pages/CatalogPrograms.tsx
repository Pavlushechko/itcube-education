// src/pages/CatalogPrograms.tsx

import { useEffect, useMemo, useState } from 'react'
import { api } from '../lib/api'
import type { Program } from '../lib/types'
import { Link } from 'react-router-dom'
import { getIdentity } from '../lib/auth'
import { AdminCreateProgram } from '../ui/AdminCreateProgram'

export function CatalogPrograms() {
  const ident = getIdentity()
  const isStaff = ident.role === 'admin' || ident.role === 'moderator'

  const [publicItems, setPublicItems] = useState<Program[]>([])
  const [staffItems, setStaffItems] = useState<Program[]>([])
  const [err, setErr] = useState<string>('')

  useEffect(() => {
    setErr('')
    api.listPrograms()
      .then(setPublicItems)
      .catch(e => setErr(String(e.message || e)))
  }, [])

  useEffect(() => {
    if (!isStaff) return
    api.listProgramsAdmin()
      .then(setStaffItems)
      .catch(e => setErr(String(e.message || e)))
  }, [isStaff])

  const drafts = useMemo(
    () => staffItems.filter(p => p.Status === 'draft'),
    [staffItems]
  )

  return (
    <div style={{ padding: 12 }}>
      <h2>Каталог программ</h2>

      {ident.role === 'admin' && (
        <AdminCreateProgram onCreated={(id) => (window.location.href = `/program/${id}`)} />
      )}

      {err && <div style={{ color: 'crimson' }}>{err}</div>}

      {isStaff && (
        <>
          <h3>Черновики (видно сотрудникам)</h3>
          {drafts.length === 0 ? (
            <div style={{ opacity: 0.75 }}>Нет черновиков</div>
          ) : (
            <ul>
              {drafts.map(p => (
                <li key={p.ID}>
                  <Link to={`/program/${p.ID}`}>
                    {p.Title}
                    {isStaff ? ` (${p.ID})` : ''}
                  </Link>{' '}
                  <span style={{ opacity: 0.75 }}>(draft)</span>
                </li>
              ))}
            </ul>
          )}
        </>
      )}

      <h3>Опубликованные</h3>
      <ul>
        {publicItems.map(p => (
          <li key={p.ID}>
            <Link to={`/program/${p.ID}`}>
              {p.Title}
              {isStaff ? ` (${p.ID})` : ''}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
