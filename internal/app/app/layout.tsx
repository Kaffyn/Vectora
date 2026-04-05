import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Vectora',
  description: 'A private NotebookLM that runs entirely on your machine',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="pt-BR">
      <body className="bg-zinc-950 text-zinc-50">{children}</body>
    </html>
  )
}
