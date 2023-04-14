import { styledTheme } from '@pluralsh/design-system'

const theme = {
  ...styledTheme,
  partials: {
    ...styledTheme.partials,
    draggable: { '--wails-draggable': 'drag' },
  },
}

export { theme as styledTheme }
