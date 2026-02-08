// src/app/Layout.tsx

import { Link, Outlet } from 'react-router-dom'
import { DevUserSwitcher } from '../ui/DevUserSwitcher'

export function Layout() {
  return (
    <div>
      <DevUserSwitcher />
      <div style={{ padding: 12, display: 'flex', gap: 12 }}>
        <Link to="/">Каталог</Link>
        <Link to="/me/applications">Мои заявки</Link>
        <Link to="/teacher/groups">Мои группы (препод)</Link>
      </div>

      <Outlet />
    </div>
  )
}
