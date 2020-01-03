module get.porter.sh/porter

go 1.12

require (
	cloud.google.com/go v0.39.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/hcsshim v0.8.6
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/bugsnag/bugsnag-go v1.5.0
	github.com/bugsnag/panicwrap v1.2.0
	github.com/carolynvs/datetime-printer v0.2.0
	github.com/cbroglie/mustache v1.0.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cloudflare/cfssl v1.4.1
	github.com/cnabio/cnab-go v0.8.0-beta1
	github.com/containerd/cgroups v0.0.0-20191125132625-80b32e3c75c9
	github.com/containerd/containerd v1.3.0
	github.com/containerd/continuity v0.0.0-20181203112020-004b46473808
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c
	github.com/containerd/ttrpc v0.0.0-20191028202541-4f1b8fe65a5c
	github.com/containerd/typeurl v0.0.0-20190911142611-5eb25027c9fd
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/cnab-to-oci v0.3.0-beta3
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.4.2-0.20181229214054-f76d6a078d88
	github.com/docker/go v1.5.1-1
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gogo/googleapis v1.3.0
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-containerregistry v0.0.0-20191015185424-71da34e4d9b3
	github.com/gorilla/mux v1.7.3
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed
	github.com/hashicorp/go-hclog v0.10.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-plugin v0.0.0-00010101000000-000000000000
	github.com/hashicorp/go-version v1.1.0
	github.com/jinzhu/gorm v1.9.11
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2
	github.com/lib/pq v1.2.0
	github.com/miekg/pkcs11 v1.0.3
	github.com/mmcdole/gofeed v1.0.0-beta2
	github.com/mmcdole/goxpp v0.0.0-20181012175147-0068e33feabf // indirect
	github.com/morikuni/aec v1.0.0
	github.com/oklog/ulid v1.3.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/runc v0.1.1
	github.com/opencontainers/runtime-spec v1.0.1
	github.com/pivotal/image-relocation v0.0.0-20191111101224-e94aff6df06c
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/qri-io/jsonschema v0.1.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2
	github.com/theupdateframework/notary v0.6.1
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.0.0-20191028145041-f83a4685e152
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/dancannon/gorethink.v3 v3.0.5
	gopkg.in/fatih/pool.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191016110408-35e52d86657a
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

replace github.com/hashicorp/go-plugin => github.com/carolynvs/go-plugin v1.0.1-acceptstdin
