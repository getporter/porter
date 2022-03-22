module get.porter.sh/porter

// I am not bumping v0.38 beyond compatibility with 1.16 because it causes go mod tidy to fail with
// go.opentelemetry.io/otel/semconv: module go.opentelemetry.io/otel@latest found (v1.4.1), but does not contain package go.opentelemetry.io/otel/semconv
// It's okay to _build with_ Go 1.17+ but it needs to be in the older compatibility mode
go 1.16

replace (
	// This points to a tag off of the porter-stable branch, since cnab-go main has diverged to support Porter v1.
	// The tagging scheme is LATEST_TAG_FROM_CNABGO-porter.N where N allows for us to make multiple
	// tags based on the same underlying version of cnab-go.
	github.com/cnabio/cnab-go => github.com/getporter/cnab-go v0.19.0-porter.5

	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Also includes a branch we haven't PR'd yet: capture-yamux-logs
	// Tagged from v1.4.0, the improved-configuration branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.0-improved-configuration.1

	// local-keyword-registry
	github.com/qri-io/jsonschema => github.com/carolynvs/jsonschema v0.2.1-0.20210602145235-283986347fba
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/carolynvs/aferox v0.3.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.6.0
	github.com/cbroglie/mustache v1.0.1
	github.com/cnabio/cnab-go v0.23.1
	github.com/cnabio/cnab-to-oci v0.3.3
	github.com/containerd/containerd v1.6.1
	github.com/docker/cli v20.10.13+incompatible
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.13+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/packr/v2 v2.8.3
	github.com/google/go-containerregistry v0.5.1
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.0
	github.com/magefile/mage v1.11.0
	github.com/mikefarah/yq/v3 v3.0.0-20201020025845-ccb718cd0f59
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.8.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/moby/buildkit v0.10.0 // indirect
)
