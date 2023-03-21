import { DataRouteObject } from 'react-router-dom'

import { Installer } from './Installer'

const route: DataRouteObject = {
  id: 'installer',
  path: '/',
  element: <Installer />,
}

export { route as installerRoute }
