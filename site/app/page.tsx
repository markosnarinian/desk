import DocPage, { generateMetadata as docMetadata } from "./(docs)/[section]/[...slug]/page"
import { DocsShell } from "@/components/docs-shell"

const params = Promise.resolve({ section: "docs", slug: ["introduction"] })

export function generateMetadata() {
  return docMetadata({ params })
}

// Render the introduction directly, including on static hosts without redirects.
export default function Home() {
  return (
    <DocsShell>
      <DocPage params={params} />
    </DocsShell>
  )
}
