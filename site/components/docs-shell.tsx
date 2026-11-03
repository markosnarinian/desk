"use client"

import { Badge, Box, Button, CloseButton, Drawer, Flex, HStack, IconButton, Link, Portal, Stack } from "@chakra-ui/react"
import NextLink from "next/link"
import { usePathname } from "next/navigation"
import { useEffect, useState } from "react"
import { LuGithub, LuMenu } from "react-icons/lu"
import { href, sectionHome, sections } from "@/lib/nav"
import { iconFor } from "./icons"
import { Logo } from "./logo"
import { Search } from "./search"
import { ColorModeButton } from "./ui/color-mode"

export const githubUrl = "https://github.com/markosnarinian/desk"

const trim = (path: string) => path.replace(/\/$/, "")

function useSection() {
  const pathname = usePathname()
  return sections.find((s) => pathname.startsWith(`/${s.id}/`)) ?? sections[0]
}

function SidebarNav() {
  const pathname = trim(usePathname())
  const section = useSection()
  return (
    <Stack as="nav" aria-label={section.label} gap="6">
      {section.groups.map((group) => {
        const Icon = iconFor(group.icon)
        return (
          <Stack key={group.label} gap="0.5">
            <HStack px="3" py="1.5" gap="2.5" fontWeight="semibold" textStyle="sm">
              <Icon aria-hidden />
              {group.label}
            </HStack>
            {group.items.map((item) => {
              const to = href(section.id, item.slug)
              const current = pathname === trim(to)
              return (
                <Link
                  key={item.slug}
                  asChild
                  px="3"
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
                    {item.badge && (
                      <Badge size="xs" colorPalette={item.badge.color} fontFamily="mono" minW="12" justifyContent="center">
                        {item.badge.text}
                      </Badge>
                    )}
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

function SectionTabs() {
  const section = useSection()
  return (
    <HStack as="nav" aria-label="Sections" gap="6" h="full" hideBelow="lg">
      {sections.map((s) => {
        const current = s.id === section.id
        return (
          <Link
            key={s.id}
            asChild
            h="full"
            textStyle="sm"
            fontWeight="medium"
            textDecoration="none"
            color={current ? "teal.fg" : "fg.muted"}
            borderBottomWidth="2px"
            borderColor={current ? "teal.solid" : "transparent"}
            _hover={current ? undefined : { color: "fg" }}
          >
            <NextLink href={sectionHome(s.id)}>{s.label}</NextLink>
          </Link>
        )
      })}
    </HStack>
  )
}

function MobileNav() {
  const [open, setOpen] = useState(false)
  const pathname = usePathname()
  const current = useSection()
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
              <HStack mb="6" gap="2">
                {sections.map((s) => (
                  <Button key={s.id} asChild size="xs" variant={s.id === current.id ? "subtle" : "ghost"} colorPalette="teal">
                    <NextLink href={sectionHome(s.id)}>{s.label}</NextLink>
                  </Button>
                ))}
              </HStack>
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
            <SectionTabs />
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
