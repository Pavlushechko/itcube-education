// src/shared/ui/Markdown.tsx

'use client'

import ReactMarkdown from 'react-markdown'

export function Markdown({ text }: { text: string }) {
  return <ReactMarkdown>{text || ''}</ReactMarkdown>
}
