import { ApolloProvider } from '@apollo/client'
import { GraphQLToast, LoadingSpinner } from '@pluralsh/design-system'
import React, { Suspense, useContext } from 'react'
import { DataRouteObject, Outlet, RouteObject } from 'react-router-dom'
import styled from 'styled-components'

import { WailsContext } from '../context/wails'
import Header from '../layout/Header'
import { useApolloClient } from '../services/apollo'

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
  const { token } = useContext(WailsContext)
  const { client, error } = useApolloClient(token)

  return (
    <div {...props}>
      <ApolloProvider client={client!}>
        <Header />
        <div className="content"><Outlet /></div>
        {error && (
          <GraphQLToast
            margin="large"
            error={{ graphQLErrors: [...error] }}
            header="Error"
            closeTimeout={10000}
          />
        )}
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
