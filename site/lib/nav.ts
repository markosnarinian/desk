type NavItem = { section: "docs" | "commands"; slug: string; label: string }
type NavGroup = { label: string; icon: string; items: NavItem[] }

// Navigation is unified; section remains part of each page's existing URL.
export const groups: NavGroup[] = [
  {
    label: "Getting started",
    icon: "book",
    items: [
      { section: "docs", slug: "introduction", label: "Introduction" },
      { section: "docs", slug: "installation", label: "Installation" },
      { section: "docs", slug: "quickstart", label: "Quickstart" },
    ],
  },
  {
    label: "Concepts",
    icon: "layers",
    items: [
      { section: "docs", slug: "workspaces", label: "Workspaces" },
      { section: "docs", slug: "shell-integration", label: "Shell integration" },
    ],
  },
  {
    label: "Use cases",
    icon: "compass",
    items: [
      { section: "docs", slug: "repository-workspaces", label: "Repository workspaces" },
      { section: "docs", slug: "amp", label: "Test Amp work locally" },
      { section: "docs", slug: "cleanup", label: "Clear away finished work" },
    ],
  },
  {
    label: "Reference",
    icon: "settings",
    items: [
      { section: "docs", slug: "configuration", label: "Configuration" },
      { section: "commands", slug: "overview", label: "Commands" },
      { section: "commands", slug: "new", label: "desk new" },
      { section: "commands", slug: "switch", label: "desk switch" },
      { section: "commands", slug: "info", label: "desk info" },
      { section: "commands", slug: "list", label: "desk list" },
      { section: "commands", slug: "rename", label: "desk rename" },
      { section: "commands", slug: "finish", label: "desk finish" },
      { section: "commands", slug: "cleanup", label: "desk cleanup" },
      { section: "commands", slug: "rm", label: "desk rm" },
      { section: "commands", slug: "amp", label: "desk amp" },
      { section: "commands", slug: "init", label: "desk init" },
      { section: "commands", slug: "version", label: "desk version" },
    ],
  },
]

export const pages = groups.flatMap((group) => group.items.map((item) => ({ ...item, group: group.label })))

export const href = (section: string, slug: string) => `/${section}/${slug}/`
