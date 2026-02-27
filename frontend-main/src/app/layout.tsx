// src/app/layout.tsx
import type { ReactNode } from 'react'
import { DevUserSwitcher } from '@/shared/dev/DevUserSwitcher'
import Link from 'next/link'

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <div>
          <Link href="/">Каталог</Link>{' '}
        </div>
        <DevUserSwitcher />
        {children}
      </body>
    </html>
  )
}
