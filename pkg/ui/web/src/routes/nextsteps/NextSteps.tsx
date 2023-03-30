import { Button, Codeline } from '@pluralsh/design-system'
import React from 'react'
import styled from 'styled-components'

const NextSteps = styled(NextStepsUnstyled)(({ theme }) => ({
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'center',
  gap: theme.spacing.medium,

  '.title': {
    ...theme.partials.text.title1,

    alignSelf: 'center',
    paddingBottom: theme.spacing.xxxlarge,
  },

  '.description': {
    ...theme.partials.text.body1,

    alignSelf: 'center',
    // paddingBottom: theme.spacing.xxxlarge,
  },

  '.codeline': {
    ...theme.partials.text.body2,

    paddingTop: theme.spacing.xxlarge,
  },

  '.actions': {
    display: 'flex',
    justifyContent: 'space-between',
    paddingTop: theme.spacing.xlarge,
  },
}))

function NextStepsUnstyled({ ...props }): React.ReactElement {
  return (
    <div {...props}>
      <span className="title">Lorem ipsum dolor!</span>
      {/* <span className="description">You can now build and deploy your applications.</span> */}

      <span className="codeline">Copy and run below command to build your applications:</span>
      <Codeline width="100%">plural build</Codeline>

      <span className="codeline">Copy and run below command to deploy your applications:</span>
      <Codeline width="100%">plural deploy --commit "Installed few apps with Plural"</Codeline>

      <div className="actions">
        <Button secondary>Close</Button>
        <Button>Install more</Button>
      </div>
    </div>
  )
}

export default NextSteps
