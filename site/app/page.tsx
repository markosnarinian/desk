import {
  Badge,
  Box,
  Button,
  Card,
  Code,
  Container,
  Flex,
  Heading,
  HStack,
  Icon,
  Link,
  List,
  SimpleGrid,
  Stack,
  Text,
} from "@chakra-ui/react"
import NextLink from "next/link"
import { LuArrowRight, LuCircleCheck, LuGithub } from "react-icons/lu"
import { CopyButton } from "@/components/copy-button"
import { githubUrl } from "@/components/docs-shell"
import { iconFor } from "@/components/icons"
import { Logo } from "@/components/logo"
import { ColorModeButton } from "@/components/ui/color-mode"

export const metadata = { alternates: { canonical: "/" } }

const features = [
  { icon: "scratch", title: "Scratch workspaces", text: "Get an empty directory for an idea in one command. No naming, no deciding where it lives." },
  { icon: "git", title: "Repository workspaces", text: "Clone owner/repo, any Git URL, or a local path into its own managed checkout, and resume it later." },
  { icon: "terminal", title: "Your shell, your tools", text: "desk opens your $SHELL inside the workspace. Exit to go back. There is no daemon and no shell hook." },
  { icon: "list", title: "One list", text: "desk list shows every workspace, its state, its path, and the reason it would be kept." },
  { icon: "archive", title: "Quarantine, not delete", text: "Cleanup moves workspaces to quarantine first. Restore anything with one command." },
  { icon: "shield", title: "Proven safe or kept", text: "Uncommitted, untracked, ignored, or unpushed work keeps a workspace around." },
]

const steps = [
  { title: "Start", command: "desk scratch experiment", text: "desk creates a workspace and opens a subshell in it." },
  { title: "Work and leave", command: "exit", text: "Leaving never deletes anything. Come back with desk enter." },
  { title: "Finish deliberately", command: "desk done ID", text: "Marking a workspace done makes it a cleanup candidate." },
  { title: "Clean up safely", command: "desk gc", text: "Preview first, then quarantine only what passes every check." },
]

const retained = [
  "Pinned workspaces and workspaces with an active session",
  "Scratch workspaces that contain any file, including dotfiles",
  "Tracked changes, untracked files, and ignored files",
  "Branches, tags, commits, stashes, and reflogs that origin does not have",
  "Submodules, Git LFS and other content filters, extra worktrees",
  "Anything desk cannot check, such as an unreachable origin",
]

const session = `$ desk gc
bb5ff6487d99 retained: not marked done
702378b08f9e would quarantine (requires --apply --quiescent)
bf22fab17609 retained: scratch contains files`

function Navbar() {
  return (
    <Box as="header" borderBottomWidth="1px" position="sticky" top="0" zIndex="sticky" bg="bg">
      <Container maxW="6xl">
        <HStack h="14" justify="space-between">
          <Logo />
          <HStack gap="1">
            <Button asChild variant="ghost" size="sm" hideBelow="sm">
              <NextLink href="/docs/introduction/">Documentation</NextLink>
            </Button>
            <Button asChild variant="ghost" size="sm" hideBelow="sm">
              <NextLink href="/commands/overview/">Commands</NextLink>
            </Button>
            <Button asChild variant="ghost" size="sm" aria-label="GitHub">
              <a href={githubUrl}>
                <LuGithub />
              </a>
            </Button>
            <ColorModeButton />
            <Button asChild size="sm" colorPalette="teal" ms="2">
              <NextLink href="/docs/quickstart/">Get started</NextLink>
            </Button>
          </HStack>
        </HStack>
      </Container>
    </Box>
  )
}

function SectionHeader({ tag, title, text }: { tag: string; title: string; text: string }) {
  return (
    <Stack gap="3" textAlign="center" align="center" maxW="2xl" mx="auto">
      <Text textStyle="sm" fontWeight="semibold" color="teal.fg">
        {tag}
      </Text>
      <Heading as="h2" size={{ base: "2xl", md: "3xl" }}>
        {title}
      </Heading>
      <Text color="fg.muted" textStyle="lg">
        {text}
      </Text>
    </Stack>
  )
}

