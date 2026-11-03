import { Box, Flex, Heading, HStack, IconButton, Link, LinkBox, LinkOverlay, SimpleGrid, Stack, Text } from "@chakra-ui/react"
import type { Metadata } from "next"
import NextLink from "next/link"
import { LuChevronLeft, LuChevronRight, LuGithub } from "react-icons/lu"
import { CopyButton } from "@/components/copy-button"
import { githubUrl } from "@/components/docs-shell"
import { mdxComponents } from "@/components/mdx"
import { TableOfContents } from "@/components/toc"
import { Prose } from "@/components/ui/prose"
import { getPage } from "@/lib/content"
import { href, pages } from "@/lib/nav"

type Props = { params: Promise<{ section: string; slug: string[] }> }

export const dynamicParams = false

export function generateStaticParams() {
  return pages.map((p) => ({ section: p.section, slug: [p.slug] }))
}

async function pageFor(params: Props["params"]) {
  const { section, slug } = await params
  return getPage(section, slug.join("/"))
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const page = await pageFor(params)
  return { title: page.title, description: page.description, alternates: { canonical: href(page.section, page.slug) } }
}

function Pagination({ prev, next }: { prev?: { label: string; section: string; slug: string }; next?: typeof prev }) {
  const item = (link: typeof prev, direction: "prev" | "next") =>
    link ? (
      <LinkBox borderWidth="1px" rounded="lg" p="4" textAlign={direction === "next" ? "end" : "start"} _hover={{ borderColor: "teal.solid" }}>
        <HStack justify={direction === "next" ? "flex-end" : "flex-start"} gap="1" textStyle="xs" color="fg.muted">
          {direction === "prev" && <LuChevronLeft />}
          {direction === "prev" ? "Previous" : "Next"}
          {direction === "next" && <LuChevronRight />}
        </HStack>
        <LinkOverlay asChild display="block" mt="1" fontWeight="medium" textStyle="sm">
          <NextLink href={href(link.section, link.slug)}>{link.label}</NextLink>
        </LinkOverlay>
      </LinkBox>
    ) : (
      <Box hideBelow="sm" />
    )
  return (
    <SimpleGrid as="nav" aria-label="Pagination" columns={{ base: 1, sm: 2 }} gap="4" mt="12">
      {item(prev, "prev")}
      {item(next, "next")}
    </SimpleGrid>
  )
}

export default async function DocPage({ params }: Props) {
  const page = await pageFor(params)
  const { Content } = page
  return (
    <Flex gap="10" px={{ base: "5", md: "10" }} justify="center">
      <Box as="main" flex="1" minW="0" maxW="3xl" py={{ base: "8", md: "10" }}>
        <Flex as="header" gap="6" align="flex-start" justify="space-between" mb="8">
          <Stack gap="2">
            <Text textStyle="sm" fontWeight="medium" color="teal.fg">
              {page.group}
            </Text>
            <Heading as="h1" size="3xl">
              {page.title}
            </Heading>
            <Text textStyle="lg" color="fg.muted">
              {page.description}
            </Text>
          </Stack>
          <Box flexShrink="0" hideBelow="md" mt="6">
            <CopyButton value={page.source} label="Copy page as Markdown">
              Copy page
            </CopyButton>
          </Box>
        </Flex>
        <Prose size="lg" maxW="none" css={{ "& :where(h2, h3)": { scrollMarginTop: "5rem" }, "& :where(td)": { verticalAlign: "top" } }}>
          <Content components={mdxComponents} />
        </Prose>
        <Pagination prev={page.prev} next={page.next} />
        <HStack as="footer" justify="space-between" mt="12" pt="6" borderTopWidth="1px" textStyle="sm" color="fg.muted">
          <Text>
            Built with{" "}
            <Link href="https://chakra-ui.com" color="fg" fontWeight="medium">
              Chakra UI
            </Link>
          </Text>
          <IconButton asChild variant="ghost" size="sm" aria-label="desk on GitHub">
            <a href={githubUrl}>
              <LuGithub />
            </a>
          </IconButton>
        </HStack>
      </Box>
      <Box as="aside" hideBelow="xl" w="56" flexShrink="0">
        <Box position="sticky" top="14" pt="10" maxH="calc(100dvh - 3.5rem)" overflowY="auto">
          <TableOfContents headings={page.headings} />
        </Box>
      </Box>
    </Flex>
  )
}
