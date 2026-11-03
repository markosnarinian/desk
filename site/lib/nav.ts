export type Badge = { text: string; color: "green" | "blue" | "gray" | "purple" | "orange" | "red" }
export type NavItem = { slug: string; label: string; badge?: Badge }
export type NavGroup = { label: string; icon: string; items: NavItem[] }
export type Section = { id: "docs" | "commands"; label: string; groups: NavGroup[] }

const create: Badge = { text: "NEW", color: "green" }
const move: Badge = { text: "MOVE", color: "blue" }
const read: Badge = { text: "READ", color: "gray" }
const setup: Badge = { text: "SETUP", color: "purple" }
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
          { slug: "shell-integration", label: "Shell integration" },
        ],
      },
      {
        label: "Guides",
        icon: "compass",
        items: [
          { slug: "repository-workspaces", label: "Repository workspaces" },
          { slug: "amp", label: "Test Amp work locally" },
          { slug: "cleanup", label: "Clear away finished work" },
        ],
      },
      {
        label: "Reference",
        icon: "settings",
        items: [{ slug: "configuration", label: "Configuration" }],
      },
    ],
  },
  {
    id: "commands",
    label: "Commands",
    groups: [
      { label: "Overview", icon: "info", items: [{ slug: "overview", label: "Overview" }] },
      {
        label: "Workspaces",
        icon: "terminal",
        items: [
          { slug: "new", label: "desk new", badge: create },
          { slug: "switch", label: "desk switch", badge: move },
          { slug: "info", label: "desk info", badge: read },
          { slug: "list", label: "desk list", badge: read },
          { slug: "rename", label: "desk rename", badge: move },
          { slug: "finish", label: "desk finish" },
          { slug: "cleanup", label: "desk cleanup", badge: remove },
          { slug: "rm", label: "desk rm", badge: remove },
        ],
      },
      {
        label: "Other",
        icon: "hash",
        items: [
          { slug: "amp", label: "desk amp", badge: create },
          { slug: "init", label: "desk init", badge: setup },
          { slug: "version", label: "desk version", badge: read },
        ],
      },
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
