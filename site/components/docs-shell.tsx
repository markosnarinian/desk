"use client"

import { Box, Button, CloseButton, Drawer, Flex, HStack, IconButton, Link, Portal, Stack, Text } from "@chakra-ui/react"
import NextLink from "next/link"
import { usePathname } from "next/navigation"
import { useEffect, useState } from "react"
import { LuGithub, LuMenu } from "react-icons/lu"
import { groups, href } from "@/lib/nav"
import { iconFor } from "./icons"
import { Logo } from "./logo"
import { Search } from "./search"
import { ColorModeButton } from "./ui/color-mode"

export const githubUrl = "https://github.com/markosnarinian/desk"

const trim = (path: string) => path.replace(/\/$/, "")

function SidebarNav() {
  const path = usePathname()
  const pathname = trim(path === "/" ? "/docs/introduction/" : path)
  return (
    <Stack as="nav" aria-label="Documentation" gap="5">
      {groups.map((group) => {
        const Icon = iconFor(group.icon)
        return (
          <Stack key={group.label} gap="0.5">
            <HStack px="3" py="1.5" gap="2.5" fontWeight="semibold" textStyle="sm">
              <Icon aria-hidden />
              {group.label}
            </HStack>
            {group.items.map((item) => {
              const to = href(item.section, item.slug)
              const current = pathname === trim(to)
              return (
                <Link
                  key={item.slug}
                  asChild
                  px="3"
                  ps={item.section === "commands" && item.slug !== "overview" ? "7" : "3"}
                  py="1.5"
                  rounded="md"
                  textStyle="sm"
                  textDecoration="none"
                  color={current ? "teal.fg" : "fg.muted"}
                  bg={current ? "teal.subtle" : undefined}
                  fontWeight={current ? "medium" : undefined}
                  _hover={current ? undefined : { color: "fg", bg: "bg.muted" }}
                >
                  <NextLink href={to} aria-current={current ? "page" : undefined}>
                    {item.label}
                  </NextLink>
                </Link>
              )
            })}
          </Stack>
        )
      })}
    </Stack>
  )
}

function MobileNav() {
  const [open, setOpen] = useState(false)
  const pathname = usePathname()
  useEffect(() => setOpen(false), [pathname])
  return (
    <Drawer.Root open={open} onOpenChange={(e) => setOpen(e.open)} placement="start" size="xs">
      <Drawer.Trigger asChild>
        <IconButton variant="ghost" size="sm" aria-label="Open navigation" hideFrom="lg">
          <LuMenu />
        </IconButton>
      </Drawer.Trigger>
      <Portal>
        <Drawer.Backdrop />
        <Drawer.Positioner>
          <Drawer.Content>
            <Drawer.Header>
              <Drawer.Title asChild>
                <Box>
                  <Logo />
                </Box>
              </Drawer.Title>
            </Drawer.Header>
            <Drawer.Body>
              <SidebarNav />
            </Drawer.Body>
            <Drawer.CloseTrigger asChild>
              <CloseButton size="sm" />
            </Drawer.CloseTrigger>
          </Drawer.Content>
        </Drawer.Positioner>
      </Portal>
    </Drawer.Root>
  )
}

export function DocsShell({ children }: { children: React.ReactNode }) {
  return (
    <Box minH="100dvh">
      <Flex
        as="aside"
        hideBelow="lg"
        direction="column"
        position="fixed"
        insetY="0"
        insetStart="0"
        w="64"
        borderEndWidth="1px"
        bg="bg"
      >
        <Flex h="14" align="center" px="5" flexShrink="0">
          <Logo />
        </Flex>
        <Box px="4" pt="2" pb="4">
          <Search />
        </Box>
        <Box flex="1" overflowY="auto" px="3" pb="8">
          <SidebarNav />
        </Box>
      </Flex>
      <Box ms={{ lg: "64" }}>
        <Box as="header" position="sticky" top="0" zIndex="sticky" h="14" px={{ base: "4", md: "8" }} borderBottomWidth="1px" bg="bg">
          <HStack h="full" justify="space-between" gap="4">
            <HStack gap="1" hideFrom="lg">
              <MobileNav />
              <Logo />
            </HStack>
            <Text hideBelow="lg" textStyle="sm" color="fg.muted">Documentation</Text>
            <HStack gap="1">
              <Box hideFrom="lg">
                <Search compact />
              </Box>
              <Button asChild variant="ghost" size="sm" hideBelow="sm">
                <a href={githubUrl}>
                  <LuGithub /> GitHub
                </a>
              </Button>
              <Button asChild size="sm" colorPalette="teal" hideBelow="sm">
                <NextLink href="/docs/quickstart/">Get started</NextLink>
              </Button>
              <ColorModeButton />
            </HStack>
          </HStack>
        </Box>
        {children}
      </Box>
    </Box>
  )
}
