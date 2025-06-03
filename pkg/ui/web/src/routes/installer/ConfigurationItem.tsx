import { FormField, Input, useActive } from '@pluralsh/design-system'
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
import { InstallerContext } from './context'

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

/**
 * Creates validator for domain uniqueness check.
 *
 * @param ctx - object that maps field name to an object field with value, validity, etc.
 * @param fieldName - field name being checked
 * @param appName - active application name
 * @param registeredDomains - a set of domains used by already installed applications
 * @param usedDomains - object that maps key (appName-fieldName) to the domain name.
 *                      It is basically a list of unique domains used by the installer locally.
 */
const createUniqueDomainValidator
  = (
    ctx: Record<string, any>,
    fieldName: string,
    appName: string,
    registeredDomains: Set<string>,
    usedDomains: Record<string, string>
  ) => (value): { valid: boolean; message: string } => {
    const domains = new Set<string>(registeredDomains)

    Object.entries(ctx)
      .filter(([name, field]) => field.type === Datatype.Domain
            && name !== fieldName
            && field.value?.length > 0)
      .forEach(([_, field]) => domains.add(field.value))

    Object.entries(usedDomains)
      .filter(([key]) => key !== domainFieldKey(appName, fieldName))
      .forEach(([_, domain]) => domains.add(domain))

    return {
      valid: !domains.has(value),
      message: `Domain ${value} already used.`,
    }
  }

const domainFieldKey = (appName, fieldName) => `${appName}-${fieldName}`

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
  const { project, context } = useContext(WailsContext)
  const { domains, setDomains } = useContext(InstallerContext)
  const { active } = useActive()

  const value = useMemo(() => ctx[name]?.value, [ctx, name])
  const validators = useMemo(() => [
    createValidator(new RegExp(validation?.regex ? `^${validation?.regex}$` : /.*/),
      config.optional,
      validation?.message),
    ...(type === Datatype.Domain
      ? [
        createUniqueDomainValidator(
          ctx,
          name,
            active.label!,
            new Set<string>(context.domains ?? []),
            domains
        ),
      ]
      : []),
  ],
  [active.label, config.optional, context.domains, ctx, domains, name, type, validation?.message, validation?.regex])
  const { valid, message } = useMemo(() => {
    for (const validator of validators) {
      const result = validator(value)

      if (!result.valid) return result
    }

    return { valid: true, message: '' }
  }, [validators, value])
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

  useEffect(() => {
    if (type !== Datatype.Domain || !value) return

    setDomains(domains => ({
      ...domains,
      ...{ [domainFieldKey(active.label, name)]: value },
    }))
  }, [active.label, name, setDomains, type, value])

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
