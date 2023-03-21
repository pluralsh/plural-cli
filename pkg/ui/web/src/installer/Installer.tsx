import {
  AppsIcon,
  FormField,
  GlobeIcon,
  Input,
  InstallIcon,
  Wizard,
  WizardInstaller,
  WizardNavigation,
  WizardPicker,
  WizardStep,
  WizardStepConfig,
  WizardStepper,
  useActive,
} from '@pluralsh/design-system'
import { P } from 'honorable'
import React, {
  ReactElement,
  useEffect,
  useMemo,
  useState,
} from 'react'

interface FormData {
  domain: string
}

function Application({ ...props }: any): ReactElement {
  const { active, setData } = useActive<FormData>()
  const [domain, setDomain] = useState<string>(active?.data?.domain ?? '')

  // Build our form data
  const data = useMemo<FormData>(() => ({ domain }), [domain])

  // Update step data on change
  useEffect(() => setData(data), [domain, setData, data])

  return (
    <WizardStep
      valid={domain?.length > 0}
      data={data}
      {...props}
    >
      <P
        overline
        color="text-xlight"
        paddingBottom="medium"
      >configure {active.label}
      </P>
      <FormField
        label="Domain"
        required
      >
        <Input
          placeholder="https://{domain}.onplural.sh"
          value={domain}
          onChange={event => setDomain(event.target.value)}
        />
      </FormField>
    </WizardStep>
  )
}

const PICKER_ITEMS: Array<WizardStepConfig> = [
  {
    key: 'airflow',
    label: 'Airflow',
    Icon: GlobeIcon,
    node: <Application key="airflow" />,
  },
  {
    key: 'airbyte',
    label: 'Airbyte',
    Icon: GlobeIcon,
    node: <Application key="airbyte" />,
  },
  {
    key: 'console',
    label: 'Console',
    Icon: GlobeIcon,
    node: <Application key="console" />,
  },
  {
    key: 'crossplane',
    label: 'Crossplane',
    Icon: GlobeIcon,
    node: <Application key="crossplane" />,
  },
  {
    key: 'grafana',
    label: 'Grafana',
    Icon: GlobeIcon,
    node: <Application key="grafana" />,
  },
  {
    key: 'mongodb',
    label: 'MongoDB',
    Icon: GlobeIcon,
    node: <Application key="mongodb" />,
  },
  {
    key: 'datadog',
    label: 'Datadog',
    Icon: GlobeIcon,
    node: <Application key="datadog" />,
  },
]

const DEFAULT_STEPS: Array<WizardStepConfig> = [
  {
    key: 'apps',
    label: 'Apps',
    Icon: AppsIcon,
    node: <WizardPicker items={PICKER_ITEMS} />,
    isDefault: true,
  },
  {
    key: 'placeholder',
    isPlaceholder: true,
  },
  {
    key: 'install',
    label: 'Install',
    Icon: InstallIcon,
    node: <WizardInstaller />,
    isDefault: true,
  },
]

function Installer(): React.ReactElement {
  return (
    <Wizard defaultSteps={DEFAULT_STEPS}>
      {{
        stepper: <WizardStepper />,
        navigation: <WizardNavigation onInstall={() => {}} />,
      }}
    </Wizard>
  )
}

export { Installer }
