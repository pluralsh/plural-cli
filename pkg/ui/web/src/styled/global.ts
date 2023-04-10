import { createGlobalStyle } from 'styled-components'

// language=SCSS
const GlobalStyles = createGlobalStyle(({ theme }) => `
  html, body, #root {
    // Layout
    height: 100vh;
    padding: 0;
    margin: 0;
    
    // Fonts
    font-size: 14px;
    font-family: Inter, Helvetica, Arial, "sans-serif";
    line-height: 20px;
    
    // Theming
    background: ${theme.colors['fill-zero']};
    color: ${theme.colors.text};    
  }
`)

export { GlobalStyles }
