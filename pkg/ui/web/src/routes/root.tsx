import { ApolloProvider } from '@apollo/client'
import { LoadingSpinner } from '@pluralsh/design-system'
import React, { Suspense, useMemo } from 'react'
import { DataRouteObject, Outlet, RouteObject } from 'react-router-dom'
import styled from 'styled-components'

import { useWailsQuery } from '../hooks/useWails'

import Header from '../layout/Header'
import { newApolloClient } from '../services/apollo'
import { Binding } from '../services/wails'

import { Routes } from './routes'

const Root = styled(RootUnstyled)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column' as const,
  height: '100%',
  overflow: 'hidden',

  '.content': {
    padding: theme.spacing.xxlarge,
    flexGrow: 1,
    overflowY: 'auto',
  },
}))

function RootUnstyled({ ...props }): React.ReactElement {
  const { data: token, loading } = useWailsQuery<string>(Binding.Token)
  const client = useMemo(() => {
    if (!token) return undefined

    return newApolloClient(token)
  }, [token])

  if (loading || !token) return <LoadingSpinner />

  return (
    <div {...props}>
      <ApolloProvider client={client!}>
        <Header />
        <div className="content"><Outlet /></div>
      </ApolloProvider>
    </div>
  )
}

const route = (children: Array<DataRouteObject>): RouteObject => ({
  path: Routes.Root,
  element: <Suspense fallback={<LoadingSpinner />}><Root /></Suspense>,
  children,
})

export { route as createRootRoute }
