module get.porter.sh/porter

go 1.21

toolchain go1.21.3

replace (
	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Also includes branches we haven't PR'd yet: capture-yamux-logs, context-cancellation
	// Tagged from v1.4.4, the porter branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.4-porter.1

	// Fixes https://github.com/spf13/viper/issues/761
	github.com/spf13/viper => github.com/getporter/viper v1.7.1-porter.2.0.20210514172839-3ea827168363

)

require (
	get.porter.sh/magefiles v0.6.2
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/carolynvs/aferox v0.3.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.9.0
	github.com/cbroglie/mustache v1.4.0
	github.com/cnabio/cnab-go v0.25.1
	github.com/cnabio/cnab-to-oci v0.4.0
	github.com/cnabio/image-relocation v0.9.0
	github.com/containerd/containerd v1.7.11
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/buildx v0.11.2
	github.com/docker/cli v24.0.5+incompatible
	github.com/docker/distribution v2.8.2+incompatible
	github.com/docker/docker v24.0.7+incompatible
	github.com/dustin/go-humanize v1.0.1
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.19.0
	github.com/google/uuid v1.3.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/magefile/mage v1.15.0
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-isatty v0.0.19
	github.com/mikefarah/yq/v3 v3.0.0-20201202084205-8846255d1c37
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mmcdole/gofeed v1.2.1
	github.com/moby/buildkit v0.12.5
	github.com/moby/term v0.5.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/opencontainers/go-digest v1.0.0
	github.com/osteele/liquid v1.3.0
	github.com/pelletier/go-toml v1.9.5
	github.com/prometheus/client_golang v1.16.0
	github.com/spf13/afero v1.9.5
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.14.0
	github.com/stretchr/testify v1.8.4
	github.com/tidwall/gjson v1.15.0
	github.com/tidwall/sjson v1.2.5
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.13.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.0
	go.opentelemetry.io/otel v1.20.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.19.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.19.0
	go.opentelemetry.io/otel/trace v1.20.0
	go.uber.org/zap v1.24.0
	golang.org/x/sync v0.5.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/PaesslerAG/gval v1.2.2 // indirect
	github.com/PuerkitoBio/goquery v1.8.1 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20230117174420-439a4b8ba167 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.18.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.27 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.26 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.35 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.28 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.2 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/continuity v0.4.2 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/containerd/typeurl/v2 v2.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20230701045847-91eb5f1b7744 // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/goccy/go-yaml v1.11.0 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/in-toto/in-toto-golang v0.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jeremywohl/flatten v1.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kr/pty v1.1.5 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mholt/archiver/v3 v3.5.1 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mmcdole/goxpp v1.1.0 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.6.2 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc4 // indirect
	github.com/opencontainers/runc v1.1.12 // indirect
	github.com/opencontainers/runtime-spec v1.1.0 // indirect
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.0 // indirect
	github.com/qri-io/jsonpointer v0.1.1 // indirect
	github.com/qri-io/jsonschema v0.2.2-0.20210831022256-780655b2ba0e // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.4.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/theupdateframework/notary v0.7.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20230629203738-36ef4d8c0dbb // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/tonistiigi/vt100 v0.0.0-20230623042737-f9a4f7ef6531 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/vbatts/tar-split v0.11.3 // indirect
	github.com/weppos/publicsuffix-go v0.5.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	github.com/zmap/zcrypto v0.0.0-20190729165852-9051775e6a2e // indirect
	github.com/zmap/zlint v0.0.0-20190806154020-fd021b4cfbeb // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.44.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.20.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20231012201019-e917dd12ba7a // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.29.1 // indirect
	k8s.io/apimachinery v0.29.1 // indirect
	k8s.io/client-go v0.29.1 // indirect
	k8s.io/klog/v2 v2.110.1 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

//replace get.porter.sh/magefiles v0.3.3 => ../magefiles
