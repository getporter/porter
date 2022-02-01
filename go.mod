module get.porter.sh/porter

go 1.16

replace (
	// install-spec-wip
	github.com/cnabio/cnab-go => github.com/carolynvs/cnab-go v0.20.2-0.20210805155536-9a543e0636f4

	// return-copy-error
	// See https://github.com/cnabio/cnab-to-oci/pull/113
	github.com/cnabio/cnab-to-oci => github.com/getporter/cnab-to-oci v0.3.1-porter.1

	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Also includes a branch we haven't PR'd yet: capture-yamux-logs
	// Tagged from v1.4.0, the improved-configuration branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.0-improved-configuration.1

	// go.mod doesn't propogate replacements in the dependency graph so I'm copying this from github.com/moby/buildkit
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

	// expose-ast
	// https://github.com/osteele/liquid/pull/59
	github.com/osteele/liquid => github.com/carolynvs/liquid v1.2.5-0.20220131221838-2e107bef298f

	// Fixes https://github.com/spf13/viper/issues/761
	github.com/spf13/viper => github.com/getporter/viper v1.7.1-porter.2.0.20210514172839-3ea827168363
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/carolynvs/aferox v0.3.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.6.0
	github.com/cbroglie/mustache v1.0.1
	github.com/cnabio/cnab-go v0.21.0
	github.com/cnabio/cnab-to-oci v0.3.1-beta1.0.20210614060230-e4d2bd5441c8
	github.com/containerd/console v1.0.2
	github.com/containerd/containerd v1.5.3
	github.com/docker/buildx v0.5.1
	github.com/docker/cli v20.10.7+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/logger v1.0.4 // indirect
	github.com/gobuffalo/packr/v2 v2.8.1 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/go-containerregistry v0.1.2
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-colorable v0.1.7
	github.com/mattn/go-isatty v0.0.12
	github.com/mikefarah/yq/v3 v3.0.0-20201020025845-ccb718cd0f59
	github.com/mitchellh/mapstructure v1.3.3
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/moby/buildkit v0.9.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/go-digest v1.0.0
	github.com/osteele/liquid v1.2.4
	github.com/pelletier/go-toml v1.9.1
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.7.1
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.1.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	google.golang.org/grpc v1.41.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
