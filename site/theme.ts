import { createSystem, defaultConfig, defineConfig } from "@chakra-ui/react"

const config = defineConfig({
  globalCss: {
    html: { scrollPaddingTop: "5rem" },
    // Shiki emits both code themes as CSS variables; pick one per color mode.
    ".shiki span": { color: "var(--shiki-light)" },
    ".dark .shiki span": { color: "var(--shiki-dark)" },
  },
  theme: {
    tokens: {
      fonts: {
        heading: { value: "var(--font-inter), Inter, sans-serif" },
        body: { value: "var(--font-inter), Inter, sans-serif" },
      },
    },
  },
})

export const system = createSystem(defaultConfig, config)
