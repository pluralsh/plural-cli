import React from 'react'
import { DataRouteObject } from 'react-router-dom'

const Installer = React.lazy(() => import('./Installer'))

const route: DataRouteObject = {
  id: 'installer',
  index: true,
  element: <Installer />,
}

export { route as installerRoute }
