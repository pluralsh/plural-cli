import React from 'react'
import { createRoot } from 'react-dom/client'

import Plural from './Plural'

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(<React.StrictMode><Plural /></React.StrictMode>)
