// src/app/layout.tsx
import type { ReactNode } from 'react'
import { DevUserSwitcher } from '@/shared/dev/DevUserSwitcher'

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <DevUserSwitcher />
        {children}
      </body>
    </html>
  )
}
