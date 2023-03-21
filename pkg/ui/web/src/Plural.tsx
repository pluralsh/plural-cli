import { styledTheme, theme } from '@pluralsh/design-system'
import { ThemeProvider as HonorableThemeProvider } from 'honorable'
import React from 'react'
import { RouterProvider } from 'react-router-dom'
import { ThemeProvider } from 'styled-components'

import { router } from './routes/router'
import { FontStyles } from './styled/fonts'
import { GlobalStyles } from './styled/global'

// function App() {
//   const [resultText, setResultText] = useState('Please enter your name below 👇')
//   const [name, setName] = useState('')
//   const updateName = (e: any) => setName(e.target.value)
//   const updateResultText = (result: string) => setResultText(result)
//
//   function greet() {
//     Greet(name).then(updateResultText)
//   }
//
//   return (
//     <div id="App">
//       {/* <div style={{"--wails-draggable": "drag"}}>drag me</div> */}
//       <img
//         src={logo}
//         id="logo"
//         alt="logo"
//       />
//       <div
//         id="result"
//         className="result"
//       >{resultText}
//       </div>
//       <div
//         id="input"
//         className="input-box"
//       >
//         <input
//           id="name"
//           className="input"
//           onChange={updateName}
//           autoComplete="off"
//           name="input"
//           type="text"
//         />
//         <button
//           type="button"
//           className="btn"
//           onClick={greet}
//         >Greet
//         </button>
//       </div>
//     </div>
//   )
// }

function Plural(): React.ReactElement {
  return (
    <HonorableThemeProvider theme={theme}>
      <ThemeProvider theme={styledTheme}>
        <GlobalStyles />
        <FontStyles />
        <RouterProvider router={router} />
      </ThemeProvider>
    </HonorableThemeProvider>
  )
}

export default Plural
