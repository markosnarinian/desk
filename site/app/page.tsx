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
import { CodeBlock, Tab, Tabs } from "@/components/mdx"
import { ColorModeButton } from "@/components/ui/color-mode"

export const metadata = { alternates: { canonical: "/" } }

const installationMethods = [
  { title: "Homebrew", command: "brew install markosnarinian/tap/desk && desk init", text: "macOS and Linux. Homebrew handles updates." },
  { title: "One-liner", command: "curl -fsSL https://raw.githubusercontent.com/markosnarinian/desk/main/site/public/install.sh | sh", text: "No Go or sudo. Verifies the download and sets up your shell." },
  { title: "Go", command: 'GOBIN="$HOME/.local/bin" go install github.com/markosnarinian/desk/cmd/desk@latest\n"$HOME/.local/bin/desk" init', text: "Go 1.24 or newer. Installs the latest tagged version." },
]

const features = [
  { icon: "scratch", title: "Build a playground", text: "Try a library, generate a throwaway scaffold, or prototype an idea. Start with desk new, then use your usual tools." },
  { icon: "git", title: "Explore a repository", text: "Choose an existing workspace or a fresh clone with desk new owner/repo. Read the code, debug, or try a change without naming a task first." },
  { icon: "terminal", title: "Test cloud work locally", text: "Use desk amp to copy an Amp task or follow it live in a fresh local clone. Uncommitted and unpushed work comes along—no push required." },
]

const steps = [
  { title: "Create", command: "desk new", text: "Start an experiment. Add owner/repo to work on a clone instead." },
  { title: "Find", command: "desk list", text: "See a tree of workspaces grouped by source repository, with a separate scratch group." },
  { title: "Return", command: "desk switch misty", text: "Go to a workspace by name or a unique prefix." },
  { title: "Clear away", command: "desk cleanup", text: "Preview safe workspaces inactive for 72 hours, then confirm. Mark scratch work disposable with desk finish first." },
]

const shellFeatures = [
  "The workspace name and your path inside it, in front of your prompt",
  "desk new and desk switch change your directory",
  "Tab completion for commands and workspace names",
  "Works with your existing prompt and theme",
  "desk_prompt_info for placing it in a custom prompt",
]

