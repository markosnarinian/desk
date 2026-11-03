"use client"

import { ChakraProvider } from "@chakra-ui/react"
import { system } from "@/theme"
import { ColorModeProvider, type ColorModeProviderProps } from "./color-mode"
import { EmotionRegistry } from "./emotion-registry"

export function Provider(props: ColorModeProviderProps) {
  return (
    <EmotionRegistry>
      <ChakraProvider value={system}>
        <ColorModeProvider {...props} />
      </ChakraProvider>
    </EmotionRegistry>
  )
}
