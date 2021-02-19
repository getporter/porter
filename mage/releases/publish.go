package releases

import (
	"context"
	"encoding/json"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/pkg/errors"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	ContainerName = "releases"
	mixinFeedBlob = "mixins/atom.xml"
	mixinFeedFile = "bin/mixins/atom.xml"
	VolatileCache = "max-age=300"    // 5 minutes
	StaticCache   = "max-age=604800" // 1 week
)

// Publish a mixin's binaries.
func PublishMixin(mixin string, version string, permalink string) {
	binDir := filepath.Join("bin/mixins/", mixin)
	versionDir := filepath.Join(binDir, version)
	if permalink == "latest" {
		must.RunV("az", "storage", "blob", "upload-batch", "-d", path.Join(ContainerName, "mixins", mixin, version), "-s", versionDir, "--content-cache-control", StaticCache)
	}
	must.RunV("az", "storage", "blob", "upload-batch", "-d", path.Join(ContainerName, "mixins", mixin, permalink), "-s", versionDir, "--content-cache-control", VolatileCache)
}

// Generate an updated mixin feed and releases it.
func PublishMixinFeed(ctx context.Context) {
	leaseId, err := lockMixinFeed(ctx)
	mgx.Must(err)
	defer shx.RunE("az", "storage", "blob", "lease", "release", "-c", ContainerName, "-b", mixinFeedBlob, "--lease-id", leaseId)

	must.RunE("az", "storage", "blob", "download", "-c", ContainerName, "-n", mixinFeedBlob, "-f", mixinFeedFile, "--lease-id", leaseId)
	must.RunV("bin/porter", "mixins", "feed", "generate", "-d", filepath.Dir(mixinFeedFile), "-f", mixinFeedFile, "-t", "build/atom-template.xml")
	must.RunV("az", "storage", "blob", "upload", "-c", ContainerName, "-n", mixinFeedBlob, "-f", mixinFeedFile, "--content-cache-control", VolatileCache, "--lease-id", leaseId)
}

// Tries to get a lock on the mixin feed in blob storage, returning the lease id
func lockMixinFeed(ctx context.Context) (string, error) {
	var leaseJson string
	var err error

	timeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	for {
		select {
		case <-timeout.Done():
			return "", errors.New("timeout while trying to acquire lease on the mixin feed")
		default:
			leaseJson, err = shx.OutputE("az", "storage", "blob", "lease", "acquire", "-c", ContainerName, "-b", mixinFeedBlob, "--lease-duration", "60", "-o=json")
			if err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "there is currently a lease on the blob") {
					time.Sleep(2 * time.Second)
					continue
				}
				mgx.Must(err)
			}

			var leaseId string
			err = json.Unmarshal([]byte(leaseJson), &leaseId)
			return leaseId, errors.Wrapf(err, "error parsing lease id %s as a json string", leaseJson)
		}
	}
}
