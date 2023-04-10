// import original module declarations
import 'styled-components'
import { theme } from '../styled/theme'

import { DEFAULT_THEME } from '../theme'

type StyledTheme = typeof theme & typeof DEFAULT_THEME

// and extend them!
declare module 'styled-components' {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface
  export interface DefaultTheme extends StyledTheme {}
}
