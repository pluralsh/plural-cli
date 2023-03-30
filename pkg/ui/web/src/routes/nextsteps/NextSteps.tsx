import { Button, Codeline } from '@pluralsh/design-system'
import React, { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import styled from 'styled-components'

import { Close } from '../../../wailsjs/go/ui/Window'
import { Routes } from '../routes'

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
  const onClose = useCallback(Close, [])
  const navigate = useNavigate()

  return (
    <div {...props}>
      <span className="title">Lorem ipsum dolor!</span>

      <span className="codeline">Copy and run below command to build your applications:</span>
      <Codeline width="100%">plural build</Codeline>

      <span className="codeline">Copy and run below command to deploy your applications:</span>
      <Codeline width="100%">plural deploy --commit "Installed few apps with Plural"</Codeline>

      <div className="actions">
        <Button
          secondary
          onClick={onClose}
        >Close
        </Button>
        <Button onClick={() => navigate(Routes.Root)}>Install more</Button>
      </div>
    </div>
  )
}

export default NextSteps
