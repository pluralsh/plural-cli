import { Button, CloseIcon } from '@pluralsh/design-system'
import PluralLogoFull from '@pluralsh/design-system/dist/components/icons/logo/PluralLogoFull'
import React, { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import styled from 'styled-components'

import { Close } from '../../wailsjs/go/ui/Window'
import { Routes } from '../routes/routes'

const Header = styled(HeaderUnstyled)(({ theme }) => ({
  // Make window draggable via header
  ...theme.partials.draggable,

  // Layout
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',

  // Spacing
  padding: `${theme.spacing.medium}px ${theme.spacing.large}px`,

  // Theming
  background: theme.colors['fill-one'],
  borderBottom: theme.borders.default,
}))

function HeaderUnstyled({ ...props }): React.ReactElement {
  const onClose = useCallback(Close, [])
  const navigate = useNavigate()

  return (
    <div {...props}>
      <PluralLogoFull
        color="white"
        height={24}
      />
      <Button
        secondary
        small
        width={24}
        height={24}
        onClick={() => navigate(Routes.Next)}
      ><CloseIcon />
      </Button>
      <Button
        secondary
        small
        width={24}
        height={24}
        onClick={onClose}
      ><CloseIcon />
      </Button>
    </div>
  )
}

export default Header
