"use client"

import { Button, Clipboard, IconButton } from "@chakra-ui/react"

/** Icon-only by default; pass `children` for a labelled button such as "Copy page". */
export function CopyButton({ value, label, children }: { value: string; label: string; children?: React.ReactNode }) {
  return (
    <Clipboard.Root value={value}>
      <Clipboard.Trigger asChild>
        {children ? (
          <Button variant="outline" size="sm">
            <Clipboard.Indicator />
            <Clipboard.Indicator copied="Copied">{children}</Clipboard.Indicator>
          </Button>
        ) : (
          <IconButton variant="ghost" size="xs" aria-label={label} color="fg.muted">
            <Clipboard.Indicator />
          </IconButton>
        )}
      </Clipboard.Trigger>
    </Clipboard.Root>
  )
}
