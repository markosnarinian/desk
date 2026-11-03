import { Button, Heading, HStack, Stack, Text } from "@chakra-ui/react"
import NextLink from "next/link"
import { DocsShell } from "@/components/docs-shell"

export const metadata = { title: "Page not found" }

export default function NotFound() {
  return (
    <DocsShell>
      <Stack as="main" align="center" textAlign="center" gap="4" px="6" py="32">
        <Text textStyle="sm" fontWeight="medium" color="teal.fg">
          404
        </Text>
        <Heading as="h1" size="3xl">
          Page not found
        </Heading>
        <Text color="fg.muted">The page you are looking for does not exist or has moved.</Text>
        <HStack mt="4" gap="3">
          <Button asChild colorPalette="teal">
            <NextLink href="/docs/introduction/">Documentation</NextLink>
          </Button>
          <Button asChild variant="outline">
            <NextLink href="/commands/overview/">Commands</NextLink>
          </Button>
        </HStack>
      </Stack>
    </DocsShell>
  )
}
