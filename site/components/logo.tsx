import { Center, HStack, Text } from "@chakra-ui/react"
import NextLink from "next/link"
import { LuTerminal } from "react-icons/lu"

export function Logo() {
  return (
    <HStack asChild gap="2">
      <NextLink href="/" aria-label="desk home">
        <Center boxSize="8" rounded="lg" bg="teal.solid" color="teal.contrast">
          <LuTerminal aria-hidden />
        </Center>
        <Text as="span" textStyle="lg" fontWeight="semibold" color="fg">
          desk
        </Text>
      </NextLink>
    </HStack>
  )
}
