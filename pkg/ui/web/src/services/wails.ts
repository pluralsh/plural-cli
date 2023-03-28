import { Context, Project, Token } from '../../wailsjs/go/ui/Client'
import {
  Client,
  ClientBinding,
  PluralContext,
  PluralProject,
} from '../types/client'

/**
 * List of supported client methods based on API Go client.
 * @see pkg/api/client.go
 */
enum Binding {
  Token = 'Token',
  Project = 'Project',
  Context = 'Context',
}

/**
 * Client mapping from defined bindings to exposed Go backend methods.
 * Abstracts the backend calls and wraps them with proper return types
 * to simplify usage in the UI.
 * @see Binding
 */
const Plural: Client = {
  [Binding.Token]: (): Promise<string> => Token(),
  [Binding.Project]: (): Promise<PluralProject> => Project() as Promise<PluralProject>,
  [Binding.Context]: (): Promise<PluralContext> => Context() as Promise<PluralContext>,
}

/**
 * Factory that simplifies getting wrapped client binding methods.
 * @param binding
 * @constructor
 */
function ClientBindingFactory<TResult>(binding: Binding): ClientBinding<TResult> {
  const bindingFn: ClientBinding<TResult> = Plural[binding] as ClientBinding<TResult>

  if (!bindingFn) throw new Error(`Unsupported client endpoint: ${binding}`)

  return bindingFn
}

export { ClientBindingFactory, Binding }
