import { FormField, Input } from '@pluralsh/design-system'
import { Switch } from 'honorable'
import StartCase from 'lodash/startCase'
import {
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'

import { WailsContext } from '../../context/wails'
import { Datatype } from '../../graphql/generated/graphql'
import { PluralProject } from '../../types/client'

import ConfigurationFileInput from './ConfigurationFileInput'

type ModifierFunction = (value: string, trim?: boolean) => string

const modifierFactory = (type: Datatype, project: PluralProject): ModifierFunction => {
  switch (type) {
  case Datatype.String:
  case Datatype.Int:
  case Datatype.Password:
    return stringModifier
  case Datatype.Bucket:
    return bucketModifier.bind({ project })
  case Datatype.Domain:
    return domainModifier.bind({ project })
  }

  return stringModifier
}

const stringModifier = value => value

function bucketModifier(this: {project: PluralProject}, value: string, trim = false) {
  const { project } = this
  const bucketPrefix = project?.bucketPrefix
  const cluster = project?.cluster
  const prefix = `${bucketPrefix}-${cluster}-`

  if (trim) return value?.replace(prefix, '')

  return bucketPrefix && cluster ? `${prefix}${value}` : value
}
function domainModifier(this: {project: PluralProject}, value: string, trim = false) {
  const { project } = this
  const subdomain = project?.network?.subdomain || ''
  const suffix = subdomain ? `.${subdomain}` : ''

  if (trim) return value?.replace(suffix, '')

  return subdomain ? `${value}${suffix}` : value
}

const createValidator = (regex: RegExp, optional: boolean, error: string) => (value): {valid: boolean, message: string} => ({
  valid: (value ? regex.test(value) : optional),
  message: error,
})

function ConfigurationField({
  config, ctx, setValue,
}) {
  const {
    name,
    default: defaultValue,
    placeholder,
    documentation,
    validation,
    optional,
    type,
  } = config
  const { project } = useContext(WailsContext)

  const value = useMemo(() => ctx[name]?.value, [ctx, name])
  const validator = useMemo(() => createValidator(new RegExp(validation?.regex ? `^${validation?.regex}$` : /.*/),
    config.optional,
    validation?.message),
  [config.optional, validation?.message, validation?.regex])
  const { valid, message } = useMemo(() => validator(value), [validator, value])
  const modifier = useMemo(() => modifierFactory(config.type, project),
    [config.type, project])

  const isFile = type === Datatype.File

  const [local, setLocal] = useState(modifier(value, true) || (isFile ? null : defaultValue))

  useEffect(() => (local
    ? setValue(
      name, modifier(local), valid, type
    )
    : setValue(
      name, local, valid, type
    )),
  [local, modifier, name, setValue, type, valid])

  const isInt = type === Datatype.Int
  const isPassword
      = type === Datatype.Password
      || ['private_key', 'public_key'].includes(config.name)

  const inputFieldType = isInt
    ? 'number'
    : isPassword
      ? 'password'
      : 'text'

  return (
    <FormField
      label={StartCase(name)}
      hint={message || documentation}
      error={!valid}
      required={!optional}
    >
      {isFile ? (
        <ConfigurationFileInput
          value={local ?? ''}
          onChange={val => {
            setLocal(val?.text ?? '')
          }}
        />
      ) : (
        <Input
          placeholder={placeholder}
          value={local}
          type={inputFieldType}
          error={!valid}
          prefix={
            config.type === Datatype.Bucket
              ? `${project?.bucketPrefix}-`
              : ''
          }
          suffix={
            config.type === Datatype.Domain
              ? `.${project?.network?.subdomain}`
              : ''
          }
          onChange={({ target: { value } }) => setLocal(value)}
        />
      )}
    </FormField>
  )
}

function BoolConfiguration({ config: { name, default: def }, ctx, setValue }) {
  const value: boolean = `${ctx[name]?.value}`.toLowerCase() === 'true'

  useEffect(() => {
    if (value === undefined && def) {
      setValue(name, def)
    }
  }, [value, def, name, setValue])

  return (
    <Switch
      checked={value}
      onChange={({ target: { checked } }) => setValue(name, checked)}
    >
      {StartCase(name)}
    </Switch>
  )
}

export function ConfigurationItem({ config, ctx, setValue }) {
  switch (config.type) {
  case Datatype.Bool:
    return (
      <BoolConfiguration
        config={config}
        ctx={ctx}
        setValue={setValue}
      />
    )
  default:
    return (
      <ConfigurationField
        config={config}
        ctx={ctx}
        setValue={setValue}
      />
    )
  }
}
