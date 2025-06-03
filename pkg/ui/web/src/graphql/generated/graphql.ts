/* eslint-disable */
import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  /**
   * The `DateTime` scalar type represents a date and time in the UTC
   * timezone. The DateTime appears in a JSON response as an ISO8601 formatted
   * string, including UTC timezone ("Z"). The parsed date and time string will
   * be converted to UTC if there is an offset.
   */
  DateTime: Date;
  Map: Map<string, unknown>;
  UploadOrUrl: string;
  Yaml: unknown;
};

export type RootQueryType = {
  __typename?: 'RootQueryType';
  account?: Maybe<Account>;
  auditMetrics?: Maybe<Array<Maybe<GeoMetric>>>;
  audits?: Maybe<AuditConnection>;
  categories?: Maybe<Array<Maybe<CategoryInfo>>>;
  category?: Maybe<CategoryInfo>;
  chart?: Maybe<Chart>;
  chartInstallations?: Maybe<ChartInstallationConnection>;
  charts?: Maybe<ChartConnection>;
  closure?: Maybe<Array<Maybe<ClosureItem>>>;
  /** Get a cluster by its ID. */
  cluster?: Maybe<Cluster>;
  /** Get a list of clusters owned by the current account. */
  clusters?: Maybe<ClusterConnection>;
  configuration?: Maybe<PluralConfiguration>;
  deferredUpdates?: Maybe<DeferredUpdateConnection>;
  demoProject?: Maybe<DemoProject>;
  dnsDomain?: Maybe<DnsDomain>;
  dnsDomains?: Maybe<DnsDomainConnection>;
  dnsRecords?: Maybe<DnsRecordConnection>;
  dockerImage?: Maybe<DockerImage>;
  dockerImages?: Maybe<DockerImageConnection>;
  dockerRepositories?: Maybe<DockerRepositoryConnection>;
  eabCredential?: Maybe<EabCredential>;
  eabCredentials?: Maybe<Array<Maybe<EabCredential>>>;
  groupMembers?: Maybe<GroupMemberConnection>;
  groups?: Maybe<GroupConnection>;
  helpQuestion?: Maybe<Scalars['String']>;
  incident?: Maybe<Incident>;
  incidents?: Maybe<IncidentConnection>;
  installation?: Maybe<Installation>;
  installations?: Maybe<InstallationConnection>;
  integrations?: Maybe<IntegrationConnection>;
  integrationWebhook?: Maybe<IntegrationWebhook>;
  integrationWebhooks?: Maybe<IntegrationWebhookConnection>;
  invite?: Maybe<Invite>;
  invites?: Maybe<InviteConnection>;
  invoices?: Maybe<InvoiceConnection>;
  keyBackup?: Maybe<KeyBackup>;
  keyBackups?: Maybe<KeyBackupConnection>;
  loginMethod?: Maybe<LoginMethodResponse>;
  loginMetrics?: Maybe<Array<Maybe<GeoMetric>>>;
  me?: Maybe<User>;
  notifications?: Maybe<NotificationConnection>;
  oauthConsent?: Maybe<Repository>;
  oauthIntegrations?: Maybe<Array<Maybe<OauthIntegration>>>;
  oauthLogin?: Maybe<Repository>;
  oauthUrls?: Maybe<Array<Maybe<OauthInfo>>>;
  oidcConsent?: Maybe<OidcStepResponse>;
  oidcLogin?: Maybe<OidcStepResponse>;
  oidcLogins?: Maybe<OidcLoginConnection>;
  platformMetrics?: Maybe<PlatformMetrics>;
  platformPlans?: Maybe<Array<Maybe<PlatformPlan>>>;
  platformSubscription?: Maybe<PlatformSubscription>;
  publicKeys?: Maybe<PublicKeyConnection>;
  publisher?: Maybe<Publisher>;
  publishers?: Maybe<PublisherConnection>;
  recipe?: Maybe<Recipe>;
  recipes?: Maybe<RecipeConnection>;
  repositories?: Maybe<RepositoryConnection>;
  /** Get an application by its ID or name. */
  repository?: Maybe<Repository>;
  repositorySubscription?: Maybe<RepositorySubscription>;
  resetToken?: Maybe<ResetToken>;
  role?: Maybe<Role>;
  roles?: Maybe<RoleConnection>;
  rollouts?: Maybe<RolloutConnection>;
  scaffold?: Maybe<Array<Maybe<ScaffoldFile>>>;
  scmAuthorization?: Maybe<Array<Maybe<AuthorizationUrl>>>;
  scmToken?: Maybe<Scalars['String']>;
  searchRepositories?: Maybe<RepositoryConnection>;
  searchUsers?: Maybe<UserConnection>;
  shell?: Maybe<CloudShell>;
  shellConfiguration?: Maybe<ShellConfiguration>;
  stack?: Maybe<Stack>;
  stacks?: Maybe<StackConnection>;
  subscriptions?: Maybe<RepositorySubscriptionConnection>;
  tags?: Maybe<GroupedTagConnection>;
  terraform?: Maybe<TerraformConnection>;
  terraformInstallations?: Maybe<TerraformInstallationConnection>;
  terraformModule?: Maybe<Terraform>;
  terraformProvider?: Maybe<TerraformProvider>;
  terraformProviders?: Maybe<Array<Maybe<Provider>>>;
  test?: Maybe<Test>;
  testLogs?: Maybe<Scalars['String']>;
  tests?: Maybe<TestConnection>;
  token?: Maybe<PersistedToken>;
  tokens?: Maybe<PersistedTokenConnection>;
  upgradeQueue?: Maybe<UpgradeQueue>;
  upgradeQueues?: Maybe<Array<Maybe<UpgradeQueue>>>;
  users?: Maybe<UserConnection>;
  versions?: Maybe<VersionConnection>;
  webhooks?: Maybe<WebhookConnection>;
};


export type RootQueryTypeAuditsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeCategoryArgs = {
  name: Category;
};


export type RootQueryTypeChartArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeChartInstallationsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeChartsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeClosureArgs = {
  id: Scalars['ID'];
  type: DependencyType;
};


export type RootQueryTypeClusterArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeClustersArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeDeferredUpdatesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  chartInstallationId?: InputMaybe<Scalars['ID']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  terraformInstallationId?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeDemoProjectArgs = {
  id?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeDnsDomainArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeDnsDomainsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeDnsRecordsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  cluster?: InputMaybe<Scalars['String']>;
  domainId?: InputMaybe<Scalars['ID']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  provider?: InputMaybe<Provider>;
};


export type RootQueryTypeDockerImageArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeDockerImagesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  dockerRepositoryId: Scalars['ID'];
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeDockerRepositoriesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeEabCredentialArgs = {
  cluster: Scalars['String'];
  provider: Provider;
};


export type RootQueryTypeGroupMembersArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  groupId: Scalars['ID'];
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeGroupsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeHelpQuestionArgs = {
  prompt: Scalars['String'];
};


export type RootQueryTypeIncidentArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeIncidentsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  filters?: InputMaybe<Array<InputMaybe<IncidentFilter>>>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  order?: InputMaybe<Order>;
  q?: InputMaybe<Scalars['String']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
  sort?: InputMaybe<IncidentSort>;
  supports?: InputMaybe<Scalars['Boolean']>;
};


