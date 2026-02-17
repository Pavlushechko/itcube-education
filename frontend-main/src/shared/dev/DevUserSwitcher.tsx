// src/shared/dev/DevUserSwitcher.tsx
'use client'

import { useEffect, useState } from 'react'
import { DEV_IDENTITIES, getIdentity, setIdentity, type DevIdentity } from './auth'

export function DevUserSwitcher() {
  const [cur, setCur] = useState<DevIdentity | null>(null)

  useEffect(() => setCur(getIdentity()), [])

  if (!cur) return null

  return (
    <div style={{ padding: 12, borderBottom: '1px solid #eee', display: 'flex', gap: 12, alignItems: 'center' }}>
      <b>DEV вход:</b>

      <select
        value={cur.userId}
        onChange={(e) => {
          const next = DEV_IDENTITIES.find(x => x.userId === e.target.value) || DEV_IDENTITIES[0]
          setIdentity(next)
          setCur(next)
          window.location.reload()
        }}
      >
        {DEV_IDENTITIES.map(x => (
          <option key={x.userId} value={x.userId}>
            {x.label}
          </option>
        ))}
      </select>

      <span style={{ opacity: 0.75 }}>
        user_id: {cur.userId} | role: {cur.role}
      </span>
    </div>
  )
}
