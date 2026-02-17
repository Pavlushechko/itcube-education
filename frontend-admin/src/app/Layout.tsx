// src/app/Layout.tsx
import { Link, Outlet } from 'react-router-dom'
import { DevUserSwitcher } from '../ui/DevUserSwitcher'
import { getIdentity } from '../lib/auth'

export function Layout() {
  const ident = getIdentity()
  const isStaff = ident.role === 'admin' || ident.role === 'moderator'

  return (
    <div>
      <DevUserSwitcher />

      <div style={{ padding: 12, display: 'flex', gap: 12 }}>
        <Link to="/catalog">Каталог</Link>
        {isStaff ? <Link to="/applications">Все заявки</Link> : null}
      </div>

      <Outlet />
    </div>
  )
}
