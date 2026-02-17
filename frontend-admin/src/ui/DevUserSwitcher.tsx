// src/ui/DevUserSwitcher.tsx

import { DEV_IDENTITIES, getIdentity, setIdentity } from '../lib/auth'
import { useState } from 'react'

export function DevUserSwitcher() {
  const [current, setCurrent] = useState(getIdentity())

  return (
    <div style={{ padding: 12, borderBottom: '1px solid #ddd', display: 'flex', gap: 12, alignItems: 'center' }}>
      <strong>DEV вход:</strong>
      <select
        value={current.label}
        onChange={(e) => {
          const next = DEV_IDENTITIES.find(x => x.label === e.target.value)!
          setIdentity(next)
          setCurrent(next)
          window.location.reload()
        }}
      >
        {DEV_IDENTITIES.map(x => (
          <option key={x.label} value={x.label}>{x.label}</option>
        ))}
      </select>

      <span style={{ opacity: 0.75 }}>
        user_id: {current.userId} | role: {current.role}
      </span>
    </div>
  )
}
