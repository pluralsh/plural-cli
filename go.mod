module github.com/pluralsh/plural-cli

go 1.22.5

require (
	cloud.google.com/go/compute v1.23.3
	cloud.google.com/go/resourcemanager v1.9.4
	cloud.google.com/go/serviceusage v1.6.0
	cloud.google.com/go/storage v1.35.1
	filippo.io/age v1.1.1
	github.com/AlecAivazis/survey/v2 v2.3.5
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.7.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.3.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage v1.3.0
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/Yamashou/gqlgenc v0.23.2
	github.com/aws/aws-sdk-go-v2 v1.27.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.32.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.6
	github.com/briandowns/spinner v1.23.0
	github.com/buger/goterm v1.0.4
	github.com/chartmuseum/helm-push v0.10.3
	github.com/databus23/helm-diff/v3 v3.6.0
	github.com/fatih/color v1.16.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-git/go-git/v5 v5.11.0
	github.com/gofrs/flock v0.8.1
	github.com/google/go-github/v45 v45.2.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hashicorp/go-retryablehttp v0.7.4
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.24.1
	github.com/helm/helm-mapkubeapis v0.4.1
	github.com/imdario/mergo v0.3.16
	github.com/inancgumus/screen v0.0.0-20190314163918-06e984b86ed3
	github.com/joho/godotenv v1.3.0
	github.com/ktrysmt/go-bitbucket v0.9.55
	github.com/likexian/doh v0.7.1
	github.com/linode/linodego v1.26.0
	github.com/mikesmitty/edkey v0.0.0-20170222072505-3356ea4e686a
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/norwoodj/helm-docs v1.11.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/packethost/packngo v0.29.0
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pluralsh/console/go/client v1.17.1-0.20240918005717-8285a4b181b1
	github.com/pluralsh/console/go/controller v0.0.0-20240918005717-8285a4b181b1
	github.com/pluralsh/gqlclient v1.12.2
	github.com/pluralsh/plural-operator v0.5.5
	github.com/pluralsh/polly v0.1.10
	github.com/posthog/posthog-go v0.0.0-20230801140217-d607812dee69
	github.com/rodaine/hclencoder v0.0.1
	github.com/samber/lo v1.46.0
	github.com/spf13/viper v1.18.2
	github.com/urfave/cli v1.22.14
	github.com/wailsapp/wails/v2 v2.4.1
	github.com/xanzy/go-gitlab v0.73.1
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64
	go.mercari.io/hcledit v0.0.8
	golang.org/x/crypto v0.25.0
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc
	golang.org/x/mod v0.18.0
	golang.org/x/oauth2 v0.19.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.14.3
	k8s.io/api v0.30.1
	k8s.io/apimachinery v0.31.0-beta.0
	k8s.io/client-go v0.30.1
	layeh.com/gopher-luar v1.0.10
	sigs.k8s.io/application v0.8.3
	sigs.k8s.io/controller-runtime v0.18.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/longrunning v0.5.4 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/99designs/gqlgen v0.17.49 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.1.1 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.3.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.17.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.4 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/bep/debounce v1.2.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/containerd/errdefs v0.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/cli v27.1.0+incompatible // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker v27.1.0+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.0 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.12.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-resty/resty/v2 v2.9.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hcl/v2 v2.15.0 // indirect
	github.com/jchv/go-winloader v0.0.0-20210711035445-715c2860da7e // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labstack/echo/v4 v4.9.0 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/leaanthony/go-ansi-parser v1.0.1 // indirect
	github.com/leaanthony/gosod v1.0.3 // indirect
	github.com/leaanthony/slicer v1.5.0 // indirect
	github.com/likexian/gokit v0.25.15 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/osteele/liquid v1.4.0 // indirect
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.0 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pluralsh/controller-reconcile-helper v0.0.4 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/skeema/knownhosts v1.2.1 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tkrajina/go-reflector v0.5.5 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/vektah/gqlparser/v2 v2.5.16 // indirect
	github.com/wailsapp/mimetype v1.4.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/otel v1.22.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.22.0 // indirect
	go.opentelemetry.io/otel/sdk v1.22.0 // indirect
	go.opentelemetry.io/otel/trace v1.22.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20231211222908-989df2bf70f3 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
)

require (
	cloud.google.com/go v0.110.10 // indirect
	cloud.google.com/go/iam v1.1.5
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.23 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.11
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.156.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.55.0
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cli/browser v1.0.0 // indirect
	github.com/cli/safeexec v1.0.0 // indirect
	github.com/containerd/containerd v1.7.20 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/go-cmp v0.6.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-log v0.7.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.4.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/oleiade/reflections v1.0.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pluralsh/oauth v0.9.2
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.52.2 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rubenv/sql-migrate v1.5.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/schollz/progressbar/v3 v3.14.1 // indirect
	github.com/sergi/go-diff v1.3.1 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.6-0.20210604193023-d5e0c0615ace // indirect
	github.com/stretchr/testify v1.9.0
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.starlark.net v0.0.0-20230525235612-a134d8f9ddca // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/term v0.22.0
	golang.org/x/text v0.16.0
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/api v0.153.0
	google.golang.org/grpc v1.60.1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/apiextensions-apiserver v0.30.0 // indirect
	k8s.io/apiserver v0.30.0 // indirect
	k8s.io/cli-runtime v0.29.3
	k8s.io/component-base v0.30.0 // indirect
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kubectl v0.29.3
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8 // indirect
	oras.land/oras-go v1.2.4 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.13.5-0.20230601165947-6ce0bf390ce3 // indirect
	sigs.k8s.io/kustomize/kyaml v0.17.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)

replace (
	github.com/docker/cli => github.com/docker/cli v24.0.6+incompatible
	github.com/docker/distribution => github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker => github.com/docker/docker v25.0.6+incompatible
	github.com/docker/go-connections => github.com/docker/go-connections v0.4.0
	github.com/norwoodj/helm-docs v1.11.2 => github.com/pluralsh/helm-docs v1.11.3-0.20230914191425-6d14ebab8817
	go.etcd.io/etcd/pkg/v3 => go.etcd.io/etcd/pkg/v3 v3.5.0-alpha.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.29.5
	k8s.io/client-go => k8s.io/client-go v0.29.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340
	oras.land/oras-go => oras.land/oras-go v1.2.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.16.6
	sigs.k8s.io/kustomize/api => sigs.k8s.io/kustomize/api v0.17.1
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.17.0
)