export type RootQueryTypeInstallationArgs = {
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeInstallationsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeIntegrationsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
  repositoryName?: InputMaybe<Scalars['String']>;
  tag?: InputMaybe<Scalars['String']>;
  type?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeIntegrationWebhookArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeIntegrationWebhooksArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeInviteArgs = {
  id: Scalars['String'];
};


export type RootQueryTypeInvitesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeInvoicesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeKeyBackupArgs = {
  name: Scalars['String'];
};


export type RootQueryTypeKeyBackupsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeLoginMethodArgs = {
  email: Scalars['String'];
  host?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeNotificationsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  cli?: InputMaybe<Scalars['Boolean']>;
  first?: InputMaybe<Scalars['Int']>;
  incidentId?: InputMaybe<Scalars['ID']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeOauthConsentArgs = {
  challenge: Scalars['String'];
};


export type RootQueryTypeOauthLoginArgs = {
  challenge: Scalars['String'];
};


export type RootQueryTypeOauthUrlsArgs = {
  host?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeOidcConsentArgs = {
  challenge: Scalars['String'];
};


export type RootQueryTypeOidcLoginArgs = {
  challenge: Scalars['String'];
};


export type RootQueryTypeOidcLoginsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypePublicKeysArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  emails?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypePublisherArgs = {
  id?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypePublishersArgs = {
  accountId?: InputMaybe<Scalars['ID']>;
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  publishable?: InputMaybe<Scalars['Boolean']>;
};


export type RootQueryTypeRecipeArgs = {
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
  repo?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeRecipesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  provider?: InputMaybe<Provider>;
  repositoryId?: InputMaybe<Scalars['ID']>;
  repositoryName?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeRepositoriesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  categories?: InputMaybe<Array<InputMaybe<Category>>>;
  category?: InputMaybe<Category>;
  first?: InputMaybe<Scalars['Int']>;
  installed?: InputMaybe<Scalars['Boolean']>;
  last?: InputMaybe<Scalars['Int']>;
  provider?: InputMaybe<Provider>;
  publisherId?: InputMaybe<Scalars['ID']>;
  publishers?: InputMaybe<Array<InputMaybe<Scalars['ID']>>>;
  q?: InputMaybe<Scalars['String']>;
  supports?: InputMaybe<Scalars['Boolean']>;
  tag?: InputMaybe<Scalars['String']>;
  tags?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};


export type RootQueryTypeRepositoryArgs = {
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeRepositorySubscriptionArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeResetTokenArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeRoleArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeRolesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
  userId?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeRolloutsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeScaffoldArgs = {
  application: Scalars['String'];
  category: Category;
  ingress?: InputMaybe<Scalars['Boolean']>;
  postgres?: InputMaybe<Scalars['Boolean']>;
  publisher: Scalars['String'];
};


export type RootQueryTypeScmTokenArgs = {
  code: Scalars['String'];
  provider: ScmProvider;
};


export type RootQueryTypeSearchRepositoriesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  query: Scalars['String'];
};


export type RootQueryTypeSearchUsersArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  incidentId: Scalars['ID'];
  last?: InputMaybe<Scalars['Int']>;
  q: Scalars['String'];
};


export type RootQueryTypeStackArgs = {
  name: Scalars['String'];
  provider: Provider;
};


export type RootQueryTypeStacksArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  featured?: InputMaybe<Scalars['Boolean']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeSubscriptionsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeTagsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  id?: InputMaybe<Scalars['ID']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
  type: TagGroup;
};


export type RootQueryTypeTerraformArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeTerraformInstallationsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId: Scalars['ID'];
};


export type RootQueryTypeTerraformModuleArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeTerraformProviderArgs = {
  name: Provider;
  vsn?: InputMaybe<Scalars['String']>;
};


export type RootQueryTypeTestArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeTestLogsArgs = {
  id: Scalars['ID'];
  step: Scalars['ID'];
};


export type RootQueryTypeTestsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
  versionId?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeTokenArgs = {
  id: Scalars['ID'];
};


export type RootQueryTypeTokensArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type RootQueryTypeUpgradeQueueArgs = {
  id?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeUsersArgs = {
  after?: InputMaybe<Scalars['String']>;
  all?: InputMaybe<Scalars['Boolean']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
  serviceAccount?: InputMaybe<Scalars['Boolean']>;
};


export type RootQueryTypeVersionsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  chartId?: InputMaybe<Scalars['ID']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  terraformId?: InputMaybe<Scalars['ID']>;
};


export type RootQueryTypeWebhooksArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type Account = {
  __typename?: 'Account';
  availableFeatures?: Maybe<PlanFeatures>;
  backgroundColor?: Maybe<Scalars['String']>;
  billingAddress?: Maybe<Address>;
  billingCustomerId?: Maybe<Scalars['String']>;
  clusterCount?: Maybe<Scalars['String']>;
  delinquentAt?: Maybe<Scalars['DateTime']>;
  domainMappings?: Maybe<Array<Maybe<DomainMapping>>>;
  grandfatheredUntil?: Maybe<Scalars['DateTime']>;
  icon?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name?: Maybe<Scalars['String']>;
  rootUser?: Maybe<User>;
  subscription?: Maybe<PlatformSubscription>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  userCount?: Maybe<Scalars['String']>;
  workosConnectionId?: Maybe<Scalars['String']>;
};

export type PlanFeatures = {
  __typename?: 'PlanFeatures';
  audit?: Maybe<Scalars['Boolean']>;
  userManagement?: Maybe<Scalars['Boolean']>;
  vpn?: Maybe<Scalars['Boolean']>;
};

export type Address = {
  __typename?: 'Address';
  city?: Maybe<Scalars['String']>;
  country?: Maybe<Scalars['String']>;
  line1?: Maybe<Scalars['String']>;
  line2?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  state?: Maybe<Scalars['String']>;
  zip?: Maybe<Scalars['String']>;
};

export type DomainMapping = {
  __typename?: 'DomainMapping';
  account?: Maybe<Account>;
  domain: Scalars['String'];
  enableSso?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type User = {
  __typename?: 'User';
  account: Account;
  address?: Maybe<Address>;
  avatar?: Maybe<Scalars['String']>;
  backgroundColor?: Maybe<Scalars['String']>;
  boundRoles?: Maybe<Array<Maybe<Role>>>;
  cards?: Maybe<CardConnection>;
  defaultQueueId?: Maybe<Scalars['ID']>;
  /** If a user has reached the demo project usage limit. */
  demoed?: Maybe<Scalars['Boolean']>;
  demoing?: Maybe<Scalars['Boolean']>;
  email: Scalars['String'];
  emailConfirmBy?: Maybe<Scalars['DateTime']>;
  emailConfirmed?: Maybe<Scalars['Boolean']>;
  hasInstallations?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  impersonationPolicy?: Maybe<ImpersonationPolicy>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  jwt?: Maybe<Scalars['String']>;
  loginMethod?: Maybe<LoginMethod>;
  name: Scalars['String'];
  onboarding?: Maybe<OnboardingState>;
  onboardingChecklist?: Maybe<OnboardingChecklist>;
  phone?: Maybe<Scalars['String']>;
  provider?: Maybe<Provider>;
  publisher?: Maybe<Publisher>;
  roles?: Maybe<Roles>;
  serviceAccount?: Maybe<Scalars['Boolean']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};


export type UserCardsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type Role = {
  __typename?: 'Role';
  account?: Maybe<Account>;
  description?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  permissions?: Maybe<Array<Maybe<Permission>>>;
  repositories?: Maybe<Array<Maybe<Scalars['String']>>>;
  roleBindings?: Maybe<Array<Maybe<RoleBinding>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum Permission {
  Billing = 'BILLING',
  Install = 'INSTALL',
  Integrations = 'INTEGRATIONS',
  Publish = 'PUBLISH',
  Support = 'SUPPORT',
  Users = 'USERS'
}

export type RoleBinding = {
  __typename?: 'RoleBinding';
  group?: Maybe<Group>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type Group = {
  __typename?: 'Group';
  description?: Maybe<Scalars['String']>;
  global?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type CardConnection = {
  __typename?: 'CardConnection';
  edges?: Maybe<Array<Maybe<CardEdge>>>;
  pageInfo: PageInfo;
};

export type CardEdge = {
  __typename?: 'CardEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Card>;
};

export type Card = {
  __typename?: 'Card';
  brand: Scalars['String'];
  expMonth: Scalars['Int'];
  expYear: Scalars['Int'];
  id: Scalars['ID'];
  last4: Scalars['String'];
  name?: Maybe<Scalars['String']>;
};

export type PageInfo = {
  __typename?: 'PageInfo';
  /** When paginating forwards, the cursor to continue. */
  endCursor?: Maybe<Scalars['String']>;
  /** When paginating forwards, are there more items? */
  hasNextPage: Scalars['Boolean'];
  /** When paginating backwards, are there more items? */
  hasPreviousPage: Scalars['Boolean'];
  /** When paginating backwards, the cursor to continue. */
  startCursor?: Maybe<Scalars['String']>;
};

export type ImpersonationPolicy = {
  __typename?: 'ImpersonationPolicy';
  bindings?: Maybe<Array<Maybe<ImpersonationPolicyBinding>>>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type ImpersonationPolicyBinding = {
  __typename?: 'ImpersonationPolicyBinding';
  group?: Maybe<Group>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export enum LoginMethod {
  Github = 'GITHUB',
  Google = 'GOOGLE',
  Password = 'PASSWORD',
  Passwordless = 'PASSWORDLESS',
  Sso = 'SSO'
}

export enum OnboardingState {
  Active = 'ACTIVE',
  Installed = 'INSTALLED',
  New = 'NEW',
  Onboarded = 'ONBOARDED'
}

export type OnboardingChecklist = {
  __typename?: 'OnboardingChecklist';
  dismissed?: Maybe<Scalars['Boolean']>;
  status?: Maybe<OnboardingChecklistState>;
};

export enum OnboardingChecklistState {
  Configured = 'CONFIGURED',
  ConsoleInstalled = 'CONSOLE_INSTALLED',
  Finished = 'FINISHED',
  New = 'NEW'
}

export enum Provider {
  Aws = 'AWS',
  Azure = 'AZURE',
  Custom = 'CUSTOM',
  Equinix = 'EQUINIX',
  Gcp = 'GCP',
  Generic = 'GENERIC',
  Kind = 'KIND',
  Kubernetes = 'KUBERNETES'
}

export type Publisher = {
  __typename?: 'Publisher';
  address?: Maybe<Address>;
  avatar?: Maybe<Scalars['String']>;
  backgroundColor?: Maybe<Scalars['String']>;
  billingAccountId?: Maybe<Scalars['String']>;
  community?: Maybe<Community>;
  description?: Maybe<Scalars['String']>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  owner?: Maybe<User>;
  phone?: Maybe<Scalars['String']>;
  repositories?: Maybe<Array<Maybe<Repository>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type Community = {
  __typename?: 'Community';
  discord?: Maybe<Scalars['String']>;
  gitUrl?: Maybe<Scalars['String']>;
  homepage?: Maybe<Scalars['String']>;
  slack?: Maybe<Scalars['String']>;
  twitter?: Maybe<Scalars['String']>;
  videos?: Maybe<Array<Maybe<Scalars['String']>>>;
};

/** Container for all resources to create an application. */
export type Repository = {
  __typename?: 'Repository';
  /** The artifacts of the application. */
  artifacts?: Maybe<Array<Maybe<Artifact>>>;
  /** The category of the application. */
  category?: Maybe<Category>;
  /** The community links of the application. */
  community?: Maybe<Community>;
  darkIcon?: Maybe<Scalars['String']>;
  /** The default tag to deploy. */
  defaultTag?: Maybe<Scalars['String']>;
  /** The description of the application. */
  description?: Maybe<Scalars['String']>;
  /** The documentation of the application. */
  docs?: Maybe<Array<Maybe<FileContent>>>;
  /** The documentation of the application. */
  documentation?: Maybe<Scalars['String']>;
  /** If the application can be edited by the current user. */
  editable?: Maybe<Scalars['Boolean']>;
  /** The git URL of the application. */
  gitUrl?: Maybe<Scalars['String']>;
  /** The homepage of the application. */
  homepage?: Maybe<Scalars['String']>;
  icon?: Maybe<Scalars['String']>;
  /** The application's ID. */
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  /** The installation of the application by a user. */
  installation?: Maybe<Installation>;
  /** The license of the application. */
  license?: Maybe<License>;
  /** The main branch of the application. */
  mainBranch?: Maybe<Scalars['String']>;
  /** The name of the application. */
  name: Scalars['String'];
  /** Notes about the application rendered after deploying and displayed to the user. */
  notes?: Maybe<Scalars['String']>;
  /** The OAuth settings for the application. */
  oauthSettings?: Maybe<OauthSettings>;
  /** The available plans for the application. */
  plans?: Maybe<Array<Maybe<Plan>>>;
  /** Whether the application is private. */
  private?: Maybe<Scalars['Boolean']>;
  /** The application's public key. */
  publicKey?: Maybe<Scalars['String']>;
  /** The application publisher. */
  publisher?: Maybe<Publisher>;
  /** The README of the application. */
  readme?: Maybe<Scalars['String']>;
  /** The recipes used to install the application. */
  recipes?: Maybe<Array<Maybe<Recipe>>>;
  /** A map of secrets of the application. */
  secrets?: Maybe<Scalars['Map']>;
  /** The tags of the application. */
  tags?: Maybe<Array<Maybe<Tag>>>;
  /** Whether the application is trending. */
  trending?: Maybe<Scalars['Boolean']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  /** Whether the application is verified. */
  verified?: Maybe<Scalars['Boolean']>;
};

export type Artifact = {
  __typename?: 'Artifact';
  arch?: Maybe<Scalars['String']>;
  blob?: Maybe<Scalars['String']>;
  filesize?: Maybe<Scalars['Int']>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  name?: Maybe<Scalars['String']>;
  platform?: Maybe<ArtifactPlatform>;
  readme?: Maybe<Scalars['String']>;
  sha?: Maybe<Scalars['String']>;
  type?: Maybe<ArtifactType>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum ArtifactPlatform {
  Android = 'ANDROID',
  Freebsd = 'FREEBSD',
  Linux = 'LINUX',
  Mac = 'MAC',
  Openbsd = 'OPENBSD',
  Solaris = 'SOLARIS',
  Windows = 'WINDOWS'
}

export enum ArtifactType {
  Cli = 'CLI',
  Desktop = 'DESKTOP',
  Mobile = 'MOBILE'
}

/** Application categories. */
export enum Category {
  Data = 'DATA',
  Database = 'DATABASE',
  Devops = 'DEVOPS',
  Messaging = 'MESSAGING',
  Network = 'NETWORK',
  Productivity = 'PRODUCTIVITY',
  Security = 'SECURITY',
  Storage = 'STORAGE'
}

export type FileContent = {
  __typename?: 'FileContent';
  content: Scalars['String'];
  path: Scalars['String'];
};

/** An installation of an application. */
export type Installation = {
  __typename?: 'Installation';
  acmeKeyId?: Maybe<Scalars['String']>;
  acmeSecret?: Maybe<Scalars['String']>;
  /** Whether the application should auto upgrade. */
  autoUpgrade?: Maybe<Scalars['Boolean']>;
  /** A YAML object of context. */
  context?: Maybe<Scalars['Map']>;
  /** The installation's ID. */
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  license?: Maybe<Scalars['String']>;
  /** The license key for the application. */
  licenseKey?: Maybe<Scalars['String']>;
  /** The OIDC provider for the application. */
  oidcProvider?: Maybe<OidcProvider>;
  /** The last ping time of an installed application. */
  pingedAt?: Maybe<Scalars['DateTime']>;
  /** The application that was installed. */
  repository?: Maybe<Repository>;
  /** The subscription for the application. */
  subscription?: Maybe<RepositorySubscription>;
  /** The tag to track for auto upgrades. */
  trackTag: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
  /** The user that installed the application. */
  user?: Maybe<User>;
};

export type OidcProvider = {
  __typename?: 'OidcProvider';
  authMethod: OidcAuthMethod;
  bindings?: Maybe<Array<Maybe<OidcProviderBinding>>>;
  clientId: Scalars['String'];
  clientSecret: Scalars['String'];
  configuration?: Maybe<OuathConfiguration>;
  consent?: Maybe<ConsentRequest>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  redirectUris?: Maybe<Array<Maybe<Scalars['String']>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

/** Supported OIDC authentication methods. */
export enum OidcAuthMethod {
  Basic = 'BASIC',
  Post = 'POST'
}

export type OidcProviderBinding = {
  __typename?: 'OidcProviderBinding';
  group?: Maybe<Group>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type OuathConfiguration = {
  __typename?: 'OuathConfiguration';
  authorizationEndpoint?: Maybe<Scalars['String']>;
  issuer?: Maybe<Scalars['String']>;
  jwksUri?: Maybe<Scalars['String']>;
  tokenEndpoint?: Maybe<Scalars['String']>;
  userinfoEndpoint?: Maybe<Scalars['String']>;
};

export type ConsentRequest = {
  __typename?: 'ConsentRequest';
  requestedScope?: Maybe<Array<Maybe<Scalars['String']>>>;
  skip?: Maybe<Scalars['Boolean']>;
};

export type RepositorySubscription = {
  __typename?: 'RepositorySubscription';
  customerId?: Maybe<Scalars['String']>;
  externalId?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  installation?: Maybe<Installation>;
  invoices?: Maybe<InvoiceConnection>;
  lineItems?: Maybe<SubscriptionLineItems>;
  plan?: Maybe<Plan>;
};


export type RepositorySubscriptionInvoicesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type InvoiceConnection = {
  __typename?: 'InvoiceConnection';
  edges?: Maybe<Array<Maybe<InvoiceEdge>>>;
  pageInfo: PageInfo;
};

export type InvoiceEdge = {
  __typename?: 'InvoiceEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Invoice>;
};

export type Invoice = {
  __typename?: 'Invoice';
  amountDue: Scalars['Int'];
  amountPaid: Scalars['Int'];
  createdAt?: Maybe<Scalars['DateTime']>;
  currency: Scalars['String'];
  hostedInvoiceUrl?: Maybe<Scalars['String']>;
  lines?: Maybe<Array<Maybe<InvoiceItem>>>;
  number: Scalars['String'];
  status?: Maybe<Scalars['String']>;
};

export type InvoiceItem = {
  __typename?: 'InvoiceItem';
  amount: Scalars['Int'];
  currency: Scalars['String'];
  description?: Maybe<Scalars['String']>;
};

export type SubscriptionLineItems = {
  __typename?: 'SubscriptionLineItems';
  items?: Maybe<Array<Maybe<Limit>>>;
};

export type Limit = {
  __typename?: 'Limit';
  dimension: Scalars['String'];
  quantity: Scalars['Int'];
};

export type Plan = {
  __typename?: 'Plan';
  cost: Scalars['Int'];
  default?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  lineItems?: Maybe<PlanLineItems>;
  metadata?: Maybe<PlanMetadata>;
  name: Scalars['String'];
  period?: Maybe<Scalars['String']>;
  serviceLevels?: Maybe<Array<Maybe<ServiceLevel>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  visible: Scalars['Boolean'];
};

export type PlanLineItems = {
  __typename?: 'PlanLineItems';
  included?: Maybe<Array<Maybe<Limit>>>;
  items?: Maybe<Array<Maybe<LineItem>>>;
};

export type LineItem = {
  __typename?: 'LineItem';
  cost: Scalars['Int'];
  dimension: Scalars['String'];
  name: Scalars['String'];
  period?: Maybe<Scalars['String']>;
  type?: Maybe<PlanType>;
};

export enum PlanType {
  Licensed = 'LICENSED',
  Metered = 'METERED'
}

export type PlanMetadata = {
  __typename?: 'PlanMetadata';
  features?: Maybe<Array<Maybe<PlanFeature>>>;
  freeform?: Maybe<Scalars['Map']>;
};

export type PlanFeature = {
  __typename?: 'PlanFeature';
  description: Scalars['String'];
  name: Scalars['String'];
};

export type ServiceLevel = {
  __typename?: 'ServiceLevel';
  maxSeverity?: Maybe<Scalars['Int']>;
  minSeverity?: Maybe<Scalars['Int']>;
  responseTime?: Maybe<Scalars['Int']>;
};

export type License = {
  __typename?: 'License';
  name?: Maybe<Scalars['String']>;
  url?: Maybe<Scalars['String']>;
};

export type OauthSettings = {
  __typename?: 'OauthSettings';
  authMethod: OidcAuthMethod;
  uriFormat: Scalars['String'];
};

export type Recipe = {
  __typename?: 'Recipe';
  description?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  oidcEnabled?: Maybe<Scalars['Boolean']>;
  oidcSettings?: Maybe<OidcSettings>;
  primary?: Maybe<Scalars['Boolean']>;
  private?: Maybe<Scalars['Boolean']>;
  provider?: Maybe<Provider>;
  recipeDependencies?: Maybe<Array<Maybe<Recipe>>>;
  recipeSections?: Maybe<Array<Maybe<RecipeSection>>>;
  repository?: Maybe<Repository>;
  restricted?: Maybe<Scalars['Boolean']>;
  tests?: Maybe<Array<Maybe<RecipeTest>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type OidcSettings = {
  __typename?: 'OidcSettings';
  authMethod: OidcAuthMethod;
  domainKey?: Maybe<Scalars['String']>;
  subdomain?: Maybe<Scalars['Boolean']>;
  uriFormat?: Maybe<Scalars['String']>;
  uriFormats?: Maybe<Array<Maybe<Scalars['String']>>>;
};

export type RecipeSection = {
  __typename?: 'RecipeSection';
  configuration?: Maybe<Array<Maybe<RecipeConfiguration>>>;
  id?: Maybe<Scalars['ID']>;
  index?: Maybe<Scalars['Int']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  recipe?: Maybe<Recipe>;
  recipeItems?: Maybe<Array<Maybe<RecipeItem>>>;
  repository?: Maybe<Repository>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type RecipeConfiguration = {
  __typename?: 'RecipeConfiguration';
  args?: Maybe<Array<Maybe<Scalars['String']>>>;
  condition?: Maybe<RecipeCondition>;
  default?: Maybe<Scalars['String']>;
  documentation?: Maybe<Scalars['String']>;
  functionName?: Maybe<Scalars['String']>;
  longform?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  optional?: Maybe<Scalars['Boolean']>;
  placeholder?: Maybe<Scalars['String']>;
  type?: Maybe<Datatype>;
  validation?: Maybe<RecipeValidation>;
};

export type RecipeCondition = {
  __typename?: 'RecipeCondition';
  field: Scalars['String'];
  operation: Operation;
  value?: Maybe<Scalars['String']>;
};

export enum Operation {
  Eq = 'EQ',
  Gt = 'GT',
  Gte = 'GTE',
  Lt = 'LT',
  Lte = 'LTE',
  Not = 'NOT',
  Prefix = 'PREFIX',
  Suffix = 'SUFFIX'
}

export enum Datatype {
  Bool = 'BOOL',
  Bucket = 'BUCKET',
  Domain = 'DOMAIN',
  File = 'FILE',
  Function = 'FUNCTION',
  Int = 'INT',
  Password = 'PASSWORD',
  String = 'STRING'
}

export type RecipeValidation = {
  __typename?: 'RecipeValidation';
  message: Scalars['String'];
  regex?: Maybe<Scalars['String']>;
  type: ValidationType;
};

export enum ValidationType {
  Regex = 'REGEX'
}

export type RecipeItem = {
  __typename?: 'RecipeItem';
  chart?: Maybe<Chart>;
  configuration?: Maybe<Array<Maybe<RecipeConfiguration>>>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  recipeSection?: Maybe<RecipeSection>;
  terraform?: Maybe<Terraform>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type Chart = {
  __typename?: 'Chart';
  dependencies?: Maybe<Dependencies>;
  description?: Maybe<Scalars['String']>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  installation?: Maybe<ChartInstallation>;
  latestVersion?: Maybe<Scalars['String']>;
  name: Scalars['String'];
  repository?: Maybe<Repository>;
  tags?: Maybe<Array<Maybe<VersionTag>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type Dependencies = {
  __typename?: 'Dependencies';
  application?: Maybe<Scalars['Boolean']>;
  breaking?: Maybe<Scalars['Boolean']>;
  cliVsn?: Maybe<Scalars['String']>;
  dependencies?: Maybe<Array<Maybe<Dependency>>>;
  instructions?: Maybe<ChangeInstructions>;
  outputs?: Maybe<Scalars['Map']>;
  providers?: Maybe<Array<Maybe<Provider>>>;
  providerVsn?: Maybe<Scalars['String']>;
  providerWirings?: Maybe<Scalars['Map']>;
  secrets?: Maybe<Array<Maybe<Scalars['String']>>>;
  wait?: Maybe<Scalars['Boolean']>;
  wirings?: Maybe<Wirings>;
};

export type Dependency = {
  __typename?: 'Dependency';
  name?: Maybe<Scalars['String']>;
  optional?: Maybe<Scalars['Boolean']>;
  repo?: Maybe<Scalars['String']>;
  type?: Maybe<DependencyType>;
  version?: Maybe<Scalars['String']>;
};

export enum DependencyType {
  Helm = 'HELM',
  Terraform = 'TERRAFORM'
}

export type ChangeInstructions = {
  __typename?: 'ChangeInstructions';
  instructions?: Maybe<Scalars['String']>;
  script?: Maybe<Scalars['String']>;
};

export type Wirings = {
  __typename?: 'Wirings';
  helm?: Maybe<Scalars['Map']>;
  terraform?: Maybe<Scalars['Map']>;
};

export type ChartInstallation = {
  __typename?: 'ChartInstallation';
  chart?: Maybe<Chart>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  installation?: Maybe<Installation>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  version?: Maybe<Version>;
};

/** The version of a package. */
export type Version = {
  __typename?: 'Version';
  chart?: Maybe<Chart>;
  crds?: Maybe<Array<Maybe<Crd>>>;
  dependencies?: Maybe<Dependencies>;
  helm?: Maybe<Scalars['Map']>;
  id: Scalars['ID'];
  imageDependencies?: Maybe<Array<Maybe<ImageDependency>>>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  package?: Maybe<Scalars['String']>;
  readme?: Maybe<Scalars['String']>;
  scan?: Maybe<PackageScan>;
  tags?: Maybe<Array<Maybe<VersionTag>>>;
  /** The template engine used to render the valuesTemplate. */
  templateType?: Maybe<TemplateType>;
  terraform?: Maybe<Terraform>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  valuesTemplate?: Maybe<Scalars['String']>;
  version: Scalars['String'];
};

export type Crd = {
  __typename?: 'Crd';
  blob?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type ImageDependency = {
  __typename?: 'ImageDependency';
  id: Scalars['ID'];
  image: DockerImage;
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  version: Version;
};

export type DockerImage = {
  __typename?: 'DockerImage';
  digest: Scalars['String'];
  dockerRepository?: Maybe<DockerRepository>;
  grade?: Maybe<ImageGrade>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  scanCompletedAt?: Maybe<Scalars['DateTime']>;
  scannedAt?: Maybe<Scalars['DateTime']>;
  tag?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  vulnerabilities?: Maybe<Array<Maybe<Vulnerability>>>;
};

export type DockerRepository = {
  __typename?: 'DockerRepository';
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  metrics?: Maybe<Array<Maybe<Metric>>>;
  name: Scalars['String'];
  public?: Maybe<Scalars['Boolean']>;
  repository?: Maybe<Repository>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};


export type DockerRepositoryMetricsArgs = {
  offset?: InputMaybe<Scalars['String']>;
  precision?: InputMaybe<Scalars['String']>;
  tag?: InputMaybe<Scalars['String']>;
};

export type Metric = {
  __typename?: 'Metric';
  name: Scalars['String'];
  tags?: Maybe<Array<Maybe<MetricTag>>>;
  values?: Maybe<Array<Maybe<MetricValue>>>;
};

export type MetricTag = {
  __typename?: 'MetricTag';
  name: Scalars['String'];
  value: Scalars['String'];
};

export type MetricValue = {
  __typename?: 'MetricValue';
  time?: Maybe<Scalars['DateTime']>;
  value?: Maybe<Scalars['Int']>;
};

export enum ImageGrade {
  A = 'A',
  B = 'B',
  C = 'C',
  D = 'D',
  F = 'F'
}

export type Vulnerability = {
  __typename?: 'Vulnerability';
  cvss?: Maybe<Cvss>;
  description?: Maybe<Scalars['String']>;
  fixedVersion?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  installedVersion?: Maybe<Scalars['String']>;
  layer?: Maybe<ImageLayer>;
  package?: Maybe<Scalars['String']>;
  score?: Maybe<Scalars['Float']>;
  severity?: Maybe<VulnGrade>;
  source?: Maybe<Scalars['String']>;
  title?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  url?: Maybe<Scalars['String']>;
  vulnerabilityId?: Maybe<Scalars['String']>;
};

export type Cvss = {
  __typename?: 'Cvss';
  attackComplexity?: Maybe<VulnGrade>;
  attackVector?: Maybe<VulnVector>;
  availability?: Maybe<VulnGrade>;
  confidentiality?: Maybe<VulnGrade>;
  integrity?: Maybe<VulnGrade>;
  privilegesRequired?: Maybe<VulnGrade>;
  userInteraction?: Maybe<VulnRequirement>;
};

export enum VulnGrade {
  Critical = 'CRITICAL',
  High = 'HIGH',
  Low = 'LOW',
  Medium = 'MEDIUM',
  None = 'NONE'
}

export enum VulnVector {
  Adjacent = 'ADJACENT',
  Local = 'LOCAL',
  Network = 'NETWORK',
  Physical = 'PHYSICAL'
}

export enum VulnRequirement {
  None = 'NONE',
  Required = 'REQUIRED'
}

export type ImageLayer = {
  __typename?: 'ImageLayer';
  diffId?: Maybe<Scalars['String']>;
  digest?: Maybe<Scalars['String']>;
};

export type PackageScan = {
  __typename?: 'PackageScan';
  errors?: Maybe<Array<Maybe<ScanError>>>;
  grade?: Maybe<ImageGrade>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  violations?: Maybe<Array<Maybe<ScanViolation>>>;
};

export type ScanError = {
  __typename?: 'ScanError';
  message?: Maybe<Scalars['String']>;
};

export type ScanViolation = {
  __typename?: 'ScanViolation';
  category?: Maybe<Scalars['String']>;
  description?: Maybe<Scalars['String']>;
  file?: Maybe<Scalars['String']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  line?: Maybe<Scalars['Int']>;
  resourceName?: Maybe<Scalars['String']>;
  resourceType?: Maybe<Scalars['String']>;
  ruleId?: Maybe<Scalars['String']>;
  ruleName?: Maybe<Scalars['String']>;
  severity?: Maybe<VulnGrade>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type VersionTag = {
  __typename?: 'VersionTag';
  chart?: Maybe<Chart>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  tag: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
  version?: Maybe<Version>;
};

/** Template engines that can be used at build time. */
export enum TemplateType {
  Gotemplate = 'GOTEMPLATE',
  Javascript = 'JAVASCRIPT',
  Lua = 'LUA'
}

export type Terraform = {
  __typename?: 'Terraform';
  dependencies?: Maybe<Dependencies>;
  description?: Maybe<Scalars['String']>;
  editable?: Maybe<Scalars['Boolean']>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  installation?: Maybe<TerraformInstallation>;
  latestVersion?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  package?: Maybe<Scalars['String']>;
  readme?: Maybe<Scalars['String']>;
  repository?: Maybe<Repository>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  valuesTemplate?: Maybe<Scalars['String']>;
};

export type TerraformInstallation = {
  __typename?: 'TerraformInstallation';
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  installation?: Maybe<Installation>;
  terraform?: Maybe<Terraform>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  version?: Maybe<Version>;
};

export type RecipeTest = {
  __typename?: 'RecipeTest';
  args?: Maybe<Array<Maybe<TestArgument>>>;
  message?: Maybe<Scalars['String']>;
  name: Scalars['String'];
  type: TestType;
};

export type TestArgument = {
  __typename?: 'TestArgument';
  key: Scalars['String'];
  name: Scalars['String'];
  repo: Scalars['String'];
};

export enum TestType {
  Git = 'GIT'
}

export type Tag = {
  __typename?: 'Tag';
  id: Scalars['ID'];
  tag: Scalars['String'];
};

export type Roles = {
  __typename?: 'Roles';
  admin?: Maybe<Scalars['Boolean']>;
};

export type PlatformSubscription = {
  __typename?: 'PlatformSubscription';
  externalId?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  lineItems?: Maybe<Array<Maybe<PlatformSubscriptionLineItems>>>;
  plan?: Maybe<PlatformPlan>;
};

export type PlatformSubscriptionLineItems = {
  __typename?: 'PlatformSubscriptionLineItems';
  dimension: LineItemDimension;
  externalId?: Maybe<Scalars['String']>;
  quantity: Scalars['Int'];
};

export enum LineItemDimension {
  Cluster = 'CLUSTER',
  User = 'USER'
}

export type PlatformPlan = {
  __typename?: 'PlatformPlan';
  cost: Scalars['Int'];
  enterprise?: Maybe<Scalars['Boolean']>;
  features?: Maybe<PlanFeatures>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  lineItems?: Maybe<Array<Maybe<PlatformPlanItem>>>;
  name: Scalars['String'];
  period: PaymentPeriod;
  updatedAt?: Maybe<Scalars['DateTime']>;
  visible: Scalars['Boolean'];
};

export type PlatformPlanItem = {
  __typename?: 'PlatformPlanItem';
  cost: Scalars['Int'];
  dimension: LineItemDimension;
  externalId?: Maybe<Scalars['String']>;
  name: Scalars['String'];
  period: PaymentPeriod;
};

export enum PaymentPeriod {
  Monthly = 'MONTHLY',
  Yearly = 'YEARLY'
}

export type GeoMetric = {
  __typename?: 'GeoMetric';
  count?: Maybe<Scalars['Int']>;
  country?: Maybe<Scalars['String']>;
};

export type AuditConnection = {
  __typename?: 'AuditConnection';
  edges?: Maybe<Array<Maybe<AuditEdge>>>;
  pageInfo: PageInfo;
};

export type AuditEdge = {
  __typename?: 'AuditEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Audit>;
};

export type Audit = {
  __typename?: 'Audit';
  action: Scalars['String'];
  actor?: Maybe<User>;
  city?: Maybe<Scalars['String']>;
  country?: Maybe<Scalars['String']>;
  group?: Maybe<Group>;
  id: Scalars['ID'];
  image?: Maybe<DockerImage>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  integrationWebhook?: Maybe<IntegrationWebhook>;
  ip?: Maybe<Scalars['String']>;
  latitude?: Maybe<Scalars['String']>;
  longitude?: Maybe<Scalars['String']>;
  repository?: Maybe<Repository>;
  role?: Maybe<Role>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
  version?: Maybe<Version>;
};

export type IntegrationWebhook = {
  __typename?: 'IntegrationWebhook';
  account?: Maybe<Account>;
  actions?: Maybe<Array<Maybe<Scalars['String']>>>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  logs?: Maybe<WebhookLogConnection>;
  name: Scalars['String'];
  secret: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
  url: Scalars['String'];
};


export type IntegrationWebhookLogsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type WebhookLogConnection = {
  __typename?: 'WebhookLogConnection';
  edges?: Maybe<Array<Maybe<WebhookLogEdge>>>;
  pageInfo: PageInfo;
};

export type WebhookLogEdge = {
  __typename?: 'WebhookLogEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<WebhookLog>;
};

export type WebhookLog = {
  __typename?: 'WebhookLog';
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  payload?: Maybe<Scalars['Map']>;
  response?: Maybe<Scalars['String']>;
  state: WebhookLogState;
  status?: Maybe<Scalars['Int']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  webhook?: Maybe<IntegrationWebhook>;
};

export enum WebhookLogState {
  Delivered = 'DELIVERED',
  Failed = 'FAILED',
  Sending = 'SENDING'
}

export type CategoryInfo = {
  __typename?: 'CategoryInfo';
  category?: Maybe<Category>;
  count?: Maybe<Scalars['Int']>;
  tags?: Maybe<GroupedTagConnection>;
};


export type CategoryInfoTagsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  q?: InputMaybe<Scalars['String']>;
};

export type GroupedTagConnection = {
  __typename?: 'GroupedTagConnection';
  edges?: Maybe<Array<Maybe<GroupedTagEdge>>>;
  pageInfo: PageInfo;
};

export type GroupedTagEdge = {
  __typename?: 'GroupedTagEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<GroupedTag>;
};

export type GroupedTag = {
  __typename?: 'GroupedTag';
  count: Scalars['Int'];
  tag: Scalars['String'];
};

export type ChartInstallationConnection = {
  __typename?: 'ChartInstallationConnection';
  edges?: Maybe<Array<Maybe<ChartInstallationEdge>>>;
  pageInfo: PageInfo;
};

export type ChartInstallationEdge = {
  __typename?: 'ChartInstallationEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<ChartInstallation>;
};

export type ChartConnection = {
  __typename?: 'ChartConnection';
  edges?: Maybe<Array<Maybe<ChartEdge>>>;
  pageInfo: PageInfo;
};

export type ChartEdge = {
  __typename?: 'ChartEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Chart>;
};

export type ClosureItem = {
  __typename?: 'ClosureItem';
  dep?: Maybe<Dependency>;
  helm?: Maybe<Chart>;
  terraform?: Maybe<Terraform>;
};

/** A Kubernetes cluster that can be used to deploy applications on with Plural. */
export type Cluster = {
  __typename?: 'Cluster';
  /** The account that the cluster belongs to. */
  account?: Maybe<Account>;
  /** The URL of the console running on the cluster. */
  consoleUrl?: Maybe<Scalars['String']>;
  /** The domain name used for applications deployed on the cluster. */
  domain?: Maybe<Scalars['String']>;
  /** The git repository URL for the cluster. */
  gitUrl?: Maybe<Scalars['String']>;
  /** The ID of the cluster. */
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  /** The name of the cluster. */
  name: Scalars['String'];
  /** The user that owns the cluster. */
  owner?: Maybe<User>;
  /** The last time the cluster was pinged. */
  pingedAt?: Maybe<Scalars['DateTime']>;
  /** The cluster's cloud provider. */
  provider: Provider;
  /** The upgrade queue for applications running on the cluster. */
  queue?: Maybe<UpgradeQueue>;
  /** The source of the cluster. */
  source?: Maybe<Source>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  /** pending upgrades for each installed app */
  upgradeInfo?: Maybe<Array<Maybe<UpgradeInfo>>>;
};

export type UpgradeQueue = {
  __typename?: 'UpgradeQueue';
  acked?: Maybe<Scalars['ID']>;
  domain?: Maybe<Scalars['String']>;
  git?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name?: Maybe<Scalars['String']>;
  pingedAt?: Maybe<Scalars['DateTime']>;
  provider?: Maybe<Provider>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  upgrades?: Maybe<UpgradeConnection>;
  user: User;
};


export type UpgradeQueueUpgradesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type UpgradeConnection = {
  __typename?: 'UpgradeConnection';
  edges?: Maybe<Array<Maybe<UpgradeEdge>>>;
  pageInfo: PageInfo;
};

export type UpgradeEdge = {
  __typename?: 'UpgradeEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Upgrade>;
};

export type Upgrade = {
  __typename?: 'Upgrade';
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  message?: Maybe<Scalars['String']>;
  repository?: Maybe<Repository>;
  type?: Maybe<UpgradeType>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum UpgradeType {
  Approval = 'APPROVAL',
  Bounce = 'BOUNCE',
  Dedicated = 'DEDICATED',
  Deploy = 'DEPLOY'
}

/** Possible cluster sources. */
export enum Source {
  Default = 'DEFAULT',
  Demo = 'DEMO',
  Shell = 'SHELL'
}

/** The pending upgrades for a repository */
export type UpgradeInfo = {
  __typename?: 'UpgradeInfo';
  count?: Maybe<Scalars['Int']>;
  installation?: Maybe<Installation>;
};

export type ClusterConnection = {
  __typename?: 'ClusterConnection';
  edges?: Maybe<Array<Maybe<ClusterEdge>>>;
  pageInfo: PageInfo;
};

export type ClusterEdge = {
  __typename?: 'ClusterEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Cluster>;
};

export type PluralConfiguration = {
  __typename?: 'PluralConfiguration';
  gitCommit?: Maybe<Scalars['String']>;
  registry?: Maybe<Scalars['String']>;
  stripeConnectId?: Maybe<Scalars['String']>;
  stripePublishableKey?: Maybe<Scalars['String']>;
};

export type DeferredUpdateConnection = {
  __typename?: 'DeferredUpdateConnection';
  edges?: Maybe<Array<Maybe<DeferredUpdateEdge>>>;
  pageInfo: PageInfo;
};

export type DeferredUpdateEdge = {
  __typename?: 'DeferredUpdateEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<DeferredUpdate>;
};

export type DeferredUpdate = {
  __typename?: 'DeferredUpdate';
  attempts?: Maybe<Scalars['Int']>;
  chartInstallation?: Maybe<ChartInstallation>;
  dequeueAt?: Maybe<Scalars['DateTime']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  messages?: Maybe<Array<Maybe<DeferredReason>>>;
  pending?: Maybe<Scalars['Boolean']>;
  terraformInstallation?: Maybe<TerraformInstallation>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  version?: Maybe<Version>;
};

export type DeferredReason = {
  __typename?: 'DeferredReason';
  message?: Maybe<Scalars['String']>;
  package?: Maybe<Scalars['String']>;
  repository?: Maybe<Scalars['String']>;
};

export type DemoProject = {
  __typename?: 'DemoProject';
  credentials?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  projectId: Scalars['String'];
  ready?: Maybe<Scalars['Boolean']>;
  state?: Maybe<DemoProjectState>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum DemoProjectState {
  Created = 'CREATED',
  Enabled = 'ENABLED',
  Ready = 'READY'
}

export type DnsDomain = {
  __typename?: 'DnsDomain';
  accessPolicy?: Maybe<DnsAccessPolicy>;
  account?: Maybe<Account>;
  creator?: Maybe<User>;
  dnsRecords?: Maybe<DnsRecordConnection>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};


export type DnsDomainDnsRecordsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type DnsAccessPolicy = {
  __typename?: 'DnsAccessPolicy';
  bindings?: Maybe<Array<Maybe<PolicyBinding>>>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type PolicyBinding = {
  __typename?: 'PolicyBinding';
  group?: Maybe<Group>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type DnsRecordConnection = {
  __typename?: 'DnsRecordConnection';
  edges?: Maybe<Array<Maybe<DnsRecordEdge>>>;
  pageInfo: PageInfo;
};

export type DnsRecordEdge = {
  __typename?: 'DnsRecordEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<DnsRecord>;
};

export type DnsRecord = {
  __typename?: 'DnsRecord';
  cluster: Scalars['String'];
  creator?: Maybe<User>;
  domain?: Maybe<DnsDomain>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  provider: Provider;
  records?: Maybe<Array<Maybe<Scalars['String']>>>;
  type: DnsRecordType;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum DnsRecordType {
  A = 'A',
  Aaaa = 'AAAA',
  Cname = 'CNAME',
  Txt = 'TXT'
}

export type DnsDomainConnection = {
  __typename?: 'DnsDomainConnection';
  edges?: Maybe<Array<Maybe<DnsDomainEdge>>>;
  pageInfo: PageInfo;
};

export type DnsDomainEdge = {
  __typename?: 'DnsDomainEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<DnsDomain>;
};

export type DockerImageConnection = {
  __typename?: 'DockerImageConnection';
  edges?: Maybe<Array<Maybe<DockerImageEdge>>>;
  pageInfo: PageInfo;
};

export type DockerImageEdge = {
  __typename?: 'DockerImageEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<DockerImage>;
};

export type DockerRepositoryConnection = {
  __typename?: 'DockerRepositoryConnection';
  edges?: Maybe<Array<Maybe<DockerRepositoryEdge>>>;
  pageInfo: PageInfo;
};

export type DockerRepositoryEdge = {
  __typename?: 'DockerRepositoryEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<DockerRepository>;
};

export type EabCredential = {
  __typename?: 'EabCredential';
  cluster: Scalars['String'];
  hmacKey: Scalars['String'];
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  keyId: Scalars['String'];
  provider: Provider;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type GroupMemberConnection = {
  __typename?: 'GroupMemberConnection';
  edges?: Maybe<Array<Maybe<GroupMemberEdge>>>;
  pageInfo: PageInfo;
};

export type GroupMemberEdge = {
  __typename?: 'GroupMemberEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<GroupMember>;
};

export type GroupMember = {
  __typename?: 'GroupMember';
  group?: Maybe<Group>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type GroupConnection = {
  __typename?: 'GroupConnection';
  edges?: Maybe<Array<Maybe<GroupEdge>>>;
  pageInfo: PageInfo;
};

export type GroupEdge = {
  __typename?: 'GroupEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Group>;
};

export type Incident = {
  __typename?: 'Incident';
  clusterInformation?: Maybe<ClusterInformation>;
  creator: User;
  description?: Maybe<Scalars['String']>;
  files?: Maybe<FileConnection>;
  follower?: Maybe<Follower>;
  followers?: Maybe<FollowerConnection>;
  history?: Maybe<IncidentHistoryConnection>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  messages?: Maybe<IncidentMessageConnection>;
  nextResponseAt?: Maybe<Scalars['DateTime']>;
  notificationCount?: Maybe<Scalars['Int']>;
  owner?: Maybe<User>;
  postmortem?: Maybe<Postmortem>;
  repository: Repository;
  severity: Scalars['Int'];
  status: IncidentStatus;
  subscription?: Maybe<SlimSubscription>;
  tags?: Maybe<Array<Maybe<Tag>>>;
  title: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};


export type IncidentFilesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type IncidentFollowersArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type IncidentHistoryArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type IncidentMessagesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type ClusterInformation = {
  __typename?: 'ClusterInformation';
  gitCommit?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  platform?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  version?: Maybe<Scalars['String']>;
};

export type FileConnection = {
  __typename?: 'FileConnection';
  edges?: Maybe<Array<Maybe<FileEdge>>>;
  pageInfo: PageInfo;
};

export type FileEdge = {
  __typename?: 'FileEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<File>;
};

export type File = {
  __typename?: 'File';
  blob: Scalars['String'];
  contentType?: Maybe<Scalars['String']>;
  filename?: Maybe<Scalars['String']>;
  filesize?: Maybe<Scalars['Int']>;
  height?: Maybe<Scalars['Int']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  mediaType?: Maybe<MediaType>;
  message: IncidentMessage;
  updatedAt?: Maybe<Scalars['DateTime']>;
  width?: Maybe<Scalars['Int']>;
};

export enum MediaType {
  Audio = 'AUDIO',
  Image = 'IMAGE',
  Other = 'OTHER',
  Pdf = 'PDF',
  Video = 'VIDEO'
}

export type IncidentMessage = {
  __typename?: 'IncidentMessage';
  creator: User;
  entities?: Maybe<Array<Maybe<MessageEntity>>>;
  file?: Maybe<File>;
  id: Scalars['ID'];
  incident: Incident;
  insertedAt?: Maybe<Scalars['DateTime']>;
  reactions?: Maybe<Array<Maybe<Reaction>>>;
  text: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type MessageEntity = {
  __typename?: 'MessageEntity';
  endIndex?: Maybe<Scalars['Int']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  startIndex?: Maybe<Scalars['Int']>;
  text?: Maybe<Scalars['String']>;
  type: MessageEntityType;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export enum MessageEntityType {
  Emoji = 'EMOJI',
  Mention = 'MENTION'
}

export type Reaction = {
  __typename?: 'Reaction';
  creator: User;
  insertedAt?: Maybe<Scalars['DateTime']>;
  message: IncidentMessage;
  name: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type Follower = {
  __typename?: 'Follower';
  id: Scalars['ID'];
  incident?: Maybe<Incident>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  preferences?: Maybe<NotificationPreferences>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user: User;
};

export type NotificationPreferences = {
  __typename?: 'NotificationPreferences';
  incidentUpdate?: Maybe<Scalars['Boolean']>;
  mention?: Maybe<Scalars['Boolean']>;
  message?: Maybe<Scalars['Boolean']>;
};

export type FollowerConnection = {
  __typename?: 'FollowerConnection';
  edges?: Maybe<Array<Maybe<FollowerEdge>>>;
  pageInfo: PageInfo;
};

export type FollowerEdge = {
  __typename?: 'FollowerEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Follower>;
};

export type IncidentHistoryConnection = {
  __typename?: 'IncidentHistoryConnection';
  edges?: Maybe<Array<Maybe<IncidentHistoryEdge>>>;
  pageInfo: PageInfo;
};

export type IncidentHistoryEdge = {
  __typename?: 'IncidentHistoryEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<IncidentHistory>;
};

export type IncidentHistory = {
  __typename?: 'IncidentHistory';
  action: IncidentAction;
  actor: User;
  changes?: Maybe<Array<Maybe<IncidentChange>>>;
  id: Scalars['ID'];
  incident: Incident;
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum IncidentAction {
  Accept = 'ACCEPT',
  Complete = 'COMPLETE',
  Create = 'CREATE',
  Edit = 'EDIT',
  Severity = 'SEVERITY',
  Status = 'STATUS'
}

export type IncidentChange = {
  __typename?: 'IncidentChange';
  key: Scalars['String'];
  next?: Maybe<Scalars['String']>;
  prev?: Maybe<Scalars['String']>;
};

export type IncidentMessageConnection = {
  __typename?: 'IncidentMessageConnection';
  edges?: Maybe<Array<Maybe<IncidentMessageEdge>>>;
  pageInfo: PageInfo;
};

export type IncidentMessageEdge = {
  __typename?: 'IncidentMessageEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<IncidentMessage>;
};

export type Postmortem = {
  __typename?: 'Postmortem';
  actionItems?: Maybe<Array<Maybe<ActionItem>>>;
  content: Scalars['String'];
  creator: User;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type ActionItem = {
  __typename?: 'ActionItem';
  link: Scalars['String'];
  type: ActionItemType;
};

export enum ActionItemType {
  Blog = 'BLOG',
  Issue = 'ISSUE',
  Pull = 'PULL'
}

export enum IncidentStatus {
  Complete = 'COMPLETE',
  InProgress = 'IN_PROGRESS',
  Open = 'OPEN',
  Resolved = 'RESOLVED'
}

export type SlimSubscription = {
  __typename?: 'SlimSubscription';
  id: Scalars['ID'];
  lineItems?: Maybe<SubscriptionLineItems>;
  plan?: Maybe<Plan>;
};

export type IncidentFilter = {
  statuses?: InputMaybe<Array<InputMaybe<IncidentStatus>>>;
  type: IncidentFilterType;
  value?: InputMaybe<Scalars['String']>;
};

export enum IncidentFilterType {
  Following = 'FOLLOWING',
  Notifications = 'NOTIFICATIONS',
  Status = 'STATUS',
  Tag = 'TAG'
}

export enum Order {
  Asc = 'ASC',
  Desc = 'DESC'
}

export enum IncidentSort {
  InsertedAt = 'INSERTED_AT',
  Severity = 'SEVERITY',
  Status = 'STATUS',
  Title = 'TITLE'
}

export type IncidentConnection = {
  __typename?: 'IncidentConnection';
  edges?: Maybe<Array<Maybe<IncidentEdge>>>;
  pageInfo: PageInfo;
};

export type IncidentEdge = {
  __typename?: 'IncidentEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Incident>;
};

export type InstallationConnection = {
  __typename?: 'InstallationConnection';
  edges?: Maybe<Array<Maybe<InstallationEdge>>>;
  pageInfo: PageInfo;
};

export type InstallationEdge = {
  __typename?: 'InstallationEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Installation>;
};

export type IntegrationConnection = {
  __typename?: 'IntegrationConnection';
  edges?: Maybe<Array<Maybe<IntegrationEdge>>>;
  pageInfo: PageInfo;
};

export type IntegrationEdge = {
  __typename?: 'IntegrationEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Integration>;
};

export type Integration = {
  __typename?: 'Integration';
  description?: Maybe<Scalars['String']>;
  icon?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  publisher?: Maybe<Publisher>;
  repository?: Maybe<Repository>;
  sourceUrl?: Maybe<Scalars['String']>;
  spec?: Maybe<Scalars['Map']>;
  tags?: Maybe<Array<Maybe<Tag>>>;
  type?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type IntegrationWebhookConnection = {
  __typename?: 'IntegrationWebhookConnection';
  edges?: Maybe<Array<Maybe<IntegrationWebhookEdge>>>;
  pageInfo: PageInfo;
};

export type IntegrationWebhookEdge = {
  __typename?: 'IntegrationWebhookEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<IntegrationWebhook>;
};

export type Invite = {
  __typename?: 'Invite';
  account?: Maybe<Account>;
  email?: Maybe<Scalars['String']>;
  existing: Scalars['Boolean'];
  expiresAt?: Maybe<Scalars['DateTime']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  secureId?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type InviteConnection = {
  __typename?: 'InviteConnection';
  edges?: Maybe<Array<Maybe<InviteEdge>>>;
  pageInfo: PageInfo;
};

export type InviteEdge = {
  __typename?: 'InviteEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Invite>;
};

export type KeyBackup = {
  __typename?: 'KeyBackup';
  digest: Scalars['String'];
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  repositories?: Maybe<Array<Scalars['String']>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user: User;
  value: Scalars['String'];
};

export type KeyBackupConnection = {
  __typename?: 'KeyBackupConnection';
  edges?: Maybe<Array<Maybe<KeyBackupEdge>>>;
  pageInfo: PageInfo;
};

export type KeyBackupEdge = {
  __typename?: 'KeyBackupEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<KeyBackup>;
};

export type LoginMethodResponse = {
  __typename?: 'LoginMethodResponse';
  authorizeUrl?: Maybe<Scalars['String']>;
  loginMethod: LoginMethod;
  token?: Maybe<Scalars['String']>;
};

export type NotificationConnection = {
  __typename?: 'NotificationConnection';
  edges?: Maybe<Array<Maybe<NotificationEdge>>>;
  pageInfo: PageInfo;
};

export type NotificationEdge = {
  __typename?: 'NotificationEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Notification>;
};

export type Notification = {
  __typename?: 'Notification';
  actor: User;
  id: Scalars['ID'];
  incident?: Maybe<Incident>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  message?: Maybe<IncidentMessage>;
  msg?: Maybe<Scalars['String']>;
  repository?: Maybe<Repository>;
  type: NotificationType;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user: User;
};

export enum NotificationType {
  IncidentUpdate = 'INCIDENT_UPDATE',
  Locked = 'LOCKED',
  Mention = 'MENTION',
  Message = 'MESSAGE'
}

export type OauthIntegration = {
  __typename?: 'OauthIntegration';
  account?: Maybe<Account>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  service: OauthService;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum OauthService {
  Zoom = 'ZOOM'
}

export type OauthInfo = {
  __typename?: 'OauthInfo';
  authorizeUrl: Scalars['String'];
  provider: OauthProvider;
};

export enum OauthProvider {
  Github = 'GITHUB',
  Gitlab = 'GITLAB',
  Google = 'GOOGLE'
}

export type OidcStepResponse = {
  __typename?: 'OidcStepResponse';
  consent?: Maybe<ConsentRequest>;
  login?: Maybe<LoginRequest>;
  repository?: Maybe<Repository>;
};

export type LoginRequest = {
  __typename?: 'LoginRequest';
  requestedScope?: Maybe<Array<Maybe<Scalars['String']>>>;
  subject?: Maybe<Scalars['String']>;
};

export type OidcLoginConnection = {
  __typename?: 'OidcLoginConnection';
  edges?: Maybe<Array<Maybe<OidcLoginEdge>>>;
  pageInfo: PageInfo;
};

export type OidcLoginEdge = {
  __typename?: 'OidcLoginEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<OidcLogin>;
};

export type OidcLogin = {
  __typename?: 'OidcLogin';
  city?: Maybe<Scalars['String']>;
  country?: Maybe<Scalars['String']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  ip?: Maybe<Scalars['String']>;
  latitude?: Maybe<Scalars['String']>;
  longitude?: Maybe<Scalars['String']>;
  owner?: Maybe<User>;
  repository?: Maybe<Repository>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user?: Maybe<User>;
};

export type PlatformMetrics = {
  __typename?: 'PlatformMetrics';
  clusters?: Maybe<Scalars['Int']>;
  publishers?: Maybe<Scalars['Int']>;
  repositories?: Maybe<Scalars['Int']>;
  rollouts?: Maybe<Scalars['Int']>;
};

export type PublicKeyConnection = {
  __typename?: 'PublicKeyConnection';
  edges?: Maybe<Array<Maybe<PublicKeyEdge>>>;
  pageInfo: PageInfo;
};

export type PublicKeyEdge = {
  __typename?: 'PublicKeyEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<PublicKey>;
};

export type PublicKey = {
  __typename?: 'PublicKey';
  content: Scalars['String'];
  digest: Scalars['String'];
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
  user: User;
};

export type PublisherConnection = {
  __typename?: 'PublisherConnection';
  edges?: Maybe<Array<Maybe<PublisherEdge>>>;
  pageInfo: PageInfo;
};

export type PublisherEdge = {
  __typename?: 'PublisherEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Publisher>;
};

export type RecipeConnection = {
  __typename?: 'RecipeConnection';
  edges?: Maybe<Array<Maybe<RecipeEdge>>>;
  pageInfo: PageInfo;
};

export type RecipeEdge = {
  __typename?: 'RecipeEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Recipe>;
};

export type RepositoryConnection = {
  __typename?: 'RepositoryConnection';
  edges?: Maybe<Array<Maybe<RepositoryEdge>>>;
  pageInfo: PageInfo;
};

export type RepositoryEdge = {
  __typename?: 'RepositoryEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Repository>;
};

export type ResetToken = {
  __typename?: 'ResetToken';
  email: Scalars['String'];
  externalId: Scalars['ID'];
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  type: ResetTokenType;
  updatedAt?: Maybe<Scalars['DateTime']>;
  user: User;
};

export enum ResetTokenType {
  Email = 'EMAIL',
  Password = 'PASSWORD'
}

export type RoleConnection = {
  __typename?: 'RoleConnection';
  edges?: Maybe<Array<Maybe<RoleEdge>>>;
  pageInfo: PageInfo;
};

export type RoleEdge = {
  __typename?: 'RoleEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Role>;
};

export type RolloutConnection = {
  __typename?: 'RolloutConnection';
  edges?: Maybe<Array<Maybe<RolloutEdge>>>;
  pageInfo: PageInfo;
};

export type RolloutEdge = {
  __typename?: 'RolloutEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Rollout>;
};

export type Rollout = {
  __typename?: 'Rollout';
  count?: Maybe<Scalars['Int']>;
  cursor?: Maybe<Scalars['ID']>;
  event?: Maybe<Scalars['String']>;
  heartbeat?: Maybe<Scalars['DateTime']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  repository?: Maybe<Repository>;
  status: RolloutStatus;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum RolloutStatus {
  Finished = 'FINISHED',
  Queued = 'QUEUED',
  Running = 'RUNNING'
}

export type ScaffoldFile = {
  __typename?: 'ScaffoldFile';
  content?: Maybe<Scalars['String']>;
  path?: Maybe<Scalars['String']>;
};

export type AuthorizationUrl = {
  __typename?: 'AuthorizationUrl';
  provider: ScmProvider;
  url: Scalars['String'];
};

export enum ScmProvider {
  Demo = 'DEMO',
  Github = 'GITHUB',
  Gitlab = 'GITLAB',
  Manual = 'MANUAL'
}

export type UserConnection = {
  __typename?: 'UserConnection';
  edges?: Maybe<Array<Maybe<UserEdge>>>;
  pageInfo: PageInfo;
};

export type UserEdge = {
  __typename?: 'UserEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<User>;
};

export type CloudShell = {
  __typename?: 'CloudShell';
  aesKey: Scalars['String'];
  alive: Scalars['Boolean'];
  cluster: Scalars['String'];
  gitUrl: Scalars['String'];
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  missing?: Maybe<Array<Maybe<Scalars['String']>>>;
  provider: Provider;
  region: Scalars['String'];
  status?: Maybe<ShellStatus>;
  subdomain: Scalars['String'];
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type ShellStatus = {
  __typename?: 'ShellStatus';
  containersReady?: Maybe<Scalars['Boolean']>;
  initialized?: Maybe<Scalars['Boolean']>;
  podScheduled?: Maybe<Scalars['Boolean']>;
  ready?: Maybe<Scalars['Boolean']>;
};

export type ShellConfiguration = {
  __typename?: 'ShellConfiguration';
  buckets?: Maybe<Array<Maybe<Scalars['String']>>>;
  contextConfiguration?: Maybe<Scalars['Map']>;
  domains?: Maybe<Array<Maybe<Scalars['String']>>>;
  git?: Maybe<GitConfiguration>;
  workspace?: Maybe<ShellWorkspace>;
};

export type GitConfiguration = {
  __typename?: 'GitConfiguration';
  branch?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  root?: Maybe<Scalars['String']>;
  url?: Maybe<Scalars['String']>;
};

export type ShellWorkspace = {
  __typename?: 'ShellWorkspace';
  bucketPrefix?: Maybe<Scalars['String']>;
  cluster?: Maybe<Scalars['String']>;
  network?: Maybe<NetworkConfiguration>;
  region?: Maybe<Scalars['String']>;
};

export type NetworkConfiguration = {
  __typename?: 'NetworkConfiguration';
  pluralDns?: Maybe<Scalars['Boolean']>;
  subdomain?: Maybe<Scalars['String']>;
};

export type Stack = {
  __typename?: 'Stack';
  bundles?: Maybe<Array<Maybe<Recipe>>>;
  collections?: Maybe<Array<Maybe<StackCollection>>>;
  community?: Maybe<Community>;
  creator?: Maybe<User>;
  description?: Maybe<Scalars['String']>;
  displayName?: Maybe<Scalars['String']>;
  featured?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  sections?: Maybe<Array<Maybe<RecipeSection>>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type StackCollection = {
  __typename?: 'StackCollection';
  bundles?: Maybe<Array<Maybe<StackRecipe>>>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  provider: Provider;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type StackRecipe = {
  __typename?: 'StackRecipe';
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  recipe: Recipe;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type StackConnection = {
  __typename?: 'StackConnection';
  edges?: Maybe<Array<Maybe<StackEdge>>>;
  pageInfo: PageInfo;
};

export type StackEdge = {
  __typename?: 'StackEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Stack>;
};

export type RepositorySubscriptionConnection = {
  __typename?: 'RepositorySubscriptionConnection';
  edges?: Maybe<Array<Maybe<RepositorySubscriptionEdge>>>;
  pageInfo: PageInfo;
};

export type RepositorySubscriptionEdge = {
  __typename?: 'RepositorySubscriptionEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<RepositorySubscription>;
};

export enum TagGroup {
  Integrations = 'INTEGRATIONS',
  Repositories = 'REPOSITORIES'
}

export type TerraformConnection = {
  __typename?: 'TerraformConnection';
  edges?: Maybe<Array<Maybe<TerraformEdge>>>;
  pageInfo: PageInfo;
};

export type TerraformEdge = {
  __typename?: 'TerraformEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Terraform>;
};

export type TerraformInstallationConnection = {
  __typename?: 'TerraformInstallationConnection';
  edges?: Maybe<Array<Maybe<TerraformInstallationEdge>>>;
  pageInfo: PageInfo;
};

export type TerraformInstallationEdge = {
  __typename?: 'TerraformInstallationEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<TerraformInstallation>;
};

export type TerraformProvider = {
  __typename?: 'TerraformProvider';
  content?: Maybe<Scalars['String']>;
  name?: Maybe<Provider>;
};

export type Test = {
  __typename?: 'Test';
  creator?: Maybe<User>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name?: Maybe<Scalars['String']>;
  promoteTag: Scalars['String'];
  repository?: Maybe<Repository>;
  sourceTag: Scalars['String'];
  status: TestStatus;
  steps?: Maybe<Array<Maybe<TestStep>>>;
  tags?: Maybe<Array<Scalars['String']>>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export enum TestStatus {
  Failed = 'FAILED',
  Queued = 'QUEUED',
  Running = 'RUNNING',
  Succeeded = 'SUCCEEDED'
}

export type TestStep = {
  __typename?: 'TestStep';
  description: Scalars['String'];
  hasLogs?: Maybe<Scalars['Boolean']>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  name: Scalars['String'];
  status: TestStatus;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type TestConnection = {
  __typename?: 'TestConnection';
  edges?: Maybe<Array<Maybe<TestEdge>>>;
  pageInfo: PageInfo;
};

export type TestEdge = {
  __typename?: 'TestEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Test>;
};

export type PersistedToken = {
  __typename?: 'PersistedToken';
  audits?: Maybe<PersistedTokenAuditConnection>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  metrics?: Maybe<Array<Maybe<GeoMetric>>>;
  token?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};


export type PersistedTokenAuditsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

export type PersistedTokenAuditConnection = {
  __typename?: 'PersistedTokenAuditConnection';
  edges?: Maybe<Array<Maybe<PersistedTokenAuditEdge>>>;
  pageInfo: PageInfo;
};

export type PersistedTokenAuditEdge = {
  __typename?: 'PersistedTokenAuditEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<PersistedTokenAudit>;
};

export type PersistedTokenAudit = {
  __typename?: 'PersistedTokenAudit';
  city?: Maybe<Scalars['String']>;
  count?: Maybe<Scalars['Int']>;
  country?: Maybe<Scalars['String']>;
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  ip?: Maybe<Scalars['String']>;
  latitude?: Maybe<Scalars['String']>;
  longitude?: Maybe<Scalars['String']>;
  timestamp?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type PersistedTokenConnection = {
  __typename?: 'PersistedTokenConnection';
  edges?: Maybe<Array<Maybe<PersistedTokenEdge>>>;
  pageInfo: PageInfo;
};

export type PersistedTokenEdge = {
  __typename?: 'PersistedTokenEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<PersistedToken>;
};

export type VersionConnection = {
  __typename?: 'VersionConnection';
  edges?: Maybe<Array<Maybe<VersionEdge>>>;
  pageInfo: PageInfo;
};

export type VersionEdge = {
  __typename?: 'VersionEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Version>;
};

export type WebhookConnection = {
  __typename?: 'WebhookConnection';
  edges?: Maybe<Array<Maybe<WebhookEdge>>>;
  pageInfo: PageInfo;
};

export type WebhookEdge = {
  __typename?: 'WebhookEdge';
  cursor?: Maybe<Scalars['String']>;
  node?: Maybe<Webhook>;
};

export type Webhook = {
  __typename?: 'Webhook';
  id?: Maybe<Scalars['ID']>;
  insertedAt?: Maybe<Scalars['DateTime']>;
  secret?: Maybe<Scalars['String']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
  url?: Maybe<Scalars['String']>;
  user?: Maybe<User>;
};

export type RootMutationType = {
  __typename?: 'RootMutationType';
  acceptIncident?: Maybe<Incident>;
  acceptLogin?: Maybe<OauthResponse>;
  acquireLock?: Maybe<ApplyLock>;
  cancelPlatformSubscription?: Maybe<PlatformSubscription>;
  completeIncident?: Maybe<Incident>;
  createArtifact?: Maybe<Artifact>;
  createCard?: Maybe<Account>;
  /** Create a new cluster. */
  createCluster?: Maybe<Cluster>;
  /** adds a dependency for this cluster to gate future upgrades */
  createClusterDependency?: Maybe<ClusterDependency>;
  createCrd?: Maybe<Crd>;
  createDemoProject?: Maybe<DemoProject>;
  createDnsRecord?: Maybe<DnsRecord>;
  createDomain?: Maybe<DnsDomain>;
  createGroup?: Maybe<Group>;
  createGroupMember?: Maybe<GroupMember>;
  createIncident?: Maybe<Incident>;
  createInstallation?: Maybe<Installation>;
  createIntegration?: Maybe<Integration>;
  createIntegrationWebhook?: Maybe<IntegrationWebhook>;
  createInvite?: Maybe<Invite>;
  createKeyBackup?: Maybe<KeyBackup>;
  createMessage?: Maybe<IncidentMessage>;
  createOauthIntegration?: Maybe<OauthIntegration>;
  createOidcProvider?: Maybe<OidcProvider>;
  createPlan?: Maybe<Plan>;
  createPlatformSubscription?: Maybe<PlatformSubscription>;
  createPublicKey?: Maybe<PublicKey>;
  createPublisher?: Maybe<Publisher>;
  createQueue?: Maybe<UpgradeQueue>;
  createReaction?: Maybe<IncidentMessage>;
  createRecipe?: Maybe<Recipe>;
  createRepository?: Maybe<Repository>;
  createResetToken?: Maybe<Scalars['Boolean']>;
  createRole?: Maybe<Role>;
  createServiceAccount?: Maybe<User>;
  createShell?: Maybe<CloudShell>;
  createStack?: Maybe<Stack>;
  createSubscription?: Maybe<RepositorySubscription>;
  createTerraform?: Maybe<Terraform>;
  createTest?: Maybe<Test>;
  createToken?: Maybe<PersistedToken>;
  createUserEvent?: Maybe<Scalars['Boolean']>;
  createWebhook?: Maybe<Webhook>;
  createZoom?: Maybe<ZoomMeeting>;
  deleteCard?: Maybe<Account>;
  deleteChartInstallation?: Maybe<ChartInstallation>;
  /** Delete a cluster. */
  deleteCluster?: Maybe<Cluster>;
  deleteDemoProject?: Maybe<DemoProject>;
  deleteDnsRecord?: Maybe<DnsRecord>;
  deleteDomain?: Maybe<DnsDomain>;
  deleteEabKey?: Maybe<EabCredential>;
  deleteGroup?: Maybe<Group>;
  deleteGroupMember?: Maybe<GroupMember>;
  deleteIncident?: Maybe<Incident>;
  deleteInstallation?: Maybe<Installation>;
  deleteIntegrationWebhook?: Maybe<IntegrationWebhook>;
  deleteInvite?: Maybe<Invite>;
  deleteKeyBackup?: Maybe<KeyBackup>;
  deleteMessage?: Maybe<IncidentMessage>;
  deletePlatformSubscription?: Maybe<Account>;
  deletePublicKey?: Maybe<PublicKey>;
  deleteReaction?: Maybe<IncidentMessage>;
  deleteRecipe?: Maybe<Recipe>;
  deleteRepository?: Maybe<Repository>;
  deleteRole?: Maybe<Role>;
  deleteShell?: Maybe<CloudShell>;
  deleteStack?: Maybe<Stack>;
  deleteTerraform?: Maybe<Terraform>;
  deleteToken?: Maybe<PersistedToken>;
  deleteUser?: Maybe<User>;
  destroyCluster?: Maybe<Scalars['Boolean']>;
  deviceLogin?: Maybe<DeviceLogin>;
  externalToken?: Maybe<Scalars['String']>;
  followIncident?: Maybe<Follower>;
  impersonateServiceAccount?: Maybe<User>;
  installBundle?: Maybe<Array<Maybe<Installation>>>;
  installChart?: Maybe<ChartInstallation>;
  installRecipe?: Maybe<Array<Maybe<Installation>>>;
  installStack?: Maybe<Array<Maybe<Recipe>>>;
  installStackShell?: Maybe<Array<Maybe<Recipe>>>;
  installTerraform?: Maybe<TerraformInstallation>;
  linkPublisher?: Maybe<Publisher>;
  login?: Maybe<User>;
  loginToken?: Maybe<User>;
  oauthCallback?: Maybe<User>;
  oauthConsent?: Maybe<OauthResponse>;
  passwordlessLogin?: Maybe<User>;
  pingWebhook?: Maybe<WebhookResponse>;
  /** moves up the upgrade waterline for a user */
  promote?: Maybe<User>;
  provisionDomain?: Maybe<DnsDomain>;
  publishLogs?: Maybe<TestStep>;
  quickStack?: Maybe<Stack>;
  readNotifications?: Maybe<Scalars['Int']>;
  realizeInvite?: Maybe<User>;
  realizeResetToken?: Maybe<Scalars['Boolean']>;
  rebootShell?: Maybe<CloudShell>;
  releaseLock?: Maybe<ApplyLock>;
  resetInstallations?: Maybe<Scalars['Int']>;
  restartShell?: Maybe<Scalars['Boolean']>;
  setupShell?: Maybe<CloudShell>;
  signup?: Maybe<User>;
  ssoCallback?: Maybe<User>;
  stopShell?: Maybe<Scalars['Boolean']>;
  transferDemoProject?: Maybe<DemoProject>;
  unfollowIncident?: Maybe<Follower>;
  uninstallTerraform?: Maybe<TerraformInstallation>;
  unlockRepository?: Maybe<Scalars['Int']>;
  updateAccount?: Maybe<Account>;
  updateChart?: Maybe<Chart>;
  updateChartInstallation?: Maybe<ChartInstallation>;
  updateDockerRepository?: Maybe<DockerRepository>;
  updateDomain?: Maybe<DnsDomain>;
  updateGroup?: Maybe<Group>;
  updateIncident?: Maybe<Incident>;
  updateInstallation?: Maybe<Installation>;
  updateIntegrationWebhook?: Maybe<IntegrationWebhook>;
  updateLineItem?: Maybe<RepositorySubscription>;
  updateMessage?: Maybe<IncidentMessage>;
  updateOidcProvider?: Maybe<OidcProvider>;
  updatePlan?: Maybe<RepositorySubscription>;
  updatePlanAttributes?: Maybe<Plan>;
  updatePlatformPlan?: Maybe<PlatformSubscription>;
  updatePublisher?: Maybe<Publisher>;
  updateRepository?: Maybe<Repository>;
  updateRole?: Maybe<Role>;
  updateServiceAccount?: Maybe<User>;
  updateShell?: Maybe<CloudShell>;
  updateShellConfiguration?: Maybe<Scalars['Boolean']>;
  updateStep?: Maybe<TestStep>;
  updateTerraform?: Maybe<Terraform>;
  updateTest?: Maybe<Test>;
  updateUser?: Maybe<User>;
  updateVersion?: Maybe<Version>;
  uploadTerraform?: Maybe<Terraform>;
  upsertOidcProvider?: Maybe<OidcProvider>;
  upsertRepository?: Maybe<Repository>;
};


export type RootMutationTypeAcceptIncidentArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeAcceptLoginArgs = {
  challenge: Scalars['String'];
};


export type RootMutationTypeAcquireLockArgs = {
  repository: Scalars['String'];
};


export type RootMutationTypeCompleteIncidentArgs = {
  id: Scalars['ID'];
  postmortem: PostmortemAttributes;
};


export type RootMutationTypeCreateArtifactArgs = {
  attributes: ArtifactAttributes;
  repositoryId?: InputMaybe<Scalars['ID']>;
  repositoryName?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeCreateCardArgs = {
  address?: InputMaybe<AddressAttributes>;
  source: Scalars['String'];
};


export type RootMutationTypeCreateClusterArgs = {
  attributes: ClusterAttributes;
};


export type RootMutationTypeCreateClusterDependencyArgs = {
  destId: Scalars['ID'];
  sourceId: Scalars['ID'];
};


export type RootMutationTypeCreateCrdArgs = {
  attributes: CrdAttributes;
  chartId?: InputMaybe<Scalars['ID']>;
  chartName?: InputMaybe<ChartName>;
};


export type RootMutationTypeCreateDnsRecordArgs = {
  attributes: DnsRecordAttributes;
  cluster: Scalars['String'];
  provider: Provider;
};


export type RootMutationTypeCreateDomainArgs = {
  attributes: DnsDomainAttributes;
};


export type RootMutationTypeCreateGroupArgs = {
  attributes: GroupAttributes;
};


export type RootMutationTypeCreateGroupMemberArgs = {
  groupId: Scalars['ID'];
  userId: Scalars['ID'];
};


export type RootMutationTypeCreateIncidentArgs = {
  attributes: IncidentAttributes;
  repository?: InputMaybe<Scalars['String']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeCreateInstallationArgs = {
  repositoryId: Scalars['ID'];
};


export type RootMutationTypeCreateIntegrationArgs = {
  attributes: IntegrationAttributes;
  repositoryName: Scalars['String'];
};


export type RootMutationTypeCreateIntegrationWebhookArgs = {
  attributes: IntegrationWebhookAttributes;
};


export type RootMutationTypeCreateInviteArgs = {
  attributes: InviteAttributes;
};


export type RootMutationTypeCreateKeyBackupArgs = {
  attributes: KeyBackupAttributes;
};


export type RootMutationTypeCreateMessageArgs = {
  attributes: IncidentMessageAttributes;
  incidentId: Scalars['ID'];
};


export type RootMutationTypeCreateOauthIntegrationArgs = {
  attributes: OauthAttributes;
};


export type RootMutationTypeCreateOidcProviderArgs = {
  attributes: OidcAttributes;
  installationId: Scalars['ID'];
};


export type RootMutationTypeCreatePlanArgs = {
  attributes: PlanAttributes;
  repositoryId: Scalars['ID'];
};


export type RootMutationTypeCreatePlatformSubscriptionArgs = {
  planId: Scalars['ID'];
};


export type RootMutationTypeCreatePublicKeyArgs = {
  attributes: PublicKeyAttributes;
};


export type RootMutationTypeCreatePublisherArgs = {
  attributes: PublisherAttributes;
};


export type RootMutationTypeCreateQueueArgs = {
  attributes: UpgradeQueueAttributes;
};


export type RootMutationTypeCreateReactionArgs = {
  messageId: Scalars['ID'];
  name: Scalars['String'];
};


export type RootMutationTypeCreateRecipeArgs = {
  attributes: RecipeAttributes;
  repositoryId?: InputMaybe<Scalars['String']>;
  repositoryName?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeCreateRepositoryArgs = {
  attributes: RepositoryAttributes;
  id?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeCreateResetTokenArgs = {
  attributes: ResetTokenAttributes;
};


export type RootMutationTypeCreateRoleArgs = {
  attributes: RoleAttributes;
};


export type RootMutationTypeCreateServiceAccountArgs = {
  attributes: ServiceAccountAttributes;
};


export type RootMutationTypeCreateShellArgs = {
  attributes: CloudShellAttributes;
};


export type RootMutationTypeCreateStackArgs = {
  attributes: StackAttributes;
};


export type RootMutationTypeCreateSubscriptionArgs = {
  attributes?: InputMaybe<SubscriptionAttributes>;
  installationId: Scalars['ID'];
  planId: Scalars['ID'];
};


export type RootMutationTypeCreateTerraformArgs = {
  attributes: TerraformAttributes;
  repositoryId: Scalars['ID'];
};


export type RootMutationTypeCreateTestArgs = {
  attributes: TestAttributes;
  name?: InputMaybe<Scalars['String']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeCreateUserEventArgs = {
  attributes: UserEventAttributes;
};


export type RootMutationTypeCreateWebhookArgs = {
  attributes: WebhookAttributes;
};


export type RootMutationTypeCreateZoomArgs = {
  attributes: MeetingAttributes;
};


export type RootMutationTypeDeleteCardArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteChartInstallationArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteClusterArgs = {
  name: Scalars['String'];
  provider: Provider;
};


export type RootMutationTypeDeleteDnsRecordArgs = {
  name: Scalars['String'];
  type: DnsRecordType;
};


export type RootMutationTypeDeleteDomainArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteEabKeyArgs = {
  cluster?: InputMaybe<Scalars['String']>;
  id?: InputMaybe<Scalars['ID']>;
  provider?: InputMaybe<Provider>;
};


export type RootMutationTypeDeleteGroupArgs = {
  groupId: Scalars['ID'];
};


export type RootMutationTypeDeleteGroupMemberArgs = {
  groupId: Scalars['ID'];
  userId: Scalars['ID'];
};


export type RootMutationTypeDeleteIncidentArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteInstallationArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteIntegrationWebhookArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteInviteArgs = {
  id?: InputMaybe<Scalars['ID']>;
  secureId?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeDeleteKeyBackupArgs = {
  name: Scalars['String'];
};


export type RootMutationTypeDeleteMessageArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeletePublicKeyArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteReactionArgs = {
  messageId: Scalars['ID'];
  name: Scalars['String'];
};


export type RootMutationTypeDeleteRecipeArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteRepositoryArgs = {
  repositoryId: Scalars['ID'];
};


export type RootMutationTypeDeleteRoleArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteStackArgs = {
  name: Scalars['String'];
};


export type RootMutationTypeDeleteTerraformArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteTokenArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDeleteUserArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeDestroyClusterArgs = {
  domain: Scalars['String'];
  name: Scalars['String'];
  provider: Provider;
};


export type RootMutationTypeFollowIncidentArgs = {
  attributes: FollowerAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeImpersonateServiceAccountArgs = {
  email?: InputMaybe<Scalars['String']>;
  id?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeInstallBundleArgs = {
  context: ContextAttributes;
  name: Scalars['String'];
  oidc: Scalars['Boolean'];
  repo: Scalars['String'];
};


export type RootMutationTypeInstallChartArgs = {
  attributes: ChartInstallationAttributes;
  installationId: Scalars['ID'];
};


export type RootMutationTypeInstallRecipeArgs = {
  context: Scalars['Map'];
  recipeId: Scalars['ID'];
};


export type RootMutationTypeInstallStackArgs = {
  name: Scalars['String'];
  provider: Provider;
};


export type RootMutationTypeInstallStackShellArgs = {
  context: ContextAttributes;
  name: Scalars['String'];
  oidc: Scalars['Boolean'];
};


export type RootMutationTypeInstallTerraformArgs = {
  attributes: TerraformInstallationAttributes;
  installationId: Scalars['ID'];
};


export type RootMutationTypeLinkPublisherArgs = {
  token: Scalars['String'];
};


export type RootMutationTypeLoginArgs = {
  deviceToken?: InputMaybe<Scalars['String']>;
  email: Scalars['String'];
  password: Scalars['String'];
};


export type RootMutationTypeLoginTokenArgs = {
  deviceToken?: InputMaybe<Scalars['String']>;
  token: Scalars['String'];
};


export type RootMutationTypeOauthCallbackArgs = {
  code: Scalars['String'];
  deviceToken?: InputMaybe<Scalars['String']>;
  host?: InputMaybe<Scalars['String']>;
  provider: OauthProvider;
};


export type RootMutationTypeOauthConsentArgs = {
  challenge: Scalars['String'];
  scopes?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};


export type RootMutationTypePasswordlessLoginArgs = {
  token: Scalars['String'];
};


export type RootMutationTypePingWebhookArgs = {
  id: Scalars['ID'];
  message?: InputMaybe<Scalars['String']>;
  repo: Scalars['String'];
};


export type RootMutationTypeProvisionDomainArgs = {
  name: Scalars['String'];
};


export type RootMutationTypePublishLogsArgs = {
  id: Scalars['ID'];
  logs: Scalars['String'];
};


export type RootMutationTypeQuickStackArgs = {
  provider: Provider;
  repositoryIds?: InputMaybe<Array<InputMaybe<Scalars['ID']>>>;
};


export type RootMutationTypeReadNotificationsArgs = {
  incidentId?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeRealizeInviteArgs = {
  id: Scalars['String'];
};


export type RootMutationTypeRealizeResetTokenArgs = {
  attributes: ResetTokenRealization;
  id: Scalars['ID'];
};


export type RootMutationTypeReleaseLockArgs = {
  attributes: LockAttributes;
  repository: Scalars['String'];
};


export type RootMutationTypeSignupArgs = {
  account?: InputMaybe<AccountAttributes>;
  attributes: UserAttributes;
  deviceToken?: InputMaybe<Scalars['String']>;
  inviteId?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeSsoCallbackArgs = {
  code: Scalars['String'];
  deviceToken?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeTransferDemoProjectArgs = {
  organizationId: Scalars['String'];
};


export type RootMutationTypeUnfollowIncidentArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeUninstallTerraformArgs = {
  id: Scalars['ID'];
};


export type RootMutationTypeUnlockRepositoryArgs = {
  name: Scalars['String'];
};


export type RootMutationTypeUpdateAccountArgs = {
  attributes: AccountAttributes;
};


export type RootMutationTypeUpdateChartArgs = {
  attributes: ChartAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateChartInstallationArgs = {
  attributes: ChartInstallationAttributes;
  chartInstallationId: Scalars['ID'];
};


export type RootMutationTypeUpdateDockerRepositoryArgs = {
  attributes: DockerRepositoryAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateDomainArgs = {
  attributes: DnsDomainAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateGroupArgs = {
  attributes: GroupAttributes;
  groupId: Scalars['ID'];
};


export type RootMutationTypeUpdateIncidentArgs = {
  attributes: IncidentAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateInstallationArgs = {
  attributes: InstallationAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateIntegrationWebhookArgs = {
  attributes: IntegrationWebhookAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateLineItemArgs = {
  attributes: LimitAttributes;
  subscriptionId: Scalars['ID'];
};


export type RootMutationTypeUpdateMessageArgs = {
  attributes: IncidentMessageAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateOidcProviderArgs = {
  attributes: OidcAttributes;
  installationId: Scalars['ID'];
};


export type RootMutationTypeUpdatePlanArgs = {
  planId: Scalars['ID'];
  subscriptionId: Scalars['ID'];
};


export type RootMutationTypeUpdatePlanAttributesArgs = {
  attributes: UpdatablePlanAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdatePlatformPlanArgs = {
  planId: Scalars['ID'];
};


export type RootMutationTypeUpdatePublisherArgs = {
  attributes: PublisherAttributes;
};


export type RootMutationTypeUpdateRepositoryArgs = {
  attributes: RepositoryAttributes;
  repositoryId?: InputMaybe<Scalars['ID']>;
  repositoryName?: InputMaybe<Scalars['String']>;
};


export type RootMutationTypeUpdateRoleArgs = {
  attributes: RoleAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateServiceAccountArgs = {
  attributes: ServiceAccountAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateShellArgs = {
  attributes: CloudShellAttributes;
};


export type RootMutationTypeUpdateShellConfigurationArgs = {
  context: Scalars['Map'];
};


export type RootMutationTypeUpdateStepArgs = {
  attributes: TestStepAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateTerraformArgs = {
  attributes: TerraformAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateTestArgs = {
  attributes: TestAttributes;
  id: Scalars['ID'];
};


export type RootMutationTypeUpdateUserArgs = {
  attributes: UserAttributes;
  id?: InputMaybe<Scalars['ID']>;
};


export type RootMutationTypeUpdateVersionArgs = {
  attributes: VersionAttributes;
  id?: InputMaybe<Scalars['ID']>;
  spec?: InputMaybe<VersionSpec>;
};


export type RootMutationTypeUploadTerraformArgs = {
  attributes: TerraformAttributes;
  name: Scalars['String'];
  repositoryName: Scalars['String'];
};


export type RootMutationTypeUpsertOidcProviderArgs = {
  attributes: OidcAttributes;
  installationId: Scalars['ID'];
};


export type RootMutationTypeUpsertRepositoryArgs = {
  attributes: RepositoryAttributes;
  name: Scalars['String'];
  publisher: Scalars['String'];
};

export type OauthResponse = {
  __typename?: 'OauthResponse';
  redirectTo: Scalars['String'];
};

export type ApplyLock = {
  __typename?: 'ApplyLock';
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  lock?: Maybe<Scalars['String']>;
  owner?: Maybe<User>;
  repository?: Maybe<Repository>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type PostmortemAttributes = {
  actionItems?: InputMaybe<Array<InputMaybe<ActionItemAttributes>>>;
  content: Scalars['String'];
};

export type ActionItemAttributes = {
  link: Scalars['String'];
  type: ActionItemType;
};

export type ArtifactAttributes = {
  arch?: InputMaybe<Scalars['String']>;
  blob?: InputMaybe<Scalars['UploadOrUrl']>;
  name: Scalars['String'];
  platform: Scalars['String'];
  readme: Scalars['String'];
  type: Scalars['String'];
};

export type AddressAttributes = {
  city: Scalars['String'];
  country: Scalars['String'];
  line1: Scalars['String'];
  line2: Scalars['String'];
  name?: InputMaybe<Scalars['String']>;
  state: Scalars['String'];
  zip: Scalars['String'];
};

/** Input for creating or updating a cluster. */
export type ClusterAttributes = {
  /** The URL of the console running on the cluster. */
  consoleUrl?: InputMaybe<Scalars['String']>;
  /** The domain name used for applications deployed on the cluster. */
  domain?: InputMaybe<Scalars['String']>;
  /** The git repository URL for the cluster. */
  gitUrl?: InputMaybe<Scalars['String']>;
  /** The name of the cluster. */
  name: Scalars['String'];
  /** The cluster's cloud provider. */
  provider: Provider;
  /** The source of the cluster. */
  source?: InputMaybe<Source>;
};

/** A dependncy reference between clusters */
export type ClusterDependency = {
  __typename?: 'ClusterDependency';
  /** the cluster holding this dependency */
  cluster?: Maybe<Cluster>;
  /** the source cluster of this dependency */
  dependency?: Maybe<Cluster>;
  id: Scalars['ID'];
  insertedAt?: Maybe<Scalars['DateTime']>;
  updatedAt?: Maybe<Scalars['DateTime']>;
};

export type CrdAttributes = {
  blob?: InputMaybe<Scalars['UploadOrUrl']>;
  name: Scalars['String'];
};

export type ChartName = {
  chart?: InputMaybe<Scalars['String']>;
  repo?: InputMaybe<Scalars['String']>;
};

export type DnsRecordAttributes = {
  name: Scalars['String'];
  records?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  type: DnsRecordType;
};

export type DnsDomainAttributes = {
  accessPolicy?: InputMaybe<DnsAccessPolicyAttributes>;
  name?: InputMaybe<Scalars['String']>;
};

export type DnsAccessPolicyAttributes = {
  bindings?: InputMaybe<Array<InputMaybe<BindingAttributes>>>;
  id?: InputMaybe<Scalars['ID']>;
};

export type BindingAttributes = {
  groupId?: InputMaybe<Scalars['ID']>;
  id?: InputMaybe<Scalars['ID']>;
  userId?: InputMaybe<Scalars['ID']>;
};

export type GroupAttributes = {
  description?: InputMaybe<Scalars['String']>;
  global?: InputMaybe<Scalars['Boolean']>;
  name: Scalars['String'];
};

export type IncidentAttributes = {
  clusterInformation?: InputMaybe<ClusterInformationAttributes>;
  description?: InputMaybe<Scalars['String']>;
  severity?: InputMaybe<Scalars['Int']>;
  status?: InputMaybe<IncidentStatus>;
  tags?: InputMaybe<Array<InputMaybe<TagAttributes>>>;
  title?: InputMaybe<Scalars['String']>;
};

export type ClusterInformationAttributes = {
  gitCommit?: InputMaybe<Scalars['String']>;
  platform?: InputMaybe<Scalars['String']>;
  version?: InputMaybe<Scalars['String']>;
};

export type TagAttributes = {
  tag: Scalars['String'];
};

export type IntegrationAttributes = {
  description?: InputMaybe<Scalars['String']>;
  icon?: InputMaybe<Scalars['UploadOrUrl']>;
  name: Scalars['String'];
  sourceUrl?: InputMaybe<Scalars['String']>;
  spec?: InputMaybe<Scalars['Yaml']>;
  tags?: InputMaybe<Array<InputMaybe<TagAttributes>>>;
  type?: InputMaybe<Scalars['String']>;
};

export type IntegrationWebhookAttributes = {
  actions?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  name: Scalars['String'];
  url: Scalars['String'];
};

export type InviteAttributes = {
  email?: InputMaybe<Scalars['String']>;
};

export type KeyBackupAttributes = {
  key: Scalars['String'];
  name: Scalars['String'];
  repositories?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type IncidentMessageAttributes = {
  entities?: InputMaybe<Array<InputMaybe<EntityAttributes>>>;
  file?: InputMaybe<FileAttributes>;
  text: Scalars['String'];
};

export type EntityAttributes = {
  endIndex?: InputMaybe<Scalars['Int']>;
  startIndex?: InputMaybe<Scalars['Int']>;
  text?: InputMaybe<Scalars['String']>;
  type: MessageEntityType;
  userId?: InputMaybe<Scalars['ID']>;
};

export type FileAttributes = {
  blob?: InputMaybe<Scalars['UploadOrUrl']>;
};

export type OauthAttributes = {
  code?: InputMaybe<Scalars['String']>;
  redirectUri?: InputMaybe<Scalars['String']>;
  service?: InputMaybe<OauthService>;
};

/** Input for creating or updating the OIDC attributes of an application installation. */
export type OidcAttributes = {
  /** The authentication method for the OIDC provider. */
  authMethod: OidcAuthMethod;
  /** The users or groups that can login through the OIDC provider. */
  bindings?: InputMaybe<Array<InputMaybe<BindingAttributes>>>;
  /** The redirect URIs for the OIDC provider. */
  redirectUris?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type PlanAttributes = {
  cost: Scalars['Int'];
  default?: InputMaybe<Scalars['Boolean']>;
  lineItems?: InputMaybe<PlanLineItemAttributes>;
  metadata?: InputMaybe<PlanMetadataAttributes>;
  name: Scalars['String'];
  period: Scalars['String'];
  serviceLevels?: InputMaybe<Array<InputMaybe<ServiceLevelAttributes>>>;
};

export type PlanLineItemAttributes = {
  included?: InputMaybe<Array<InputMaybe<LimitAttributes>>>;
  items?: InputMaybe<Array<InputMaybe<LineItemAttributes>>>;
};

export type LimitAttributes = {
  dimension: Scalars['String'];
  quantity: Scalars['Int'];
};

export type LineItemAttributes = {
  cost: Scalars['Int'];
  dimension: Scalars['String'];
  name: Scalars['String'];
  period: Scalars['String'];
  type?: InputMaybe<PlanType>;
};

export type PlanMetadataAttributes = {
  features?: InputMaybe<Array<InputMaybe<PlanFeatureAttributes>>>;
  freeform?: InputMaybe<Scalars['Yaml']>;
};

export type PlanFeatureAttributes = {
  description: Scalars['String'];
  name: Scalars['String'];
};

export type ServiceLevelAttributes = {
  maxSeverity?: InputMaybe<Scalars['Int']>;
  minSeverity?: InputMaybe<Scalars['Int']>;
  responseTime?: InputMaybe<Scalars['Int']>;
};

export type PublicKeyAttributes = {
  content: Scalars['String'];
  name: Scalars['String'];
};

export type PublisherAttributes = {
  address?: InputMaybe<AddressAttributes>;
  avatar?: InputMaybe<Scalars['UploadOrUrl']>;
  community?: InputMaybe<CommunityAttributes>;
  description?: InputMaybe<Scalars['String']>;
  name?: InputMaybe<Scalars['String']>;
  phone?: InputMaybe<Scalars['String']>;
};

/** Input for creating or updating the community links of an application. */
export type CommunityAttributes = {
  /** The application's Discord server. */
  discord?: InputMaybe<Scalars['String']>;
  /** The application's git URL. */
  gitUrl?: InputMaybe<Scalars['String']>;
  /** The application's homepage. */
  homepage?: InputMaybe<Scalars['String']>;
  /** The application's Slack channel. */
  slack?: InputMaybe<Scalars['String']>;
  /** The application's Twitter account. */
  twitter?: InputMaybe<Scalars['String']>;
  /** The videos of the application. */
  videos?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type UpgradeQueueAttributes = {
  domain?: InputMaybe<Scalars['String']>;
  git?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  provider?: InputMaybe<Provider>;
};

export type RecipeAttributes = {
  dependencies?: InputMaybe<Array<InputMaybe<RecipeReference>>>;
  description?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  oidcSettings?: InputMaybe<OidcSettingsAttributes>;
  primary?: InputMaybe<Scalars['Boolean']>;
  private?: InputMaybe<Scalars['Boolean']>;
  provider?: InputMaybe<Provider>;
  restricted?: InputMaybe<Scalars['Boolean']>;
  sections?: InputMaybe<Array<InputMaybe<RecipeSectionAttributes>>>;
  tests?: InputMaybe<Array<InputMaybe<RecipeTestAttributes>>>;
};

export type RecipeReference = {
  name: Scalars['String'];
  repo: Scalars['String'];
};

export type OidcSettingsAttributes = {
  authMethod: OidcAuthMethod;
  domainKey?: InputMaybe<Scalars['String']>;
  subdomain?: InputMaybe<Scalars['Boolean']>;
  uriFormat?: InputMaybe<Scalars['String']>;
  uriFormats?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type RecipeSectionAttributes = {
  configuration?: InputMaybe<Array<InputMaybe<RecipeConfigurationAttributes>>>;
  items?: InputMaybe<Array<InputMaybe<RecipeItemAttributes>>>;
  name: Scalars['String'];
};

export type RecipeConfigurationAttributes = {
  condition?: InputMaybe<RecipeConditionAttributes>;
  default?: InputMaybe<Scalars['String']>;
  documentation?: InputMaybe<Scalars['String']>;
  functionName?: InputMaybe<Scalars['String']>;
  longform?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  optional?: InputMaybe<Scalars['Boolean']>;
  placeholder?: InputMaybe<Scalars['String']>;
  type: Datatype;
  validation?: InputMaybe<RecipeValidationAttributes>;
};

export type RecipeConditionAttributes = {
  field: Scalars['String'];
  operation: Operation;
  value?: InputMaybe<Scalars['String']>;
};

export type RecipeValidationAttributes = {
  message: Scalars['String'];
  regex?: InputMaybe<Scalars['String']>;
  type: ValidationType;
};

export type RecipeItemAttributes = {
  configuration?: InputMaybe<Array<InputMaybe<RecipeConfigurationAttributes>>>;
  name: Scalars['String'];
  type: RecipeItemType;
};

export enum RecipeItemType {
  Helm = 'HELM',
  Terraform = 'TERRAFORM'
}

export type RecipeTestAttributes = {
  args?: InputMaybe<Array<InputMaybe<TestArgumentAttributes>>>;
  message?: InputMaybe<Scalars['String']>;
  name: Scalars['String'];
  type: TestType;
};

export type TestArgumentAttributes = {
  key: Scalars['String'];
  name: Scalars['String'];
  repo: Scalars['String'];
};

/** Input for creating or updating an application's attributes. */
export type RepositoryAttributes = {
  /** The category of the application. */
  category?: InputMaybe<Category>;
  /** The application's community links. */
  community?: InputMaybe<CommunityAttributes>;
  /** The application's dark icon. */
  darkIcon?: InputMaybe<Scalars['UploadOrUrl']>;
  /** The default tag to use when deploying the application. */
  defaultTag?: InputMaybe<Scalars['String']>;
  /** A short description of the application. */
  description?: InputMaybe<Scalars['String']>;
  /** The application's documentation. */
  docs?: InputMaybe<Scalars['UploadOrUrl']>;
  /** A link to the application's documentation. */
  documentation?: InputMaybe<Scalars['String']>;
  /** The application's git URL. */
  gitUrl?: InputMaybe<Scalars['String']>;
  /** The application's homepage. */
  homepage?: InputMaybe<Scalars['String']>;
  /** The application's icon. */
  icon?: InputMaybe<Scalars['UploadOrUrl']>;
  /** The application's integration resource definition. */
  integrationResourceDefinition?: InputMaybe<ResourceDefinitionAttributes>;
  /** The name of the application. */
  name?: InputMaybe<Scalars['String']>;
  /** Notes about the application rendered after deploying and displayed to the user. */
  notes?: InputMaybe<Scalars['String']>;
  /** The application's OAuth settings. */
  oauthSettings?: InputMaybe<OauthSettingsAttributes>;
  /** Whether the application is private. */
  private?: InputMaybe<Scalars['Boolean']>;
  /** The application's README. */
  readme?: InputMaybe<Scalars['String']>;
  /** A YAML object of secrets. */
  secrets?: InputMaybe<Scalars['Yaml']>;
  /** The application's tags. */
  tags?: InputMaybe<Array<InputMaybe<TagAttributes>>>;
  /** Whether the application is trending. */
  trending?: InputMaybe<Scalars['Boolean']>;
  /** Whether the application is verified. */
  verified?: InputMaybe<Scalars['Boolean']>;
};

export type ResourceDefinitionAttributes = {
  name: Scalars['String'];
  spec?: InputMaybe<Array<InputMaybe<SpecificationAttributes>>>;
};

export type SpecificationAttributes = {
  inner?: InputMaybe<SpecDatatype>;
  name: Scalars['String'];
  required?: InputMaybe<Scalars['Boolean']>;
  spec?: InputMaybe<Array<InputMaybe<SpecificationAttributes>>>;
  type: SpecDatatype;
};

export enum SpecDatatype {
  Bool = 'BOOL',
  Float = 'FLOAT',
  Int = 'INT',
  List = 'LIST',
  Object = 'OBJECT',
  String = 'STRING'
}

/** Input for the application's OAuth settings. */
export type OauthSettingsAttributes = {
  /** The authentication method for the OAuth provider. */
  authMethod: OidcAuthMethod;
  /** The URI format for the OAuth provider. */
  uriFormat: Scalars['String'];
};

export type ResetTokenAttributes = {
  email?: InputMaybe<Scalars['String']>;
  type: ResetTokenType;
};

export type RoleAttributes = {
  description?: InputMaybe<Scalars['String']>;
  name?: InputMaybe<Scalars['String']>;
  permissions?: InputMaybe<Array<InputMaybe<Permission>>>;
  repositories?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  roleBindings?: InputMaybe<Array<InputMaybe<BindingAttributes>>>;
};

export type ServiceAccountAttributes = {
  email?: InputMaybe<Scalars['String']>;
  impersonationPolicy?: InputMaybe<ImpersonationPolicyAttributes>;
  name?: InputMaybe<Scalars['String']>;
};

export type ImpersonationPolicyAttributes = {
  bindings?: InputMaybe<Array<InputMaybe<ImpersonationPolicyBindingAttributes>>>;
  id?: InputMaybe<Scalars['ID']>;
};

export type ImpersonationPolicyBindingAttributes = {
  groupId?: InputMaybe<Scalars['ID']>;
  id?: InputMaybe<Scalars['ID']>;
  userId?: InputMaybe<Scalars['ID']>;
};

export type CloudShellAttributes = {
  credentials: ShellCredentialsAttributes;
  demoId?: InputMaybe<Scalars['ID']>;
  provider?: InputMaybe<Provider>;
  scm?: InputMaybe<ScmAttributes>;
  workspace: WorkspaceAttributes;
};

export type ShellCredentialsAttributes = {
  aws?: InputMaybe<AwsShellCredentialsAttributes>;
  azure?: InputMaybe<AzureShellCredentialsAttributes>;
  gcp?: InputMaybe<GcpShellCredentialsAttributes>;
};

export type AwsShellCredentialsAttributes = {
  accessKeyId: Scalars['String'];
  secretAccessKey: Scalars['String'];
};

export type AzureShellCredentialsAttributes = {
  clientId: Scalars['String'];
  clientSecret: Scalars['String'];
  storageAccount: Scalars['String'];
  subscriptionId: Scalars['String'];
  tenantId: Scalars['String'];
};

export type GcpShellCredentialsAttributes = {
  applicationCredentials: Scalars['String'];
};

export type ScmAttributes = {
  gitUrl?: InputMaybe<Scalars['String']>;
  name?: InputMaybe<Scalars['String']>;
  org?: InputMaybe<Scalars['String']>;
  privateKey?: InputMaybe<Scalars['String']>;
  provider?: InputMaybe<ScmProvider>;
  publicKey?: InputMaybe<Scalars['String']>;
  token?: InputMaybe<Scalars['String']>;
};

export type WorkspaceAttributes = {
  bucketPrefix: Scalars['String'];
  cluster: Scalars['String'];
  project?: InputMaybe<Scalars['String']>;
  region: Scalars['String'];
  subdomain: Scalars['String'];
};

export type StackAttributes = {
  collections?: InputMaybe<Array<InputMaybe<StackCollectionAttributes>>>;
  community?: InputMaybe<CommunityAttributes>;
  description?: InputMaybe<Scalars['String']>;
  displayName?: InputMaybe<Scalars['String']>;
  featured?: InputMaybe<Scalars['Boolean']>;
  name: Scalars['String'];
};

export type StackCollectionAttributes = {
  bundles?: InputMaybe<Array<InputMaybe<RecipeReference>>>;
  provider: Provider;
};

export type SubscriptionAttributes = {
  lineItems?: InputMaybe<SubscriptionLineItemAttributes>;
};

export type SubscriptionLineItemAttributes = {
  items?: InputMaybe<Array<InputMaybe<LimitAttributes>>>;
};

export type TerraformAttributes = {
  dependencies?: InputMaybe<Scalars['Yaml']>;
  description?: InputMaybe<Scalars['String']>;
  name?: InputMaybe<Scalars['String']>;
  package?: InputMaybe<Scalars['UploadOrUrl']>;
  version?: InputMaybe<Scalars['String']>;
};

export type TestAttributes = {
  name?: InputMaybe<Scalars['String']>;
  promoteTag?: InputMaybe<Scalars['String']>;
  status?: InputMaybe<TestStatus>;
  steps?: InputMaybe<Array<InputMaybe<TestStepAttributes>>>;
  tags?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type TestStepAttributes = {
  description?: InputMaybe<Scalars['String']>;
  id?: InputMaybe<Scalars['ID']>;
  logs?: InputMaybe<Scalars['UploadOrUrl']>;
  name?: InputMaybe<Scalars['String']>;
  status?: InputMaybe<TestStatus>;
};

export type UserEventAttributes = {
  data?: InputMaybe<Scalars['String']>;
  event: Scalars['String'];
  status?: InputMaybe<UserEventStatus>;
};

export enum UserEventStatus {
  Error = 'ERROR',
  Ok = 'OK'
}

export type WebhookAttributes = {
  url: Scalars['String'];
};

export type MeetingAttributes = {
  incidentId?: InputMaybe<Scalars['ID']>;
  topic: Scalars['String'];
};

export type ZoomMeeting = {
  __typename?: 'ZoomMeeting';
  joinUrl: Scalars['String'];
  password?: Maybe<Scalars['String']>;
};

export type DeviceLogin = {
  __typename?: 'DeviceLogin';
  deviceToken: Scalars['String'];
  loginUrl: Scalars['String'];
};

export type FollowerAttributes = {
  preferences?: InputMaybe<NotificationPreferencesAttributes>;
};

export type NotificationPreferencesAttributes = {
  incidentUpdate: Scalars['Boolean'];
  mention: Scalars['Boolean'];
  message: Scalars['Boolean'];
};

export type ContextAttributes = {
  buckets?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
  configuration: Scalars['Map'];
  domains?: InputMaybe<Array<InputMaybe<Scalars['String']>>>;
};

export type ChartInstallationAttributes = {
  chartId?: InputMaybe<Scalars['ID']>;
  versionId?: InputMaybe<Scalars['ID']>;
};

export type TerraformInstallationAttributes = {
  terraformId?: InputMaybe<Scalars['ID']>;
  versionId?: InputMaybe<Scalars['ID']>;
};

export type WebhookResponse = {
  __typename?: 'WebhookResponse';
  body?: Maybe<Scalars['String']>;
  headers?: Maybe<Scalars['Map']>;
  statusCode: Scalars['Int'];
};

export type ResetTokenRealization = {
  password?: InputMaybe<Scalars['String']>;
};

export type LockAttributes = {
  lock: Scalars['String'];
};

export type AccountAttributes = {
  billingAddress?: InputMaybe<AddressAttributes>;
  domainMappings?: InputMaybe<Array<InputMaybe<DomainMappingInput>>>;
  icon?: InputMaybe<Scalars['UploadOrUrl']>;
  name?: InputMaybe<Scalars['String']>;
};

export type DomainMappingInput = {
  domain?: InputMaybe<Scalars['String']>;
  enableSso?: InputMaybe<Scalars['Boolean']>;
  id?: InputMaybe<Scalars['ID']>;
};

export type UserAttributes = {
  avatar?: InputMaybe<Scalars['UploadOrUrl']>;
  confirm?: InputMaybe<Scalars['String']>;
  email?: InputMaybe<Scalars['String']>;
  loginMethod?: InputMaybe<LoginMethod>;
  name?: InputMaybe<Scalars['String']>;
  onboarding?: InputMaybe<OnboardingState>;
  onboardingChecklist?: InputMaybe<OnboardingChecklistAttributes>;
  password?: InputMaybe<Scalars['String']>;
  roles?: InputMaybe<RolesAttributes>;
};

export type OnboardingChecklistAttributes = {
  dismissed?: InputMaybe<Scalars['Boolean']>;
  status?: InputMaybe<OnboardingChecklistState>;
};

export type RolesAttributes = {
  admin?: InputMaybe<Scalars['Boolean']>;
};

export type ChartAttributes = {
  tags?: InputMaybe<Array<InputMaybe<VersionTagAttributes>>>;
};

export type VersionTagAttributes = {
  tag: Scalars['String'];
  versionId?: InputMaybe<Scalars['ID']>;
};

export type DockerRepositoryAttributes = {
  public: Scalars['Boolean'];
};

/** Input for creating or updating the tag attributes of an application installation. */
export type InstallationAttributes = {
  /** Whether the application should auto upgrade. */
  autoUpgrade?: InputMaybe<Scalars['Boolean']>;
  /** A YAML object of context. */
  context?: InputMaybe<Scalars['Yaml']>;
  /** The tag to track for auto upgrades. */
  trackTag?: InputMaybe<Scalars['String']>;
};

export type UpdatablePlanAttributes = {
  default?: InputMaybe<Scalars['Boolean']>;
  serviceLevels?: InputMaybe<Array<InputMaybe<ServiceLevelAttributes>>>;
};

export type VersionAttributes = {
  tags?: InputMaybe<Array<InputMaybe<VersionTagAttributes>>>;
};

export type VersionSpec = {
  chart?: InputMaybe<Scalars['String']>;
  repository?: InputMaybe<Scalars['String']>;
  terraform?: InputMaybe<Scalars['String']>;
  version?: InputMaybe<Scalars['String']>;
};

export type RootSubscriptionType = {
  __typename?: 'RootSubscriptionType';
  incidentDelta?: Maybe<IncidentDelta>;
  incidentMessageDelta?: Maybe<IncidentMessageDelta>;
  notification?: Maybe<Notification>;
  rolloutDelta?: Maybe<RolloutDelta>;
  testDelta?: Maybe<TestDelta>;
  testLogs?: Maybe<StepLogs>;
  upgrade?: Maybe<Upgrade>;
  upgradeQueueDelta?: Maybe<UpgradeQueueDelta>;
};


export type RootSubscriptionTypeIncidentDeltaArgs = {
  incidentId?: InputMaybe<Scalars['ID']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
};


export type RootSubscriptionTypeIncidentMessageDeltaArgs = {
  incidentId?: InputMaybe<Scalars['ID']>;
};


export type RootSubscriptionTypeRolloutDeltaArgs = {
  repositoryId: Scalars['ID'];
};


export type RootSubscriptionTypeTestDeltaArgs = {
  repositoryId: Scalars['ID'];
};


export type RootSubscriptionTypeTestLogsArgs = {
  testId: Scalars['ID'];
};


export type RootSubscriptionTypeUpgradeArgs = {
  id?: InputMaybe<Scalars['ID']>;
};

export type IncidentDelta = {
  __typename?: 'IncidentDelta';
  delta?: Maybe<Delta>;
  payload?: Maybe<Incident>;
};

export enum Delta {
  Create = 'CREATE',
  Delete = 'DELETE',
  Update = 'UPDATE'
}

export type IncidentMessageDelta = {
  __typename?: 'IncidentMessageDelta';
  delta?: Maybe<Delta>;
  payload?: Maybe<IncidentMessage>;
};

export type RolloutDelta = {
  __typename?: 'RolloutDelta';
  delta?: Maybe<Delta>;
  payload?: Maybe<Rollout>;
};

export type TestDelta = {
  __typename?: 'TestDelta';
  delta?: Maybe<Delta>;
  payload?: Maybe<Test>;
};

export type StepLogs = {
  __typename?: 'StepLogs';
  logs?: Maybe<Array<Maybe<Scalars['String']>>>;
  step?: Maybe<TestStep>;
};

export type UpgradeQueueDelta = {
  __typename?: 'UpgradeQueueDelta';
  delta?: Maybe<Delta>;
  payload?: Maybe<UpgradeQueue>;
};

export type AuditFragment = { __typename?: 'Audit', id: string, action: string, ip?: string | null, country?: string | null, city?: string | null, latitude?: string | null, longitude?: string | null, insertedAt?: Date | null, actor?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null, integrationWebhook?: { __typename?: 'IntegrationWebhook', id: string, name: string, url: string, secret: string, actions?: Array<string | null> | null } | null, role?: { __typename?: 'Role', id: string, name: string, description?: string | null, repositories?: Array<string | null> | null, permissions?: Array<Permission | null> | null, roleBindings?: Array<{ __typename?: 'RoleBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, image?: { __typename?: 'DockerImage', id: string, tag?: string | null, dockerRepository?: { __typename?: 'DockerRepository', name: string } | null } | null };

export type PolicyBindingFragment = { __typename?: 'PolicyBinding', id: string, group?: { __typename?: 'Group', id: string, name: string } | null, user?: { __typename?: 'User', id: string, name: string, email: string } | null };

export type DnsDomainFragment = { __typename?: 'DnsDomain', id: string, name: string, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, accessPolicy?: { __typename?: 'DnsAccessPolicy', id: string, bindings?: Array<{ __typename?: 'PolicyBinding', id: string, group?: { __typename?: 'Group', id: string, name: string } | null, user?: { __typename?: 'User', id: string, name: string, email: string } | null } | null> | null } | null };

export type InviteFragment = { __typename?: 'Invite', id: string, secureId?: string | null, email?: string | null, insertedAt?: Date | null };

export type OidcLoginFragment = { __typename?: 'OidcLogin', ip?: string | null, country?: string | null, city?: string | null, latitude?: string | null, longitude?: string | null, insertedAt?: Date | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null };

export type UpdateAccountMutationVariables = Exact<{
  attributes: AccountAttributes;
}>;


export type UpdateAccountMutation = { __typename?: 'RootMutationType', updateAccount?: { __typename?: 'Account', id: string, name?: string | null, billingCustomerId?: string | null, backgroundColor?: string | null, domainMappings?: Array<{ __typename?: 'DomainMapping', id: string, domain: string, enableSso?: boolean | null } | null> | null } | null };

export type ArtifactFragment = { __typename?: 'Artifact', id?: string | null, name?: string | null, blob?: string | null, type?: ArtifactType | null, platform?: ArtifactPlatform | null, arch?: string | null, filesize?: number | null, sha?: string | null, readme?: string | null, insertedAt?: Date | null, updatedAt?: Date | null };

export type ListArtifactsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type ListArtifactsQuery = { __typename?: 'RootQueryType', repository?: { __typename?: 'Repository', artifacts?: Array<{ __typename?: 'Artifact', id?: string | null, name?: string | null, blob?: string | null, type?: ArtifactType | null, platform?: ArtifactPlatform | null, arch?: string | null, filesize?: number | null, sha?: string | null, readme?: string | null, insertedAt?: Date | null, updatedAt?: Date | null } | null> | null } | null };

export type CreateArtifactMutationVariables = Exact<{
  repoName: Scalars['String'];
  name: Scalars['String'];
  readme: Scalars['String'];
  artifactType: Scalars['String'];
  platform: Scalars['String'];
  blob: Scalars['UploadOrUrl'];
  arch?: InputMaybe<Scalars['String']>;
}>;


export type CreateArtifactMutation = { __typename?: 'RootMutationType', createArtifact?: { __typename?: 'Artifact', id?: string | null, name?: string | null, blob?: string | null, type?: ArtifactType | null, platform?: ArtifactPlatform | null, arch?: string | null, filesize?: number | null, sha?: string | null, readme?: string | null, insertedAt?: Date | null, updatedAt?: Date | null } | null };

export type ChartFragment = { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null };

export type CrdFragment = { __typename?: 'Crd', id: string, name: string, blob?: string | null };

export type ChartInstallationFragment = { __typename?: 'ChartInstallation', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null };

export type ScanViolationFragment = { __typename?: 'ScanViolation', ruleName?: string | null, description?: string | null, ruleId?: string | null, severity?: VulnGrade | null, category?: string | null, resourceName?: string | null, resourceType?: string | null, file?: string | null, line?: number | null };

export type ScanErrorFragment = { __typename?: 'ScanError', message?: string | null };

export type PackageScanFragment = { __typename?: 'PackageScan', id: string, grade?: ImageGrade | null, violations?: Array<{ __typename?: 'ScanViolation', ruleName?: string | null, description?: string | null, ruleId?: string | null, severity?: VulnGrade | null, category?: string | null, resourceName?: string | null, resourceType?: string | null, file?: string | null, line?: number | null } | null> | null, errors?: Array<{ __typename?: 'ScanError', message?: string | null } | null> | null };

export type GetChartsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetChartsQuery = { __typename?: 'RootQueryType', charts?: { __typename?: 'ChartConnection', edges?: Array<{ __typename?: 'ChartEdge', node?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null> | null } | null };

export type GetVersionsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetVersionsQuery = { __typename?: 'RootQueryType', versions?: { __typename?: 'VersionConnection', edges?: Array<{ __typename?: 'VersionEdge', node?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null> | null } | null };

export type GetChartInstallationsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetChartInstallationsQuery = { __typename?: 'RootQueryType', chartInstallations?: { __typename?: 'ChartInstallationConnection', edges?: Array<{ __typename?: 'ChartInstallationEdge', node?: { __typename?: 'ChartInstallation', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null } | null> | null } | null };

export type GetPackageInstallationsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetPackageInstallationsQuery = { __typename?: 'RootQueryType', chartInstallations?: { __typename?: 'ChartInstallationConnection', edges?: Array<{ __typename?: 'ChartInstallationEdge', node?: { __typename?: 'ChartInstallation', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null } | null> | null } | null, terraformInstallations?: { __typename?: 'TerraformInstallationConnection', edges?: Array<{ __typename?: 'TerraformInstallationEdge', node?: { __typename?: 'TerraformInstallation', id?: string | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null } | null> | null } | null };

export type CreateCrdMutationVariables = Exact<{
  chartName: ChartName;
  name: Scalars['String'];
  blob: Scalars['UploadOrUrl'];
}>;


export type CreateCrdMutation = { __typename?: 'RootMutationType', createCrd?: { __typename?: 'Crd', id: string } | null };

export type UninstallChartMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type UninstallChartMutation = { __typename?: 'RootMutationType', deleteChartInstallation?: { __typename?: 'ChartInstallation', id?: string | null } | null };

export type DnsRecordFragment = { __typename?: 'DnsRecord', id: string, name: string, type: DnsRecordType, records?: Array<string | null> | null, cluster: string, provider: Provider, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null };

export type GetDnsRecordsQueryVariables = Exact<{
  cluster: Scalars['String'];
  provider: Provider;
}>;


export type GetDnsRecordsQuery = { __typename?: 'RootQueryType', dnsRecords?: { __typename?: 'DnsRecordConnection', edges?: Array<{ __typename?: 'DnsRecordEdge', node?: { __typename?: 'DnsRecord', id: string, name: string, type: DnsRecordType, records?: Array<string | null> | null, cluster: string, provider: Provider, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null } | null> | null } | null };

export type CreateDnsRecordMutationVariables = Exact<{
  cluster: Scalars['String'];
  provider: Provider;
  attributes: DnsRecordAttributes;
}>;


export type CreateDnsRecordMutation = { __typename?: 'RootMutationType', createDnsRecord?: { __typename?: 'DnsRecord', id: string, name: string, type: DnsRecordType, records?: Array<string | null> | null, cluster: string, provider: Provider, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null };

export type DeleteDnsRecordMutationVariables = Exact<{
  name: Scalars['String'];
  type: DnsRecordType;
}>;


export type DeleteDnsRecordMutation = { __typename?: 'RootMutationType', deleteDnsRecord?: { __typename?: 'DnsRecord', id: string, name: string, type: DnsRecordType, records?: Array<string | null> | null, cluster: string, provider: Provider, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null };

export type DockerRepoFragment = { __typename?: 'DockerRepository', id: string, name: string, public?: boolean | null, insertedAt?: Date | null, updatedAt?: Date | null, repository?: { __typename?: 'Repository', id: string, name: string } | null };

export type DockerRepositoryFragment = { __typename?: 'DockerRepository', id: string, name: string, public?: boolean | null, insertedAt?: Date | null, updatedAt?: Date | null, repository?: { __typename?: 'Repository', id: string, name: string, editable?: boolean | null } | null };

export type DockerImageFragment = { __typename?: 'DockerImage', id: string, tag?: string | null, digest: string, scannedAt?: Date | null, grade?: ImageGrade | null, insertedAt?: Date | null, updatedAt?: Date | null };

export type VulnerabilityFragment = { __typename?: 'Vulnerability', id: string, title?: string | null, description?: string | null, vulnerabilityId?: string | null, package?: string | null, installedVersion?: string | null, fixedVersion?: string | null, source?: string | null, url?: string | null, severity?: VulnGrade | null, score?: number | null, cvss?: { __typename?: 'Cvss', attackVector?: VulnVector | null, attackComplexity?: VulnGrade | null, privilegesRequired?: VulnGrade | null, userInteraction?: VulnRequirement | null, confidentiality?: VulnGrade | null, integrity?: VulnGrade | null, availability?: VulnGrade | null } | null, layer?: { __typename?: 'ImageLayer', digest?: string | null, diffId?: string | null } | null };

export type CreateDomainMutationVariables = Exact<{
  name: Scalars['String'];
}>;


export type CreateDomainMutation = { __typename?: 'RootMutationType', provisionDomain?: { __typename?: 'DnsDomain', id: string, name: string, insertedAt?: Date | null, creator?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, accessPolicy?: { __typename?: 'DnsAccessPolicy', id: string, bindings?: Array<{ __typename?: 'PolicyBinding', id: string, group?: { __typename?: 'Group', id: string, name: string } | null, user?: { __typename?: 'User', id: string, name: string, email: string } | null } | null> | null } | null } | null };

export type GroupMembersQueryVariables = Exact<{
  cursor?: InputMaybe<Scalars['String']>;
  id: Scalars['ID'];
}>;


export type GroupMembersQuery = { __typename?: 'RootQueryType', groupMembers?: { __typename?: 'GroupMemberConnection', pageInfo: { __typename?: 'PageInfo', endCursor?: string | null, hasNextPage: boolean }, edges?: Array<{ __typename?: 'GroupMemberEdge', node?: { __typename?: 'GroupMember', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null } | null> | null } | null };

export type CreateGroupMemberMutationVariables = Exact<{
  groupId: Scalars['ID'];
  userId: Scalars['ID'];
}>;


export type CreateGroupMemberMutation = { __typename?: 'RootMutationType', createGroupMember?: { __typename?: 'GroupMember', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null };

export type DeleteGroupMemberMutationVariables = Exact<{
  groupId: Scalars['ID'];
  userId: Scalars['ID'];
}>;


export type DeleteGroupMemberMutation = { __typename?: 'RootMutationType', deleteGroupMember?: { __typename?: 'GroupMember', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null };

export type CreateGroupMutationVariables = Exact<{
  attributes: GroupAttributes;
}>;


export type CreateGroupMutation = { __typename?: 'RootMutationType', createGroup?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null };

export type UpdateGroupMutationVariables = Exact<{
  id: Scalars['ID'];
  attributes: GroupAttributes;
}>;


export type UpdateGroupMutation = { __typename?: 'RootMutationType', updateGroup?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null };

export type DeleteGroupMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type DeleteGroupMutation = { __typename?: 'RootMutationType', deleteGroup?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null };

export type GroupsQueryVariables = Exact<{
  q?: InputMaybe<Scalars['String']>;
  cursor?: InputMaybe<Scalars['String']>;
}>;


export type GroupsQuery = { __typename?: 'RootQueryType', groups?: { __typename?: 'GroupConnection', pageInfo: { __typename?: 'PageInfo', endCursor?: string | null, hasNextPage: boolean }, edges?: Array<{ __typename?: 'GroupEdge', node?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null } | null };

export type PostmortemFragment = { __typename?: 'Postmortem', id: string, content: string, actionItems?: Array<{ __typename?: 'ActionItem', type: ActionItemType, link: string } | null> | null };

export type FollowerFragment = { __typename?: 'Follower', id: string, incident?: { __typename?: 'Incident', id: string } | null, user: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null }, preferences?: { __typename?: 'NotificationPreferences', message?: boolean | null, incidentUpdate?: boolean | null, mention?: boolean | null } | null };

export type SlimSubscriptionFragment = { __typename?: 'SlimSubscription', id: string, lineItems?: { __typename?: 'SubscriptionLineItems', items?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null } | null, plan?: { __typename?: 'Plan', id: string, name: string, cost: number, period?: string | null, serviceLevels?: Array<{ __typename?: 'ServiceLevel', minSeverity?: number | null, maxSeverity?: number | null, responseTime?: number | null } | null> | null, lineItems?: { __typename?: 'PlanLineItems', included?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null, items?: Array<{ __typename?: 'LineItem', name: string, dimension: string, cost: number, period?: string | null, type?: PlanType | null } | null> | null } | null, metadata?: { __typename?: 'PlanMetadata', features?: Array<{ __typename?: 'PlanFeature', name: string, description: string } | null> | null } | null } | null };

export type ClusterInformationFragment = { __typename?: 'ClusterInformation', version?: string | null, gitCommit?: string | null, platform?: string | null };

export type IncidentFragment = { __typename?: 'Incident', id: string, title: string, description?: string | null, severity: number, status: IncidentStatus, notificationCount?: number | null, nextResponseAt?: Date | null, insertedAt?: Date | null, creator: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null }, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, repository: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null }, subscription?: { __typename?: 'SlimSubscription', id: string, lineItems?: { __typename?: 'SubscriptionLineItems', items?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null } | null, plan?: { __typename?: 'Plan', id: string, name: string, cost: number, period?: string | null, serviceLevels?: Array<{ __typename?: 'ServiceLevel', minSeverity?: number | null, maxSeverity?: number | null, responseTime?: number | null } | null> | null, lineItems?: { __typename?: 'PlanLineItems', included?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null, items?: Array<{ __typename?: 'LineItem', name: string, dimension: string, cost: number, period?: string | null, type?: PlanType | null } | null> | null } | null, metadata?: { __typename?: 'PlanMetadata', features?: Array<{ __typename?: 'PlanFeature', name: string, description: string } | null> | null } | null } | null } | null, clusterInformation?: { __typename?: 'ClusterInformation', version?: string | null, gitCommit?: string | null, platform?: string | null } | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null };

export type IncidentHistoryFragment = { __typename?: 'IncidentHistory', id: string, action: IncidentAction, insertedAt?: Date | null, changes?: Array<{ __typename?: 'IncidentChange', key: string, prev?: string | null, next?: string | null } | null> | null, actor: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } };

export type FileFragment = { __typename?: 'File', id: string, blob: string, mediaType?: MediaType | null, contentType?: string | null, filesize?: number | null, filename?: string | null };

export type IncidentMessageFragment = { __typename?: 'IncidentMessage', id: string, text: string, insertedAt?: Date | null, creator: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null }, reactions?: Array<{ __typename?: 'Reaction', name: string, creator: { __typename?: 'User', id: string, email: string } } | null> | null, file?: { __typename?: 'File', id: string, blob: string, mediaType?: MediaType | null, contentType?: string | null, filesize?: number | null, filename?: string | null } | null, entities?: Array<{ __typename?: 'MessageEntity', type: MessageEntityType, text?: string | null, startIndex?: number | null, endIndex?: number | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null> | null };

export type NotificationFragment = { __typename?: 'Notification', id: string, type: NotificationType, msg?: string | null, insertedAt?: Date | null, actor: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null }, incident?: { __typename?: 'Incident', id: string, title: string, repository: { __typename?: 'Repository', id: string, name: string, icon?: string | null, darkIcon?: string | null } } | null, message?: { __typename?: 'IncidentMessage', text: string } | null, repository?: { __typename?: 'Repository', id: string, name: string, icon?: string | null, darkIcon?: string | null } | null };

export type InstallationFragment = { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null };

export type GetInstallationQueryVariables = Exact<{
  name?: InputMaybe<Scalars['String']>;
}>;


export type GetInstallationQuery = { __typename?: 'RootQueryType', installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null };

export type GetInstallationByIdQueryVariables = Exact<{
  id?: InputMaybe<Scalars['ID']>;
}>;


export type GetInstallationByIdQuery = { __typename?: 'RootQueryType', installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null };

export type GetInstallationsQueryVariables = Exact<{
  first?: InputMaybe<Scalars['Int']>;
}>;


export type GetInstallationsQuery = { __typename?: 'RootQueryType', installations?: { __typename?: 'InstallationConnection', edges?: Array<{ __typename?: 'InstallationEdge', node?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null } | null> | null } | null };

export type UpsertOidcProviderMutationVariables = Exact<{
  id: Scalars['ID'];
  attributes: OidcAttributes;
}>;


export type UpsertOidcProviderMutation = { __typename?: 'RootMutationType', upsertOidcProvider?: { __typename?: 'OidcProvider', id: string } | null };

export type IntegrationWebhookFragment = { __typename?: 'IntegrationWebhook', id: string, name: string, url: string, secret: string, actions?: Array<string | null> | null };

export type WebhookLogFragment = { __typename?: 'WebhookLog', id: string, state: WebhookLogState, status?: number | null, payload?: Map<string, unknown> | null, response?: string | null, insertedAt?: Date | null };

export type OauthIntegrationFragment = { __typename?: 'OauthIntegration', id: string, service: OauthService, insertedAt?: Date | null };

export type ZoomMeetingFragment = { __typename?: 'ZoomMeeting', joinUrl: string, password?: string | null };

export type SignupInviteMutationVariables = Exact<{
  attributes: UserAttributes;
  inviteId: Scalars['String'];
}>;


export type SignupInviteMutation = { __typename?: 'RootMutationType', signup?: { __typename?: 'User', jwt?: string | null } | null };

export type RealizeInviteMutationVariables = Exact<{
  id: Scalars['String'];
}>;


export type RealizeInviteMutation = { __typename?: 'RootMutationType', realizeInvite?: { __typename?: 'User', jwt?: string | null } | null };

export type InviteQueryVariables = Exact<{
  id: Scalars['String'];
}>;


export type InviteQuery = { __typename?: 'RootQueryType', invite?: { __typename?: 'Invite', id: string, email?: string | null, existing: boolean, account?: { __typename?: 'Account', id: string, name?: string | null, billingCustomerId?: string | null, backgroundColor?: string | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, account: { __typename?: 'Account', id: string, name?: string | null, billingCustomerId?: string | null, backgroundColor?: string | null }, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null };

export type MetricFragment = { __typename?: 'Metric', name: string, tags?: Array<{ __typename?: 'MetricTag', name: string, value: string } | null> | null, values?: Array<{ __typename?: 'MetricValue', time?: Date | null, value?: number | null } | null> | null };

export type PageInfoFragment = { __typename?: 'PageInfo', endCursor?: string | null, hasNextPage: boolean };

export type OidcProviderFragment = { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null };

export type OAuthInfoFragment = { __typename?: 'OauthInfo', provider: OauthProvider, authorizeUrl: string };

export type LimitFragment = { __typename?: 'Limit', dimension: string, quantity: number };

export type LineItemFragment = { __typename?: 'LineItem', name: string, dimension: string, cost: number, period?: string | null, type?: PlanType | null };

export type ServiceLevelFragment = { __typename?: 'ServiceLevel', minSeverity?: number | null, maxSeverity?: number | null, responseTime?: number | null };

export type PlanFragment = { __typename?: 'Plan', id: string, name: string, cost: number, period?: string | null, serviceLevels?: Array<{ __typename?: 'ServiceLevel', minSeverity?: number | null, maxSeverity?: number | null, responseTime?: number | null } | null> | null, lineItems?: { __typename?: 'PlanLineItems', included?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null, items?: Array<{ __typename?: 'LineItem', name: string, dimension: string, cost: number, period?: string | null, type?: PlanType | null } | null> | null } | null, metadata?: { __typename?: 'PlanMetadata', features?: Array<{ __typename?: 'PlanFeature', name: string, description: string } | null> | null } | null };

export type SubscriptionFragment = { __typename?: 'RepositorySubscription', id: string, plan?: { __typename?: 'Plan', id: string, name: string, cost: number, period?: string | null, serviceLevels?: Array<{ __typename?: 'ServiceLevel', minSeverity?: number | null, maxSeverity?: number | null, responseTime?: number | null } | null> | null, lineItems?: { __typename?: 'PlanLineItems', included?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null, items?: Array<{ __typename?: 'LineItem', name: string, dimension: string, cost: number, period?: string | null, type?: PlanType | null } | null> | null } | null, metadata?: { __typename?: 'PlanMetadata', features?: Array<{ __typename?: 'PlanFeature', name: string, description: string } | null> | null } | null } | null, lineItems?: { __typename?: 'SubscriptionLineItems', items?: Array<{ __typename?: 'Limit', dimension: string, quantity: number } | null> | null } | null };

export type InvoiceItemFragment = { __typename?: 'InvoiceItem', amount: number, currency: string, description?: string | null };

export type InvoiceFragment = { __typename?: 'Invoice', number: string, amountDue: number, amountPaid: number, currency: string, status?: string | null, createdAt?: Date | null, hostedInvoiceUrl?: string | null, lines?: Array<{ __typename?: 'InvoiceItem', amount: number, currency: string, description?: string | null } | null> | null };

export type CardFragment = { __typename?: 'Card', id: string, last4: string, expMonth: number, expYear: number, name?: string | null, brand: string };

export type SubscriptionQueryVariables = Exact<{ [key: string]: never; }>;


export type SubscriptionQuery = { __typename?: 'RootQueryType', account?: { __typename?: 'Account', billingCustomerId?: string | null, grandfatheredUntil?: Date | null, delinquentAt?: Date | null, userCount?: string | null, clusterCount?: string | null, availableFeatures?: { __typename?: 'PlanFeatures', userManagement?: boolean | null, audit?: boolean | null } | null, subscription?: { __typename?: 'PlatformSubscription', id: string, plan?: { __typename?: 'PlatformPlan', id: string, period: PaymentPeriod, lineItems?: Array<{ __typename?: 'PlatformPlanItem', dimension: LineItemDimension, cost: number } | null> | null } | null } | null, billingAddress?: { __typename?: 'Address', name?: string | null, line1?: string | null, line2?: string | null, zip?: string | null, state?: string | null, city?: string | null, country?: string | null } | null } | null };

export type UpdateAccountBillingMutationVariables = Exact<{
  attributes: AccountAttributes;
}>;


export type UpdateAccountBillingMutation = { __typename?: 'RootMutationType', updateAccount?: { __typename?: 'Account', id: string } | null };

export type UpgradeToProfessionalPlanMutationVariables = Exact<{
  planId: Scalars['ID'];
}>;


export type UpgradeToProfessionalPlanMutation = { __typename?: 'RootMutationType', createPlatformSubscription?: { __typename?: 'PlatformSubscription', id: string } | null };

export type DowngradeToFreePlanMutationMutationVariables = Exact<{ [key: string]: never; }>;


export type DowngradeToFreePlanMutationMutation = { __typename?: 'RootMutationType', deletePlatformSubscription?: { __typename?: 'Account', id: string } | null };

export type CardsQueryVariables = Exact<{ [key: string]: never; }>;


export type CardsQuery = { __typename?: 'RootQueryType', me?: { __typename?: 'User', id: string, cards?: { __typename?: 'CardConnection', edges?: Array<{ __typename?: 'CardEdge', node?: { __typename?: 'Card', id: string, last4: string, expMonth: number, expYear: number, name?: string | null, brand: string } | null } | null> | null } | null } | null };

export type CreateCardMutationVariables = Exact<{
  source: Scalars['String'];
  address?: InputMaybe<AddressAttributes>;
}>;


export type CreateCardMutation = { __typename?: 'RootMutationType', createCard?: { __typename?: 'Account', id: string } | null };

export type DeleteCardMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type DeleteCardMutation = { __typename?: 'RootMutationType', deleteCard?: { __typename?: 'Account', id: string } | null };

export type RecipeFragment = { __typename?: 'Recipe', id: string, name: string, description?: string | null, restricted?: boolean | null, provider?: Provider | null, tests?: Array<{ __typename?: 'RecipeTest', type: TestType, name: string, message?: string | null, args?: Array<{ __typename?: 'TestArgument', name: string, repo: string, key: string } | null> | null } | null> | null, repository?: { __typename?: 'Repository', id: string, name: string } | null, oidcSettings?: { __typename?: 'OidcSettings', uriFormat?: string | null, uriFormats?: Array<string | null> | null, authMethod: OidcAuthMethod, domainKey?: string | null, subdomain?: boolean | null } | null, recipeSections?: Array<{ __typename?: 'RecipeSection', index?: number | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, recipeItems?: Array<{ __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null };

export type RecipeItemFragment = { __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null };

export type RecipeSectionFragment = { __typename?: 'RecipeSection', index?: number | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, recipeItems?: Array<{ __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null };

export type RecipeConfigurationFragment = { __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null };

export type StackFragment = { __typename?: 'Stack', id: string, name: string, displayName?: string | null, description?: string | null, featured?: boolean | null, creator?: { __typename?: 'User', id: string, name: string } | null, collections?: Array<{ __typename?: 'StackCollection', id: string, provider: Provider, bundles?: Array<{ __typename?: 'StackRecipe', recipe: { __typename?: 'Recipe', repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null } } | null> | null } | null> | null };

export type GetRecipeQueryVariables = Exact<{
  repo?: InputMaybe<Scalars['String']>;
  name?: InputMaybe<Scalars['String']>;
  id?: InputMaybe<Scalars['ID']>;
}>;


export type GetRecipeQuery = { __typename?: 'RootQueryType', recipe?: { __typename?: 'Recipe', id: string, name: string, description?: string | null, restricted?: boolean | null, provider?: Provider | null, recipeDependencies?: Array<{ __typename?: 'Recipe', id: string, name: string, description?: string | null, restricted?: boolean | null, provider?: Provider | null, tests?: Array<{ __typename?: 'RecipeTest', type: TestType, name: string, message?: string | null, args?: Array<{ __typename?: 'TestArgument', name: string, repo: string, key: string } | null> | null } | null> | null, repository?: { __typename?: 'Repository', id: string, name: string } | null, oidcSettings?: { __typename?: 'OidcSettings', uriFormat?: string | null, uriFormats?: Array<string | null> | null, authMethod: OidcAuthMethod, domainKey?: string | null, subdomain?: boolean | null } | null, recipeSections?: Array<{ __typename?: 'RecipeSection', index?: number | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, recipeItems?: Array<{ __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null } | null> | null, tests?: Array<{ __typename?: 'RecipeTest', type: TestType, name: string, message?: string | null, args?: Array<{ __typename?: 'TestArgument', name: string, repo: string, key: string } | null> | null } | null> | null, repository?: { __typename?: 'Repository', id: string, name: string } | null, oidcSettings?: { __typename?: 'OidcSettings', uriFormat?: string | null, uriFormats?: Array<string | null> | null, authMethod: OidcAuthMethod, domainKey?: string | null, subdomain?: boolean | null } | null, recipeSections?: Array<{ __typename?: 'RecipeSection', index?: number | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, recipeItems?: Array<{ __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null } | null };

export type ListRecipesQueryVariables = Exact<{
  repositoryName?: InputMaybe<Scalars['String']>;
  repositoryId?: InputMaybe<Scalars['ID']>;
  provider?: InputMaybe<Provider>;
}>;


export type ListRecipesQuery = { __typename?: 'RootQueryType', recipes?: { __typename?: 'RecipeConnection', edges?: Array<{ __typename?: 'RecipeEdge', node?: { __typename?: 'Recipe', id: string, name: string, description?: string | null, restricted?: boolean | null, provider?: Provider | null, tests?: Array<{ __typename?: 'RecipeTest', type: TestType, name: string, message?: string | null, args?: Array<{ __typename?: 'TestArgument', name: string, repo: string, key: string } | null> | null } | null> | null, repository?: { __typename?: 'Repository', id: string, name: string } | null, oidcSettings?: { __typename?: 'OidcSettings', uriFormat?: string | null, uriFormats?: Array<string | null> | null, authMethod: OidcAuthMethod, domainKey?: string | null, subdomain?: boolean | null } | null, recipeSections?: Array<{ __typename?: 'RecipeSection', index?: number | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, recipeItems?: Array<{ __typename?: 'RecipeItem', id?: string | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null, configuration?: Array<{ __typename?: 'RecipeConfiguration', name?: string | null, type?: Datatype | null, default?: string | null, documentation?: string | null, optional?: boolean | null, placeholder?: string | null, functionName?: string | null, condition?: { __typename?: 'RecipeCondition', field: string, operation: Operation, value?: string | null } | null, validation?: { __typename?: 'RecipeValidation', type: ValidationType, regex?: string | null, message: string } | null } | null> | null } | null> | null } | null } | null> | null } | null };

export type CreateRecipeMutationVariables = Exact<{
  name: Scalars['String'];
  attributes: RecipeAttributes;
}>;


export type CreateRecipeMutation = { __typename?: 'RootMutationType', createRecipe?: { __typename?: 'Recipe', id: string } | null };

export type InstallRecipeMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type InstallRecipeMutation = { __typename?: 'RootMutationType', installRecipe?: Array<{ __typename?: 'Installation', id: string } | null> | null };

export type CreateStackMutationVariables = Exact<{
  attributes: StackAttributes;
}>;


export type CreateStackMutation = { __typename?: 'RootMutationType', createStack?: { __typename?: 'Stack', id: string } | null };

export type GetStackQueryVariables = Exact<{
  name: Scalars['String'];
  provider: Provider;
}>;


export type GetStackQuery = { __typename?: 'RootQueryType', stack?: { __typename?: 'Stack', id: string, name: string, displayName?: string | null, description?: string | null, featured?: boolean | null, creator?: { __typename?: 'User', id: string, name: string } | null, collections?: Array<{ __typename?: 'StackCollection', id: string, provider: Provider, bundles?: Array<{ __typename?: 'StackRecipe', recipe: { __typename?: 'Recipe', repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null } } | null> | null } | null> | null } | null };

export type ListStacksQueryVariables = Exact<{
  featured?: InputMaybe<Scalars['Boolean']>;
  cursor?: InputMaybe<Scalars['String']>;
}>;


export type ListStacksQuery = { __typename?: 'RootQueryType', stacks?: { __typename?: 'StackConnection', edges?: Array<{ __typename?: 'StackEdge', node?: { __typename?: 'Stack', id: string, name: string, displayName?: string | null, description?: string | null, featured?: boolean | null, creator?: { __typename?: 'User', id: string, name: string } | null, collections?: Array<{ __typename?: 'StackCollection', id: string, provider: Provider, bundles?: Array<{ __typename?: 'StackRecipe', recipe: { __typename?: 'Recipe', repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null } } | null> | null } | null> | null } | null } | null> | null } | null };

export type CreateQuickStackMutationVariables = Exact<{
  applicationIds?: InputMaybe<Array<InputMaybe<Scalars['ID']>> | InputMaybe<Scalars['ID']>>;
  provider: Provider;
}>;


export type CreateQuickStackMutation = { __typename?: 'RootMutationType', quickStack?: { __typename?: 'Stack', id: string, name: string } | null };

export type InstallStackShellMutationVariables = Exact<{
  name: Scalars['String'];
  context: ContextAttributes;
  oidc: Scalars['Boolean'];
}>;


export type InstallStackShellMutation = { __typename?: 'RootMutationType', installStackShell?: Array<{ __typename?: 'Recipe', id: string, name: string } | null> | null };

export type ApplyLockFragment = { __typename?: 'ApplyLock', id: string, lock?: string | null };

export type CategoryFragment = { __typename?: 'CategoryInfo', category?: Category | null, count?: number | null };

export type RepoFragment = { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null };

export type RepositoryFragment = { __typename?: 'Repository', id: string, name: string, notes?: string | null, icon?: string | null, darkIcon?: string | null, description?: string | null, publisher?: { __typename?: 'Publisher', name: string } | null, recipes?: Array<{ __typename?: 'Recipe', name: string } | null> | null };

export type DependenciesFragment = { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null };

export type IntegrationFragment = { __typename?: 'Integration', id: string, name: string, icon?: string | null, sourceUrl?: string | null, description?: string | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null };

export type RepositoryQueryVariables = Exact<{
  id?: InputMaybe<Scalars['ID']>;
  name?: InputMaybe<Scalars['String']>;
}>;


export type RepositoryQuery = { __typename?: 'RootQueryType', repository?: { __typename?: 'Repository', editable?: boolean | null, publicKey?: string | null, secrets?: Map<string, unknown> | null, readme?: string | null, mainBranch?: string | null, gitUrl?: string | null, homepage?: string | null, documentation?: string | null, id: string, name: string, notes?: string | null, description?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, artifacts?: Array<{ __typename?: 'Artifact', id?: string | null, name?: string | null, blob?: string | null, type?: ArtifactType | null, platform?: ArtifactPlatform | null, arch?: string | null, filesize?: number | null, sha?: string | null, readme?: string | null, insertedAt?: Date | null, updatedAt?: Date | null } | null> | null, installation?: { __typename?: 'Installation', id: string, context?: Map<string, unknown> | null, license?: string | null, licenseKey?: string | null, acmeKeyId?: string | null, acmeSecret?: string | null, autoUpgrade?: boolean | null, trackTag: string, oidcProvider?: { __typename?: 'OidcProvider', id: string, clientId: string, authMethod: OidcAuthMethod, clientSecret: string, redirectUris?: Array<string | null> | null, bindings?: Array<{ __typename?: 'OidcProviderBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null, configuration?: { __typename?: 'OuathConfiguration', issuer?: string | null, authorizationEndpoint?: string | null, tokenEndpoint?: string | null, jwksUri?: string | null, userinfoEndpoint?: string | null } | null } | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null } | null, tags?: Array<{ __typename?: 'Tag', tag: string } | null> | null, license?: { __typename?: 'License', name?: string | null, url?: string | null } | null, community?: { __typename?: 'Community', discord?: string | null, slack?: string | null, homepage?: string | null, gitUrl?: string | null, twitter?: string | null } | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null };

export type CreateResourceDefinitionMutationVariables = Exact<{
  name: Scalars['String'];
  input: ResourceDefinitionAttributes;
}>;


export type CreateResourceDefinitionMutation = { __typename?: 'RootMutationType', updateRepository?: { __typename?: 'Repository', id: string } | null };

export type CreateIntegrationMutationVariables = Exact<{
  name: Scalars['String'];
  attrs: IntegrationAttributes;
}>;


export type CreateIntegrationMutation = { __typename?: 'RootMutationType', createIntegration?: { __typename?: 'Integration', id: string } | null };

export type UpdateRepositoryMutationVariables = Exact<{
  name: Scalars['String'];
  attrs: RepositoryAttributes;
}>;


export type UpdateRepositoryMutation = { __typename?: 'RootMutationType', updateRepository?: { __typename?: 'Repository', id: string } | null };

export type CreateRepositoryMutationVariables = Exact<{
  name: Scalars['String'];
  publisher: Scalars['String'];
  attributes: RepositoryAttributes;
}>;


export type CreateRepositoryMutation = { __typename?: 'RootMutationType', upsertRepository?: { __typename?: 'Repository', id: string } | null };

export type AcquireLockMutationVariables = Exact<{
  name: Scalars['String'];
}>;


export type AcquireLockMutation = { __typename?: 'RootMutationType', acquireLock?: { __typename?: 'ApplyLock', id: string, lock?: string | null } | null };

export type ReleaseLockMutationVariables = Exact<{
  name: Scalars['String'];
  attrs: LockAttributes;
}>;


export type ReleaseLockMutation = { __typename?: 'RootMutationType', releaseLock?: { __typename?: 'ApplyLock', id: string, lock?: string | null } | null };

export type UnlockRepositoryMutationVariables = Exact<{
  name: Scalars['String'];
}>;


export type UnlockRepositoryMutation = { __typename?: 'RootMutationType', unlockRepository?: number | null };

export type ListRepositoriesQueryVariables = Exact<{
  q?: InputMaybe<Scalars['String']>;
  installed?: InputMaybe<Scalars['Boolean']>;
  first?: InputMaybe<Scalars['Int']>;
  provider?: InputMaybe<Provider>;
}>;


export type ListRepositoriesQuery = { __typename?: 'RootQueryType', repositories?: { __typename?: 'RepositoryConnection', edges?: Array<{ __typename?: 'RepositoryEdge', node?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, icon?: string | null, darkIcon?: string | null, description?: string | null, publisher?: { __typename?: 'Publisher', name: string } | null, recipes?: Array<{ __typename?: 'Recipe', name: string } | null> | null } | null } | null> | null } | null };

export type ScaffoldsQueryVariables = Exact<{
  app: Scalars['String'];
  pub: Scalars['String'];
  cat: Category;
  ing?: InputMaybe<Scalars['Boolean']>;
  pg?: InputMaybe<Scalars['Boolean']>;
}>;


export type ScaffoldsQuery = { __typename?: 'RootQueryType', scaffold?: Array<{ __typename?: 'ScaffoldFile', path?: string | null, content?: string | null } | null> | null };

export type DeleteRepositoryMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type DeleteRepositoryMutation = { __typename?: 'RootMutationType', deleteRepository?: { __typename?: 'Repository', id: string } | null };

export type GetTfProvidersQueryVariables = Exact<{ [key: string]: never; }>;


export type GetTfProvidersQuery = { __typename?: 'RootQueryType', terraformProviders?: Array<Provider | null> | null };

export type GetTfProviderScaffoldQueryVariables = Exact<{
  name: Provider;
  vsn?: InputMaybe<Scalars['String']>;
}>;


export type GetTfProviderScaffoldQuery = { __typename?: 'RootQueryType', terraformProvider?: { __typename?: 'TerraformProvider', name?: Provider | null, content?: string | null } | null };

export type CloudShellFragment = { __typename?: 'CloudShell', id: string, aesKey: string, gitUrl: string, alive: boolean, provider: Provider, subdomain: string, cluster: string, status?: { __typename?: 'ShellStatus', ready?: boolean | null, initialized?: boolean | null, containersReady?: boolean | null, podScheduled?: boolean | null } | null };

export type DemoProjectFragment = { __typename?: 'DemoProject', id: string, projectId: string, credentials?: string | null, ready?: boolean | null, state?: DemoProjectState | null };

export type GetShellQueryVariables = Exact<{ [key: string]: never; }>;


export type GetShellQuery = { __typename?: 'RootQueryType', shell?: { __typename?: 'CloudShell', id: string, aesKey: string, gitUrl: string, alive: boolean, provider: Provider, subdomain: string, cluster: string, status?: { __typename?: 'ShellStatus', ready?: boolean | null, initialized?: boolean | null, containersReady?: boolean | null, podScheduled?: boolean | null } | null } | null };

export type DeleteShellMutationVariables = Exact<{ [key: string]: never; }>;


export type DeleteShellMutation = { __typename?: 'RootMutationType', deleteShell?: { __typename?: 'CloudShell', id: string, aesKey: string, gitUrl: string, alive: boolean, provider: Provider, subdomain: string, cluster: string, status?: { __typename?: 'ShellStatus', ready?: boolean | null, initialized?: boolean | null, containersReady?: boolean | null, podScheduled?: boolean | null } | null } | null };

export type TerraformFragment = { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null };

export type TerraformInstallationFragment = { __typename?: 'TerraformInstallation', id?: string | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null };

export type GetTerraformQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetTerraformQuery = { __typename?: 'RootQueryType', terraform?: { __typename?: 'TerraformConnection', edges?: Array<{ __typename?: 'TerraformEdge', node?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null> | null } | null };

export type GetTerraformInstallationsQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type GetTerraformInstallationsQuery = { __typename?: 'RootQueryType', terraformInstallations?: { __typename?: 'TerraformInstallationConnection', edges?: Array<{ __typename?: 'TerraformInstallationEdge', node?: { __typename?: 'TerraformInstallation', id?: string | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, version?: { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null } | null } | null> | null } | null };

export type UploadTerraformMutationVariables = Exact<{
  repoName: Scalars['String'];
  name: Scalars['String'];
  uploadOrUrl: Scalars['UploadOrUrl'];
}>;


export type UploadTerraformMutation = { __typename?: 'RootMutationType', uploadTerraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null, readme?: string | null, package?: string | null, description?: string | null, latestVersion?: string | null, valuesTemplate?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null };

export type UninstallTerraformMutationVariables = Exact<{
  id: Scalars['ID'];
}>;


export type UninstallTerraformMutation = { __typename?: 'RootMutationType', uninstallTerraform?: { __typename?: 'TerraformInstallation', id?: string | null } | null };

export type StepFragment = { __typename?: 'TestStep', id: string, name: string, status: TestStatus, hasLogs?: boolean | null, description: string, insertedAt?: Date | null, updatedAt?: Date | null };

export type TestFragment = { __typename?: 'Test', id: string, name?: string | null, promoteTag: string, status: TestStatus, insertedAt?: Date | null, updatedAt?: Date | null, steps?: Array<{ __typename?: 'TestStep', id: string, name: string, status: TestStatus, hasLogs?: boolean | null, description: string, insertedAt?: Date | null, updatedAt?: Date | null } | null> | null };

export type CreateTestMutationVariables = Exact<{
  name: Scalars['String'];
  attrs: TestAttributes;
}>;


export type CreateTestMutation = { __typename?: 'RootMutationType', createTest?: { __typename?: 'Test', id: string, name?: string | null, promoteTag: string, status: TestStatus, insertedAt?: Date | null, updatedAt?: Date | null, steps?: Array<{ __typename?: 'TestStep', id: string, name: string, status: TestStatus, hasLogs?: boolean | null, description: string, insertedAt?: Date | null, updatedAt?: Date | null } | null> | null } | null };

export type UpdateTestMutationVariables = Exact<{
  id: Scalars['ID'];
  attrs: TestAttributes;
}>;


export type UpdateTestMutation = { __typename?: 'RootMutationType', updateTest?: { __typename?: 'Test', id: string, name?: string | null, promoteTag: string, status: TestStatus, insertedAt?: Date | null, updatedAt?: Date | null, steps?: Array<{ __typename?: 'TestStep', id: string, name: string, status: TestStatus, hasLogs?: boolean | null, description: string, insertedAt?: Date | null, updatedAt?: Date | null } | null> | null } | null };

export type UpdateStepMutationVariables = Exact<{
  id: Scalars['ID'];
  logs: Scalars['UploadOrUrl'];
}>;


export type UpdateStepMutation = { __typename?: 'RootMutationType', updateStep?: { __typename?: 'TestStep', id: string } | null };

export type PublishLogsMutationVariables = Exact<{
  id: Scalars['ID'];
  logs: Scalars['String'];
}>;


export type PublishLogsMutation = { __typename?: 'RootMutationType', publishLogs?: { __typename?: 'TestStep', id: string } | null };

export type UpgradeQueueFragment = { __typename?: 'UpgradeQueue', id: string, acked?: string | null, name?: string | null, domain?: string | null, git?: string | null, pingedAt?: Date | null, provider?: Provider | null };

export type RolloutFragment = { __typename?: 'Rollout', id: string, event?: string | null, cursor?: string | null, count?: number | null, status: RolloutStatus, heartbeat?: Date | null };

export type UpgradeFragment = { __typename?: 'Upgrade', id: string, message?: string | null, insertedAt?: Date | null, repository?: { __typename?: 'Repository', id: string, name: string, notes?: string | null, description?: string | null, documentation?: string | null, icon?: string | null, darkIcon?: string | null, private?: boolean | null, trending?: boolean | null, verified?: boolean | null, category?: Category | null, oauthSettings?: { __typename?: 'OauthSettings', uriFormat: string, authMethod: OidcAuthMethod } | null, publisher?: { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, recipes?: Array<{ __typename?: 'Recipe', name: string, provider?: Provider | null, description?: string | null } | null> | null } | null };

export type DeferredUpdateFragment = { __typename?: 'DeferredUpdate', id: string, dequeueAt?: Date | null, attempts?: number | null, insertedAt?: Date | null, version?: { __typename?: 'Version', version: string } | null };

export type AccountFragment = { __typename?: 'Account', id: string, name?: string | null, billingCustomerId?: string | null, backgroundColor?: string | null };

export type GroupFragment = { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null };

export type UserFragment = { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null };

export type ImpersonationPolicyFragment = { __typename?: 'ImpersonationPolicy', id: string, bindings?: Array<{ __typename?: 'ImpersonationPolicyBinding', id: string, group?: { __typename?: 'Group', id: string, name: string } | null, user?: { __typename?: 'User', id: string, name: string, email: string } | null } | null> | null };

export type GroupMemberFragment = { __typename?: 'GroupMember', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null };

export type TokenFragment = { __typename?: 'PersistedToken', id?: string | null, token?: string | null, insertedAt?: Date | null };

export type TokenAuditFragment = { __typename?: 'PersistedTokenAudit', ip?: string | null, timestamp?: Date | null, count?: number | null, country?: string | null, city?: string | null, latitude?: string | null, longitude?: string | null };

export type AddressFragment = { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null };

export type PublisherFragment = { __typename?: 'Publisher', id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null };

export type WebhookFragment = { __typename?: 'Webhook', id?: string | null, url?: string | null, secret?: string | null, insertedAt?: Date | null };

export type RoleBindingFragment = { __typename?: 'RoleBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null };

export type RoleFragment = { __typename?: 'Role', id: string, name: string, description?: string | null, repositories?: Array<string | null> | null, permissions?: Array<Permission | null> | null, roleBindings?: Array<{ __typename?: 'RoleBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null };

export type PublicKeyFragment = { __typename?: 'PublicKey', id: string, name: string, digest: string, insertedAt?: Date | null, content: string, user: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } };

export type EabCredentialFragment = { __typename?: 'EabCredential', id: string, keyId: string, hmacKey: string, cluster: string, provider: Provider, insertedAt?: Date | null };

export type MeQueryVariables = Exact<{ [key: string]: never; }>;


export type MeQuery = { __typename?: 'RootQueryType', me?: { __typename?: 'User', loginMethod?: LoginMethod | null, hasInstallations?: boolean | null, id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, account: { __typename?: 'Account', id: string, name?: string | null, billingCustomerId?: string | null, backgroundColor?: string | null, rootUser?: { __typename?: 'User', id: string, name: string, email: string } | null, domainMappings?: Array<{ __typename?: 'DomainMapping', id: string, domain: string, enableSso?: boolean | null } | null> | null }, publisher?: { __typename?: 'Publisher', billingAccountId?: string | null, id?: string | null, name: string, phone?: string | null, avatar?: string | null, description?: string | null, backgroundColor?: string | null, owner?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, address?: { __typename?: 'Address', line1?: string | null, line2?: string | null, city?: string | null, country?: string | null, state?: string | null, zip?: string | null } | null } | null, boundRoles?: Array<{ __typename?: 'Role', id: string, name: string, description?: string | null, repositories?: Array<string | null> | null, permissions?: Array<Permission | null> | null, roleBindings?: Array<{ __typename?: 'RoleBinding', id: string, user?: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, group?: { __typename?: 'Group', id: string, name: string, global?: boolean | null, description?: string | null } | null } | null> | null } | null> | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } | null, configuration?: { __typename?: 'PluralConfiguration', stripeConnectId?: string | null, stripePublishableKey?: string | null, registry?: string | null, gitCommit?: string | null } | null };

export type GetLoginMethodQueryVariables = Exact<{
  email: Scalars['String'];
}>;


export type GetLoginMethodQuery = { __typename?: 'RootQueryType', loginMethod?: { __typename?: 'LoginMethodResponse', loginMethod: LoginMethod, token?: string | null } | null };

export type ListTokensQueryVariables = Exact<{ [key: string]: never; }>;


export type ListTokensQuery = { __typename?: 'RootQueryType', tokens?: { __typename?: 'PersistedTokenConnection', edges?: Array<{ __typename?: 'PersistedTokenEdge', node?: { __typename?: 'PersistedToken', token?: string | null } | null } | null> | null } | null };

export type ListKeysQueryVariables = Exact<{
  emails?: InputMaybe<Array<InputMaybe<Scalars['String']>> | InputMaybe<Scalars['String']>>;
}>;


export type ListKeysQuery = { __typename?: 'RootQueryType', publicKeys?: { __typename?: 'PublicKeyConnection', edges?: Array<{ __typename?: 'PublicKeyEdge', node?: { __typename?: 'PublicKey', id: string, name: string, digest: string, insertedAt?: Date | null, content: string, user: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } } | null } | null> | null } | null };

export type GetEabCredentialQueryVariables = Exact<{
  cluster: Scalars['String'];
  provider: Provider;
}>;


export type GetEabCredentialQuery = { __typename?: 'RootQueryType', eabCredential?: { __typename?: 'EabCredential', id: string, keyId: string, hmacKey: string, cluster: string, provider: Provider, insertedAt?: Date | null } | null };

export type DevLoginMutationVariables = Exact<{ [key: string]: never; }>;


export type DevLoginMutation = { __typename?: 'RootMutationType', deviceLogin?: { __typename?: 'DeviceLogin', loginUrl: string, deviceToken: string } | null };

export type LoginMutationVariables = Exact<{
  email: Scalars['String'];
  password: Scalars['String'];
  deviceToken?: InputMaybe<Scalars['String']>;
}>;


export type LoginMutation = { __typename?: 'RootMutationType', login?: { __typename?: 'User', jwt?: string | null } | null };

export type ImpersonateServiceAccountMutationVariables = Exact<{
  email?: InputMaybe<Scalars['String']>;
}>;


export type ImpersonateServiceAccountMutation = { __typename?: 'RootMutationType', impersonateServiceAccount?: { __typename?: 'User', jwt?: string | null, email: string } | null };

export type CreateAccessTokenMutationVariables = Exact<{ [key: string]: never; }>;


export type CreateAccessTokenMutation = { __typename?: 'RootMutationType', createToken?: { __typename?: 'PersistedToken', token?: string | null } | null };

export type CreateKeyMutationVariables = Exact<{
  key: Scalars['String'];
  name: Scalars['String'];
}>;


export type CreateKeyMutation = { __typename?: 'RootMutationType', createPublicKey?: { __typename?: 'PublicKey', id: string } | null };

export type DeleteEabCredentialMutationVariables = Exact<{
  cluster: Scalars['String'];
  provider: Provider;
}>;


export type DeleteEabCredentialMutation = { __typename?: 'RootMutationType', deleteEabKey?: { __typename?: 'EabCredential', id: string } | null };

export type CreateEventMutationVariables = Exact<{
  attrs: UserEventAttributes;
}>;


export type CreateEventMutation = { __typename?: 'RootMutationType', createUserEvent?: boolean | null };

export type LoginMethodQueryVariables = Exact<{
  email: Scalars['String'];
  host?: InputMaybe<Scalars['String']>;
}>;


export type LoginMethodQuery = { __typename?: 'RootQueryType', loginMethod?: { __typename?: 'LoginMethodResponse', loginMethod: LoginMethod, token?: string | null, authorizeUrl?: string | null } | null };

export type SignupMutationVariables = Exact<{
  attributes: UserAttributes;
  account?: InputMaybe<AccountAttributes>;
  deviceToken?: InputMaybe<Scalars['String']>;
}>;


export type SignupMutation = { __typename?: 'RootMutationType', signup?: { __typename?: 'User', jwt?: string | null, onboarding?: OnboardingState | null } | null };

export type PasswordlessLoginMutationVariables = Exact<{
  token: Scalars['String'];
}>;


export type PasswordlessLoginMutation = { __typename?: 'RootMutationType', passwordlessLogin?: { __typename?: 'User', jwt?: string | null } | null };

export type PollLoginTokenMutationVariables = Exact<{
  token: Scalars['String'];
  deviceToken?: InputMaybe<Scalars['String']>;
}>;


export type PollLoginTokenMutation = { __typename?: 'RootMutationType', loginToken?: { __typename?: 'User', jwt?: string | null } | null };

export type OauthUrlsQueryVariables = Exact<{
  host?: InputMaybe<Scalars['String']>;
}>;


export type OauthUrlsQuery = { __typename?: 'RootQueryType', oauthUrls?: Array<{ __typename?: 'OauthInfo', provider: OauthProvider, authorizeUrl: string } | null> | null };

export type AcceptLoginMutationVariables = Exact<{
  challenge: Scalars['String'];
}>;


export type AcceptLoginMutation = { __typename?: 'RootMutationType', acceptLogin?: { __typename?: 'OauthResponse', redirectTo: string } | null };

export type CreateResetTokenMutationVariables = Exact<{
  attributes: ResetTokenAttributes;
}>;


export type CreateResetTokenMutation = { __typename?: 'RootMutationType', createResetToken?: boolean | null };

export type RealizeResetTokenMutationVariables = Exact<{
  id: Scalars['ID'];
  attributes: ResetTokenRealization;
}>;


export type RealizeResetTokenMutation = { __typename?: 'RootMutationType', realizeResetToken?: boolean | null };

export type ResetTokenQueryVariables = Exact<{
  id: Scalars['ID'];
}>;


export type ResetTokenQuery = { __typename?: 'RootQueryType', resetToken?: { __typename?: 'ResetToken', type: ResetTokenType, user: { __typename?: 'User', id: string, name: string, email: string, avatar?: string | null, provider?: Provider | null, demoing?: boolean | null, demoed?: boolean | null, onboarding?: OnboardingState | null, emailConfirmed?: boolean | null, emailConfirmBy?: Date | null, backgroundColor?: string | null, serviceAccount?: boolean | null, onboardingChecklist?: { __typename?: 'OnboardingChecklist', dismissed?: boolean | null, status?: OnboardingChecklistState | null } | null, roles?: { __typename?: 'Roles', admin?: boolean | null } | null } } | null };

export type VersionTagFragment = { __typename?: 'VersionTag', id: string, tag: string, version?: { __typename?: 'Version', id: string } | null };

export type VersionFragment = { __typename?: 'Version', id: string, helm?: Map<string, unknown> | null, readme?: string | null, valuesTemplate?: string | null, version: string, insertedAt?: Date | null, package?: string | null, crds?: Array<{ __typename?: 'Crd', id: string, name: string, blob?: string | null } | null> | null, chart?: { __typename?: 'Chart', id?: string | null, name: string, description?: string | null, latestVersion?: string | null, insertedAt?: Date | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null } | null, terraform?: { __typename?: 'Terraform', id?: string | null, name?: string | null } | null, dependencies?: { __typename?: 'Dependencies', wait?: boolean | null, application?: boolean | null, providers?: Array<Provider | null> | null, secrets?: Array<string | null> | null, providerWirings?: Map<string, unknown> | null, outputs?: Map<string, unknown> | null, dependencies?: Array<{ __typename?: 'Dependency', name?: string | null, repo?: string | null, type?: DependencyType | null, version?: string | null, optional?: boolean | null } | null> | null, wirings?: { __typename?: 'Wirings', terraform?: Map<string, unknown> | null, helm?: Map<string, unknown> | null } | null } | null };

export type UpdateVersionMutationVariables = Exact<{
  spec?: InputMaybe<VersionSpec>;
  attributes: VersionAttributes;
}>;


export type UpdateVersionMutation = { __typename?: 'RootMutationType', updateVersion?: { __typename?: 'Version', id: string } | null };

export const UserFragmentDoc = gql`
    fragment User on User {
  id
  name
  email
  avatar
  provider
  demoing
  demoed
  onboarding
  onboardingChecklist {
    dismissed
    status
  }
  emailConfirmed
  emailConfirmBy
  backgroundColor
  serviceAccount
  roles {
    admin
  }
}
    `;
export const AddressFragmentDoc = gql`
    fragment Address on Address {
  line1
  line2
  city
  country
  state
  zip
}
    `;
export const PublisherFragmentDoc = gql`
    fragment Publisher on Publisher {
  id
  name
  phone
  avatar
  description
  backgroundColor
  owner {
    ...User
  }
  address {
    ...Address
  }
}
    ${UserFragmentDoc}
${AddressFragmentDoc}`;
export const RepoFragmentDoc = gql`
    fragment Repo on Repository {
  id
  name
  notes
  description
  documentation
  icon
  darkIcon
  private
  trending
  verified
  category
  oauthSettings {
    uriFormat
    authMethod
  }
  publisher {
    ...Publisher
  }
  recipes {
    name
    provider
    description
  }
}
    ${PublisherFragmentDoc}`;
export const GroupFragmentDoc = gql`
    fragment Group on Group {
  id
  name
  global
  description
}
    `;
export const IntegrationWebhookFragmentDoc = gql`
    fragment IntegrationWebhook on IntegrationWebhook {
  id
  name
  url
  secret
  actions
}
    `;
export const RoleBindingFragmentDoc = gql`
    fragment RoleBinding on RoleBinding {
  id
  user {
    ...User
  }
  group {
    ...Group
  }
}
    ${UserFragmentDoc}
${GroupFragmentDoc}`;
export const RoleFragmentDoc = gql`
    fragment Role on Role {
  id
  name
  description
  repositories
  permissions
  roleBindings {
    ...RoleBinding
  }
}
    ${RoleBindingFragmentDoc}`;
export const CrdFragmentDoc = gql`
    fragment Crd on Crd {
  id
  name
  blob
}
    `;
export const DependenciesFragmentDoc = gql`
    fragment Dependencies on Dependencies {
  dependencies {
    name
    repo
    type
    version
    optional
  }
  wait
  application
  providers
  secrets
  wirings {
    terraform
    helm
  }
  providerWirings
  outputs
}
    `;
export const ChartFragmentDoc = gql`
    fragment Chart on Chart {
  id
  name
  description
  latestVersion
  dependencies {
    ...Dependencies
  }
  insertedAt
}
    ${DependenciesFragmentDoc}`;
export const VersionFragmentDoc = gql`
    fragment Version on Version {
  id
  helm
  readme
  valuesTemplate
  version
  insertedAt
  package
  crds {
    ...Crd
  }
  chart {
    ...Chart
  }
  terraform {
    id
    name
  }
  dependencies {
    ...Dependencies
  }
}
    ${CrdFragmentDoc}
${ChartFragmentDoc}
${DependenciesFragmentDoc}`;
export const AuditFragmentDoc = gql`
    fragment Audit on Audit {
  id
  action
  ip
  country
  city
  latitude
  longitude
  actor {
    ...User
  }
  repository {
    ...Repo
  }
  group {
    ...Group
  }
  integrationWebhook {
    ...IntegrationWebhook
  }
  role {
    ...Role
  }
  version {
    ...Version
  }
  image {
    id
    tag
    dockerRepository {
      name
    }
  }
  insertedAt
}
    ${UserFragmentDoc}
${RepoFragmentDoc}
${GroupFragmentDoc}
${IntegrationWebhookFragmentDoc}
${RoleFragmentDoc}
${VersionFragmentDoc}`;
export const PolicyBindingFragmentDoc = gql`
    fragment PolicyBinding on PolicyBinding {
  id
  group {
    id
    name
  }
  user {
    id
    name
    email
  }
}
    `;
export const DnsDomainFragmentDoc = gql`
    fragment DnsDomain on DnsDomain {
  id
  name
  creator {
    ...User
  }
  accessPolicy {
    id
    bindings {
      ...PolicyBinding
    }
  }
  insertedAt
}
    ${UserFragmentDoc}
${PolicyBindingFragmentDoc}`;
export const InviteFragmentDoc = gql`
    fragment Invite on Invite {
  id
  secureId
  email
  insertedAt
}
    `;
export const OidcLoginFragmentDoc = gql`
    fragment OidcLogin on OidcLogin {
  ip
  country
  city
  latitude
  longitude
  user {
    ...User
  }
  owner {
    ...User
  }
  repository {
    ...Repo
  }
  insertedAt
}
    ${UserFragmentDoc}
${RepoFragmentDoc}`;
export const ArtifactFragmentDoc = gql`
    fragment Artifact on Artifact {
  id
  name
  blob
  type
  platform
  arch
  filesize
  sha
  readme
  insertedAt
  updatedAt
}
    `;
export const ChartInstallationFragmentDoc = gql`
    fragment ChartInstallation on ChartInstallation {
  id
  chart {
    ...Chart
    dependencies {
      ...Dependencies
    }
  }
  version {
    ...Version
  }
}
    ${ChartFragmentDoc}
${DependenciesFragmentDoc}
${VersionFragmentDoc}`;
export const ScanViolationFragmentDoc = gql`
    fragment ScanViolation on ScanViolation {
  ruleName
  description
  ruleId
  severity
  category
  resourceName
  resourceType
  file
  line
}
    `;
export const ScanErrorFragmentDoc = gql`
    fragment ScanError on ScanError {
  message
}
    `;
export const PackageScanFragmentDoc = gql`
    fragment PackageScan on PackageScan {
  id
  grade
  violations {
    ...ScanViolation
  }
  errors {
    ...ScanError
  }
}
    ${ScanViolationFragmentDoc}
${ScanErrorFragmentDoc}`;
export const DnsRecordFragmentDoc = gql`
    fragment DnsRecord on DnsRecord {
  id
  name
  type
  records
  cluster
  provider
  creator {
    ...User
  }
  insertedAt
}
    ${UserFragmentDoc}`;
export const DockerRepoFragmentDoc = gql`
    fragment DockerRepo on DockerRepository {
  id
  name
  public
  repository {
    id
    name
  }
  insertedAt
  updatedAt
}
    `;
export const DockerRepositoryFragmentDoc = gql`
    fragment DockerRepository on DockerRepository {
  id
  name
  public
  repository {
    id
    name
    editable
  }
  insertedAt
  updatedAt
}
    `;
export const DockerImageFragmentDoc = gql`
    fragment DockerImage on DockerImage {
  id
  tag
  digest
  scannedAt
  grade
  insertedAt
  updatedAt
}
    `;
export const VulnerabilityFragmentDoc = gql`
    fragment Vulnerability on Vulnerability {
  id
  title
  description
  vulnerabilityId
  package
  installedVersion
  fixedVersion
  source
  url
  severity
  score
  cvss {
    attackVector
    attackComplexity
    privilegesRequired
    userInteraction
    confidentiality
    integrity
    availability
  }
  layer {
    digest
    diffId
  }
}
    `;
export const PostmortemFragmentDoc = gql`
    fragment Postmortem on Postmortem {
  id
  content
  actionItems {
    type
    link
  }
}
    `;
export const FollowerFragmentDoc = gql`
    fragment Follower on Follower {
  id
  incident {
    id
  }
  user {
    ...User
  }
  preferences {
    message
    incidentUpdate
    mention
  }
}
    ${UserFragmentDoc}`;
export const ServiceLevelFragmentDoc = gql`
    fragment ServiceLevel on ServiceLevel {
  minSeverity
  maxSeverity
  responseTime
}
    `;
export const LimitFragmentDoc = gql`
    fragment Limit on Limit {
  dimension
  quantity
}
    `;
export const LineItemFragmentDoc = gql`
    fragment LineItem on LineItem {
  name
  dimension
  cost
  period
  type
}
    `;
export const PlanFragmentDoc = gql`
    fragment Plan on Plan {
  id
  name
  cost
  period
  serviceLevels {
    ...ServiceLevel
  }
  lineItems {
    included {
      ...Limit
    }
    items {
      ...LineItem
    }
  }
  metadata {
    features {
      name
      description
    }
  }
}
    ${ServiceLevelFragmentDoc}
${LimitFragmentDoc}
${LineItemFragmentDoc}`;
export const SlimSubscriptionFragmentDoc = gql`
    fragment SlimSubscription on SlimSubscription {
  id
  lineItems {
    items {
      dimension
      quantity
    }
  }
  plan {
    ...Plan
  }
}
    ${PlanFragmentDoc}`;
export const ClusterInformationFragmentDoc = gql`
    fragment ClusterInformation on ClusterInformation {
  version
  gitCommit
  platform
}
    `;
export const IncidentFragmentDoc = gql`
    fragment Incident on Incident {
  id
  title
  description
  severity
  status
  notificationCount
  nextResponseAt
  creator {
    ...User
  }
  owner {
    ...User
  }
  repository {
    ...Repo
  }
  subscription {
    ...SlimSubscription
  }
  clusterInformation {
    ...ClusterInformation
  }
  tags {
    tag
  }
  insertedAt
}
    ${UserFragmentDoc}
${RepoFragmentDoc}
${SlimSubscriptionFragmentDoc}
${ClusterInformationFragmentDoc}`;
export const IncidentHistoryFragmentDoc = gql`
    fragment IncidentHistory on IncidentHistory {
  id
  action
  changes {
    key
    prev
    next
  }
  actor {
    ...User
  }
  insertedAt
}
    ${UserFragmentDoc}`;
export const FileFragmentDoc = gql`
    fragment File on File {
  id
  blob
  mediaType
  contentType
  filesize
  filename
}
    `;
export const IncidentMessageFragmentDoc = gql`
    fragment IncidentMessage on IncidentMessage {
  id
  text
  creator {
    ...User
  }
  reactions {
    name
    creator {
      id
      email
    }
  }
  file {
    ...File
  }
  entities {
    type
    user {
      ...User
    }
    text
    startIndex
    endIndex
  }
  insertedAt
}
    ${UserFragmentDoc}
${FileFragmentDoc}`;
export const NotificationFragmentDoc = gql`
    fragment Notification on Notification {
  id
  type
  msg
  actor {
    ...User
  }
  incident {
    id
    title
    repository {
      id
      name
      icon
      darkIcon
    }
  }
  message {
    text
  }
  repository {
    id
    name
    icon
    darkIcon
  }
  insertedAt
}
    ${UserFragmentDoc}`;
export const WebhookLogFragmentDoc = gql`
    fragment WebhookLog on WebhookLog {
  id
  state
  status
  payload
  response
  insertedAt
}
    `;
export const OauthIntegrationFragmentDoc = gql`
    fragment OauthIntegration on OauthIntegration {
  id
  service
  insertedAt
}
    `;
export const ZoomMeetingFragmentDoc = gql`
    fragment ZoomMeeting on ZoomMeeting {
  joinUrl
  password
}
    `;
export const MetricFragmentDoc = gql`
    fragment Metric on Metric {
  name
  tags {
    name
    value
  }
  values {
    time
    value
  }
}
    `;
export const PageInfoFragmentDoc = gql`
    fragment PageInfo on PageInfo {
  endCursor
  hasNextPage
}
    `;
export const OAuthInfoFragmentDoc = gql`
    fragment OAuthInfo on OauthInfo {
  provider
  authorizeUrl
}
    `;
export const SubscriptionFragmentDoc = gql`
    fragment Subscription on RepositorySubscription {
  id
  plan {
    ...Plan
  }
  lineItems {
    items {
      ...Limit
    }
  }
}
    ${PlanFragmentDoc}
${LimitFragmentDoc}`;
export const InvoiceItemFragmentDoc = gql`
    fragment InvoiceItem on InvoiceItem {
  amount
  currency
  description
}
    `;
export const InvoiceFragmentDoc = gql`
    fragment Invoice on Invoice {
  number
  amountDue
  amountPaid
  currency
  status
  createdAt
  hostedInvoiceUrl
  lines {
    ...InvoiceItem
  }
}
    ${InvoiceItemFragmentDoc}`;
export const CardFragmentDoc = gql`
    fragment Card on Card {
  id
  last4
  expMonth
  expYear
  name
  brand
}
    `;
export const OidcProviderFragmentDoc = gql`
    fragment OIDCProvider on OidcProvider {
  id
  clientId
  authMethod
  clientSecret
  redirectUris
  bindings {
    id
    user {
      ...User
    }
    group {
      ...Group
    }
  }
  configuration {
    issuer
    authorizationEndpoint
    tokenEndpoint
    jwksUri
    userinfoEndpoint
  }
}
    ${UserFragmentDoc}
${GroupFragmentDoc}`;
export const InstallationFragmentDoc = gql`
    fragment Installation on Installation {
  id
  context
  license
  licenseKey
  acmeKeyId
  acmeSecret
  autoUpgrade
  trackTag
  repository {
    ...Repo
  }
  user {
    ...User
  }
  oidcProvider {
    ...OIDCProvider
  }
}
    ${RepoFragmentDoc}
${UserFragmentDoc}
${OidcProviderFragmentDoc}`;
export const TerraformFragmentDoc = gql`
    fragment Terraform on Terraform {
  id
  name
  readme
  package
  description
  latestVersion
  dependencies {
    ...Dependencies
  }
  valuesTemplate
  insertedAt
}
    ${DependenciesFragmentDoc}`;
export const RecipeConfigurationFragmentDoc = gql`
    fragment RecipeConfiguration on RecipeConfiguration {
  name
  type
  default
  documentation
  optional
  placeholder
  functionName
  condition {
    field
    operation
    value
  }
  validation {
    type
    regex
    message
  }
}
    `;
export const RecipeItemFragmentDoc = gql`
    fragment RecipeItem on RecipeItem {
  id
  chart {
    ...Chart
  }
  terraform {
    ...Terraform
  }
  configuration {
    ...RecipeConfiguration
  }
}
    ${ChartFragmentDoc}
${TerraformFragmentDoc}
${RecipeConfigurationFragmentDoc}`;
export const RecipeSectionFragmentDoc = gql`
    fragment RecipeSection on RecipeSection {
  index
  repository {
    ...Repo
    installation {
      ...Installation
    }
  }
  recipeItems {
    ...RecipeItem
  }
  configuration {
    ...RecipeConfiguration
  }
}
    ${RepoFragmentDoc}
${InstallationFragmentDoc}
${RecipeItemFragmentDoc}
${RecipeConfigurationFragmentDoc}`;
export const RecipeFragmentDoc = gql`
    fragment Recipe on Recipe {
  id
  name
  description
  restricted
  provider
  tests {
    type
    name
    message
    args {
      name
      repo
      key
    }
  }
  repository {
    id
    name
  }
  oidcSettings {
    uriFormat
    uriFormats
    authMethod
    domainKey
    subdomain
  }
  recipeSections {
    ...RecipeSection
  }
}
    ${RecipeSectionFragmentDoc}`;
export const StackFragmentDoc = gql`
    fragment Stack on Stack {
  id
  name
  displayName
  description
  featured
  creator {
    id
    name
  }
  collections {
    id
    provider
    bundles {
      recipe {
        repository {
          ...Repo
          tags {
            tag
          }
        }
      }
    }
  }
}
    ${RepoFragmentDoc}`;
export const ApplyLockFragmentDoc = gql`
    fragment ApplyLock on ApplyLock {
  id
  lock
}
    `;
export const CategoryFragmentDoc = gql`
    fragment Category on CategoryInfo {
  category
  count
}
    `;
export const RepositoryFragmentDoc = gql`
    fragment Repository on Repository {
  id
  name
  notes
  icon
  darkIcon
  description
  publisher {
    name
  }
  recipes {
    name
  }
}
    `;
export const IntegrationFragmentDoc = gql`
    fragment Integration on Integration {
  id
  name
  icon
  sourceUrl
  description
  tags {
    tag
  }
  publisher {
    ...Publisher
  }
}
    ${PublisherFragmentDoc}`;
export const CloudShellFragmentDoc = gql`
    fragment CloudShell on CloudShell {
  id
  aesKey
  gitUrl
  alive
  provider
  subdomain
  cluster
  status {
    ready
    initialized
    containersReady
    podScheduled
  }
}
    `;
export const DemoProjectFragmentDoc = gql`
    fragment DemoProject on DemoProject {
  id
  projectId
  credentials
  ready
  state
}
    `;
export const TerraformInstallationFragmentDoc = gql`
    fragment TerraformInstallation on TerraformInstallation {
  id
  terraform {
    ...Terraform
  }
  version {
    ...Version
  }
}
    ${TerraformFragmentDoc}
${VersionFragmentDoc}`;
export const StepFragmentDoc = gql`
    fragment Step on TestStep {
  id
  name
  status
  hasLogs
  description
  insertedAt
  updatedAt
}
    `;
export const TestFragmentDoc = gql`
    fragment Test on Test {
  id
  name
  promoteTag
  status
  insertedAt
  updatedAt
  steps {
    ...Step
  }
}
    ${StepFragmentDoc}`;
export const UpgradeQueueFragmentDoc = gql`
    fragment UpgradeQueue on UpgradeQueue {
  id
  acked
  name
  domain
  git
  pingedAt
  provider
}
    `;
export const RolloutFragmentDoc = gql`
    fragment Rollout on Rollout {
  id
  event
  cursor
  count
  status
  heartbeat
}
    `;
export const UpgradeFragmentDoc = gql`
    fragment Upgrade on Upgrade {
  id
  message
  repository {
    ...Repo
  }
  insertedAt
}
    ${RepoFragmentDoc}`;
export const DeferredUpdateFragmentDoc = gql`
    fragment DeferredUpdate on DeferredUpdate {
  id
  dequeueAt
  attempts
  version {
    version
  }
  insertedAt
}
    `;
export const AccountFragmentDoc = gql`
    fragment Account on Account {
  id
  name
  billingCustomerId
  backgroundColor
}
    `;
export const ImpersonationPolicyFragmentDoc = gql`
    fragment ImpersonationPolicy on ImpersonationPolicy {
  id
  bindings {
    id
    group {
      id
      name
    }
    user {
      id
      name
      email
    }
  }
}
    `;
export const GroupMemberFragmentDoc = gql`
    fragment GroupMember on GroupMember {
  id
  user {
    ...User
  }
}
    ${UserFragmentDoc}`;
export const TokenFragmentDoc = gql`
    fragment Token on PersistedToken {
  id
  token
  insertedAt
}
    `;
export const TokenAuditFragmentDoc = gql`
    fragment TokenAudit on PersistedTokenAudit {
  ip
  timestamp
  count
  country
  city
  latitude
  longitude
}
    `;
export const WebhookFragmentDoc = gql`
    fragment Webhook on Webhook {
  id
  url
  secret
  insertedAt
}
    `;
export const PublicKeyFragmentDoc = gql`
    fragment PublicKey on PublicKey {
  id
  name
  digest
  insertedAt
  content
  user {
    ...User
  }
}
    ${UserFragmentDoc}`;
export const EabCredentialFragmentDoc = gql`
    fragment EabCredential on EabCredential {
  id
  keyId
  hmacKey
  cluster
  provider
  insertedAt
}
    `;
export const VersionTagFragmentDoc = gql`
    fragment VersionTag on VersionTag {
  id
  tag
  version {
    id
  }
}
    `;
export const UpdateAccountDocument = gql`
    mutation UpdateAccount($attributes: AccountAttributes!) {
  updateAccount(attributes: $attributes) {
    ...Account
    domainMappings {
      id
      domain
      enableSso
    }
  }
}
    ${AccountFragmentDoc}`;
export type UpdateAccountMutationFn = Apollo.MutationFunction<UpdateAccountMutation, UpdateAccountMutationVariables>;

/**
 * __useUpdateAccountMutation__
 *
 * To run a mutation, you first call `useUpdateAccountMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAccountMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAccountMutation, { data, loading, error }] = useUpdateAccountMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useUpdateAccountMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAccountMutation, UpdateAccountMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAccountMutation, UpdateAccountMutationVariables>(UpdateAccountDocument, options);
      }
export type UpdateAccountMutationHookResult = ReturnType<typeof useUpdateAccountMutation>;
export type UpdateAccountMutationResult = Apollo.MutationResult<UpdateAccountMutation>;
export type UpdateAccountMutationOptions = Apollo.BaseMutationOptions<UpdateAccountMutation, UpdateAccountMutationVariables>;
export const ListArtifactsDocument = gql`
    query ListArtifacts($id: ID!) {
  repository(id: $id) {
    artifacts {
      ...Artifact
    }
  }
}
    ${ArtifactFragmentDoc}`;

/**
 * __useListArtifactsQuery__
 *
 * To run a query within a React component, call `useListArtifactsQuery` and pass it any options that fit your needs.
 * When your component renders, `useListArtifactsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListArtifactsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useListArtifactsQuery(baseOptions: Apollo.QueryHookOptions<ListArtifactsQuery, ListArtifactsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListArtifactsQuery, ListArtifactsQueryVariables>(ListArtifactsDocument, options);
      }
export function useListArtifactsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListArtifactsQuery, ListArtifactsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListArtifactsQuery, ListArtifactsQueryVariables>(ListArtifactsDocument, options);
        }
export type ListArtifactsQueryHookResult = ReturnType<typeof useListArtifactsQuery>;
export type ListArtifactsLazyQueryHookResult = ReturnType<typeof useListArtifactsLazyQuery>;
export type ListArtifactsQueryResult = Apollo.QueryResult<ListArtifactsQuery, ListArtifactsQueryVariables>;
export const CreateArtifactDocument = gql`
    mutation CreateArtifact($repoName: String!, $name: String!, $readme: String!, $artifactType: String!, $platform: String!, $blob: UploadOrUrl!, $arch: String) {
  createArtifact(repositoryName: $repoName, attributes: {name: $name, blob: $blob, readme: $readme, type: $artifactType, platform: $platform, arch: $arch}) {
    ...Artifact
  }
}
    ${ArtifactFragmentDoc}`;
export type CreateArtifactMutationFn = Apollo.MutationFunction<CreateArtifactMutation, CreateArtifactMutationVariables>;

/**
 * __useCreateArtifactMutation__
 *
 * To run a mutation, you first call `useCreateArtifactMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateArtifactMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createArtifactMutation, { data, loading, error }] = useCreateArtifactMutation({
 *   variables: {
 *      repoName: // value for 'repoName'
 *      name: // value for 'name'
 *      readme: // value for 'readme'
 *      artifactType: // value for 'artifactType'
 *      platform: // value for 'platform'
 *      blob: // value for 'blob'
 *      arch: // value for 'arch'
 *   },
 * });
 */
export function useCreateArtifactMutation(baseOptions?: Apollo.MutationHookOptions<CreateArtifactMutation, CreateArtifactMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateArtifactMutation, CreateArtifactMutationVariables>(CreateArtifactDocument, options);
      }
export type CreateArtifactMutationHookResult = ReturnType<typeof useCreateArtifactMutation>;
export type CreateArtifactMutationResult = Apollo.MutationResult<CreateArtifactMutation>;
export type CreateArtifactMutationOptions = Apollo.BaseMutationOptions<CreateArtifactMutation, CreateArtifactMutationVariables>;
export const GetChartsDocument = gql`
    query GetCharts($id: ID!) {
  charts(repositoryId: $id, first: 100) {
    edges {
      node {
        ...Chart
      }
    }
  }
}
    ${ChartFragmentDoc}`;

/**
 * __useGetChartsQuery__
 *
 * To run a query within a React component, call `useGetChartsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetChartsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetChartsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetChartsQuery(baseOptions: Apollo.QueryHookOptions<GetChartsQuery, GetChartsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetChartsQuery, GetChartsQueryVariables>(GetChartsDocument, options);
      }
export function useGetChartsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetChartsQuery, GetChartsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetChartsQuery, GetChartsQueryVariables>(GetChartsDocument, options);
        }
export type GetChartsQueryHookResult = ReturnType<typeof useGetChartsQuery>;
export type GetChartsLazyQueryHookResult = ReturnType<typeof useGetChartsLazyQuery>;
export type GetChartsQueryResult = Apollo.QueryResult<GetChartsQuery, GetChartsQueryVariables>;
export const GetVersionsDocument = gql`
    query GetVersions($id: ID!) {
  versions(chartId: $id, first: 100) {
    edges {
      node {
        ...Version
      }
    }
  }
}
    ${VersionFragmentDoc}`;

/**
 * __useGetVersionsQuery__
 *
 * To run a query within a React component, call `useGetVersionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetVersionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetVersionsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetVersionsQuery(baseOptions: Apollo.QueryHookOptions<GetVersionsQuery, GetVersionsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetVersionsQuery, GetVersionsQueryVariables>(GetVersionsDocument, options);
      }
export function useGetVersionsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetVersionsQuery, GetVersionsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetVersionsQuery, GetVersionsQueryVariables>(GetVersionsDocument, options);
        }
export type GetVersionsQueryHookResult = ReturnType<typeof useGetVersionsQuery>;
export type GetVersionsLazyQueryHookResult = ReturnType<typeof useGetVersionsLazyQuery>;
export type GetVersionsQueryResult = Apollo.QueryResult<GetVersionsQuery, GetVersionsQueryVariables>;
export const GetChartInstallationsDocument = gql`
    query GetChartInstallations($id: ID!) {
  chartInstallations(repositoryId: $id, first: 100) {
    edges {
      node {
        ...ChartInstallation
      }
    }
  }
}
    ${ChartInstallationFragmentDoc}`;

/**
 * __useGetChartInstallationsQuery__
 *
 * To run a query within a React component, call `useGetChartInstallationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetChartInstallationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetChartInstallationsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetChartInstallationsQuery(baseOptions: Apollo.QueryHookOptions<GetChartInstallationsQuery, GetChartInstallationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetChartInstallationsQuery, GetChartInstallationsQueryVariables>(GetChartInstallationsDocument, options);
      }
export function useGetChartInstallationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetChartInstallationsQuery, GetChartInstallationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetChartInstallationsQuery, GetChartInstallationsQueryVariables>(GetChartInstallationsDocument, options);
        }
export type GetChartInstallationsQueryHookResult = ReturnType<typeof useGetChartInstallationsQuery>;
export type GetChartInstallationsLazyQueryHookResult = ReturnType<typeof useGetChartInstallationsLazyQuery>;
export type GetChartInstallationsQueryResult = Apollo.QueryResult<GetChartInstallationsQuery, GetChartInstallationsQueryVariables>;
export const GetPackageInstallationsDocument = gql`
    query GetPackageInstallations($id: ID!) {
  chartInstallations(repositoryId: $id, first: 100) {
    edges {
      node {
        ...ChartInstallation
      }
    }
  }
  terraformInstallations(repositoryId: $id, first: 100) {
    edges {
      node {
        ...TerraformInstallation
      }
    }
  }
}
    ${ChartInstallationFragmentDoc}
${TerraformInstallationFragmentDoc}`;

/**
 * __useGetPackageInstallationsQuery__
 *
 * To run a query within a React component, call `useGetPackageInstallationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetPackageInstallationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetPackageInstallationsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetPackageInstallationsQuery(baseOptions: Apollo.QueryHookOptions<GetPackageInstallationsQuery, GetPackageInstallationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetPackageInstallationsQuery, GetPackageInstallationsQueryVariables>(GetPackageInstallationsDocument, options);
      }
export function useGetPackageInstallationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetPackageInstallationsQuery, GetPackageInstallationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetPackageInstallationsQuery, GetPackageInstallationsQueryVariables>(GetPackageInstallationsDocument, options);
        }
export type GetPackageInstallationsQueryHookResult = ReturnType<typeof useGetPackageInstallationsQuery>;
export type GetPackageInstallationsLazyQueryHookResult = ReturnType<typeof useGetPackageInstallationsLazyQuery>;
export type GetPackageInstallationsQueryResult = Apollo.QueryResult<GetPackageInstallationsQuery, GetPackageInstallationsQueryVariables>;
export const CreateCrdDocument = gql`
    mutation CreateCrd($chartName: ChartName!, $name: String!, $blob: UploadOrUrl!) {
  createCrd(chartName: $chartName, attributes: {name: $name, blob: $blob}) {
    id
  }
}
    `;
export type CreateCrdMutationFn = Apollo.MutationFunction<CreateCrdMutation, CreateCrdMutationVariables>;

/**
 * __useCreateCrdMutation__
 *
 * To run a mutation, you first call `useCreateCrdMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateCrdMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createCrdMutation, { data, loading, error }] = useCreateCrdMutation({
 *   variables: {
 *      chartName: // value for 'chartName'
 *      name: // value for 'name'
 *      blob: // value for 'blob'
 *   },
 * });
 */
export function useCreateCrdMutation(baseOptions?: Apollo.MutationHookOptions<CreateCrdMutation, CreateCrdMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateCrdMutation, CreateCrdMutationVariables>(CreateCrdDocument, options);
      }
export type CreateCrdMutationHookResult = ReturnType<typeof useCreateCrdMutation>;
export type CreateCrdMutationResult = Apollo.MutationResult<CreateCrdMutation>;
export type CreateCrdMutationOptions = Apollo.BaseMutationOptions<CreateCrdMutation, CreateCrdMutationVariables>;
export const UninstallChartDocument = gql`
    mutation UninstallChart($id: ID!) {
  deleteChartInstallation(id: $id) {
    id
  }
}
    `;
export type UninstallChartMutationFn = Apollo.MutationFunction<UninstallChartMutation, UninstallChartMutationVariables>;

/**
 * __useUninstallChartMutation__
 *
 * To run a mutation, you first call `useUninstallChartMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUninstallChartMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [uninstallChartMutation, { data, loading, error }] = useUninstallChartMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useUninstallChartMutation(baseOptions?: Apollo.MutationHookOptions<UninstallChartMutation, UninstallChartMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UninstallChartMutation, UninstallChartMutationVariables>(UninstallChartDocument, options);
      }
export type UninstallChartMutationHookResult = ReturnType<typeof useUninstallChartMutation>;
export type UninstallChartMutationResult = Apollo.MutationResult<UninstallChartMutation>;
export type UninstallChartMutationOptions = Apollo.BaseMutationOptions<UninstallChartMutation, UninstallChartMutationVariables>;
export const GetDnsRecordsDocument = gql`
    query GetDnsRecords($cluster: String!, $provider: Provider!) {
  dnsRecords(cluster: $cluster, provider: $provider, first: 500) {
    edges {
      node {
        ...DnsRecord
      }
    }
  }
}
    ${DnsRecordFragmentDoc}`;

/**
 * __useGetDnsRecordsQuery__
 *
 * To run a query within a React component, call `useGetDnsRecordsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDnsRecordsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDnsRecordsQuery({
 *   variables: {
 *      cluster: // value for 'cluster'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useGetDnsRecordsQuery(baseOptions: Apollo.QueryHookOptions<GetDnsRecordsQuery, GetDnsRecordsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetDnsRecordsQuery, GetDnsRecordsQueryVariables>(GetDnsRecordsDocument, options);
      }
export function useGetDnsRecordsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetDnsRecordsQuery, GetDnsRecordsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetDnsRecordsQuery, GetDnsRecordsQueryVariables>(GetDnsRecordsDocument, options);
        }
export type GetDnsRecordsQueryHookResult = ReturnType<typeof useGetDnsRecordsQuery>;
export type GetDnsRecordsLazyQueryHookResult = ReturnType<typeof useGetDnsRecordsLazyQuery>;
export type GetDnsRecordsQueryResult = Apollo.QueryResult<GetDnsRecordsQuery, GetDnsRecordsQueryVariables>;
export const CreateDnsRecordDocument = gql`
    mutation CreateDnsRecord($cluster: String!, $provider: Provider!, $attributes: DnsRecordAttributes!) {
  createDnsRecord(cluster: $cluster, provider: $provider, attributes: $attributes) {
    ...DnsRecord
  }
}
    ${DnsRecordFragmentDoc}`;
export type CreateDnsRecordMutationFn = Apollo.MutationFunction<CreateDnsRecordMutation, CreateDnsRecordMutationVariables>;

/**
 * __useCreateDnsRecordMutation__
 *
 * To run a mutation, you first call `useCreateDnsRecordMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateDnsRecordMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createDnsRecordMutation, { data, loading, error }] = useCreateDnsRecordMutation({
 *   variables: {
 *      cluster: // value for 'cluster'
 *      provider: // value for 'provider'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateDnsRecordMutation(baseOptions?: Apollo.MutationHookOptions<CreateDnsRecordMutation, CreateDnsRecordMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateDnsRecordMutation, CreateDnsRecordMutationVariables>(CreateDnsRecordDocument, options);
      }
export type CreateDnsRecordMutationHookResult = ReturnType<typeof useCreateDnsRecordMutation>;
export type CreateDnsRecordMutationResult = Apollo.MutationResult<CreateDnsRecordMutation>;
export type CreateDnsRecordMutationOptions = Apollo.BaseMutationOptions<CreateDnsRecordMutation, CreateDnsRecordMutationVariables>;
export const DeleteDnsRecordDocument = gql`
    mutation DeleteDnsRecord($name: String!, $type: DnsRecordType!) {
  deleteDnsRecord(name: $name, type: $type) {
    ...DnsRecord
  }
}
    ${DnsRecordFragmentDoc}`;
export type DeleteDnsRecordMutationFn = Apollo.MutationFunction<DeleteDnsRecordMutation, DeleteDnsRecordMutationVariables>;

/**
 * __useDeleteDnsRecordMutation__
 *
 * To run a mutation, you first call `useDeleteDnsRecordMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteDnsRecordMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteDnsRecordMutation, { data, loading, error }] = useDeleteDnsRecordMutation({
 *   variables: {
 *      name: // value for 'name'
 *      type: // value for 'type'
 *   },
 * });
 */
export function useDeleteDnsRecordMutation(baseOptions?: Apollo.MutationHookOptions<DeleteDnsRecordMutation, DeleteDnsRecordMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteDnsRecordMutation, DeleteDnsRecordMutationVariables>(DeleteDnsRecordDocument, options);
      }
export type DeleteDnsRecordMutationHookResult = ReturnType<typeof useDeleteDnsRecordMutation>;
export type DeleteDnsRecordMutationResult = Apollo.MutationResult<DeleteDnsRecordMutation>;
export type DeleteDnsRecordMutationOptions = Apollo.BaseMutationOptions<DeleteDnsRecordMutation, DeleteDnsRecordMutationVariables>;
export const CreateDomainDocument = gql`
    mutation CreateDomain($name: String!) {
  provisionDomain(name: $name) {
    ...DnsDomain
  }
}
    ${DnsDomainFragmentDoc}`;
export type CreateDomainMutationFn = Apollo.MutationFunction<CreateDomainMutation, CreateDomainMutationVariables>;

/**
 * __useCreateDomainMutation__
 *
 * To run a mutation, you first call `useCreateDomainMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateDomainMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createDomainMutation, { data, loading, error }] = useCreateDomainMutation({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useCreateDomainMutation(baseOptions?: Apollo.MutationHookOptions<CreateDomainMutation, CreateDomainMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateDomainMutation, CreateDomainMutationVariables>(CreateDomainDocument, options);
      }
export type CreateDomainMutationHookResult = ReturnType<typeof useCreateDomainMutation>;
export type CreateDomainMutationResult = Apollo.MutationResult<CreateDomainMutation>;
export type CreateDomainMutationOptions = Apollo.BaseMutationOptions<CreateDomainMutation, CreateDomainMutationVariables>;
export const GroupMembersDocument = gql`
    query GroupMembers($cursor: String, $id: ID!) {
  groupMembers(groupId: $id, after: $cursor, first: 20) {
    pageInfo {
      ...PageInfo
    }
    edges {
      node {
        ...GroupMember
      }
    }
  }
}
    ${PageInfoFragmentDoc}
${GroupMemberFragmentDoc}`;

/**
 * __useGroupMembersQuery__
 *
 * To run a query within a React component, call `useGroupMembersQuery` and pass it any options that fit your needs.
 * When your component renders, `useGroupMembersQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGroupMembersQuery({
 *   variables: {
 *      cursor: // value for 'cursor'
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGroupMembersQuery(baseOptions: Apollo.QueryHookOptions<GroupMembersQuery, GroupMembersQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GroupMembersQuery, GroupMembersQueryVariables>(GroupMembersDocument, options);
      }
export function useGroupMembersLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GroupMembersQuery, GroupMembersQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GroupMembersQuery, GroupMembersQueryVariables>(GroupMembersDocument, options);
        }
export type GroupMembersQueryHookResult = ReturnType<typeof useGroupMembersQuery>;
export type GroupMembersLazyQueryHookResult = ReturnType<typeof useGroupMembersLazyQuery>;
export type GroupMembersQueryResult = Apollo.QueryResult<GroupMembersQuery, GroupMembersQueryVariables>;
export const CreateGroupMemberDocument = gql`
    mutation CreateGroupMember($groupId: ID!, $userId: ID!) {
  createGroupMember(groupId: $groupId, userId: $userId) {
    ...GroupMember
  }
}
    ${GroupMemberFragmentDoc}`;
export type CreateGroupMemberMutationFn = Apollo.MutationFunction<CreateGroupMemberMutation, CreateGroupMemberMutationVariables>;

/**
 * __useCreateGroupMemberMutation__
 *
 * To run a mutation, you first call `useCreateGroupMemberMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateGroupMemberMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createGroupMemberMutation, { data, loading, error }] = useCreateGroupMemberMutation({
 *   variables: {
 *      groupId: // value for 'groupId'
 *      userId: // value for 'userId'
 *   },
 * });
 */
export function useCreateGroupMemberMutation(baseOptions?: Apollo.MutationHookOptions<CreateGroupMemberMutation, CreateGroupMemberMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateGroupMemberMutation, CreateGroupMemberMutationVariables>(CreateGroupMemberDocument, options);
      }
export type CreateGroupMemberMutationHookResult = ReturnType<typeof useCreateGroupMemberMutation>;
export type CreateGroupMemberMutationResult = Apollo.MutationResult<CreateGroupMemberMutation>;
export type CreateGroupMemberMutationOptions = Apollo.BaseMutationOptions<CreateGroupMemberMutation, CreateGroupMemberMutationVariables>;
export const DeleteGroupMemberDocument = gql`
    mutation DeleteGroupMember($groupId: ID!, $userId: ID!) {
  deleteGroupMember(groupId: $groupId, userId: $userId) {
    ...GroupMember
  }
}
    ${GroupMemberFragmentDoc}`;
export type DeleteGroupMemberMutationFn = Apollo.MutationFunction<DeleteGroupMemberMutation, DeleteGroupMemberMutationVariables>;

/**
 * __useDeleteGroupMemberMutation__
 *
 * To run a mutation, you first call `useDeleteGroupMemberMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteGroupMemberMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteGroupMemberMutation, { data, loading, error }] = useDeleteGroupMemberMutation({
 *   variables: {
 *      groupId: // value for 'groupId'
 *      userId: // value for 'userId'
 *   },
 * });
 */
export function useDeleteGroupMemberMutation(baseOptions?: Apollo.MutationHookOptions<DeleteGroupMemberMutation, DeleteGroupMemberMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteGroupMemberMutation, DeleteGroupMemberMutationVariables>(DeleteGroupMemberDocument, options);
      }
export type DeleteGroupMemberMutationHookResult = ReturnType<typeof useDeleteGroupMemberMutation>;
export type DeleteGroupMemberMutationResult = Apollo.MutationResult<DeleteGroupMemberMutation>;
export type DeleteGroupMemberMutationOptions = Apollo.BaseMutationOptions<DeleteGroupMemberMutation, DeleteGroupMemberMutationVariables>;
export const CreateGroupDocument = gql`
    mutation CreateGroup($attributes: GroupAttributes!) {
  createGroup(attributes: $attributes) {
    ...Group
  }
}
    ${GroupFragmentDoc}`;
export type CreateGroupMutationFn = Apollo.MutationFunction<CreateGroupMutation, CreateGroupMutationVariables>;

/**
 * __useCreateGroupMutation__
 *
 * To run a mutation, you first call `useCreateGroupMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateGroupMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createGroupMutation, { data, loading, error }] = useCreateGroupMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateGroupMutation(baseOptions?: Apollo.MutationHookOptions<CreateGroupMutation, CreateGroupMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateGroupMutation, CreateGroupMutationVariables>(CreateGroupDocument, options);
      }
export type CreateGroupMutationHookResult = ReturnType<typeof useCreateGroupMutation>;
export type CreateGroupMutationResult = Apollo.MutationResult<CreateGroupMutation>;
export type CreateGroupMutationOptions = Apollo.BaseMutationOptions<CreateGroupMutation, CreateGroupMutationVariables>;
export const UpdateGroupDocument = gql`
    mutation UpdateGroup($id: ID!, $attributes: GroupAttributes!) {
  updateGroup(groupId: $id, attributes: $attributes) {
    ...Group
  }
}
    ${GroupFragmentDoc}`;
export type UpdateGroupMutationFn = Apollo.MutationFunction<UpdateGroupMutation, UpdateGroupMutationVariables>;

/**
 * __useUpdateGroupMutation__
 *
 * To run a mutation, you first call `useUpdateGroupMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateGroupMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateGroupMutation, { data, loading, error }] = useUpdateGroupMutation({
 *   variables: {
 *      id: // value for 'id'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useUpdateGroupMutation(baseOptions?: Apollo.MutationHookOptions<UpdateGroupMutation, UpdateGroupMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateGroupMutation, UpdateGroupMutationVariables>(UpdateGroupDocument, options);
      }
export type UpdateGroupMutationHookResult = ReturnType<typeof useUpdateGroupMutation>;
export type UpdateGroupMutationResult = Apollo.MutationResult<UpdateGroupMutation>;
export type UpdateGroupMutationOptions = Apollo.BaseMutationOptions<UpdateGroupMutation, UpdateGroupMutationVariables>;
export const DeleteGroupDocument = gql`
    mutation DeleteGroup($id: ID!) {
  deleteGroup(groupId: $id) {
    ...Group
  }
}
    ${GroupFragmentDoc}`;
export type DeleteGroupMutationFn = Apollo.MutationFunction<DeleteGroupMutation, DeleteGroupMutationVariables>;

/**
 * __useDeleteGroupMutation__
 *
 * To run a mutation, you first call `useDeleteGroupMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteGroupMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteGroupMutation, { data, loading, error }] = useDeleteGroupMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteGroupMutation(baseOptions?: Apollo.MutationHookOptions<DeleteGroupMutation, DeleteGroupMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteGroupMutation, DeleteGroupMutationVariables>(DeleteGroupDocument, options);
      }
export type DeleteGroupMutationHookResult = ReturnType<typeof useDeleteGroupMutation>;
export type DeleteGroupMutationResult = Apollo.MutationResult<DeleteGroupMutation>;
export type DeleteGroupMutationOptions = Apollo.BaseMutationOptions<DeleteGroupMutation, DeleteGroupMutationVariables>;
export const GroupsDocument = gql`
    query Groups($q: String, $cursor: String) {
  groups(q: $q, first: 20, after: $cursor) {
    pageInfo {
      ...PageInfo
    }
    edges {
      node {
        ...Group
      }
    }
  }
}
    ${PageInfoFragmentDoc}
${GroupFragmentDoc}`;

/**
 * __useGroupsQuery__
 *
 * To run a query within a React component, call `useGroupsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGroupsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGroupsQuery({
 *   variables: {
 *      q: // value for 'q'
 *      cursor: // value for 'cursor'
 *   },
 * });
 */
export function useGroupsQuery(baseOptions?: Apollo.QueryHookOptions<GroupsQuery, GroupsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GroupsQuery, GroupsQueryVariables>(GroupsDocument, options);
      }
export function useGroupsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GroupsQuery, GroupsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GroupsQuery, GroupsQueryVariables>(GroupsDocument, options);
        }
export type GroupsQueryHookResult = ReturnType<typeof useGroupsQuery>;
export type GroupsLazyQueryHookResult = ReturnType<typeof useGroupsLazyQuery>;
export type GroupsQueryResult = Apollo.QueryResult<GroupsQuery, GroupsQueryVariables>;
export const GetInstallationDocument = gql`
    query GetInstallation($name: String) {
  installation(name: $name) {
    ...Installation
  }
}
    ${InstallationFragmentDoc}`;

/**
 * __useGetInstallationQuery__
 *
 * To run a query within a React component, call `useGetInstallationQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetInstallationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetInstallationQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetInstallationQuery(baseOptions?: Apollo.QueryHookOptions<GetInstallationQuery, GetInstallationQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetInstallationQuery, GetInstallationQueryVariables>(GetInstallationDocument, options);
      }
export function useGetInstallationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetInstallationQuery, GetInstallationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetInstallationQuery, GetInstallationQueryVariables>(GetInstallationDocument, options);
        }
export type GetInstallationQueryHookResult = ReturnType<typeof useGetInstallationQuery>;
export type GetInstallationLazyQueryHookResult = ReturnType<typeof useGetInstallationLazyQuery>;
export type GetInstallationQueryResult = Apollo.QueryResult<GetInstallationQuery, GetInstallationQueryVariables>;
export const GetInstallationByIdDocument = gql`
    query GetInstallationById($id: ID) {
  installation(id: $id) {
    ...Installation
  }
}
    ${InstallationFragmentDoc}`;

/**
 * __useGetInstallationByIdQuery__
 *
 * To run a query within a React component, call `useGetInstallationByIdQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetInstallationByIdQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetInstallationByIdQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetInstallationByIdQuery(baseOptions?: Apollo.QueryHookOptions<GetInstallationByIdQuery, GetInstallationByIdQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetInstallationByIdQuery, GetInstallationByIdQueryVariables>(GetInstallationByIdDocument, options);
      }
export function useGetInstallationByIdLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetInstallationByIdQuery, GetInstallationByIdQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetInstallationByIdQuery, GetInstallationByIdQueryVariables>(GetInstallationByIdDocument, options);
        }
export type GetInstallationByIdQueryHookResult = ReturnType<typeof useGetInstallationByIdQuery>;
export type GetInstallationByIdLazyQueryHookResult = ReturnType<typeof useGetInstallationByIdLazyQuery>;
export type GetInstallationByIdQueryResult = Apollo.QueryResult<GetInstallationByIdQuery, GetInstallationByIdQueryVariables>;
export const GetInstallationsDocument = gql`
    query GetInstallations($first: Int) {
  installations(first: $first) {
    edges {
      node {
        ...Installation
      }
    }
  }
}
    ${InstallationFragmentDoc}`;

/**
 * __useGetInstallationsQuery__
 *
 * To run a query within a React component, call `useGetInstallationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetInstallationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetInstallationsQuery({
 *   variables: {
 *      first: // value for 'first'
 *   },
 * });
 */
export function useGetInstallationsQuery(baseOptions?: Apollo.QueryHookOptions<GetInstallationsQuery, GetInstallationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetInstallationsQuery, GetInstallationsQueryVariables>(GetInstallationsDocument, options);
      }
export function useGetInstallationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetInstallationsQuery, GetInstallationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetInstallationsQuery, GetInstallationsQueryVariables>(GetInstallationsDocument, options);
        }
export type GetInstallationsQueryHookResult = ReturnType<typeof useGetInstallationsQuery>;
export type GetInstallationsLazyQueryHookResult = ReturnType<typeof useGetInstallationsLazyQuery>;
export type GetInstallationsQueryResult = Apollo.QueryResult<GetInstallationsQuery, GetInstallationsQueryVariables>;
export const UpsertOidcProviderDocument = gql`
    mutation UpsertOidcProvider($id: ID!, $attributes: OidcAttributes!) {
  upsertOidcProvider(installationId: $id, attributes: $attributes) {
    id
  }
}
    `;
export type UpsertOidcProviderMutationFn = Apollo.MutationFunction<UpsertOidcProviderMutation, UpsertOidcProviderMutationVariables>;

/**
 * __useUpsertOidcProviderMutation__
 *
 * To run a mutation, you first call `useUpsertOidcProviderMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpsertOidcProviderMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [upsertOidcProviderMutation, { data, loading, error }] = useUpsertOidcProviderMutation({
 *   variables: {
 *      id: // value for 'id'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useUpsertOidcProviderMutation(baseOptions?: Apollo.MutationHookOptions<UpsertOidcProviderMutation, UpsertOidcProviderMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpsertOidcProviderMutation, UpsertOidcProviderMutationVariables>(UpsertOidcProviderDocument, options);
      }
export type UpsertOidcProviderMutationHookResult = ReturnType<typeof useUpsertOidcProviderMutation>;
export type UpsertOidcProviderMutationResult = Apollo.MutationResult<UpsertOidcProviderMutation>;
export type UpsertOidcProviderMutationOptions = Apollo.BaseMutationOptions<UpsertOidcProviderMutation, UpsertOidcProviderMutationVariables>;
export const SignupInviteDocument = gql`
    mutation SignupInvite($attributes: UserAttributes!, $inviteId: String!) {
  signup(attributes: $attributes, inviteId: $inviteId) {
    jwt
  }
}
    `;
export type SignupInviteMutationFn = Apollo.MutationFunction<SignupInviteMutation, SignupInviteMutationVariables>;

/**
 * __useSignupInviteMutation__
 *
 * To run a mutation, you first call `useSignupInviteMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSignupInviteMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [signupInviteMutation, { data, loading, error }] = useSignupInviteMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *      inviteId: // value for 'inviteId'
 *   },
 * });
 */
export function useSignupInviteMutation(baseOptions?: Apollo.MutationHookOptions<SignupInviteMutation, SignupInviteMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SignupInviteMutation, SignupInviteMutationVariables>(SignupInviteDocument, options);
      }
export type SignupInviteMutationHookResult = ReturnType<typeof useSignupInviteMutation>;
export type SignupInviteMutationResult = Apollo.MutationResult<SignupInviteMutation>;
export type SignupInviteMutationOptions = Apollo.BaseMutationOptions<SignupInviteMutation, SignupInviteMutationVariables>;
export const RealizeInviteDocument = gql`
    mutation RealizeInvite($id: String!) {
  realizeInvite(id: $id) {
    jwt
  }
}
    `;
export type RealizeInviteMutationFn = Apollo.MutationFunction<RealizeInviteMutation, RealizeInviteMutationVariables>;

/**
 * __useRealizeInviteMutation__
 *
 * To run a mutation, you first call `useRealizeInviteMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRealizeInviteMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [realizeInviteMutation, { data, loading, error }] = useRealizeInviteMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useRealizeInviteMutation(baseOptions?: Apollo.MutationHookOptions<RealizeInviteMutation, RealizeInviteMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RealizeInviteMutation, RealizeInviteMutationVariables>(RealizeInviteDocument, options);
      }
export type RealizeInviteMutationHookResult = ReturnType<typeof useRealizeInviteMutation>;
export type RealizeInviteMutationResult = Apollo.MutationResult<RealizeInviteMutation>;
export type RealizeInviteMutationOptions = Apollo.BaseMutationOptions<RealizeInviteMutation, RealizeInviteMutationVariables>;
export const InviteDocument = gql`
    query Invite($id: String!) {
  invite(id: $id) {
    id
    email
    existing
    account {
      ...Account
    }
    user {
      ...User
      account {
        ...Account
      }
    }
  }
}
    ${AccountFragmentDoc}
${UserFragmentDoc}`;

/**
 * __useInviteQuery__
 *
 * To run a query within a React component, call `useInviteQuery` and pass it any options that fit your needs.
 * When your component renders, `useInviteQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useInviteQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useInviteQuery(baseOptions: Apollo.QueryHookOptions<InviteQuery, InviteQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<InviteQuery, InviteQueryVariables>(InviteDocument, options);
      }
export function useInviteLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<InviteQuery, InviteQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<InviteQuery, InviteQueryVariables>(InviteDocument, options);
        }
export type InviteQueryHookResult = ReturnType<typeof useInviteQuery>;
export type InviteLazyQueryHookResult = ReturnType<typeof useInviteLazyQuery>;
export type InviteQueryResult = Apollo.QueryResult<InviteQuery, InviteQueryVariables>;
export const SubscriptionDocument = gql`
    query Subscription {
  account {
    billingCustomerId
    grandfatheredUntil
    delinquentAt
    userCount
    clusterCount
    availableFeatures {
      userManagement
      audit
    }
    subscription {
      id
      plan {
        id
        period
        lineItems {
          dimension
          cost
        }
      }
    }
    billingAddress {
      name
      line1
      line2
      zip
      state
      city
      country
    }
  }
}
    `;

/**
 * __useSubscriptionQuery__
 *
 * To run a query within a React component, call `useSubscriptionQuery` and pass it any options that fit your needs.
 * When your component renders, `useSubscriptionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSubscriptionQuery({
 *   variables: {
 *   },
 * });
 */
export function useSubscriptionQuery(baseOptions?: Apollo.QueryHookOptions<SubscriptionQuery, SubscriptionQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SubscriptionQuery, SubscriptionQueryVariables>(SubscriptionDocument, options);
      }
export function useSubscriptionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SubscriptionQuery, SubscriptionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SubscriptionQuery, SubscriptionQueryVariables>(SubscriptionDocument, options);
        }
export type SubscriptionQueryHookResult = ReturnType<typeof useSubscriptionQuery>;
export type SubscriptionLazyQueryHookResult = ReturnType<typeof useSubscriptionLazyQuery>;
export type SubscriptionQueryResult = Apollo.QueryResult<SubscriptionQuery, SubscriptionQueryVariables>;
export const UpdateAccountBillingDocument = gql`
    mutation UpdateAccountBilling($attributes: AccountAttributes!) {
  updateAccount(attributes: $attributes) {
    id
  }
}
    `;
export type UpdateAccountBillingMutationFn = Apollo.MutationFunction<UpdateAccountBillingMutation, UpdateAccountBillingMutationVariables>;

/**
 * __useUpdateAccountBillingMutation__
 *
 * To run a mutation, you first call `useUpdateAccountBillingMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAccountBillingMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAccountBillingMutation, { data, loading, error }] = useUpdateAccountBillingMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useUpdateAccountBillingMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAccountBillingMutation, UpdateAccountBillingMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAccountBillingMutation, UpdateAccountBillingMutationVariables>(UpdateAccountBillingDocument, options);
      }
export type UpdateAccountBillingMutationHookResult = ReturnType<typeof useUpdateAccountBillingMutation>;
export type UpdateAccountBillingMutationResult = Apollo.MutationResult<UpdateAccountBillingMutation>;
export type UpdateAccountBillingMutationOptions = Apollo.BaseMutationOptions<UpdateAccountBillingMutation, UpdateAccountBillingMutationVariables>;
export const UpgradeToProfessionalPlanDocument = gql`
    mutation UpgradeToProfessionalPlan($planId: ID!) {
  createPlatformSubscription(planId: $planId) {
    id
  }
}
    `;
export type UpgradeToProfessionalPlanMutationFn = Apollo.MutationFunction<UpgradeToProfessionalPlanMutation, UpgradeToProfessionalPlanMutationVariables>;

/**
 * __useUpgradeToProfessionalPlanMutation__
 *
 * To run a mutation, you first call `useUpgradeToProfessionalPlanMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpgradeToProfessionalPlanMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [upgradeToProfessionalPlanMutation, { data, loading, error }] = useUpgradeToProfessionalPlanMutation({
 *   variables: {
 *      planId: // value for 'planId'
 *   },
 * });
 */
export function useUpgradeToProfessionalPlanMutation(baseOptions?: Apollo.MutationHookOptions<UpgradeToProfessionalPlanMutation, UpgradeToProfessionalPlanMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpgradeToProfessionalPlanMutation, UpgradeToProfessionalPlanMutationVariables>(UpgradeToProfessionalPlanDocument, options);
      }
export type UpgradeToProfessionalPlanMutationHookResult = ReturnType<typeof useUpgradeToProfessionalPlanMutation>;
export type UpgradeToProfessionalPlanMutationResult = Apollo.MutationResult<UpgradeToProfessionalPlanMutation>;
export type UpgradeToProfessionalPlanMutationOptions = Apollo.BaseMutationOptions<UpgradeToProfessionalPlanMutation, UpgradeToProfessionalPlanMutationVariables>;
export const DowngradeToFreePlanMutationDocument = gql`
    mutation DowngradeToFreePlanMutation {
  deletePlatformSubscription {
    id
  }
}
    `;
export type DowngradeToFreePlanMutationMutationFn = Apollo.MutationFunction<DowngradeToFreePlanMutationMutation, DowngradeToFreePlanMutationMutationVariables>;

/**
 * __useDowngradeToFreePlanMutationMutation__
 *
 * To run a mutation, you first call `useDowngradeToFreePlanMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDowngradeToFreePlanMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [downgradeToFreePlanMutationMutation, { data, loading, error }] = useDowngradeToFreePlanMutationMutation({
 *   variables: {
 *   },
 * });
 */
export function useDowngradeToFreePlanMutationMutation(baseOptions?: Apollo.MutationHookOptions<DowngradeToFreePlanMutationMutation, DowngradeToFreePlanMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DowngradeToFreePlanMutationMutation, DowngradeToFreePlanMutationMutationVariables>(DowngradeToFreePlanMutationDocument, options);
      }
export type DowngradeToFreePlanMutationMutationHookResult = ReturnType<typeof useDowngradeToFreePlanMutationMutation>;
export type DowngradeToFreePlanMutationMutationResult = Apollo.MutationResult<DowngradeToFreePlanMutationMutation>;
export type DowngradeToFreePlanMutationMutationOptions = Apollo.BaseMutationOptions<DowngradeToFreePlanMutationMutation, DowngradeToFreePlanMutationMutationVariables>;
export const CardsDocument = gql`
    query Cards {
  me {
    id
    cards(first: 100) {
      edges {
        node {
          ...Card
        }
      }
    }
  }
}
    ${CardFragmentDoc}`;

/**
 * __useCardsQuery__
 *
 * To run a query within a React component, call `useCardsQuery` and pass it any options that fit your needs.
 * When your component renders, `useCardsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCardsQuery({
 *   variables: {
 *   },
 * });
 */
export function useCardsQuery(baseOptions?: Apollo.QueryHookOptions<CardsQuery, CardsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CardsQuery, CardsQueryVariables>(CardsDocument, options);
      }
export function useCardsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CardsQuery, CardsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CardsQuery, CardsQueryVariables>(CardsDocument, options);
        }
export type CardsQueryHookResult = ReturnType<typeof useCardsQuery>;
export type CardsLazyQueryHookResult = ReturnType<typeof useCardsLazyQuery>;
export type CardsQueryResult = Apollo.QueryResult<CardsQuery, CardsQueryVariables>;
export const CreateCardDocument = gql`
    mutation CreateCard($source: String!, $address: AddressAttributes) {
  createCard(source: $source, address: $address) {
    id
  }
}
    `;
export type CreateCardMutationFn = Apollo.MutationFunction<CreateCardMutation, CreateCardMutationVariables>;

/**
 * __useCreateCardMutation__
 *
 * To run a mutation, you first call `useCreateCardMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateCardMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createCardMutation, { data, loading, error }] = useCreateCardMutation({
 *   variables: {
 *      source: // value for 'source'
 *      address: // value for 'address'
 *   },
 * });
 */
export function useCreateCardMutation(baseOptions?: Apollo.MutationHookOptions<CreateCardMutation, CreateCardMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateCardMutation, CreateCardMutationVariables>(CreateCardDocument, options);
      }
export type CreateCardMutationHookResult = ReturnType<typeof useCreateCardMutation>;
export type CreateCardMutationResult = Apollo.MutationResult<CreateCardMutation>;
export type CreateCardMutationOptions = Apollo.BaseMutationOptions<CreateCardMutation, CreateCardMutationVariables>;
export const DeleteCardDocument = gql`
    mutation DeleteCard($id: ID!) {
  deleteCard(id: $id) {
    id
  }
}
    `;
export type DeleteCardMutationFn = Apollo.MutationFunction<DeleteCardMutation, DeleteCardMutationVariables>;

/**
 * __useDeleteCardMutation__
 *
 * To run a mutation, you first call `useDeleteCardMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteCardMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteCardMutation, { data, loading, error }] = useDeleteCardMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteCardMutation(baseOptions?: Apollo.MutationHookOptions<DeleteCardMutation, DeleteCardMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteCardMutation, DeleteCardMutationVariables>(DeleteCardDocument, options);
      }
export type DeleteCardMutationHookResult = ReturnType<typeof useDeleteCardMutation>;
export type DeleteCardMutationResult = Apollo.MutationResult<DeleteCardMutation>;
export type DeleteCardMutationOptions = Apollo.BaseMutationOptions<DeleteCardMutation, DeleteCardMutationVariables>;
export const GetRecipeDocument = gql`
    query GetRecipe($repo: String, $name: String, $id: ID) {
  recipe(repo: $repo, name: $name, id: $id) {
    ...Recipe
    recipeDependencies {
      ...Recipe
    }
  }
}
    ${RecipeFragmentDoc}`;

/**
 * __useGetRecipeQuery__
 *
 * To run a query within a React component, call `useGetRecipeQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetRecipeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetRecipeQuery({
 *   variables: {
 *      repo: // value for 'repo'
 *      name: // value for 'name'
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetRecipeQuery(baseOptions?: Apollo.QueryHookOptions<GetRecipeQuery, GetRecipeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetRecipeQuery, GetRecipeQueryVariables>(GetRecipeDocument, options);
      }
export function useGetRecipeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetRecipeQuery, GetRecipeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetRecipeQuery, GetRecipeQueryVariables>(GetRecipeDocument, options);
        }
export type GetRecipeQueryHookResult = ReturnType<typeof useGetRecipeQuery>;
export type GetRecipeLazyQueryHookResult = ReturnType<typeof useGetRecipeLazyQuery>;
export type GetRecipeQueryResult = Apollo.QueryResult<GetRecipeQuery, GetRecipeQueryVariables>;
export const ListRecipesDocument = gql`
    query ListRecipes($repositoryName: String, $repositoryId: ID, $provider: Provider) {
  recipes(repositoryName: $repositoryName, provider: $provider, first: 500, repositoryId: $repositoryId) {
    edges {
      node {
        ...Recipe
      }
    }
  }
}
    ${RecipeFragmentDoc}`;

/**
 * __useListRecipesQuery__
 *
 * To run a query within a React component, call `useListRecipesQuery` and pass it any options that fit your needs.
 * When your component renders, `useListRecipesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListRecipesQuery({
 *   variables: {
 *      repositoryName: // value for 'repositoryName'
 *      repositoryId: // value for 'repositoryId'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useListRecipesQuery(baseOptions?: Apollo.QueryHookOptions<ListRecipesQuery, ListRecipesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListRecipesQuery, ListRecipesQueryVariables>(ListRecipesDocument, options);
      }
export function useListRecipesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListRecipesQuery, ListRecipesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListRecipesQuery, ListRecipesQueryVariables>(ListRecipesDocument, options);
        }
export type ListRecipesQueryHookResult = ReturnType<typeof useListRecipesQuery>;
export type ListRecipesLazyQueryHookResult = ReturnType<typeof useListRecipesLazyQuery>;
export type ListRecipesQueryResult = Apollo.QueryResult<ListRecipesQuery, ListRecipesQueryVariables>;
export const CreateRecipeDocument = gql`
    mutation CreateRecipe($name: String!, $attributes: RecipeAttributes!) {
  createRecipe(repositoryName: $name, attributes: $attributes) {
    id
  }
}
    `;
export type CreateRecipeMutationFn = Apollo.MutationFunction<CreateRecipeMutation, CreateRecipeMutationVariables>;

/**
 * __useCreateRecipeMutation__
 *
 * To run a mutation, you first call `useCreateRecipeMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateRecipeMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createRecipeMutation, { data, loading, error }] = useCreateRecipeMutation({
 *   variables: {
 *      name: // value for 'name'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateRecipeMutation(baseOptions?: Apollo.MutationHookOptions<CreateRecipeMutation, CreateRecipeMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateRecipeMutation, CreateRecipeMutationVariables>(CreateRecipeDocument, options);
      }
export type CreateRecipeMutationHookResult = ReturnType<typeof useCreateRecipeMutation>;
export type CreateRecipeMutationResult = Apollo.MutationResult<CreateRecipeMutation>;
export type CreateRecipeMutationOptions = Apollo.BaseMutationOptions<CreateRecipeMutation, CreateRecipeMutationVariables>;
export const InstallRecipeDocument = gql`
    mutation InstallRecipe($id: ID!) {
  installRecipe(recipeId: $id, context: "{}") {
    id
  }
}
    `;
export type InstallRecipeMutationFn = Apollo.MutationFunction<InstallRecipeMutation, InstallRecipeMutationVariables>;

/**
 * __useInstallRecipeMutation__
 *
 * To run a mutation, you first call `useInstallRecipeMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useInstallRecipeMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [installRecipeMutation, { data, loading, error }] = useInstallRecipeMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useInstallRecipeMutation(baseOptions?: Apollo.MutationHookOptions<InstallRecipeMutation, InstallRecipeMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<InstallRecipeMutation, InstallRecipeMutationVariables>(InstallRecipeDocument, options);
      }
export type InstallRecipeMutationHookResult = ReturnType<typeof useInstallRecipeMutation>;
export type InstallRecipeMutationResult = Apollo.MutationResult<InstallRecipeMutation>;
export type InstallRecipeMutationOptions = Apollo.BaseMutationOptions<InstallRecipeMutation, InstallRecipeMutationVariables>;
export const CreateStackDocument = gql`
    mutation CreateStack($attributes: StackAttributes!) {
  createStack(attributes: $attributes) {
    id
  }
}
    `;
export type CreateStackMutationFn = Apollo.MutationFunction<CreateStackMutation, CreateStackMutationVariables>;

/**
 * __useCreateStackMutation__
 *
 * To run a mutation, you first call `useCreateStackMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateStackMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createStackMutation, { data, loading, error }] = useCreateStackMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateStackMutation(baseOptions?: Apollo.MutationHookOptions<CreateStackMutation, CreateStackMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateStackMutation, CreateStackMutationVariables>(CreateStackDocument, options);
      }
export type CreateStackMutationHookResult = ReturnType<typeof useCreateStackMutation>;
export type CreateStackMutationResult = Apollo.MutationResult<CreateStackMutation>;
export type CreateStackMutationOptions = Apollo.BaseMutationOptions<CreateStackMutation, CreateStackMutationVariables>;
export const GetStackDocument = gql`
    query GetStack($name: String!, $provider: Provider!) {
  stack(name: $name, provider: $provider) {
    ...Stack
  }
}
    ${StackFragmentDoc}`;

/**
 * __useGetStackQuery__
 *
 * To run a query within a React component, call `useGetStackQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetStackQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetStackQuery({
 *   variables: {
 *      name: // value for 'name'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useGetStackQuery(baseOptions: Apollo.QueryHookOptions<GetStackQuery, GetStackQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetStackQuery, GetStackQueryVariables>(GetStackDocument, options);
      }
export function useGetStackLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetStackQuery, GetStackQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetStackQuery, GetStackQueryVariables>(GetStackDocument, options);
        }
export type GetStackQueryHookResult = ReturnType<typeof useGetStackQuery>;
export type GetStackLazyQueryHookResult = ReturnType<typeof useGetStackLazyQuery>;
export type GetStackQueryResult = Apollo.QueryResult<GetStackQuery, GetStackQueryVariables>;
export const ListStacksDocument = gql`
    query ListStacks($featured: Boolean, $cursor: String) {
  stacks(first: 100, after: $cursor, featured: $featured) {
    edges {
      node {
        ...Stack
      }
    }
  }
}
    ${StackFragmentDoc}`;

/**
 * __useListStacksQuery__
 *
 * To run a query within a React component, call `useListStacksQuery` and pass it any options that fit your needs.
 * When your component renders, `useListStacksQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListStacksQuery({
 *   variables: {
 *      featured: // value for 'featured'
 *      cursor: // value for 'cursor'
 *   },
 * });
 */
export function useListStacksQuery(baseOptions?: Apollo.QueryHookOptions<ListStacksQuery, ListStacksQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListStacksQuery, ListStacksQueryVariables>(ListStacksDocument, options);
      }
export function useListStacksLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListStacksQuery, ListStacksQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListStacksQuery, ListStacksQueryVariables>(ListStacksDocument, options);
        }
export type ListStacksQueryHookResult = ReturnType<typeof useListStacksQuery>;
export type ListStacksLazyQueryHookResult = ReturnType<typeof useListStacksLazyQuery>;
export type ListStacksQueryResult = Apollo.QueryResult<ListStacksQuery, ListStacksQueryVariables>;
export const CreateQuickStackDocument = gql`
    mutation CreateQuickStack($applicationIds: [ID], $provider: Provider!) {
  quickStack(repositoryIds: $applicationIds, provider: $provider) {
    id
    name
  }
}
    `;
export type CreateQuickStackMutationFn = Apollo.MutationFunction<CreateQuickStackMutation, CreateQuickStackMutationVariables>;

/**
 * __useCreateQuickStackMutation__
 *
 * To run a mutation, you first call `useCreateQuickStackMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateQuickStackMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createQuickStackMutation, { data, loading, error }] = useCreateQuickStackMutation({
 *   variables: {
 *      applicationIds: // value for 'applicationIds'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useCreateQuickStackMutation(baseOptions?: Apollo.MutationHookOptions<CreateQuickStackMutation, CreateQuickStackMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateQuickStackMutation, CreateQuickStackMutationVariables>(CreateQuickStackDocument, options);
      }
export type CreateQuickStackMutationHookResult = ReturnType<typeof useCreateQuickStackMutation>;
export type CreateQuickStackMutationResult = Apollo.MutationResult<CreateQuickStackMutation>;
export type CreateQuickStackMutationOptions = Apollo.BaseMutationOptions<CreateQuickStackMutation, CreateQuickStackMutationVariables>;
export const InstallStackShellDocument = gql`
    mutation InstallStackShell($name: String!, $context: ContextAttributes!, $oidc: Boolean!) {
  installStackShell(name: $name, context: $context, oidc: $oidc) {
    id
    name
  }
}
    `;
export type InstallStackShellMutationFn = Apollo.MutationFunction<InstallStackShellMutation, InstallStackShellMutationVariables>;

/**
 * __useInstallStackShellMutation__
 *
 * To run a mutation, you first call `useInstallStackShellMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useInstallStackShellMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [installStackShellMutation, { data, loading, error }] = useInstallStackShellMutation({
 *   variables: {
 *      name: // value for 'name'
 *      context: // value for 'context'
 *      oidc: // value for 'oidc'
 *   },
 * });
 */
export function useInstallStackShellMutation(baseOptions?: Apollo.MutationHookOptions<InstallStackShellMutation, InstallStackShellMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<InstallStackShellMutation, InstallStackShellMutationVariables>(InstallStackShellDocument, options);
      }
export type InstallStackShellMutationHookResult = ReturnType<typeof useInstallStackShellMutation>;
export type InstallStackShellMutationResult = Apollo.MutationResult<InstallStackShellMutation>;
export type InstallStackShellMutationOptions = Apollo.BaseMutationOptions<InstallStackShellMutation, InstallStackShellMutationVariables>;
export const RepositoryDocument = gql`
    query Repository($id: ID, $name: String) {
  repository(id: $id, name: $name) {
    ...Repo
    editable
    publicKey
    secrets
    artifacts {
      ...Artifact
    }
    installation {
      ...Installation
      oidcProvider {
        ...OIDCProvider
      }
    }
    tags {
      tag
    }
    readme
    mainBranch
    gitUrl
    homepage
    license {
      name
      url
    }
    documentation
    community {
      discord
      slack
      homepage
      gitUrl
      twitter
    }
  }
}
    ${RepoFragmentDoc}
${ArtifactFragmentDoc}
${InstallationFragmentDoc}
${OidcProviderFragmentDoc}`;

/**
 * __useRepositoryQuery__
 *
 * To run a query within a React component, call `useRepositoryQuery` and pass it any options that fit your needs.
 * When your component renders, `useRepositoryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useRepositoryQuery({
 *   variables: {
 *      id: // value for 'id'
 *      name: // value for 'name'
 *   },
 * });
 */
export function useRepositoryQuery(baseOptions?: Apollo.QueryHookOptions<RepositoryQuery, RepositoryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<RepositoryQuery, RepositoryQueryVariables>(RepositoryDocument, options);
      }
export function useRepositoryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<RepositoryQuery, RepositoryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<RepositoryQuery, RepositoryQueryVariables>(RepositoryDocument, options);
        }
export type RepositoryQueryHookResult = ReturnType<typeof useRepositoryQuery>;
export type RepositoryLazyQueryHookResult = ReturnType<typeof useRepositoryLazyQuery>;
export type RepositoryQueryResult = Apollo.QueryResult<RepositoryQuery, RepositoryQueryVariables>;
export const CreateResourceDefinitionDocument = gql`
    mutation CreateResourceDefinition($name: String!, $input: ResourceDefinitionAttributes!) {
  updateRepository(repositoryName: $name, attributes: {integrationResourceDefinition: $input}) {
    id
  }
}
    `;
export type CreateResourceDefinitionMutationFn = Apollo.MutationFunction<CreateResourceDefinitionMutation, CreateResourceDefinitionMutationVariables>;

/**
 * __useCreateResourceDefinitionMutation__
 *
 * To run a mutation, you first call `useCreateResourceDefinitionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateResourceDefinitionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createResourceDefinitionMutation, { data, loading, error }] = useCreateResourceDefinitionMutation({
 *   variables: {
 *      name: // value for 'name'
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateResourceDefinitionMutation(baseOptions?: Apollo.MutationHookOptions<CreateResourceDefinitionMutation, CreateResourceDefinitionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateResourceDefinitionMutation, CreateResourceDefinitionMutationVariables>(CreateResourceDefinitionDocument, options);
      }
export type CreateResourceDefinitionMutationHookResult = ReturnType<typeof useCreateResourceDefinitionMutation>;
export type CreateResourceDefinitionMutationResult = Apollo.MutationResult<CreateResourceDefinitionMutation>;
export type CreateResourceDefinitionMutationOptions = Apollo.BaseMutationOptions<CreateResourceDefinitionMutation, CreateResourceDefinitionMutationVariables>;
export const CreateIntegrationDocument = gql`
    mutation CreateIntegration($name: String!, $attrs: IntegrationAttributes!) {
  createIntegration(repositoryName: $name, attributes: $attrs) {
    id
  }
}
    `;
export type CreateIntegrationMutationFn = Apollo.MutationFunction<CreateIntegrationMutation, CreateIntegrationMutationVariables>;

/**
 * __useCreateIntegrationMutation__
 *
 * To run a mutation, you first call `useCreateIntegrationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateIntegrationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createIntegrationMutation, { data, loading, error }] = useCreateIntegrationMutation({
 *   variables: {
 *      name: // value for 'name'
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useCreateIntegrationMutation(baseOptions?: Apollo.MutationHookOptions<CreateIntegrationMutation, CreateIntegrationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateIntegrationMutation, CreateIntegrationMutationVariables>(CreateIntegrationDocument, options);
      }
export type CreateIntegrationMutationHookResult = ReturnType<typeof useCreateIntegrationMutation>;
export type CreateIntegrationMutationResult = Apollo.MutationResult<CreateIntegrationMutation>;
export type CreateIntegrationMutationOptions = Apollo.BaseMutationOptions<CreateIntegrationMutation, CreateIntegrationMutationVariables>;
export const UpdateRepositoryDocument = gql`
    mutation UpdateRepository($name: String!, $attrs: RepositoryAttributes!) {
  updateRepository(repositoryName: $name, attributes: $attrs) {
    id
  }
}
    `;
export type UpdateRepositoryMutationFn = Apollo.MutationFunction<UpdateRepositoryMutation, UpdateRepositoryMutationVariables>;

/**
 * __useUpdateRepositoryMutation__
 *
 * To run a mutation, you first call `useUpdateRepositoryMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateRepositoryMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateRepositoryMutation, { data, loading, error }] = useUpdateRepositoryMutation({
 *   variables: {
 *      name: // value for 'name'
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useUpdateRepositoryMutation(baseOptions?: Apollo.MutationHookOptions<UpdateRepositoryMutation, UpdateRepositoryMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateRepositoryMutation, UpdateRepositoryMutationVariables>(UpdateRepositoryDocument, options);
      }
export type UpdateRepositoryMutationHookResult = ReturnType<typeof useUpdateRepositoryMutation>;
export type UpdateRepositoryMutationResult = Apollo.MutationResult<UpdateRepositoryMutation>;
export type UpdateRepositoryMutationOptions = Apollo.BaseMutationOptions<UpdateRepositoryMutation, UpdateRepositoryMutationVariables>;
export const CreateRepositoryDocument = gql`
    mutation CreateRepository($name: String!, $publisher: String!, $attributes: RepositoryAttributes!) {
  upsertRepository(name: $name, publisher: $publisher, attributes: $attributes) {
    id
  }
}
    `;
export type CreateRepositoryMutationFn = Apollo.MutationFunction<CreateRepositoryMutation, CreateRepositoryMutationVariables>;

/**
 * __useCreateRepositoryMutation__
 *
 * To run a mutation, you first call `useCreateRepositoryMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateRepositoryMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createRepositoryMutation, { data, loading, error }] = useCreateRepositoryMutation({
 *   variables: {
 *      name: // value for 'name'
 *      publisher: // value for 'publisher'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateRepositoryMutation(baseOptions?: Apollo.MutationHookOptions<CreateRepositoryMutation, CreateRepositoryMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateRepositoryMutation, CreateRepositoryMutationVariables>(CreateRepositoryDocument, options);
      }
export type CreateRepositoryMutationHookResult = ReturnType<typeof useCreateRepositoryMutation>;
export type CreateRepositoryMutationResult = Apollo.MutationResult<CreateRepositoryMutation>;
export type CreateRepositoryMutationOptions = Apollo.BaseMutationOptions<CreateRepositoryMutation, CreateRepositoryMutationVariables>;
export const AcquireLockDocument = gql`
    mutation AcquireLock($name: String!) {
  acquireLock(repository: $name) {
    ...ApplyLock
  }
}
    ${ApplyLockFragmentDoc}`;
export type AcquireLockMutationFn = Apollo.MutationFunction<AcquireLockMutation, AcquireLockMutationVariables>;

/**
 * __useAcquireLockMutation__
 *
 * To run a mutation, you first call `useAcquireLockMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAcquireLockMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [acquireLockMutation, { data, loading, error }] = useAcquireLockMutation({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useAcquireLockMutation(baseOptions?: Apollo.MutationHookOptions<AcquireLockMutation, AcquireLockMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AcquireLockMutation, AcquireLockMutationVariables>(AcquireLockDocument, options);
      }
export type AcquireLockMutationHookResult = ReturnType<typeof useAcquireLockMutation>;
export type AcquireLockMutationResult = Apollo.MutationResult<AcquireLockMutation>;
export type AcquireLockMutationOptions = Apollo.BaseMutationOptions<AcquireLockMutation, AcquireLockMutationVariables>;
export const ReleaseLockDocument = gql`
    mutation ReleaseLock($name: String!, $attrs: LockAttributes!) {
  releaseLock(repository: $name, attributes: $attrs) {
    ...ApplyLock
  }
}
    ${ApplyLockFragmentDoc}`;
export type ReleaseLockMutationFn = Apollo.MutationFunction<ReleaseLockMutation, ReleaseLockMutationVariables>;

/**
 * __useReleaseLockMutation__
 *
 * To run a mutation, you first call `useReleaseLockMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useReleaseLockMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [releaseLockMutation, { data, loading, error }] = useReleaseLockMutation({
 *   variables: {
 *      name: // value for 'name'
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useReleaseLockMutation(baseOptions?: Apollo.MutationHookOptions<ReleaseLockMutation, ReleaseLockMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ReleaseLockMutation, ReleaseLockMutationVariables>(ReleaseLockDocument, options);
      }
export type ReleaseLockMutationHookResult = ReturnType<typeof useReleaseLockMutation>;
export type ReleaseLockMutationResult = Apollo.MutationResult<ReleaseLockMutation>;
export type ReleaseLockMutationOptions = Apollo.BaseMutationOptions<ReleaseLockMutation, ReleaseLockMutationVariables>;
export const UnlockRepositoryDocument = gql`
    mutation UnlockRepository($name: String!) {
  unlockRepository(name: $name)
}
    `;
export type UnlockRepositoryMutationFn = Apollo.MutationFunction<UnlockRepositoryMutation, UnlockRepositoryMutationVariables>;

/**
 * __useUnlockRepositoryMutation__
 *
 * To run a mutation, you first call `useUnlockRepositoryMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUnlockRepositoryMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [unlockRepositoryMutation, { data, loading, error }] = useUnlockRepositoryMutation({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useUnlockRepositoryMutation(baseOptions?: Apollo.MutationHookOptions<UnlockRepositoryMutation, UnlockRepositoryMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UnlockRepositoryMutation, UnlockRepositoryMutationVariables>(UnlockRepositoryDocument, options);
      }
export type UnlockRepositoryMutationHookResult = ReturnType<typeof useUnlockRepositoryMutation>;
export type UnlockRepositoryMutationResult = Apollo.MutationResult<UnlockRepositoryMutation>;
export type UnlockRepositoryMutationOptions = Apollo.BaseMutationOptions<UnlockRepositoryMutation, UnlockRepositoryMutationVariables>;
export const ListRepositoriesDocument = gql`
    query ListRepositories($q: String, $installed: Boolean, $first: Int = 100, $provider: Provider = null) {
  repositories(q: $q, first: $first, installed: $installed, provider: $provider) {
    edges {
      node {
        ...Repository
      }
    }
  }
}
    ${RepositoryFragmentDoc}`;

/**
 * __useListRepositoriesQuery__
 *
 * To run a query within a React component, call `useListRepositoriesQuery` and pass it any options that fit your needs.
 * When your component renders, `useListRepositoriesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListRepositoriesQuery({
 *   variables: {
 *      q: // value for 'q'
 *      installed: // value for 'installed'
 *      first: // value for 'first'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useListRepositoriesQuery(baseOptions?: Apollo.QueryHookOptions<ListRepositoriesQuery, ListRepositoriesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListRepositoriesQuery, ListRepositoriesQueryVariables>(ListRepositoriesDocument, options);
      }
export function useListRepositoriesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListRepositoriesQuery, ListRepositoriesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListRepositoriesQuery, ListRepositoriesQueryVariables>(ListRepositoriesDocument, options);
        }
export type ListRepositoriesQueryHookResult = ReturnType<typeof useListRepositoriesQuery>;
export type ListRepositoriesLazyQueryHookResult = ReturnType<typeof useListRepositoriesLazyQuery>;
export type ListRepositoriesQueryResult = Apollo.QueryResult<ListRepositoriesQuery, ListRepositoriesQueryVariables>;
export const ScaffoldsDocument = gql`
    query Scaffolds($app: String!, $pub: String!, $cat: Category!, $ing: Boolean, $pg: Boolean) {
  scaffold(application: $app, publisher: $pub, category: $cat, ingress: $ing, postgres: $pg) {
    path
    content
  }
}
    `;

/**
 * __useScaffoldsQuery__
 *
 * To run a query within a React component, call `useScaffoldsQuery` and pass it any options that fit your needs.
 * When your component renders, `useScaffoldsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useScaffoldsQuery({
 *   variables: {
 *      app: // value for 'app'
 *      pub: // value for 'pub'
 *      cat: // value for 'cat'
 *      ing: // value for 'ing'
 *      pg: // value for 'pg'
 *   },
 * });
 */
export function useScaffoldsQuery(baseOptions: Apollo.QueryHookOptions<ScaffoldsQuery, ScaffoldsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ScaffoldsQuery, ScaffoldsQueryVariables>(ScaffoldsDocument, options);
      }
export function useScaffoldsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ScaffoldsQuery, ScaffoldsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ScaffoldsQuery, ScaffoldsQueryVariables>(ScaffoldsDocument, options);
        }
export type ScaffoldsQueryHookResult = ReturnType<typeof useScaffoldsQuery>;
export type ScaffoldsLazyQueryHookResult = ReturnType<typeof useScaffoldsLazyQuery>;
export type ScaffoldsQueryResult = Apollo.QueryResult<ScaffoldsQuery, ScaffoldsQueryVariables>;
export const DeleteRepositoryDocument = gql`
    mutation DeleteRepository($id: ID!) {
  deleteRepository(repositoryId: $id) {
    id
  }
}
    `;
export type DeleteRepositoryMutationFn = Apollo.MutationFunction<DeleteRepositoryMutation, DeleteRepositoryMutationVariables>;

/**
 * __useDeleteRepositoryMutation__
 *
 * To run a mutation, you first call `useDeleteRepositoryMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteRepositoryMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteRepositoryMutation, { data, loading, error }] = useDeleteRepositoryMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteRepositoryMutation(baseOptions?: Apollo.MutationHookOptions<DeleteRepositoryMutation, DeleteRepositoryMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteRepositoryMutation, DeleteRepositoryMutationVariables>(DeleteRepositoryDocument, options);
      }
export type DeleteRepositoryMutationHookResult = ReturnType<typeof useDeleteRepositoryMutation>;
export type DeleteRepositoryMutationResult = Apollo.MutationResult<DeleteRepositoryMutation>;
export type DeleteRepositoryMutationOptions = Apollo.BaseMutationOptions<DeleteRepositoryMutation, DeleteRepositoryMutationVariables>;
export const GetTfProvidersDocument = gql`
    query GetTfProviders {
  terraformProviders
}
    `;

/**
 * __useGetTfProvidersQuery__
 *
 * To run a query within a React component, call `useGetTfProvidersQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTfProvidersQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTfProvidersQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetTfProvidersQuery(baseOptions?: Apollo.QueryHookOptions<GetTfProvidersQuery, GetTfProvidersQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTfProvidersQuery, GetTfProvidersQueryVariables>(GetTfProvidersDocument, options);
      }
export function useGetTfProvidersLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTfProvidersQuery, GetTfProvidersQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTfProvidersQuery, GetTfProvidersQueryVariables>(GetTfProvidersDocument, options);
        }
export type GetTfProvidersQueryHookResult = ReturnType<typeof useGetTfProvidersQuery>;
export type GetTfProvidersLazyQueryHookResult = ReturnType<typeof useGetTfProvidersLazyQuery>;
export type GetTfProvidersQueryResult = Apollo.QueryResult<GetTfProvidersQuery, GetTfProvidersQueryVariables>;
export const GetTfProviderScaffoldDocument = gql`
    query GetTfProviderScaffold($name: Provider!, $vsn: String) {
  terraformProvider(name: $name, vsn: $vsn) {
    name
    content
  }
}
    `;

/**
 * __useGetTfProviderScaffoldQuery__
 *
 * To run a query within a React component, call `useGetTfProviderScaffoldQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTfProviderScaffoldQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTfProviderScaffoldQuery({
 *   variables: {
 *      name: // value for 'name'
 *      vsn: // value for 'vsn'
 *   },
 * });
 */
export function useGetTfProviderScaffoldQuery(baseOptions: Apollo.QueryHookOptions<GetTfProviderScaffoldQuery, GetTfProviderScaffoldQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTfProviderScaffoldQuery, GetTfProviderScaffoldQueryVariables>(GetTfProviderScaffoldDocument, options);
      }
export function useGetTfProviderScaffoldLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTfProviderScaffoldQuery, GetTfProviderScaffoldQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTfProviderScaffoldQuery, GetTfProviderScaffoldQueryVariables>(GetTfProviderScaffoldDocument, options);
        }
export type GetTfProviderScaffoldQueryHookResult = ReturnType<typeof useGetTfProviderScaffoldQuery>;
export type GetTfProviderScaffoldLazyQueryHookResult = ReturnType<typeof useGetTfProviderScaffoldLazyQuery>;
export type GetTfProviderScaffoldQueryResult = Apollo.QueryResult<GetTfProviderScaffoldQuery, GetTfProviderScaffoldQueryVariables>;
export const GetShellDocument = gql`
    query GetShell {
  shell {
    ...CloudShell
  }
}
    ${CloudShellFragmentDoc}`;

/**
 * __useGetShellQuery__
 *
 * To run a query within a React component, call `useGetShellQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetShellQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetShellQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetShellQuery(baseOptions?: Apollo.QueryHookOptions<GetShellQuery, GetShellQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetShellQuery, GetShellQueryVariables>(GetShellDocument, options);
      }
export function useGetShellLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetShellQuery, GetShellQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetShellQuery, GetShellQueryVariables>(GetShellDocument, options);
        }
export type GetShellQueryHookResult = ReturnType<typeof useGetShellQuery>;
export type GetShellLazyQueryHookResult = ReturnType<typeof useGetShellLazyQuery>;
export type GetShellQueryResult = Apollo.QueryResult<GetShellQuery, GetShellQueryVariables>;
export const DeleteShellDocument = gql`
    mutation DeleteShell {
  deleteShell {
    ...CloudShell
  }
}
    ${CloudShellFragmentDoc}`;
export type DeleteShellMutationFn = Apollo.MutationFunction<DeleteShellMutation, DeleteShellMutationVariables>;

/**
 * __useDeleteShellMutation__
 *
 * To run a mutation, you first call `useDeleteShellMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteShellMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteShellMutation, { data, loading, error }] = useDeleteShellMutation({
 *   variables: {
 *   },
 * });
 */
export function useDeleteShellMutation(baseOptions?: Apollo.MutationHookOptions<DeleteShellMutation, DeleteShellMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteShellMutation, DeleteShellMutationVariables>(DeleteShellDocument, options);
      }
export type DeleteShellMutationHookResult = ReturnType<typeof useDeleteShellMutation>;
export type DeleteShellMutationResult = Apollo.MutationResult<DeleteShellMutation>;
export type DeleteShellMutationOptions = Apollo.BaseMutationOptions<DeleteShellMutation, DeleteShellMutationVariables>;
export const GetTerraformDocument = gql`
    query GetTerraform($id: ID!) {
  terraform(repositoryId: $id, first: 100) {
    edges {
      node {
        ...Terraform
      }
    }
  }
}
    ${TerraformFragmentDoc}`;

/**
 * __useGetTerraformQuery__
 *
 * To run a query within a React component, call `useGetTerraformQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTerraformQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTerraformQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetTerraformQuery(baseOptions: Apollo.QueryHookOptions<GetTerraformQuery, GetTerraformQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTerraformQuery, GetTerraformQueryVariables>(GetTerraformDocument, options);
      }
export function useGetTerraformLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTerraformQuery, GetTerraformQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTerraformQuery, GetTerraformQueryVariables>(GetTerraformDocument, options);
        }
export type GetTerraformQueryHookResult = ReturnType<typeof useGetTerraformQuery>;
export type GetTerraformLazyQueryHookResult = ReturnType<typeof useGetTerraformLazyQuery>;
export type GetTerraformQueryResult = Apollo.QueryResult<GetTerraformQuery, GetTerraformQueryVariables>;
export const GetTerraformInstallationsDocument = gql`
    query GetTerraformInstallations($id: ID!) {
  terraformInstallations(repositoryId: $id, first: 100) {
    edges {
      node {
        ...TerraformInstallation
      }
    }
  }
}
    ${TerraformInstallationFragmentDoc}`;

/**
 * __useGetTerraformInstallationsQuery__
 *
 * To run a query within a React component, call `useGetTerraformInstallationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetTerraformInstallationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetTerraformInstallationsQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGetTerraformInstallationsQuery(baseOptions: Apollo.QueryHookOptions<GetTerraformInstallationsQuery, GetTerraformInstallationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetTerraformInstallationsQuery, GetTerraformInstallationsQueryVariables>(GetTerraformInstallationsDocument, options);
      }
export function useGetTerraformInstallationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetTerraformInstallationsQuery, GetTerraformInstallationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetTerraformInstallationsQuery, GetTerraformInstallationsQueryVariables>(GetTerraformInstallationsDocument, options);
        }
export type GetTerraformInstallationsQueryHookResult = ReturnType<typeof useGetTerraformInstallationsQuery>;
export type GetTerraformInstallationsLazyQueryHookResult = ReturnType<typeof useGetTerraformInstallationsLazyQuery>;
export type GetTerraformInstallationsQueryResult = Apollo.QueryResult<GetTerraformInstallationsQuery, GetTerraformInstallationsQueryVariables>;
export const UploadTerraformDocument = gql`
    mutation UploadTerraform($repoName: String!, $name: String!, $uploadOrUrl: UploadOrUrl!) {
  uploadTerraform(repositoryName: $repoName, name: $name, attributes: {name: $name, package: $uploadOrUrl}) {
    ...Terraform
  }
}
    ${TerraformFragmentDoc}`;
export type UploadTerraformMutationFn = Apollo.MutationFunction<UploadTerraformMutation, UploadTerraformMutationVariables>;

/**
 * __useUploadTerraformMutation__
 *
 * To run a mutation, you first call `useUploadTerraformMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUploadTerraformMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [uploadTerraformMutation, { data, loading, error }] = useUploadTerraformMutation({
 *   variables: {
 *      repoName: // value for 'repoName'
 *      name: // value for 'name'
 *      uploadOrUrl: // value for 'uploadOrUrl'
 *   },
 * });
 */
export function useUploadTerraformMutation(baseOptions?: Apollo.MutationHookOptions<UploadTerraformMutation, UploadTerraformMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UploadTerraformMutation, UploadTerraformMutationVariables>(UploadTerraformDocument, options);
      }
export type UploadTerraformMutationHookResult = ReturnType<typeof useUploadTerraformMutation>;
export type UploadTerraformMutationResult = Apollo.MutationResult<UploadTerraformMutation>;
export type UploadTerraformMutationOptions = Apollo.BaseMutationOptions<UploadTerraformMutation, UploadTerraformMutationVariables>;
export const UninstallTerraformDocument = gql`
    mutation UninstallTerraform($id: ID!) {
  uninstallTerraform(id: $id) {
    id
  }
}
    `;
export type UninstallTerraformMutationFn = Apollo.MutationFunction<UninstallTerraformMutation, UninstallTerraformMutationVariables>;

/**
 * __useUninstallTerraformMutation__
 *
 * To run a mutation, you first call `useUninstallTerraformMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUninstallTerraformMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [uninstallTerraformMutation, { data, loading, error }] = useUninstallTerraformMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useUninstallTerraformMutation(baseOptions?: Apollo.MutationHookOptions<UninstallTerraformMutation, UninstallTerraformMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UninstallTerraformMutation, UninstallTerraformMutationVariables>(UninstallTerraformDocument, options);
      }
export type UninstallTerraformMutationHookResult = ReturnType<typeof useUninstallTerraformMutation>;
export type UninstallTerraformMutationResult = Apollo.MutationResult<UninstallTerraformMutation>;
export type UninstallTerraformMutationOptions = Apollo.BaseMutationOptions<UninstallTerraformMutation, UninstallTerraformMutationVariables>;
export const CreateTestDocument = gql`
    mutation CreateTest($name: String!, $attrs: TestAttributes!) {
  createTest(name: $name, attributes: $attrs) {
    ...Test
  }
}
    ${TestFragmentDoc}`;
export type CreateTestMutationFn = Apollo.MutationFunction<CreateTestMutation, CreateTestMutationVariables>;

/**
 * __useCreateTestMutation__
 *
 * To run a mutation, you first call `useCreateTestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateTestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createTestMutation, { data, loading, error }] = useCreateTestMutation({
 *   variables: {
 *      name: // value for 'name'
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useCreateTestMutation(baseOptions?: Apollo.MutationHookOptions<CreateTestMutation, CreateTestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateTestMutation, CreateTestMutationVariables>(CreateTestDocument, options);
      }
export type CreateTestMutationHookResult = ReturnType<typeof useCreateTestMutation>;
export type CreateTestMutationResult = Apollo.MutationResult<CreateTestMutation>;
export type CreateTestMutationOptions = Apollo.BaseMutationOptions<CreateTestMutation, CreateTestMutationVariables>;
export const UpdateTestDocument = gql`
    mutation UpdateTest($id: ID!, $attrs: TestAttributes!) {
  updateTest(id: $id, attributes: $attrs) {
    ...Test
  }
}
    ${TestFragmentDoc}`;
export type UpdateTestMutationFn = Apollo.MutationFunction<UpdateTestMutation, UpdateTestMutationVariables>;

/**
 * __useUpdateTestMutation__
 *
 * To run a mutation, you first call `useUpdateTestMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateTestMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateTestMutation, { data, loading, error }] = useUpdateTestMutation({
 *   variables: {
 *      id: // value for 'id'
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useUpdateTestMutation(baseOptions?: Apollo.MutationHookOptions<UpdateTestMutation, UpdateTestMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateTestMutation, UpdateTestMutationVariables>(UpdateTestDocument, options);
      }
export type UpdateTestMutationHookResult = ReturnType<typeof useUpdateTestMutation>;
export type UpdateTestMutationResult = Apollo.MutationResult<UpdateTestMutation>;
export type UpdateTestMutationOptions = Apollo.BaseMutationOptions<UpdateTestMutation, UpdateTestMutationVariables>;
export const UpdateStepDocument = gql`
    mutation UpdateStep($id: ID!, $logs: UploadOrUrl!) {
  updateStep(id: $id, attributes: {logs: $logs}) {
    id
  }
}
    `;
export type UpdateStepMutationFn = Apollo.MutationFunction<UpdateStepMutation, UpdateStepMutationVariables>;

/**
 * __useUpdateStepMutation__
 *
 * To run a mutation, you first call `useUpdateStepMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateStepMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateStepMutation, { data, loading, error }] = useUpdateStepMutation({
 *   variables: {
 *      id: // value for 'id'
 *      logs: // value for 'logs'
 *   },
 * });
 */
export function useUpdateStepMutation(baseOptions?: Apollo.MutationHookOptions<UpdateStepMutation, UpdateStepMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateStepMutation, UpdateStepMutationVariables>(UpdateStepDocument, options);
      }
export type UpdateStepMutationHookResult = ReturnType<typeof useUpdateStepMutation>;
export type UpdateStepMutationResult = Apollo.MutationResult<UpdateStepMutation>;
export type UpdateStepMutationOptions = Apollo.BaseMutationOptions<UpdateStepMutation, UpdateStepMutationVariables>;
export const PublishLogsDocument = gql`
    mutation PublishLogs($id: ID!, $logs: String!) {
  publishLogs(id: $id, logs: $logs) {
    id
  }
}
    `;
export type PublishLogsMutationFn = Apollo.MutationFunction<PublishLogsMutation, PublishLogsMutationVariables>;

/**
 * __usePublishLogsMutation__
 *
 * To run a mutation, you first call `usePublishLogsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePublishLogsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [publishLogsMutation, { data, loading, error }] = usePublishLogsMutation({
 *   variables: {
 *      id: // value for 'id'
 *      logs: // value for 'logs'
 *   },
 * });
 */
export function usePublishLogsMutation(baseOptions?: Apollo.MutationHookOptions<PublishLogsMutation, PublishLogsMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<PublishLogsMutation, PublishLogsMutationVariables>(PublishLogsDocument, options);
      }
export type PublishLogsMutationHookResult = ReturnType<typeof usePublishLogsMutation>;
export type PublishLogsMutationResult = Apollo.MutationResult<PublishLogsMutation>;
export type PublishLogsMutationOptions = Apollo.BaseMutationOptions<PublishLogsMutation, PublishLogsMutationVariables>;
export const MeDocument = gql`
    query Me {
  me {
    ...User
    loginMethod
    hasInstallations
    account {
      ...Account
      rootUser {
        id
        name
        email
      }
      domainMappings {
        id
        domain
        enableSso
      }
    }
    publisher {
      ...Publisher
      billingAccountId
    }
    boundRoles {
      ...Role
    }
  }
  configuration {
    stripeConnectId
    stripePublishableKey
    registry
    gitCommit
  }
}
    ${UserFragmentDoc}
${AccountFragmentDoc}
${PublisherFragmentDoc}
${RoleFragmentDoc}`;

/**
 * __useMeQuery__
 *
 * To run a query within a React component, call `useMeQuery` and pass it any options that fit your needs.
 * When your component renders, `useMeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMeQuery({
 *   variables: {
 *   },
 * });
 */
export function useMeQuery(baseOptions?: Apollo.QueryHookOptions<MeQuery, MeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MeQuery, MeQueryVariables>(MeDocument, options);
      }
export function useMeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MeQuery, MeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MeQuery, MeQueryVariables>(MeDocument, options);
        }
export type MeQueryHookResult = ReturnType<typeof useMeQuery>;
export type MeLazyQueryHookResult = ReturnType<typeof useMeLazyQuery>;
export type MeQueryResult = Apollo.QueryResult<MeQuery, MeQueryVariables>;
export const GetLoginMethodDocument = gql`
    query GetLoginMethod($email: String!) {
  loginMethod(email: $email) {
    loginMethod
    token
  }
}
    `;

/**
 * __useGetLoginMethodQuery__
 *
 * To run a query within a React component, call `useGetLoginMethodQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetLoginMethodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetLoginMethodQuery({
 *   variables: {
 *      email: // value for 'email'
 *   },
 * });
 */
export function useGetLoginMethodQuery(baseOptions: Apollo.QueryHookOptions<GetLoginMethodQuery, GetLoginMethodQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetLoginMethodQuery, GetLoginMethodQueryVariables>(GetLoginMethodDocument, options);
      }
export function useGetLoginMethodLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetLoginMethodQuery, GetLoginMethodQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetLoginMethodQuery, GetLoginMethodQueryVariables>(GetLoginMethodDocument, options);
        }
export type GetLoginMethodQueryHookResult = ReturnType<typeof useGetLoginMethodQuery>;
export type GetLoginMethodLazyQueryHookResult = ReturnType<typeof useGetLoginMethodLazyQuery>;
export type GetLoginMethodQueryResult = Apollo.QueryResult<GetLoginMethodQuery, GetLoginMethodQueryVariables>;
export const ListTokensDocument = gql`
    query ListTokens {
  tokens(first: 3) {
    edges {
      node {
        token
      }
    }
  }
}
    `;

/**
 * __useListTokensQuery__
 *
 * To run a query within a React component, call `useListTokensQuery` and pass it any options that fit your needs.
 * When your component renders, `useListTokensQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListTokensQuery({
 *   variables: {
 *   },
 * });
 */
export function useListTokensQuery(baseOptions?: Apollo.QueryHookOptions<ListTokensQuery, ListTokensQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListTokensQuery, ListTokensQueryVariables>(ListTokensDocument, options);
      }
export function useListTokensLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListTokensQuery, ListTokensQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListTokensQuery, ListTokensQueryVariables>(ListTokensDocument, options);
        }
export type ListTokensQueryHookResult = ReturnType<typeof useListTokensQuery>;
export type ListTokensLazyQueryHookResult = ReturnType<typeof useListTokensLazyQuery>;
export type ListTokensQueryResult = Apollo.QueryResult<ListTokensQuery, ListTokensQueryVariables>;
export const ListKeysDocument = gql`
    query ListKeys($emails: [String]) {
  publicKeys(emails: $emails, first: 1000) {
    edges {
      node {
        ...PublicKey
      }
    }
  }
}
    ${PublicKeyFragmentDoc}`;

/**
 * __useListKeysQuery__
 *
 * To run a query within a React component, call `useListKeysQuery` and pass it any options that fit your needs.
 * When your component renders, `useListKeysQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useListKeysQuery({
 *   variables: {
 *      emails: // value for 'emails'
 *   },
 * });
 */
export function useListKeysQuery(baseOptions?: Apollo.QueryHookOptions<ListKeysQuery, ListKeysQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ListKeysQuery, ListKeysQueryVariables>(ListKeysDocument, options);
      }
export function useListKeysLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ListKeysQuery, ListKeysQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ListKeysQuery, ListKeysQueryVariables>(ListKeysDocument, options);
        }
export type ListKeysQueryHookResult = ReturnType<typeof useListKeysQuery>;
export type ListKeysLazyQueryHookResult = ReturnType<typeof useListKeysLazyQuery>;
export type ListKeysQueryResult = Apollo.QueryResult<ListKeysQuery, ListKeysQueryVariables>;
export const GetEabCredentialDocument = gql`
    query GetEabCredential($cluster: String!, $provider: Provider!) {
  eabCredential(cluster: $cluster, provider: $provider) {
    ...EabCredential
  }
}
    ${EabCredentialFragmentDoc}`;

/**
 * __useGetEabCredentialQuery__
 *
 * To run a query within a React component, call `useGetEabCredentialQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetEabCredentialQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetEabCredentialQuery({
 *   variables: {
 *      cluster: // value for 'cluster'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useGetEabCredentialQuery(baseOptions: Apollo.QueryHookOptions<GetEabCredentialQuery, GetEabCredentialQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetEabCredentialQuery, GetEabCredentialQueryVariables>(GetEabCredentialDocument, options);
      }
export function useGetEabCredentialLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetEabCredentialQuery, GetEabCredentialQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetEabCredentialQuery, GetEabCredentialQueryVariables>(GetEabCredentialDocument, options);
        }
export type GetEabCredentialQueryHookResult = ReturnType<typeof useGetEabCredentialQuery>;
export type GetEabCredentialLazyQueryHookResult = ReturnType<typeof useGetEabCredentialLazyQuery>;
export type GetEabCredentialQueryResult = Apollo.QueryResult<GetEabCredentialQuery, GetEabCredentialQueryVariables>;
export const DevLoginDocument = gql`
    mutation DevLogin {
  deviceLogin {
    loginUrl
    deviceToken
  }
}
    `;
export type DevLoginMutationFn = Apollo.MutationFunction<DevLoginMutation, DevLoginMutationVariables>;

/**
 * __useDevLoginMutation__
 *
 * To run a mutation, you first call `useDevLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDevLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [devLoginMutation, { data, loading, error }] = useDevLoginMutation({
 *   variables: {
 *   },
 * });
 */
export function useDevLoginMutation(baseOptions?: Apollo.MutationHookOptions<DevLoginMutation, DevLoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DevLoginMutation, DevLoginMutationVariables>(DevLoginDocument, options);
      }
export type DevLoginMutationHookResult = ReturnType<typeof useDevLoginMutation>;
export type DevLoginMutationResult = Apollo.MutationResult<DevLoginMutation>;
export type DevLoginMutationOptions = Apollo.BaseMutationOptions<DevLoginMutation, DevLoginMutationVariables>;
export const LoginDocument = gql`
    mutation Login($email: String!, $password: String!, $deviceToken: String) {
  login(email: $email, password: $password, deviceToken: $deviceToken) {
    jwt
  }
}
    `;
export type LoginMutationFn = Apollo.MutationFunction<LoginMutation, LoginMutationVariables>;

/**
 * __useLoginMutation__
 *
 * To run a mutation, you first call `useLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [loginMutation, { data, loading, error }] = useLoginMutation({
 *   variables: {
 *      email: // value for 'email'
 *      password: // value for 'password'
 *      deviceToken: // value for 'deviceToken'
 *   },
 * });
 */
export function useLoginMutation(baseOptions?: Apollo.MutationHookOptions<LoginMutation, LoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<LoginMutation, LoginMutationVariables>(LoginDocument, options);
      }
export type LoginMutationHookResult = ReturnType<typeof useLoginMutation>;
export type LoginMutationResult = Apollo.MutationResult<LoginMutation>;
export type LoginMutationOptions = Apollo.BaseMutationOptions<LoginMutation, LoginMutationVariables>;
export const ImpersonateServiceAccountDocument = gql`
    mutation ImpersonateServiceAccount($email: String) {
  impersonateServiceAccount(email: $email) {
    jwt
    email
  }
}
    `;
export type ImpersonateServiceAccountMutationFn = Apollo.MutationFunction<ImpersonateServiceAccountMutation, ImpersonateServiceAccountMutationVariables>;

/**
 * __useImpersonateServiceAccountMutation__
 *
 * To run a mutation, you first call `useImpersonateServiceAccountMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useImpersonateServiceAccountMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [impersonateServiceAccountMutation, { data, loading, error }] = useImpersonateServiceAccountMutation({
 *   variables: {
 *      email: // value for 'email'
 *   },
 * });
 */
export function useImpersonateServiceAccountMutation(baseOptions?: Apollo.MutationHookOptions<ImpersonateServiceAccountMutation, ImpersonateServiceAccountMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ImpersonateServiceAccountMutation, ImpersonateServiceAccountMutationVariables>(ImpersonateServiceAccountDocument, options);
      }
export type ImpersonateServiceAccountMutationHookResult = ReturnType<typeof useImpersonateServiceAccountMutation>;
export type ImpersonateServiceAccountMutationResult = Apollo.MutationResult<ImpersonateServiceAccountMutation>;
export type ImpersonateServiceAccountMutationOptions = Apollo.BaseMutationOptions<ImpersonateServiceAccountMutation, ImpersonateServiceAccountMutationVariables>;
export const CreateAccessTokenDocument = gql`
    mutation CreateAccessToken {
  createToken {
    token
  }
}
    `;
export type CreateAccessTokenMutationFn = Apollo.MutationFunction<CreateAccessTokenMutation, CreateAccessTokenMutationVariables>;

/**
 * __useCreateAccessTokenMutation__
 *
 * To run a mutation, you first call `useCreateAccessTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateAccessTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createAccessTokenMutation, { data, loading, error }] = useCreateAccessTokenMutation({
 *   variables: {
 *   },
 * });
 */
export function useCreateAccessTokenMutation(baseOptions?: Apollo.MutationHookOptions<CreateAccessTokenMutation, CreateAccessTokenMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateAccessTokenMutation, CreateAccessTokenMutationVariables>(CreateAccessTokenDocument, options);
      }
export type CreateAccessTokenMutationHookResult = ReturnType<typeof useCreateAccessTokenMutation>;
export type CreateAccessTokenMutationResult = Apollo.MutationResult<CreateAccessTokenMutation>;
export type CreateAccessTokenMutationOptions = Apollo.BaseMutationOptions<CreateAccessTokenMutation, CreateAccessTokenMutationVariables>;
export const CreateKeyDocument = gql`
    mutation CreateKey($key: String!, $name: String!) {
  createPublicKey(attributes: {content: $key, name: $name}) {
    id
  }
}
    `;
export type CreateKeyMutationFn = Apollo.MutationFunction<CreateKeyMutation, CreateKeyMutationVariables>;

/**
 * __useCreateKeyMutation__
 *
 * To run a mutation, you first call `useCreateKeyMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateKeyMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createKeyMutation, { data, loading, error }] = useCreateKeyMutation({
 *   variables: {
 *      key: // value for 'key'
 *      name: // value for 'name'
 *   },
 * });
 */
export function useCreateKeyMutation(baseOptions?: Apollo.MutationHookOptions<CreateKeyMutation, CreateKeyMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateKeyMutation, CreateKeyMutationVariables>(CreateKeyDocument, options);
      }
export type CreateKeyMutationHookResult = ReturnType<typeof useCreateKeyMutation>;
export type CreateKeyMutationResult = Apollo.MutationResult<CreateKeyMutation>;
export type CreateKeyMutationOptions = Apollo.BaseMutationOptions<CreateKeyMutation, CreateKeyMutationVariables>;
export const DeleteEabCredentialDocument = gql`
    mutation DeleteEabCredential($cluster: String!, $provider: Provider!) {
  deleteEabKey(cluster: $cluster, provider: $provider) {
    id
  }
}
    `;
export type DeleteEabCredentialMutationFn = Apollo.MutationFunction<DeleteEabCredentialMutation, DeleteEabCredentialMutationVariables>;

/**
 * __useDeleteEabCredentialMutation__
 *
 * To run a mutation, you first call `useDeleteEabCredentialMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteEabCredentialMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteEabCredentialMutation, { data, loading, error }] = useDeleteEabCredentialMutation({
 *   variables: {
 *      cluster: // value for 'cluster'
 *      provider: // value for 'provider'
 *   },
 * });
 */
export function useDeleteEabCredentialMutation(baseOptions?: Apollo.MutationHookOptions<DeleteEabCredentialMutation, DeleteEabCredentialMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteEabCredentialMutation, DeleteEabCredentialMutationVariables>(DeleteEabCredentialDocument, options);
      }
export type DeleteEabCredentialMutationHookResult = ReturnType<typeof useDeleteEabCredentialMutation>;
export type DeleteEabCredentialMutationResult = Apollo.MutationResult<DeleteEabCredentialMutation>;
export type DeleteEabCredentialMutationOptions = Apollo.BaseMutationOptions<DeleteEabCredentialMutation, DeleteEabCredentialMutationVariables>;
export const CreateEventDocument = gql`
    mutation CreateEvent($attrs: UserEventAttributes!) {
  createUserEvent(attributes: $attrs)
}
    `;
export type CreateEventMutationFn = Apollo.MutationFunction<CreateEventMutation, CreateEventMutationVariables>;

/**
 * __useCreateEventMutation__
 *
 * To run a mutation, you first call `useCreateEventMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateEventMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createEventMutation, { data, loading, error }] = useCreateEventMutation({
 *   variables: {
 *      attrs: // value for 'attrs'
 *   },
 * });
 */
export function useCreateEventMutation(baseOptions?: Apollo.MutationHookOptions<CreateEventMutation, CreateEventMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateEventMutation, CreateEventMutationVariables>(CreateEventDocument, options);
      }
export type CreateEventMutationHookResult = ReturnType<typeof useCreateEventMutation>;
export type CreateEventMutationResult = Apollo.MutationResult<CreateEventMutation>;
export type CreateEventMutationOptions = Apollo.BaseMutationOptions<CreateEventMutation, CreateEventMutationVariables>;
export const LoginMethodDocument = gql`
    query LoginMethod($email: String!, $host: String) {
  loginMethod(email: $email, host: $host) {
    loginMethod
    token
    authorizeUrl
  }
}
    `;

/**
 * __useLoginMethodQuery__
 *
 * To run a query within a React component, call `useLoginMethodQuery` and pass it any options that fit your needs.
 * When your component renders, `useLoginMethodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useLoginMethodQuery({
 *   variables: {
 *      email: // value for 'email'
 *      host: // value for 'host'
 *   },
 * });
 */
export function useLoginMethodQuery(baseOptions: Apollo.QueryHookOptions<LoginMethodQuery, LoginMethodQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<LoginMethodQuery, LoginMethodQueryVariables>(LoginMethodDocument, options);
      }
export function useLoginMethodLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<LoginMethodQuery, LoginMethodQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<LoginMethodQuery, LoginMethodQueryVariables>(LoginMethodDocument, options);
        }
export type LoginMethodQueryHookResult = ReturnType<typeof useLoginMethodQuery>;
export type LoginMethodLazyQueryHookResult = ReturnType<typeof useLoginMethodLazyQuery>;
export type LoginMethodQueryResult = Apollo.QueryResult<LoginMethodQuery, LoginMethodQueryVariables>;
export const SignupDocument = gql`
    mutation Signup($attributes: UserAttributes!, $account: AccountAttributes, $deviceToken: String) {
  signup(attributes: $attributes, account: $account, deviceToken: $deviceToken) {
    jwt
    onboarding
  }
}
    `;
export type SignupMutationFn = Apollo.MutationFunction<SignupMutation, SignupMutationVariables>;

/**
 * __useSignupMutation__
 *
 * To run a mutation, you first call `useSignupMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSignupMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [signupMutation, { data, loading, error }] = useSignupMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *      account: // value for 'account'
 *      deviceToken: // value for 'deviceToken'
 *   },
 * });
 */
export function useSignupMutation(baseOptions?: Apollo.MutationHookOptions<SignupMutation, SignupMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SignupMutation, SignupMutationVariables>(SignupDocument, options);
      }
export type SignupMutationHookResult = ReturnType<typeof useSignupMutation>;
export type SignupMutationResult = Apollo.MutationResult<SignupMutation>;
export type SignupMutationOptions = Apollo.BaseMutationOptions<SignupMutation, SignupMutationVariables>;
export const PasswordlessLoginDocument = gql`
    mutation PasswordlessLogin($token: String!) {
  passwordlessLogin(token: $token) {
    jwt
  }
}
    `;
export type PasswordlessLoginMutationFn = Apollo.MutationFunction<PasswordlessLoginMutation, PasswordlessLoginMutationVariables>;

/**
 * __usePasswordlessLoginMutation__
 *
 * To run a mutation, you first call `usePasswordlessLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePasswordlessLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [passwordlessLoginMutation, { data, loading, error }] = usePasswordlessLoginMutation({
 *   variables: {
 *      token: // value for 'token'
 *   },
 * });
 */
export function usePasswordlessLoginMutation(baseOptions?: Apollo.MutationHookOptions<PasswordlessLoginMutation, PasswordlessLoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<PasswordlessLoginMutation, PasswordlessLoginMutationVariables>(PasswordlessLoginDocument, options);
      }
export type PasswordlessLoginMutationHookResult = ReturnType<typeof usePasswordlessLoginMutation>;
export type PasswordlessLoginMutationResult = Apollo.MutationResult<PasswordlessLoginMutation>;
export type PasswordlessLoginMutationOptions = Apollo.BaseMutationOptions<PasswordlessLoginMutation, PasswordlessLoginMutationVariables>;
export const PollLoginTokenDocument = gql`
    mutation PollLoginToken($token: String!, $deviceToken: String) {
  loginToken(token: $token, deviceToken: $deviceToken) {
    jwt
  }
}
    `;
export type PollLoginTokenMutationFn = Apollo.MutationFunction<PollLoginTokenMutation, PollLoginTokenMutationVariables>;

/**
 * __usePollLoginTokenMutation__
 *
 * To run a mutation, you first call `usePollLoginTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePollLoginTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [pollLoginTokenMutation, { data, loading, error }] = usePollLoginTokenMutation({
 *   variables: {
 *      token: // value for 'token'
 *      deviceToken: // value for 'deviceToken'
 *   },
 * });
 */
export function usePollLoginTokenMutation(baseOptions?: Apollo.MutationHookOptions<PollLoginTokenMutation, PollLoginTokenMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<PollLoginTokenMutation, PollLoginTokenMutationVariables>(PollLoginTokenDocument, options);
      }
export type PollLoginTokenMutationHookResult = ReturnType<typeof usePollLoginTokenMutation>;
export type PollLoginTokenMutationResult = Apollo.MutationResult<PollLoginTokenMutation>;
export type PollLoginTokenMutationOptions = Apollo.BaseMutationOptions<PollLoginTokenMutation, PollLoginTokenMutationVariables>;
export const OauthUrlsDocument = gql`
    query OauthUrls($host: String) {
  oauthUrls(host: $host) {
    provider
    authorizeUrl
  }
}
    `;

/**
 * __useOauthUrlsQuery__
 *
 * To run a query within a React component, call `useOauthUrlsQuery` and pass it any options that fit your needs.
 * When your component renders, `useOauthUrlsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useOauthUrlsQuery({
 *   variables: {
 *      host: // value for 'host'
 *   },
 * });
 */
export function useOauthUrlsQuery(baseOptions?: Apollo.QueryHookOptions<OauthUrlsQuery, OauthUrlsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<OauthUrlsQuery, OauthUrlsQueryVariables>(OauthUrlsDocument, options);
      }
export function useOauthUrlsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<OauthUrlsQuery, OauthUrlsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<OauthUrlsQuery, OauthUrlsQueryVariables>(OauthUrlsDocument, options);
        }
export type OauthUrlsQueryHookResult = ReturnType<typeof useOauthUrlsQuery>;
export type OauthUrlsLazyQueryHookResult = ReturnType<typeof useOauthUrlsLazyQuery>;
export type OauthUrlsQueryResult = Apollo.QueryResult<OauthUrlsQuery, OauthUrlsQueryVariables>;
export const AcceptLoginDocument = gql`
    mutation AcceptLogin($challenge: String!) {
  acceptLogin(challenge: $challenge) {
    redirectTo
  }
}
    `;
export type AcceptLoginMutationFn = Apollo.MutationFunction<AcceptLoginMutation, AcceptLoginMutationVariables>;

/**
 * __useAcceptLoginMutation__
 *
 * To run a mutation, you first call `useAcceptLoginMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAcceptLoginMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [acceptLoginMutation, { data, loading, error }] = useAcceptLoginMutation({
 *   variables: {
 *      challenge: // value for 'challenge'
 *   },
 * });
 */
export function useAcceptLoginMutation(baseOptions?: Apollo.MutationHookOptions<AcceptLoginMutation, AcceptLoginMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AcceptLoginMutation, AcceptLoginMutationVariables>(AcceptLoginDocument, options);
      }
export type AcceptLoginMutationHookResult = ReturnType<typeof useAcceptLoginMutation>;
export type AcceptLoginMutationResult = Apollo.MutationResult<AcceptLoginMutation>;
export type AcceptLoginMutationOptions = Apollo.BaseMutationOptions<AcceptLoginMutation, AcceptLoginMutationVariables>;
export const CreateResetTokenDocument = gql`
    mutation CreateResetToken($attributes: ResetTokenAttributes!) {
  createResetToken(attributes: $attributes)
}
    `;
export type CreateResetTokenMutationFn = Apollo.MutationFunction<CreateResetTokenMutation, CreateResetTokenMutationVariables>;

/**
 * __useCreateResetTokenMutation__
 *
 * To run a mutation, you first call `useCreateResetTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateResetTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createResetTokenMutation, { data, loading, error }] = useCreateResetTokenMutation({
 *   variables: {
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useCreateResetTokenMutation(baseOptions?: Apollo.MutationHookOptions<CreateResetTokenMutation, CreateResetTokenMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateResetTokenMutation, CreateResetTokenMutationVariables>(CreateResetTokenDocument, options);
      }
export type CreateResetTokenMutationHookResult = ReturnType<typeof useCreateResetTokenMutation>;
export type CreateResetTokenMutationResult = Apollo.MutationResult<CreateResetTokenMutation>;
export type CreateResetTokenMutationOptions = Apollo.BaseMutationOptions<CreateResetTokenMutation, CreateResetTokenMutationVariables>;
export const RealizeResetTokenDocument = gql`
    mutation RealizeResetToken($id: ID!, $attributes: ResetTokenRealization!) {
  realizeResetToken(id: $id, attributes: $attributes)
}
    `;
export type RealizeResetTokenMutationFn = Apollo.MutationFunction<RealizeResetTokenMutation, RealizeResetTokenMutationVariables>;

/**
 * __useRealizeResetTokenMutation__
 *
 * To run a mutation, you first call `useRealizeResetTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRealizeResetTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [realizeResetTokenMutation, { data, loading, error }] = useRealizeResetTokenMutation({
 *   variables: {
 *      id: // value for 'id'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useRealizeResetTokenMutation(baseOptions?: Apollo.MutationHookOptions<RealizeResetTokenMutation, RealizeResetTokenMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RealizeResetTokenMutation, RealizeResetTokenMutationVariables>(RealizeResetTokenDocument, options);
      }
export type RealizeResetTokenMutationHookResult = ReturnType<typeof useRealizeResetTokenMutation>;
export type RealizeResetTokenMutationResult = Apollo.MutationResult<RealizeResetTokenMutation>;
export type RealizeResetTokenMutationOptions = Apollo.BaseMutationOptions<RealizeResetTokenMutation, RealizeResetTokenMutationVariables>;
export const ResetTokenDocument = gql`
    query ResetToken($id: ID!) {
  resetToken(id: $id) {
    type
    user {
      ...User
    }
  }
}
    ${UserFragmentDoc}`;

/**
 * __useResetTokenQuery__
 *
 * To run a query within a React component, call `useResetTokenQuery` and pass it any options that fit your needs.
 * When your component renders, `useResetTokenQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResetTokenQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useResetTokenQuery(baseOptions: Apollo.QueryHookOptions<ResetTokenQuery, ResetTokenQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ResetTokenQuery, ResetTokenQueryVariables>(ResetTokenDocument, options);
      }
export function useResetTokenLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ResetTokenQuery, ResetTokenQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ResetTokenQuery, ResetTokenQueryVariables>(ResetTokenDocument, options);
        }
export type ResetTokenQueryHookResult = ReturnType<typeof useResetTokenQuery>;
export type ResetTokenLazyQueryHookResult = ReturnType<typeof useResetTokenLazyQuery>;
export type ResetTokenQueryResult = Apollo.QueryResult<ResetTokenQuery, ResetTokenQueryVariables>;
export const UpdateVersionDocument = gql`
    mutation UpdateVersion($spec: VersionSpec, $attributes: VersionAttributes!) {
  updateVersion(spec: $spec, attributes: $attributes) {
    id
  }
}
    `;
export type UpdateVersionMutationFn = Apollo.MutationFunction<UpdateVersionMutation, UpdateVersionMutationVariables>;

/**
 * __useUpdateVersionMutation__
 *
 * To run a mutation, you first call `useUpdateVersionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateVersionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateVersionMutation, { data, loading, error }] = useUpdateVersionMutation({
 *   variables: {
 *      spec: // value for 'spec'
 *      attributes: // value for 'attributes'
 *   },
 * });
 */
export function useUpdateVersionMutation(baseOptions?: Apollo.MutationHookOptions<UpdateVersionMutation, UpdateVersionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateVersionMutation, UpdateVersionMutationVariables>(UpdateVersionDocument, options);
      }
export type UpdateVersionMutationHookResult = ReturnType<typeof useUpdateVersionMutation>;
export type UpdateVersionMutationResult = Apollo.MutationResult<UpdateVersionMutation>;
export type UpdateVersionMutationOptions = Apollo.BaseMutationOptions<UpdateVersionMutation, UpdateVersionMutationVariables>;