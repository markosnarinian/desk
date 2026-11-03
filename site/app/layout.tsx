import type { Metadata } from "next"
import { Inter } from "next/font/google"
import { Provider } from "@/components/ui/provider"
import { siteUrl } from "@/lib/site"

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" })

export const metadata: Metadata = {
  metadataBase: siteUrl ? new URL(siteUrl) : undefined,
  title: { default: "desk — workspaces for the work in between", template: "%s | desk" },
  description: "desk keeps quick experiments and one-off clones in managed workspaces, and only cleans up what it can prove is safe.",
  icons: { icon: "/favicon.svg" },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.variable} suppressHydrationWarning>
      <body>
        <Provider>{children}</Provider>
      </body>
    </html>
  )
}
