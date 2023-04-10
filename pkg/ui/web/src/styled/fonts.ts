import { createGlobalStyle } from 'styled-components'

// language=SCSS
const FontStyles = createGlobalStyle(() => `
  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-regular.otf") format("opentype");
    font-weight: 400;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-medium.woff") format("woff");
    font-weight: 450;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-medium.woff") format("woff");
    font-weight: 500;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-bold.woff") format("woff");
    font-weight: 500;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-regular-italic.woff") format("woff");
    font-weight: 400;
    font-style: italic;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-medium-italic.woff") format("woff");
    font-weight: 500;
    font-style: italic;
  }

  @font-face {
    font-family: 'Monument';
    src: url("/fonts/monument-bold-italic.woff") format("woff");
    font-weight: 600;
    font-style: italic;
  }

  @font-face {
    font-family: 'Monument Semi-Mono';
    src: url("/fonts/ABCMonumentGroteskSemi-Mono-Regular.woff") format("woff");
    font-weight: 400;
  }

  @font-face {
    font-family: 'Monument Semi-Mono';
    src: url("/fonts/ABCMonumentGroteskSemi-Mono-Medium.woff") format("woff");
    font-weight: 500;
  }

  @font-face {
    font-family: 'Monument Semi-Mono';
    src: url("/fonts/ABCMonumentGroteskSemi-Mono-Heavy.woff") format("woff");
    font-weight: 600;
  }

  @font-face {
    font-family: 'Monument Mono';
    src: url("/fonts/ABCMonumentGroteskMono-Regular.woff") format("woff");
    font-weight: 400;
  }

  @font-face {
    font-family: 'Monument Mono';
    src: url("/fonts/ABCMonumentGroteskMono-Medium.woff") format("woff");
    font-weight: 500;
  }

  @font-face {
    font-family: 'Monument Mono';
    src: url("/fonts/ABCMonumentGroteskMono-Heavy.woff") format("woff");
    font-weight: 600;
  }
`)

export { FontStyles }
