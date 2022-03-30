# Plural CLI

The plural cli automates all gitops operations for your deployments of plural applications.  The core workflow should mostly be as simple as:

```bash
plural build
plural deploy
```

And if you want to teardown your infrastructure, you can simply run:

```bash
plural destroy
```

To add, update or reconfigure any applications deployed by plural.  But it goes even deeper and solves things like:

* Secret management (via a similar mechanism as git-crypt)
* Application Health checking - `plural watch APP`
* Log tailing - `plural logs list APP` and `plural logs tail APP LOGSTREAM`
* Setting up secure proxies into databases, private web UIs - `plural proxy list APP` and `plural proxy connect APP NAME`

## Installation

There are a number of means to install plural, the simplest is to use our homebrew tap if you're using mac:

```bash
brew install pluralsh/plural/plural
```

More detailed instructions for other platforms can be found at https://docs.plural.sh/getting-started#1.-install-plural-cli-and-dependencies

Plural does require a few other cli's to be installed, namely:
* helm
* terraform
* kubectl
* cloud provider cli for the infrastructure you're deploying to, like `awscli`, `gcloud`, etc
* [kind](https://kind.sigs.k8s.io/) if using kind to deploy a local cluster for testing

## Setup

The core workflow is all git based, so you should create a git repository on github or wherever you're using SCM, clone it locally, then run:

```bash
plural init
```

You'll want to then install a bundle for whatever application you'd like, we'll use https://github.com/airbytehq/airbyte as an example.  You can search for the bundles using:


```bash
plural bundle list airbyte
```

And chose one (using aws as an example cloud provider target) like:

```bash
plural bundle install airbyte airbyte-aws
```

This will set the basic configuration parameters for all the infrastructure needed to install airbyte.  Then just run:

```bash
plural build
plural deploy --commit "deploying my first plural app!"
```

To install it.


## Installing the Plural Console

We highly recommend installing the [plural console](https://github.com/pluralsh/console) alongside your plural applications.  That can be done easily with:

```bash
plural bundle install console console-aws
plural build
plural deploy --commit "deploying the plural console"
```

## Reaching Out

If you have any issues with your plural installations, or just want to show us some love, feel free to drop into our discord [here](https://discord.gg/bEBAMXV64s)