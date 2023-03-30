import { ApolloError, useApolloClient, useQuery } from '@apollo/client'
import {
  GraphQLToast,
  Wizard,
  WizardNavigation,
  WizardStepConfig,
  WizardStepper,
} from '@pluralsh/design-system'
import React, {
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { useNavigate } from 'react-router-dom'
import styled from 'styled-components'

import Loader from '../../components/loader/Loader'
import {
  ListRepositoriesDocument,
  ListRepositoriesQueryVariables,
  Provider,
  RootQueryType,
} from '../../graphql/generated/graphql'
import { Routes } from '../routes'

import { buildSteps, install, toDefaultSteps } from './helpers'

const FILTERED_APPS = ['bootstrap', 'ingress-nginx', 'postgres']
const FORCED_APPS = {
  console: 'The Plural Console will allow you to monitor, upgrade, and deploy applications easily from one centralized place.',
}

const Installer = styled(InstallerUnstyled)(() => ({
  height: '100%',
}))

function InstallerUnstyled({ ...props }): React.ReactElement {
  const navigate = useNavigate()
  const client = useApolloClient()
  // const { project: { provider } } = useContext(WailsContext)

  // TODO: Use project provider when finished testing
  const provider = Provider.Aws

  const [stepsLoading, setStepsLoading] = useState(false)
  const [steps, setSteps] = useState<Array<WizardStepConfig>>()
  const [error, setError] = useState<ApolloError | undefined>()
  const [defaultSteps, setDefaultSteps] = useState<Array<WizardStepConfig>>([])

  const { data: connection } = useQuery<Pick<RootQueryType, 'repositories'>, ListRepositoriesQueryVariables>(ListRepositoriesDocument, {
    variables: {
      installed: false,
      provider,
    },
    fetchPolicy: 'network-only',
  })

  const applications = useMemo(() => connection
    ?.repositories
    ?.edges
    ?.map(repo => repo!.node)
    .filter(app => ((!app?.private ?? true)) && !FILTERED_APPS.includes(app!.name)), [connection?.repositories?.edges])

  const onInstall = useCallback((payload: Array<WizardStepConfig>) => {
    setStepsLoading(true)

    install(client, payload)
      .then(() => navigate(Routes.Next))
      .catch(err => setError(err))
      .finally(() => setStepsLoading(false))
  }, [client, navigate])

  const onSelect = useCallback((selectedApplications: Array<WizardStepConfig>) => {
    const build = async () => {
      const steps = await buildSteps(client, provider!, selectedApplications)

      setSteps(steps)
    }

    setStepsLoading(true)
    build().finally(() => setStepsLoading(false))
  }, [client, provider])

  useEffect(() => setDefaultSteps(toDefaultSteps(applications, provider!, { ...FORCED_APPS })), [applications?.length, provider])

  if (!applications || defaultSteps.length === 0) {
    return <Loader />
  }

  return (
    <div {...props}>
      <Wizard
        onSelect={onSelect}
        defaultSteps={defaultSteps}
        dependencySteps={steps}
        limit={5}
        loading={stepsLoading}
      >
        {{
          stepper: <WizardStepper />,
          navigation: <WizardNavigation onInstall={onInstall} />,
        }}
      </Wizard>

      {error && (
        <GraphQLToast
          error={{ graphQLErrors: error.graphQLErrors ? [...error.graphQLErrors] : [{ message: error as any }] }}
          header="Error"
          onClose={() => setError(undefined)}
          margin="medium"
          closeTimeout={20000}
        />
      )}
    </div>
  )
}

export default Installer
