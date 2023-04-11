import { createGlobalStyle } from 'styled-components'

const ScrollbarStyles = createGlobalStyle(({ theme }) => theme.partials.scrollBar({ fillLevel: 0 }))

export { ScrollbarStyles }
