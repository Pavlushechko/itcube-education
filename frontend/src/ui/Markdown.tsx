// src/ui/Markdown.tsx

import ReactMarkdown from 'react-markdown'

export function Markdown({ text }: { text: string }) {
  return <ReactMarkdown>{text || ''}</ReactMarkdown>
}