export default function Home() {
  return (
    <Box minH="100dvh">
      <Navbar />
      <main>
        <Container maxW="6xl" py={{ base: "16", md: "24" }}>
          <Stack gap="6" align="center" textAlign="center" maxW="3xl" mx="auto">
            <Badge colorPalette="teal" variant="surface" size="lg" rounded="full" px="3">
              Initial release · macOS and Linux
            </Badge>
            <Heading as="h1" size={{ base: "4xl", md: "6xl" }} letterSpacing="tight">
              A place for the work in between
            </Heading>
            <Text textStyle={{ base: "lg", md: "xl" }} color="fg.muted" maxW="2xl">
              desk gives quick experiments and one-off clones their own managed workspaces, then cleans up only what it
              can prove is safe to lose.
            </Text>
            <HStack gap="3" flexWrap="wrap" justify="center">
              <Button asChild size="lg" colorPalette="teal">
                <NextLink href="/docs/quickstart/">
                  Get started <LuArrowRight />
                </NextLink>
              </Button>
              <Button asChild size="lg" variant="outline">
                <NextLink href="/commands/overview/">Browse commands</NextLink>
              </Button>
            </HStack>
          </Stack>

          <Card.Root variant="outline" maxW="2xl" mx="auto" mt="14" overflow="hidden">
            <HStack justify="space-between" ps="4" pe="2" h="10" borderBottomWidth="1px" bg="bg.subtle">
              <HStack gap="1.5" aria-hidden>
                {["red.400", "yellow.400", "green.400"].map((c) => (
                  <Box key={c} boxSize="2.5" rounded="full" bg={c} />
                ))}
              </HStack>
              <Text textStyle="xs" color="fg.muted">
                Terminal
              </Text>
              <CopyButton value="desk scratch experiment" label="Copy command" />
            </HStack>
            <Box as="pre" p="5" fontFamily="mono" textStyle="sm" lineHeight="1.8" overflowX="auto" color="fg.muted">
              <Text as="span" color="fg">
                $ desk scratch experiment
              </Text>
              {"\n7e4a9b23c810\ndesk: entering experiment (7e4a9b23c810); exit to return; nothing is deleted\n"}
              <Text as="span" color="fg">
                $ exit
              </Text>
              {"\n"}
              <Text as="span" color="fg">
                $ desk list
              </Text>
              {"\nID            NAME          STATE  RETENTION                  PATH\n7e4a9b23c810  \"experiment\"  ready  eligible with --quiescent  ~/.desk/workspaces/7e4a9b23c810"}
            </Box>
          </Card.Root>
        </Container>

        <Box bg="bg.subtle" borderYWidth="1px" py={{ base: "16", md: "24" }}>
          <Container maxW="6xl">
            <SectionHeader
              tag="Why desk"
              title="Stop collecting directories you are afraid to delete"
              text="Throwaway work tends to stick around: test-2, tmp-fix, a third clone of the same repo. desk keeps it in one place and keeps track of what is still worth keeping."
            />
            <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap="6" mt="14">
              {features.map((f) => {
                const FeatureIcon = iconFor(f.icon)
                return (
                  <Card.Root key={f.title} variant="outline">
                    <Card.Body gap="3">
                      <Flex boxSize="10" rounded="lg" bg="teal.subtle" color="teal.fg" align="center" justify="center">
                        <Icon boxSize="5">
                          <FeatureIcon />
                        </Icon>
                      </Flex>
                      <Card.Title>{f.title}</Card.Title>
                      <Card.Description>{f.text}</Card.Description>
                    </Card.Body>
                  </Card.Root>
                )
              })}
            </SimpleGrid>
          </Container>
        </Box>

        <Container maxW="6xl" py={{ base: "16", md: "24" }}>
          <SectionHeader
            tag="How it works"
            title="Four commands, from idea to cleanup"
            text="Creating a workspace and getting rid of it are separate steps. Nothing happens on a timer."
          />
          <SimpleGrid columns={{ base: 1, md: 2, lg: 4 }} gap="8" mt="14">
            {steps.map((s, i) => (
              <Stack key={s.title} gap="3">
                <Flex boxSize="8" rounded="md" bg="teal.solid" color="teal.contrast" align="center" justify="center" fontWeight="semibold">
                  {i + 1}
                </Flex>
                <Heading as="h3" size="md">
                  {s.title}
                </Heading>
                <Code size="md" alignSelf="flex-start">
                  {s.command}
                </Code>
                <Text color="fg.muted" textStyle="sm">
                  {s.text}
                </Text>
              </Stack>
            ))}
          </SimpleGrid>
        </Container>

        <Box bg="bg.subtle" borderYWidth="1px" py={{ base: "16", md: "24" }}>
          <Container maxW="6xl">
            <SimpleGrid columns={{ base: 1, lg: 2 }} gap={{ base: "10", lg: "16" }} alignItems="center">
              <Stack gap="5">
                <Text textStyle="sm" fontWeight="semibold" color="teal.fg">
                  Safety
                </Text>
                <Heading as="h2" size={{ base: "2xl", md: "3xl" }}>
                  A clean git status does not mean it is safe to delete
                </Heading>
                <Text color="fg.muted" textStyle="lg">
                  Before anything leaves your workspaces, desk asks origin what it actually has. If desk cannot prove the
                  work exists somewhere else, it keeps the workspace and tells you why.
                </Text>
                <List.Root gap="2.5" variant="plain">
                  {retained.map((item) => (
                    <List.Item key={item} alignItems="flex-start">
                      <List.Indicator asChild color="teal.fg" mt="1">
                        <LuCircleCheck />
                      </List.Indicator>
                      {item}
                    </List.Item>
                  ))}
                </List.Root>
                <Link asChild color="teal.fg" fontWeight="medium" alignSelf="flex-start">
                  <NextLink href="/docs/retention-rules/">
                    Read the retention rules <LuArrowRight />
                  </NextLink>
                </Link>
              </Stack>
              <Card.Root variant="outline" overflow="hidden">
                <Box px="4" py="2.5" borderBottomWidth="1px" textStyle="xs" color="fg.muted">
                  Dry run (the default)
                </Box>
                <Box as="pre" p="5" fontFamily="mono" textStyle="xs" lineHeight="1.9" overflowX="auto">
                  {session}
                </Box>
                <Box px="4" py="3" borderTopWidth="1px" textStyle="sm" color="fg.muted">
                  Nothing moves until you run <Code>desk gc --apply --quiescent</Code>, and even then it goes to
                  quarantine, not the trash.
                </Box>
              </Card.Root>
            </SimpleGrid>
          </Container>
        </Box>

        <Container maxW="6xl" py={{ base: "16", md: "24" }}>
          <Card.Root variant="subtle" colorPalette="teal" bg="teal.subtle" textAlign="center">
            <Card.Body alignItems="center" gap="4" py={{ base: "10", md: "14" }}>
              <Heading as="h2" size={{ base: "2xl", md: "3xl" }}>
                Try it without touching your real workspaces
              </Heading>
              <Text color="fg.muted" maxW="xl">
                Point desk at a separate data root while you get a feel for it. Unset the variable to go back.
              </Text>
              <Code size="lg" px="3" py="2">
                export DESK_HOME=&quot;$HOME/.desk-trial&quot;
              </Code>
              <HStack gap="3" mt="2" flexWrap="wrap" justify="center">
                <Button asChild colorPalette="teal">
                  <NextLink href="/docs/installation/">Install desk</NextLink>
                </Button>
                <Button asChild variant="outline" bg="bg">
                  <NextLink href="/docs/introduction/">Read the docs</NextLink>
                </Button>
              </HStack>
            </Card.Body>
          </Card.Root>
        </Container>
      </main>

      <Box as="footer" borderTopWidth="1px">
        <Container maxW="6xl" py="8">
          <Flex justify="space-between" align="center" gap="4" flexWrap="wrap" textStyle="sm" color="fg.muted">
            <Logo />
            <HStack gap="6">
              <Link asChild color="fg.muted">
                <NextLink href="/docs/introduction/">Documentation</NextLink>
              </Link>
              <Link asChild color="fg.muted">
                <NextLink href="/commands/overview/">Commands</NextLink>
              </Link>
              <Link href={githubUrl} color="fg.muted">
                GitHub
              </Link>
            </HStack>
            <Text>
              Built with{" "}
              <Link href="https://chakra-ui.com" color="fg">
                Chakra UI
              </Link>
            </Text>
          </Flex>
        </Container>
      </Box>
    </Box>
  )
}
