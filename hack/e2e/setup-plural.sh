#!/usr/bin/env bash

set -euo pipefail

PLURALDIR=$(dirname $0)/../..

cd "$PLURALDIR"
source hack/lib.sh

PLURALHOME="$HOME"/.plural
TESTDIR="$HOME"/test
SHAREDIR="$HOME"/share

mkdir -p "$PLURALHOME"
mkdir -p "$HOME"/.ssh

echodate "Creating config.yaml ..."
cat << EOF > "$PLURALHOME"/config.yml
$CLI_E2E_CONF
EOF

echodate "Creating identity ..."
cat << EOF > "$PLURALHOME"/identity
$CLI_E2E_IDENTITY_FILE
EOF

echodate "Creating key ..."
cat << EOF > "$PLURALHOME"/key
$CLI_E2E_KEY_FILE
EOF

echodate "Creating private ssh key ..."
cat << EOF > "$HOME"/.ssh/id_rsa
$CLI_E2E_PRIVATE_KEY
EOF

echodate "Creating private sharing ssh key ..."
cat << EOF > "$HOME"/.ssh/id_sharing
$CLI_E2E_SHARING_PRIVATE_KEY
EOF

echodate "Creating public ssh key ..."
cat << EOF > "$HOME"/.ssh/id_rsa.pub
$CLI_E2E_PUBLIC_KEY
EOF

echodate "Creating public sharing ssh key ..."
cat << EOF > "$HOME"/.ssh/id_sharing.pub
$CLI_E2E_SHARING_PUBLIC_KEY
EOF

chmod 600 ~/.ssh/id_rsa
chmod 600 ~/.ssh/id_sharing

git -c core.sshCommand="ssh -i ~/.ssh/id_rsa" clone git@github.com:pluralsh/cli-e2e-tests.git "$TESTDIR"
git -c core.sshCommand="ssh -i ~/.ssh/id_sharing" clone git@github.com:pluralsh/e2e-sharing.git "$SHAREDIR"
git config --global user.email cli-e2e@pluraldev.sh
git config --global user.name cli-e2e


echodate "Creating workspace.yaml ..."
cat << EOF > "$TESTDIR"/workspace.yaml
apiVersion: plural.sh/v1alpha1
kind: ProjectManifest
metadata:
  name: testcli
spec:
  cluster: testcli
  bucket: testcli-tf-state
  project: ""
  provider: kind
  region: us-east-1
  owner:
    email: cli-e2e@pluraldev.sh
  network:
    subdomain: clie2e.onplural.sh
    pluraldns: true
  bucketPrefix: test
  context: {}
EOF


echodate "Entering to work directory ..."
cd "$TESTDIR"
git config core.sshCommand 'ssh -i ~/.ssh/id_rsa'
git checkout -b origin/main


export PLURAL_CONSOLE_HOSTNAME=minio.clie2e.onplural.sh
export PLURAL_CONSOLE_CONSOLEHOSTNAME=minioui.clie2e.onplural.sh
export PLURAL_CONSOLE_CONSOLE_DNS=console.clie2e.onplural.sh
export PLURAL_CONSOLE_ADMIN_EMAIL=cli-e2e@pluraldev.sh
export PLURAL_CONSOLE_ADMIN_NAME=cli-e2e
export PLURAL_CONSOLE_GIT_USER=cli-e2e
export PLURAL_CONSOLE_GIT_EMAIL=cli-e2e@pluraldev.sh
export PLURAL_CONSOLE_PRIVATE_KEY="$HOME"/.ssh/id_rsa
export PLURAL_CONSOLE_PUBLIC_KEY="$HOME"/.ssh/id_rsa.pub
export PLURAL_CONSOLE_PASSPHRASE=" "
export PLURAL_CONSOLE_WAL_BUCKET=test-testcli-postgres-wal

export PLURAL_CONFIRM_OIDC=true
export PLURAL_REPOS_RESET_CONFIRM=true

export PLURAL_LOGIN_AFFIRM_CURRENT_USER=true
export PLURAL_INIT_AFFIRM_CURRENT_REPO=true
export PLURAL_INIT_AFFIRM_BACKUP_KEY=false

plural init
plural repos reset
plural bundle install console console-kind
plural build --force
retry 3 plural deploy --commit=""






