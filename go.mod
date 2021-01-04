module get.porter.sh/porter

go 1.13

replace (
	// jsonschema lock
	github.com/cnabio/cnab-go => github.com/carolynvs/cnab-go v0.13.4-0.20201230032116-229dd4b057af

	// See https://github.com/containerd/containerd/issues/3031
	// When I try to just use the require, go is shortening it to v2.7.1+incompatible which then fails to build...
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	// See https://github.com/hashicorp/go-plugin/pull/127 and
	// https://github.com/hashicorp/go-plugin/pull/163
	// Tagged from v1.4.0, the improved-configuration branch
	github.com/hashicorp/go-plugin => github.com/getporter/go-plugin v1.4.0-improved-configuration

	// Fork (fluent branch) that adds fluent syntax and supports running a
	// command in a directory without using chdir
	github.com/magefile/mage => github.com/carolynvs/mage v1.10.1-0.20201116013517-68243214dee0

	// jsonschema lock
	github.com/qri-io/jsonschema => github.com/carolynvs/jsonschema v0.2.1-0.20201229145510-cc593f443fdb

	// pr239-fix-memmap-rename-dir
	// https://github.com/spf13/afero/pull/239 + fixes
	github.com/spf13/afero => github.com/getporter/afero v1.2.3-0.20210106151829-9adb084dc832

	golang.org/x/sys => golang.org/x/sys v0.0.0-20190830141801-acfa387b8d69
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/carolynvs/aferox v0.2.1
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/carolynvs/magex v0.3.1-0.20201231144157-18bcbf7fb1fa
	github.com/cbroglie/mustache v1.0.1
	github.com/cnabio/cnab-go v0.15.0
	github.com/cnabio/cnab-to-oci v0.3.1-beta1
	github.com/containerd/cgroups v0.0.0-20200710171044-318312a37340 // indirect
	github.com/containerd/containerd v1.3.0
	github.com/containerd/continuity v0.0.0-20200228182428-0f16d7a0959c // indirect
	github.com/containerd/fifo v0.0.0-20191213151349-ff969a566b00 // indirect
	github.com/containerd/ttrpc v1.0.0 // indirect
	github.com/containerd/typeurl v1.0.0 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.4.2-0.20181229214054-f76d6a078d88
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gogo/googleapis v1.3.2 // indirect
	github.com/google/go-containerregistry v0.0.0-20191015185424-71da34e4d9b3
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hashicorp/go-hclog v0.14.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/magefile/mage v1.10.0
	github.com/mikefarah/yq/v3 v3.0.0-20201020025845-ccb718cd0f59
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/op/go-logging.v1 v1.0.0-20160211212156-b2cb9fa56473
	gopkg.in/yaml.v2 v2.2.4
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
)
