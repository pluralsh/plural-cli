import { createBrowserRouter } from 'react-router-dom'

import { installerRoute } from '../installer/route'

import { createRootRoute } from './root'

const router = createBrowserRouter([
  createRootRoute([
    installerRoute,
  ]),
])

export { router }
