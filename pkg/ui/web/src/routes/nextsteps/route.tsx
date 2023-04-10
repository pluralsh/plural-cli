import React from 'react'
import { DataRouteObject } from 'react-router-dom'

import { Routes } from '../routes'

const NextSteps = React.lazy(() => import('./NextSteps'))

const route: DataRouteObject = {
  id: 'next',
  path: Routes.Next,
  element: <NextSteps />,
}

export { route as nextStepsRoute }
