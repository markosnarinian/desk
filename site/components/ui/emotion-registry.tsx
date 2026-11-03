"use client"

import createCache from "@emotion/cache"
import { CacheProvider } from "@emotion/react"
import { useServerInsertedHTML } from "next/navigation"
import { useState } from "react"

/**
 * Collects Emotion styles during server rendering and emits them in <head>, so they are not
 * rendered inline in the tree, which the client does not reproduce and would fail to hydrate.
 */
export function EmotionRegistry({ children }: { children: React.ReactNode }) {
  const [registry] = useState(() => {
    const cache = createCache({ key: "css" })
    // Compat mode stores rules in cache.inserted instead of rendering <style> elements inline.
    cache.compat = true
    let inserted: { name: string; global: boolean }[] = []
    const insert = cache.insert
    cache.insert = (...args) => {
      const [selector, serialized] = args
      if (cache.inserted[serialized.name] === undefined) inserted.push({ name: serialized.name, global: !selector })
      return insert(...args)
    }
    const flush = () => {
      const flushed = inserted
      inserted = []
      return flushed
    }
    return { cache, flush }
  })

  useServerInsertedHTML(() => {
    const flushed = registry.flush()
    if (!flushed.length) return null
    const { cache } = registry
    const rules = (global: boolean) => {
      const items = flushed.filter((item) => item.global === global && typeof cache.inserted[item.name] === "string")
      return { names: items.map((i) => i.name).join(" "), css: items.map((i) => cache.inserted[i.name]).join("") }
    }
    const globals = rules(true)
    const styles = rules(false)
    return (
      <>
        {globals.css && (
          <style data-emotion={`${cache.key}-global ${globals.names}`} dangerouslySetInnerHTML={{ __html: globals.css }} />
        )}
        {styles.css && <style data-emotion={`${cache.key} ${styles.names}`} dangerouslySetInnerHTML={{ __html: styles.css }} />}
      </>
    )
  })

  return <CacheProvider value={registry.cache}>{children}</CacheProvider>
}
