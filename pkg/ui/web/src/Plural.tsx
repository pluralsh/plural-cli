import { theme } from '@pluralsh/design-system'
import { Grommet, ThemeType } from 'grommet'
import { HonorableTheme, ThemeProvider as HonorableThemeProvider } from 'honorable'
import React, { useMemo } from 'react'
import { RouterProvider } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'

import Loader from './components/loader/Loader'
import { WailsContext, WailsContextProps } from './context/wails'
import { Provider } from './graphql/generated/graphql'
import { useWailsQuery } from './hooks/useWails'
import { router } from './routes/router'
import { Binding } from './services/wails'
import { FontStyles } from './styled/fonts'
import { GlobalStyles } from './styled/global'
import { ScrollbarStyles } from './styled/scrollbar'
import { theme as styledTheme } from './styled/theme'
import { PluralProject } from './types/client'

function Plural(): React.ReactElement {
  const { data: context } = useWailsQuery(Binding.Context)
  const { data: project = {} as PluralProject } = useWailsQuery<PluralProject>(Binding.Project)
  const { data: provider } = useWailsQuery<Provider>(Binding.Provider)
  const { data: token } = useWailsQuery(Binding.Token)

  const isReady = useMemo(() => !!context && !!project && !!token, [context, project, token])
  const wailsContext = useMemo(() => ({
    context,
    project: { ...project, provider },
    token,
  } as WailsContextProps), [context, project, provider, token])

  return (
    <HonorableThemeProvider theme={theme as HonorableTheme}>
      <ThemeProvider theme={styledTheme}>
        <Grommet
          full
          theme={styledTheme as any as ThemeType}
          themeMode="dark"
        >
          <WailsContext.Provider value={wailsContext}>
            <GlobalStyles />
            <FontStyles />
            <ScrollbarStyles />
            {isReady && <RouterProvider router={router} />}
            {!isReady && <Loader />}
          </WailsContext.Provider>
        </Grommet>
      </ThemeProvider>
    </HonorableThemeProvider>
  )
}

export default Plural
