import { LoopingLogo } from '@pluralsh/design-system'
import React from 'react'
import styled from 'styled-components'

const Loader = styled(LoaderUnstyled)(() => ({
  height: '100%',
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
}))

function LoaderUnstyled({ ...props }): React.ReactElement {
  return (
    <div {...props}>
      <LoopingLogo />
    </div>
  )
}

export default Loader