const session = `➜  ~ desk switch misty
(misty-ember) ➜  misty-ember cd src/api
(misty-ember/src/api) ➜  api desk switch
(misty-ember) ➜  misty-ember desk info --path
/home/you/.desk/workspaces/misty-ember`

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
              Early and minimal · macOS and Linux
            </Badge>
            <Heading as="h1" size={{ base: "4xl", md: "6xl" }} letterSpacing="tight">
              A workspace for the local part.
            </Heading>
            <Text textStyle={{ base: "lg", md: "xl" }} color="fg.muted" maxW="2xl">
              As agents take more work to the cloud, local work becomes more fragmented.
              desk gives temporary clones, playgrounds, and prototypes a workspace you can create, return to, and remove.
            </Text>
            <HStack gap="3" flexWrap="wrap" justify="center">
              <Button asChild size="lg" colorPalette="teal">
                <NextLink href="#install">
                  Install desk <LuArrowRight />
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
              <CopyButton value="desk new" label="Copy command" />
            </HStack>
            <Box as="pre" p="5" fontFamily="mono" textStyle="sm" lineHeight="1.8" whiteSpace="pre-wrap" overflowWrap="anywhere" color="fg.muted">
              <Text as="span" color="fg">
                $ desk new
              </Text>
              {"\ndesk: created misty-ember\n"}
              <Text as="span" color="fg">
                (misty-ember) $ printf 'console.log(2 ** 10)\n' &gt; try.js
              </Text>
              {"\n"}
              <Text as="span" color="fg">
                (misty-ember) $ node try.js
              </Text>
              {"\n1024\n"}
              <Text as="span" color="fg">
                (misty-ember) $ desk rm
              </Text>
              {"\nDelete misty-ember (has files)? [y/N] y\nMoved to garbage for 72 hours."}
            </Box>
          </Card.Root>
        </Container>

        <Box bg="bg.subtle" borderYWidth="1px" py={{ base: "16", md: "24" }}>
          <Container maxW="6xl">
            <SectionHeader
              tag="Why desk"
              title="Local work is changing"
              text="Agents get cloud machines they can spin up and tear down. Humans need a local equivalent for exploration and experimentation: a clone to inspect, a throwaway scaffold, a prototype to run."
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
            title="Create it. Use it. Clear it away."
            text="Plain directories, your usual tools. Cleanup checks for local-only work before asking. Removed workspaces stay in garbage for 72 hours; expired garbage is collected when you run desk, not by a daemon."
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
                  Shell integration
                </Text>
                <Heading as="h2" size={{ base: "2xl", md: "3xl" }}>
                  Your prompt shows where you are
                </Heading>
                <Text color="fg.muted" textStyle="lg">
                  Run desk init once. It sets up directory switching, completion, and a workspace prompt
                  for zsh (including Oh My Zsh), bash, or fish. Then open a new terminal.
                </Text>
                <List.Root gap="2.5" variant="plain">
                  {shellFeatures.map((item) => (
                    <List.Item key={item} alignItems="flex-start">
                      <List.Indicator asChild color="teal.fg" mt="1">
                        <LuCircleCheck />
                      </List.Indicator>
                      {item}
                    </List.Item>
                  ))}
                </List.Root>
                <Link asChild color="teal.fg" fontWeight="medium" alignSelf="flex-start">
                  <NextLink href="/docs/shell-integration/">
                    Set up shell integration <LuArrowRight />
                  </NextLink>
                </Link>
              </Stack>
              <Card.Root variant="outline" overflow="hidden">
                <Box px="4" py="2.5" borderBottomWidth="1px" textStyle="xs" color="fg.muted">
                  Oh My Zsh, robbyrussell theme
                </Box>
                <Box as="pre" p="5" fontFamily="mono" textStyle="xs" lineHeight="1.9" whiteSpace="pre-wrap" overflowWrap="anywhere">
                  {session}
                </Box>
                <Box px="4" py="3" borderTopWidth="1px" textStyle="sm" color="fg.muted">
                  Run <Code>desk init</Code> to detect your shell, including Oh My Zsh setups.
                </Box>
              </Card.Root>
            </SimpleGrid>
          </Container>
        </Box>

        <Container id="install" maxW="6xl" py={{ base: "16", md: "24" }} scrollMarginTop="16">
          <Card.Root variant="subtle" colorPalette="teal" bg="teal.subtle" textAlign="center">
            <Card.Body alignItems="center" gap="4" py={{ base: "10", md: "14" }}>
              <Heading as="h2" size={{ base: "2xl", md: "3xl" }}>
                Install desk your way
              </Heading>
              <Text color="fg.muted" maxW="xl">
                Choose an install method, then open a new terminal.
                Zsh (including Oh My Zsh), bash, and fish.
              </Text>
              <Text textStyle="sm" color="fg.muted" maxW="xl">
                First release pending. For now, build from source using the installation guide.
              </Text>
              <Box w="full" maxW="3xl" minW="0" textAlign="start">
                <Tabs>
                  {installationMethods.map((method) => (
                    <Tab key={method.title} title={method.title}>
                      <Text color="fg.muted" textStyle="sm">{method.text}</Text>
                      <CodeBlock data-language="sh" data-code={method.command} style={{ whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}>
                        <code>{method.command}</code>
                      </CodeBlock>
                    </Tab>
                  ))}
                </Tabs>
              </Box>
              <HStack gap="3" mt="2" flexWrap="wrap" justify="center">
                <Button asChild colorPalette="teal">
                  <NextLink href="/docs/installation/">Installation guide</NextLink>
                </Button>
                <Button asChild variant="outline" bg="bg">
                  <NextLink href="/docs/quickstart/">Try your first workspace</NextLink>
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
