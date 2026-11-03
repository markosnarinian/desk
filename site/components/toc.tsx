"use client"

import { Box, HStack, Link, Stack, Text } from "@chakra-ui/react"
import { useEffect, useState } from "react"
import { LuAlignLeft } from "react-icons/lu"
import type { Heading } from "@/lib/content"

export function TableOfContents({ headings }: { headings: Heading[] }) {
  const [active, setActive] = useState<string | undefined>(headings[0]?.id)

  useEffect(() => {
    // The active heading is the last one scrolled past the sticky header.
    const update = () => {
      let current: string | undefined = headings[0]?.id
      for (const { id } of headings) {
        const el = document.getElementById(id)
        if (el && el.getBoundingClientRect().top <= 100) current = id
      }
      if (window.innerHeight + window.scrollY >= document.body.scrollHeight - 2) current = headings.at(-1)?.id
      setActive(current)
    }
    update()
    window.addEventListener("scroll", update, { passive: true })
    return () => window.removeEventListener("scroll", update)
  }, [headings])

  if (!headings.length) return null
  return (
    <Box as="nav" aria-label="On this page">
      <HStack gap="2" mb="3" textStyle="sm" fontWeight="medium">
        <LuAlignLeft aria-hidden />
        <Text>On this page</Text>
      </HStack>
      <Stack gap="0" borderStartWidth="1px">
        {headings.map((h) => {
          const current = h.id === active
          return (
            <Link
              key={h.id}
              href={`#${h.id}`}
              aria-current={current ? "location" : undefined}
              display="block"
              py="1"
              ps={h.depth === 3 ? "6" : "3"}
              ms="-1px"
              borderStartWidth="1px"
              borderColor={current ? "teal.solid" : "transparent"}
              textStyle="sm"
              textDecoration="none"
              color={current ? "teal.fg" : "fg.muted"}
              _hover={{ color: current ? undefined : "fg" }}
            >
              {h.text}
            </Link>
          )
        })}
      </Stack>
    </Box>
  )
}
