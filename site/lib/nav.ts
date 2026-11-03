export type Badge = { text: string; color: "green" | "blue" | "gray" | "purple" | "orange" | "red" }
export type NavItem = { slug: string; label: string; badge?: Badge }
export type NavGroup = { label: string; icon: string; items: NavItem[] }
export type Section = { id: "docs" | "commands"; label: string; groups: NavGroup[] }

const create: Badge = { text: "NEW", color: "green" }
const run: Badge = { text: "RUN", color: "blue" }
const read: Badge = { text: "READ", color: "gray" }
const mark: Badge = { text: "MARK", color: "purple" }
const move: Badge = { text: "MOVE", color: "orange" }
const remove: Badge = { text: "DELETE", color: "red" }

export const sections: Section[] = [
  {
    id: "docs",
    label: "Documentation",
    groups: [
      {
        label: "Getting Started",
        icon: "book",
        items: [
          { slug: "introduction", label: "Introduction" },
          { slug: "installation", label: "Installation" },
          { slug: "quickstart", label: "Quickstart" },
        ],
      },
      {
        label: "Concepts",
        icon: "layers",
        items: [
          { slug: "workspaces", label: "Workspaces" },
          { slug: "sessions", label: "Sessions" },
          { slug: "cleanup-lifecycle", label: "Cleanup lifecycle" },
          { slug: "retention-rules", label: "Retention rules" },
        ],
      },
      {
        label: "Guides",
        icon: "compass",
        items: [
          { slug: "scratch-experiments", label: "Scratch experiments" },
          { slug: "repository-workspaces", label: "Repository workspaces" },
          { slug: "recovering-work", label: "Recovering work" },
        ],
      },
      {
        label: "Reference",
        icon: "settings",
        items: [
          { slug: "configuration", label: "Configuration" },
          { slug: "troubleshooting", label: "Troubleshooting" },
        ],
      },
    ],
  },
  {
    id: "commands",
    label: "Commands",
    groups: [
      { label: "Overview", icon: "info", items: [{ slug: "overview", label: "Overview" }] },
      {
        label: "Create",
        icon: "plus",
        items: [
          { slug: "scratch", label: "desk scratch", badge: create },
          { slug: "open", label: "desk open", badge: create },
        ],
      },
      {
        label: "Work",
        icon: "terminal",
        items: [
          { slug: "enter", label: "desk enter", badge: run },
          { slug: "list", label: "desk list", badge: read },
        ],
      },
      {
        label: "Retain",
        icon: "pin",
        items: [
          { slug: "keep", label: "desk keep", badge: mark },
          { slug: "done", label: "desk done", badge: mark },
        ],
      },
      {
        label: "Clean up",
        icon: "archive",
        items: [
          { slug: "gc", label: "desk gc", badge: move },
          { slug: "remove", label: "desk remove", badge: move },
          { slug: "restore", label: "desk restore", badge: move },
          { slug: "purge", label: "desk purge", badge: remove },
        ],
      },
      { label: "Other", icon: "hash", items: [{ slug: "version", label: "desk version", badge: read }] },
    ],
  },
]

export const pages = sections.flatMap((section) =>
  section.groups.flatMap((group) => group.items.map((item) => ({ ...item, section: section.id, group: group.label }))),
)

export const href = (section: string, slug: string) => `/${section}/${slug}/`

export const sectionHome = (id: Section["id"]) => {
  const first = sections.find((s) => s.id === id)!.groups[0].items[0]
  return href(id, first.slug)
}
