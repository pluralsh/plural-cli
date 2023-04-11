import { createContext } from 'react'

import { PluralContext, PluralProject } from '../types/client'

interface WailsContextProps {
  token: string;
  project: PluralProject
  context: PluralContext
}

const WailsContext = createContext<WailsContextProps>({
  project: {},
  context: {},
} as WailsContextProps)

export type { WailsContextProps, PluralContext }
export { WailsContext }
