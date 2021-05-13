module get.porter.sh/porter

go 1.16

replace (
	// This points to a tag off of the porter branch. This branch has whatever unmerged PRs that are currently
	// outstanding. The tagging scheme is LATEST_TAG_FROM_CNABGO-porter.N where N allows for us to make multiple
	// tags based on the same underlying version of cnab-go.
	github.com/cnabio/cnab-go => github.com/getporter/cnab-go v0.19.0-porter.2

	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Also includes a branch we haven't PR'd yet: capture-yamux-logs
	// Tagged from v1.4.0, the improved-configuration branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.0-improved-configuration.1

	// go.mod doesn't propogate replacements in the dependency graph so I'm copying this from github.com/moby/buildkit
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

	// local-keyword-registry
	github.com/qri-io/jsonschema => github.com/carolynvs/jsonschema v0.2.1-0.20210120214917-11cc5e4545c8

	// Fixes https://github.com/spf13/viper/issues/761
	github.com/spf13/viper => github.com/getporter/viper v1.7.1-porter.1.0.20210512222257-5d1ed639c43c
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/carolynvs/aferox v0.3.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.6.0
	github.com/cbroglie/mustache v1.0.1
	github.com/cnabio/cnab-go v0.19.0
	github.com/cnabio/cnab-to-oci v0.3.1-beta1
	github.com/containerd/console v1.0.1
	github.com/containerd/containerd v1.5.0-beta.1
	github.com/docker/buildx v0.5.1
	github.com/docker/cli v20.10.0-beta1.0.20201029214301-1d20b15adc38+incompatible
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.6+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-containerregistry v0.1.2
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-plugin v1.4.0
	github.com/magefile/mage v1.11.0
	github.com/mikefarah/yq/v3 v3.0.0-20201020025845-ccb718cd0f59
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/moby/buildkit v0.8.1-0.20201205083753-0af7b1b9c693
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
