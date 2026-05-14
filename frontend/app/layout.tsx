import type { Metadata } from 'next'
import { Bricolage_Grotesque, IBM_Plex_Mono } from 'next/font/google'
import { Providers } from './providers'
import './globals.css'

const headingFont = Bricolage_Grotesque({
  subsets: ['latin'],
  variable: '--font-heading',
})

const monoFont = IBM_Plex_Mono({
  subsets: ['latin'],
  weight: ['400', '500'],
  variable: '--font-mono',
})

export const metadata: Metadata = {
  title: 'Live IG Display',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="th">
      <body className={`${headingFont.variable} ${monoFont.variable} min-h-screen text-stone-950`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
