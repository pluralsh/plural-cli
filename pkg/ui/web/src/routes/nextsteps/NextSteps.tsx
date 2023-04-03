import { Button, Codeline } from '@pluralsh/design-system'
import React, { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import styled from 'styled-components'

import { Close, SetClipboard } from '../../../wailsjs/go/ui/Window'
import { Routes } from '../routes'

const NextSteps = styled(NextStepsUnstyled)(({ theme }) => ({
  height: '100%',
  display: 'flex',
  flexDirection: 'column',
  justifyContent: 'center',
  gap: theme.spacing.medium,

  '.title': {
    ...theme.partials.text.title2,

    display: 'flex',
    flexDirection: 'column',
    alignSelf: 'center',
    alignItems: 'center',
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
  const onCopy = useCallback((text: string) => SetClipboard(text), [])
  const navigate = useNavigate()

  return (
    <div {...props}>
      <span className="title">
        <span>Almost done!</span>
        <span>Follow the next steps to complete your installations.</span>
      </span>

      <span className="codeline">Copy and run below command to build your applications:</span>
      <Codeline
        width="100%"
        onCopyClick={onCopy}
      >plural build
      </Codeline>

      <span className="codeline">Copy and run below command to deploy your applications:</span>
      <Codeline
        width="100%"
        onCopyClick={onCopy}
      >plural deploy --commit "Installed few apps with Plural"
      </Codeline>

      <div className="actions">
        <Button
          secondary
          onClick={() => navigate(Routes.Root)}
        >Install more
        </Button>
        <Button
          onClick={onClose}
        >Close
        </Button>
      </div>
    </div>
  )
}

export default NextSteps
