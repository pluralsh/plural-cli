import { createHashRouter } from 'react-router-dom'

import { installerRoute } from './installer/route'
import { nextStepsRoute } from './nextsteps/route'

import { createRootRoute } from './root'

const router = createHashRouter([
  createRootRoute([
    installerRoute,
    nextStepsRoute,
  ]),
])

export { router }
