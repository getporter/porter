module get.porter.sh/porter

go 1.18

replace (
	// a dependency of buildx (github.com/tonistiigi/fsutil) relies on an untagged version of Docker on the master branch
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20220226190722-8667ccd1124c+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.3-0.20220121014307-40bb9831756f+incompatible

	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Also includes branches we haven't PR'd yet: capture-yamux-logs, context-cancellation
	// Tagged from v1.4.4, the porter branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.4-porter.1

	// Fixes https://github.com/spf13/viper/issues/761
	github.com/spf13/viper => github.com/getporter/viper v1.7.1-porter.2.0.20210514172839-3ea827168363
)

require (
	get.porter.sh/magefiles v0.3.0
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/carolynvs/aferox v0.3.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.8.0
	github.com/cbroglie/mustache v1.0.1
	github.com/cnabio/cnab-go v0.23.4
	github.com/cnabio/cnab-to-oci v0.3.4
	github.com/containerd/containerd v1.6.6
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/buildx v0.8.1
	github.com/docker/cli v20.10.17+incompatible
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.17+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.7
	github.com/google/go-containerregistry v0.5.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-colorable v0.1.7
	github.com/mattn/go-isatty v0.0.14
	github.com/mikefarah/yq/v3 v3.0.0-20201020025845-ccb718cd0f59
	github.com/mitchellh/mapstructure v1.4.2
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/moby/buildkit v0.10.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/go-digest v1.0.0
	github.com/osteele/liquid v1.3.0
	github.com/pelletier/go-toml v1.9.4
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.1
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.7.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.29.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.4.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.45.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/andybalholm/brotli v1.0.0 // indirect
	github.com/andybalholm/cascadia v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/continuity v0.2.2 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.11.2 // indirect
	github.com/containerd/ttrpc v1.1.0 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20220216171627-b6721b06b4ef // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/docker/go v1.5.1-1.0.20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/goccy/go-yaml v1.8.1 // indirect
	github.com/gofrs/flock v0.7.3 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4-0.20210608040537-544b4180ac70 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jeremywohl/flatten v1.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.15.1 // indirect
	github.com/klauspost/pgzip v1.2.4 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mholt/archiver/v3 v3.5.0 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mount v0.3.0 // indirect
	github.com/moby/sys/mountinfo v0.6.0 // indirect
	github.com/moby/sys/signal v0.6.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pierrec/lz4/v4 v4.0.3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/qri-io/jsonpointer v0.1.1 // indirect
	github.com/qri-io/jsonschema v0.2.2-0.20210831022256-780655b2ba0e // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20220315205639-9ed612626da3 // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/tonistiigi/vt100 v0.0.0-20210615222946-8066bb97264f // indirect
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/weppos/publicsuffix-go v0.5.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.0.2 // indirect
	github.com/xdg-go/stringprep v1.0.2 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/zmap/zcrypto v0.0.0-20190729165852-9051775e6a2e // indirect
	github.com/zmap/zlint v0.0.0-20190806154020-fd021b4cfbeb // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.29.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.29.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.4.1 // indirect
	go.opentelemetry.io/otel/internal/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/metric v0.27.0 // indirect
	go.opentelemetry.io/proto/otlp v0.12.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220314164441-57ef72a4c106 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.56.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.24.1 // indirect
	k8s.io/apimachinery v0.24.1 // indirect
	k8s.io/client-go v0.24.1 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
