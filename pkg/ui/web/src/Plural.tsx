import { theme } from '@pluralsh/design-system'
import { ThemeProvider as HonorableThemeProvider } from 'honorable'
import React, { useMemo } from 'react'
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

function Plural(): React.ReactElement {
  const { data: context } = useWailsQuery(Binding.Context)
  const { data: project } = useWailsQuery(Binding.Project)
  const { data: token } = useWailsQuery(Binding.Token)

  const isReady = useMemo(() => !!context && !!project && !!token, [context, project, token])
  const wailsContext = useMemo(() => ({
    context,
    project,
    token,
  } as WailsContextProps), [context, project, token])

  return (
    <HonorableThemeProvider theme={theme}>
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
