import { ApolloError, useApolloClient, useQuery } from '@apollo/client'
import {
  GraphQLToast,
  LoopingLogo,
  Wizard,
  WizardNavigation,
  WizardStepConfig,
  WizardStepper,
} from '@pluralsh/design-system'
import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { useNavigate } from 'react-router-dom'

import { WailsContext } from '../../context/wails'
import {
  ListRepositoriesDocument,
  ListRepositoriesQueryVariables,
  Provider,
  RootQueryType,
} from '../../graphql/generated/graphql'
import { Routes } from '../routes'

import { buildSteps, toDefaultSteps } from './helpers'

const FILTERED_APPS = ['bootstrap', 'ingress-nginx', 'postgres']
const FORCED_APPS = {
  console: 'The Plural Console will allow you to monitor, upgrade, and deploy applications easily from one centralized place.',
}

function Installer(): React.ReactElement {
  const navigate = useNavigate()
  const client = useApolloClient()
  // const { project: { provider } } = useContext(WailsContext)

  // TODO: Use project provider when finished testing
  const provider = Provider.Aws

  const [stepsLoading, setStepsLoading] = useState(false)
  const [steps, setSteps] = useState<Array<WizardStepConfig>>([])
  const [error, setError] = useState<ApolloError | undefined>()
  const [defaultSteps, setDefaultSteps] = useState<Array<WizardStepConfig>>([])
  const [selectedApplications, setSelectedApplications] = useState<Array<string>>([])

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

  const onSelect = useCallback((selectedApplications: Array<WizardStepConfig>) => {
    const build = async () => {
      const steps = await buildSteps(client, provider!, selectedApplications)

      setSteps(steps)
    }

    setSelectedApplications(selectedApplications.map(app => app.label ?? 'unknown'))
    setStepsLoading(true)
    build().finally(() => setStepsLoading(false))
  }, [client, provider])

  useEffect(() => setDefaultSteps(toDefaultSteps(applications, provider!, { ...FORCED_APPS })), [applications?.length, provider])

  if (!applications || defaultSteps.length === 0) {
    return (
      <div style={{
        // TODO: move to styled css
        display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%',
      }}
      >
        <LoopingLogo />
      </div>
    )
  }

  return (
    <div style={{ height: '100%' }}>
      <Wizard
        onSelect={onSelect}
        defaultSteps={defaultSteps}
        dependencySteps={steps}
        limit={5}
        loading={stepsLoading}
      >
        {{
          stepper: <WizardStepper />,
          navigation: <WizardNavigation onInstall={() => navigate(Routes.Next)} />,
        }}
      </Wizard>

      {error && (
        <GraphQLToast
          error={{ graphQLErrors: [...error.graphQLErrors] }}
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
