import {
  Alert,
  Badge,
  Box,
  Center,
  Code,
  Flex,
  HStack,
  Link,
  Tabs as ChakraTabs,
  Text,
} from "@chakra-ui/react"
import type { MDXComponents } from "mdx/types"
import NextLink from "next/link"
import { Children, cloneElement, isValidElement, type ComponentProps, type ReactElement, type ReactNode } from "react"
import { CopyButton } from "./copy-button"

const status = { info: "info", tip: "success", success: "success", warning: "warning", danger: "error" } as const

export function Callout({ type = "info", title, children }: { type?: keyof typeof status; title?: string; children: ReactNode }) {
  return (
    <Alert.Root status={status[type]} variant="subtle" my="6" alignItems="flex-start">
      <Alert.Indicator mt="0.5" />
      <Alert.Content css={{ "& p": { my: 0 }, "& p + p": { mt: 2 } }}>
        {title && <Alert.Title fontWeight="semibold">{title}</Alert.Title>}
        <Alert.Description>{children}</Alert.Description>
      </Alert.Content>
    </Alert.Root>
  )
}

type StepProps = { title: string; number?: number; children?: ReactNode }

export function Steps({ children }: { children: ReactNode }) {
  const steps = Children.toArray(children).filter(isValidElement) as ReactElement<StepProps>[]
  return <Box my="8">{steps.map((step, i) => cloneElement(step, { key: i, number: i + 1 }))}</Box>
}

export function Step({ title, number, children }: StepProps) {
  return (
    <Flex gap="4" pb="8" _last={{ pb: "0", "& [data-line]": { display: "none" } }}>
      <Flex direction="column" align="center" flexShrink="0">
        <Center boxSize="7" rounded="md" bg="teal.solid" color="teal.contrast" textStyle="sm" fontWeight="semibold">
          {number}
        </Center>
        <Box data-line flex="1" w="1px" bg="border" mt="2" />
      </Flex>
      <Box flex="1" minW="0" pt="0.5" css={{ "& > div > :first-child": { mt: 2 }, "& > div > :last-child": { mb: 0 } }}>
        <Text fontWeight="semibold" color="fg">
          {title}
        </Text>
        <div>{children}</div>
      </Box>
    </Flex>
  )
}

export function ParamField({
  name,
  type,
  required,
  default: defaultValue,
  children,
}: {
  name: string
  type?: string
  required?: boolean
  default?: string
  children?: ReactNode
}) {
  return (
    <Box py="4" borderBottomWidth="1px" _first={{ borderTopWidth: "1px" }}>
      <HStack className="not-prose" gap="2" flexWrap="wrap">
        <Code variant="plain" colorPalette="teal" fontWeight="semibold" px="0">
          {name}
        </Code>
        {type && (
          <Badge variant="surface" size="sm">
            {type}
          </Badge>
        )}
        {required && (
          <Badge colorPalette="red" size="sm">
            required
          </Badge>
        )}
        {defaultValue && (
          <Text textStyle="xs" color="fg.muted">
            default: <Code size="sm">{defaultValue}</Code>
          </Text>
        )}
      </HStack>
      <Box mt="2" css={{ "& p": { my: 0 }, "& p + p": { mt: 2 } }}>
        {children}
      </Box>
    </Box>
  )
}

type TabProps = { title: string; children?: ReactNode }

export function Tabs({ children }: { children: ReactNode }) {
  const tabs = Children.toArray(children).filter(isValidElement) as ReactElement<TabProps>[]
  return (
    <ChakraTabs.Root defaultValue="0" variant="line" colorPalette="teal" my="6" size="sm">
      <ChakraTabs.List className="not-prose">
        {tabs.map((tab, i) => (
          <ChakraTabs.Trigger key={i} value={String(i)}>
            {tab.props.title}
          </ChakraTabs.Trigger>
        ))}
      </ChakraTabs.List>
      {tabs.map((tab, i) => (
        <ChakraTabs.Content key={i} value={String(i)} css={{ "& > :first-child": { mt: 0 } }}>
          {tab.props.children}
        </ChakraTabs.Content>
      ))}
    </ChakraTabs.Root>
  )
}

export function Tab({ children }: TabProps) {
  return <>{children}</>
}

const languageLabels: Record<string, string> = { sh: "Terminal", bash: "Terminal", text: "Output" }

export function CodeBlock(props: ComponentProps<"pre"> & { "data-language"?: string; "data-code"?: string; "data-title"?: string }) {
  const { children, className, style, "data-language": language = "text", "data-code": code = "", "data-title": title } = props
  return (
    <Box className="not-prose" my="6" borderWidth="1px" rounded="lg" overflow="hidden" bg="bg.subtle">
      <HStack justify="space-between" ps="4" pe="2" h="9" borderBottomWidth="1px">
        <Text textStyle="xs" color="fg.muted">
          {title ?? languageLabels[language] ?? language}
        </Text>
        <CopyButton value={code} label="Copy code" />
      </HStack>
      <Box
        as="pre"
        className={className}
        style={style}
        px="4"
        py="3"
        overflowX="auto"
        fontFamily="mono"
        textStyle="sm"
        lineHeight="1.7"
        bg="transparent!"
      >
        {children}
      </Box>
    </Box>
  )
}

function heading(Tag: "h2" | "h3") {
  return function Heading({ id, children }: ComponentProps<"h2">) {
    return (
      <Tag id={id} className="group">
        {children}
        <Link
          href={`#${id}`}
          aria-label="Link to this section"
          ms="2"
          color="fg.subtle"
          textDecoration="none!"
          opacity="0"
          _groupHover={{ opacity: 1 }}
          _focusVisible={{ opacity: 1 }}
        >
          #
        </Link>
      </Tag>
    )
  }
}

export const mdxComponents: MDXComponents = {
  Callout,
  Steps,
  Step,
  ParamField,
  Tabs,
  Tab,
  pre: CodeBlock,
  h2: heading("h2"),
  h3: heading("h3"),
  a: ({ href = "", children }) =>
    href.startsWith("/") ? (
      <Link asChild color="teal.fg!">
        <NextLink href={href}>{children}</NextLink>
      </Link>
    ) : (
      <Link href={href} color="teal.fg!" target="_blank" rel="noreferrer">
        {children}
      </Link>
    ),
  table: ({ children }) => (
    <Box overflowX="auto" my="6">
      <table style={{ margin: 0 }}>{children}</table>
    </Box>
  ),
}
