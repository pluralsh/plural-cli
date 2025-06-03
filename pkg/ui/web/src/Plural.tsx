import { theme } from '@pluralsh/design-system'
import { HonorableTheme, ThemeProvider as HonorableThemeProvider } from 'honorable'
import React, { useCallback, useMemo } from 'react'
import { RouterProvider } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'

import Loader from './components/loader/Loader'
import { WailsContext, WailsContextProps } from './context/wails'
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
  const { data: token } = useWailsQuery(Binding.Token)

  const isReady = useMemo(() => !!context && !!project && !!token, [context, project, token])
  const fixProjectProvider = useCallback((project: PluralProject) => ({ ...project, provider: project.provider?.toUpperCase() }), [])
  const wailsContext = useMemo(() => ({
    context,
    project: fixProjectProvider(project),
    token,
  } as WailsContextProps), [context, fixProjectProvider, project, token])

  return (
    <HonorableThemeProvider theme={theme as HonorableTheme}>
      <ThemeProvider theme={styledTheme}>
        <WailsContext.Provider value={wailsContext}>
          <GlobalStyles />
          <FontStyles />
          <ScrollbarStyles />
          {isReady && <RouterProvider router={router} />}
          {!isReady && <Loader />}
        </WailsContext.Provider>
      </ThemeProvider>
    </HonorableThemeProvider>
  )
}

export default Plural
