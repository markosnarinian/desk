import {
  LuArchive,
  LuBookOpen,
  LuCompass,
  LuDownload,
  LuFolderGit2,
  LuHash,
  LuInfo,
  LuLayers,
  LuLifeBuoy,
  LuList,
  LuPin,
  LuPlus,
  LuRotateCcw,
  LuRocket,
  LuSettings,
  LuShieldCheck,
  LuSquareTerminal,
  LuTerminal,
  LuTrash2,
} from "react-icons/lu"

/** Icons referenced by name from navigation config and MDX. */
export const icons = {
  archive: LuArchive,
  book: LuBookOpen,
  compass: LuCompass,
  download: LuDownload,
  git: LuFolderGit2,
  hash: LuHash,
  info: LuInfo,
  layers: LuLayers,
  help: LuLifeBuoy,
  list: LuList,
  pin: LuPin,
  plus: LuPlus,
  restore: LuRotateCcw,
  rocket: LuRocket,
  settings: LuSettings,
  shield: LuShieldCheck,
  scratch: LuSquareTerminal,
  terminal: LuTerminal,
  trash: LuTrash2,
}

export type IconName = keyof typeof icons
export const iconFor = (name: string) => icons[name as IconName] ?? LuHash
