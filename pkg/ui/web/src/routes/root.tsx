import React from 'react'
import { DataRouteObject, Outlet, RouteObject } from 'react-router-dom'
import styled from 'styled-components'

import { Header } from '../layout/Header'

const Root = styled(RootUnstyled)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column' as const,
  height: '100%',

  '.content': {
    padding: theme.spacing.xxlarge,
    flexGrow: 1,
  },
}))

function RootUnstyled({ ...props }): React.ReactElement {
  return (
    <div {...props}>
      <Header />
      <div className="content"><Outlet /></div>
    </div>
  )
}

const route = (children: Array<DataRouteObject>): RouteObject => ({
  path: '/',
  element: <Root />,
  children,
})

export { route as createRootRoute }
