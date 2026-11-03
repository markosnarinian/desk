import type { MetadataRoute } from "next"
import { href, pages } from "@/lib/nav"
import { siteUrl } from "@/lib/site"

export const dynamic = "force-static"

// Without a canonical origin there are no absolute URLs to list, so the sitemap is empty.
export default function sitemap(): MetadataRoute.Sitemap {
  if (!siteUrl) return []
  return ["/", ...pages.map((p) => href(p.section, p.slug))].map((path) => ({ url: new URL(path, siteUrl).href }))
}
