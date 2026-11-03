"use client"

import { Badge, Box, Button, CloseButton, Dialog, HStack, Input, InputGroup, Kbd, Portal, Stack, Text } from "@chakra-ui/react"
import { useRouter } from "next/navigation"
import { useEffect, useMemo, useRef, useState } from "react"
import { LuCornerDownLeft, LuFileText, LuHash, LuSearch } from "react-icons/lu"
import type { SearchEntry } from "@/lib/content"

let index: Promise<SearchEntry[]> | undefined
const loadIndex = () => (index ??= fetch("/search-index.json").then((r) => r.json()))

function search(entries: SearchEntry[], query: string) {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean)
  if (!terms.length) return []
  const scored = entries.map((entry) => {
    const title = `${entry.page} ${entry.heading ?? ""}`.toLowerCase()
    const text = entry.text.toLowerCase()
    let score = 0
    for (const term of terms) {
      if (title.includes(term)) score += 3
      else if (text.includes(term)) score += 1
      else return { entry, score: 0 }
    }
    return { entry, score }
  })
  return scored
    .filter((r) => r.score > 0)
    .sort((a, b) => b.score - a.score)
    .slice(0, 8)
    .map((r) => r.entry)
}

function excerpt(text: string, query: string) {
  const term = query.trim().split(/\s+/)[0]?.toLowerCase() ?? ""
  const at = Math.max(0, text.toLowerCase().indexOf(term))
  const start = Math.max(0, at - 40)
  return `${start > 0 ? "…" : ""}${text.slice(start, start + 140)}`
}

export function Search({ compact }: { compact?: boolean }) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [entries, setEntries] = useState<SearchEntry[]>([])
  const [active, setActive] = useState(0)
  const router = useRouter()
  const trigger = useRef<HTMLButtonElement>(null)
  const results = useMemo(() => search(entries, query), [entries, query])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      // Desktop and mobile each render a trigger; only the visible one owns the shortcut.
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k" && trigger.current?.offsetParent) {
        e.preventDefault()
        setOpen((o) => !o)
      }
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [])

  useEffect(() => {
    if (open) loadIndex().then(setEntries)
  }, [open])

  const go = (entry: SearchEntry) => {
    setOpen(false)
    setQuery("")
    router.push(entry.href)
  }

  return (
    <Dialog.Root open={open} onOpenChange={(e) => setOpen(e.open)} size="lg" placement="top">
      <Dialog.Trigger asChild>
        {compact ? (
          <Button ref={trigger} variant="ghost" size="sm" aria-label="Search documentation" px="2">
            <LuSearch />
          </Button>
        ) : (
          <Button ref={trigger} variant="outline" size="sm" w="full" justifyContent="flex-start" color="fg.muted" fontWeight="normal">
            <LuSearch />
            Search...
            <Kbd ms="auto" size="sm">
              ⌘K
            </Kbd>
          </Button>
        )}
      </Dialog.Trigger>
      <Portal>
        <Dialog.Backdrop />
        <Dialog.Positioner>
          <Dialog.Content overflow="hidden">
            <Dialog.Header p="0" borderBottomWidth="1px">
              <Dialog.Title srOnly>Search documentation</Dialog.Title>
              <InputGroup startElement={<LuSearch />}>
                <Input
                  autoFocus
                  border="0"
                  h="14"
                  pe="12"
                  outline="none"
                  placeholder="Search docs and commands"
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value)
                    setActive(0)
                  }}
                  onKeyDown={(e) => {
                    if (e.key === "ArrowDown") setActive((i) => Math.min(i + 1, results.length - 1))
                    else if (e.key === "ArrowUp") setActive((i) => Math.max(i - 1, 0))
                    else if (e.key === "Enter" && results[active]) go(results[active])
                    else return
                    e.preventDefault()
                  }}
                />
              </InputGroup>
            </Dialog.Header>
            <Dialog.Body p="2" maxH="60vh" overflowY="auto">
              {!query && (
                <Text p="6" textAlign="center" color="fg.muted" textStyle="sm">
                  Try “prompt”, “DESK_HOME”, or “desk cd”.
                </Text>
              )}
              {query && !results.length && (
                <Text p="6" textAlign="center" color="fg.muted" textStyle="sm">
                  No results for “{query}”
                </Text>
              )}
              <Stack gap="1" role="listbox" aria-label="Search results">
                {results.map((r, i) => (
                  <HStack
                    key={r.href}
                    role="option"
                    aria-selected={i === active}
                    align="flex-start"
                    gap="3"
                    p="3"
                    rounded="md"
                    cursor="pointer"
                    bg={i === active ? "teal.subtle" : undefined}
                    onMouseEnter={() => setActive(i)}
                    onClick={() => go(r)}
                  >
                    <Box mt="1" color={i === active ? "teal.fg" : "fg.muted"}>
                      {r.heading ? <LuHash /> : <LuFileText />}
                    </Box>
                    <Box flex="1" minW="0">
                      <HStack gap="2">
                        <Text fontWeight="medium" textStyle="sm">
                          {r.heading ?? r.page}
                        </Text>
                        <Badge size="xs" variant="surface">
                          {r.section}
                        </Badge>
                      </HStack>
                      <Text textStyle="xs" color="fg.muted" lineClamp={2}>
                        {r.heading && `${r.page} · `}
                        {excerpt(r.text, query)}
                      </Text>
                    </Box>
                    {i === active && (
                      <Box mt="1" color="fg.muted">
                        <LuCornerDownLeft />
                      </Box>
                    )}
                  </HStack>
                ))}
              </Stack>
            </Dialog.Body>
            <Dialog.CloseTrigger asChild top="3.5">
              <CloseButton size="sm" />
            </Dialog.CloseTrigger>
          </Dialog.Content>
        </Dialog.Positioner>
      </Portal>
    </Dialog.Root>
  )
}
