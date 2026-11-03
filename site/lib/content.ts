import fs from "node:fs"
import path from "node:path"
import { evaluate } from "@mdx-js/mdx"
import rehypeShiki from "@shikijs/rehype"
import GithubSlugger from "github-slugger"
import matter from "gray-matter"
import type { Element, Root } from "hast"
import { toString } from "hast-util-to-string"
import type { MDXContent } from "mdx/types"
import type { ShikiTransformer } from "shiki"
import * as runtime from "react/jsx-runtime"
import rehypeSlug from "rehype-slug"
import remarkGfm from "remark-gfm"
import { visit } from "unist-util-visit"
import { pages } from "./nav"

export type Heading = { id: string; text: string; depth: 2 | 3 }
export type SearchEntry = { href: string; page: string; section: string; heading?: string; text: string }
type Link = { label: string; section: string; slug: string }

export type Page = {
  section: string
  slug: string
  title: string
  description: string
  group: string
  source: string
  headings: Heading[]
  Content: MDXContent
  prev?: Link
  next?: Link
}

const contentDir = path.join(process.cwd(), "content")

function read(section: string, slug: string) {
  const { data, content } = matter(fs.readFileSync(path.join(contentDir, section, `${slug}.mdx`), "utf8"))
  return { title: String(data.title), description: String(data.description ?? ""), body: content }
}

const cache = new Map<string, Promise<Page>>()

export function getPage(section: string, slug: string) {
  const key = `${section}/${slug}`
  let page = cache.get(key)
  if (!page) cache.set(key, (page = load(section, slug)))
  return page
}

async function load(section: string, slug: string): Promise<Page> {
  const index = pages.findIndex((p) => p.section === section && p.slug === slug)
  if (index < 0) throw new Error(`Unknown page: ${section}/${slug}`)
  const { title, description, body } = read(section, slug)

  const headings: Heading[] = []
  const collectHeadings = () => (tree: Root) =>
    visit(tree, "element", (node: Element) => {
      if ((node.tagName === "h2" || node.tagName === "h3") && node.properties.id) {
        headings.push({ id: String(node.properties.id), text: toString(node), depth: node.tagName === "h2" ? 2 : 3 })
      }
    })

  const { default: Content } = await evaluate(body, {
    ...runtime,
    remarkPlugins: [remarkGfm],
    rehypePlugins: [
      rehypeSlug,
      collectHeadings,
      [
        rehypeShiki,
        {
          themes: { light: "github-light", dark: "github-dark" },
          defaultColor: false,
          defaultLanguage: "text",
          fallbackLanguage: "text",
          transformers: [
            {
              pre(node) {
                node.properties.dataLanguage = this.options.lang
                node.properties.dataCode = this.source
                // A fence title such as ```text title="Usage" labels the block.
                const title = /title="([^"]+)"/.exec(this.options.meta?.__raw ?? "")?.[1]
                if (title) node.properties.dataTitle = title
              },
            } satisfies ShikiTransformer,
          ],
        },
      ],
    ],
  })

  const link = (p?: (typeof pages)[number]) => p && { label: p.label, section: p.section, slug: p.slug }
  return {
    section,
    slug,
    title,
    description,
    group: pages[index].group,
    source: `# ${title}\n\n${description}\n\n${body.trim()}\n`,
    headings,
    Content,
    prev: link(pages[index - 1]),
    next: link(pages[index + 1]),
  }
}

// Plain text for search: drop JSX tags, link targets, and Markdown punctuation.
const plain = (s: string) =>
  s
    .replace(/<[^>]*>/g, " ")
    .replace(/\[([^\]]*)\]\([^)]*\)/g, "$1")
    .replace(/[`*|#>]/g, "")
    .replace(/\s+/g, " ")
    .trim()

/** Splits every page at its h2/h3 headings; anchors use the same slugger as rehype-slug. */
export function getSearchEntries(): SearchEntry[] {
  const entries: SearchEntry[] = []
  for (const p of pages) {
    const { title, description, body } = read(p.section, p.slug)
    const base = `/${p.section}/${p.slug}/`
    const section = p.group
    const slugger = new GithubSlugger()
    let current: SearchEntry = { href: base, page: title, section, text: description }
    entries.push(current)
    let fenced = false
    for (const line of body.split("\n")) {
      if (line.trimStart().startsWith("```")) fenced = !fenced
      const match = !fenced && /^(##|###) (.+)$/.exec(line)
      if (match) {
        const heading = plain(match[2])
        current = { href: `${base}#${slugger.slug(heading)}`, page: title, section, heading, text: "" }
        entries.push(current)
      } else {
        current.text += ` ${plain(line)}`
      }
    }
  }
  return entries.map((e) => ({ ...e, text: e.text.trim() }))
}
