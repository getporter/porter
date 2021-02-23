package releases

import (
	"context"
	"encoding/json"
	"log"
	"os"
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
	ContainerName = "porter"
	mixinFeedBlob = "mixins/atom.xml"
	mixinFeedFile = "bin/mixins/atom.xml"
	VolatileCache = "max-age=300"    // 5 minutes
	StaticCache   = "max-age=604800" // 1 week
)

// Prepares bin directory for publishing
func PrepareMixinForPublish(mixin string, version string, permalink string) {
	// Prepare the bin directory for generating a mixin feed
	// We want the bin to contain either a version directory (v1.2.3) or a canary directory.
	// We do not want a latest directory, latest entries are calculated using the most recent
	// timestamp in the atom.xml, not from an explicit entry.
	if permalink == "latest" {
		return
	}

	binDir := filepath.Join("bin/mixins/", mixin)
	// Temp hack until we have mixin.mk totally moved into mage
	if mixin == "porter" {
		binDir = "bin"
	}
	versionDir := filepath.Join(binDir, version)
	permalinkDir := filepath.Join(binDir, permalink)

	mgx.Must(os.RemoveAll(permalinkDir))
	log.Printf("mv %s %s\n", versionDir, permalinkDir)
	mgx.Must(os.Rename(versionDir, permalinkDir))
}

// Publish a mixin's binaries.
func PublishMixin(mixin string, version string, permalink string) {
	var publishDir string
	if permalink == "canary" {
		publishDir = filepath.Join("bin/mixins/", mixin, permalink)
	} else {
		publishDir = filepath.Join("bin/mixins/", mixin, version)
	}

	if permalink == "latest" {
		must.RunV("az", "storage", "blob", "upload-batch", "-d", path.Join(ContainerName, "mixins", mixin, version), "-s", publishDir, "--content-cache-control", StaticCache)
	}
	must.RunV("az", "storage", "blob", "upload-batch", "-d", path.Join(ContainerName, "mixins", mixin, permalink), "-s", publishDir, "--content-cache-control", VolatileCache)
}

// Generate an updated mixin feed and publishes it.
func PublishMixinFeed(ctx context.Context) {
	leaseId, unlock, err := lockMixinFeed(ctx)
	defer unlock()
	mgx.Must(err)

	must.RunE("az", "storage", "blob", "download", "-c", ContainerName, "-n", mixinFeedBlob, "-f", mixinFeedFile, "--lease-id", leaseId)
	GenerateMixinFeed()
	must.RunV("az", "storage", "blob", "upload", "-c", ContainerName, "-n", mixinFeedBlob, "-f", mixinFeedFile, "--content-cache-control", VolatileCache, "--lease-id", leaseId)
}

// Generate a mixin feed from any mixin versions in bin.
func GenerateMixinFeed() {
	must.RunV("bin/porter", "mixins", "feed", "generate", "-d", filepath.Dir(mixinFeedFile), "-f", mixinFeedFile, "-t", "build/atom-template.xml")
}

// Tries to get a lock on the mixin feed in blob storage, returning the lease id
func lockMixinFeed(ctx context.Context) (string, func(), error) {
	var leaseJson string
	var err error

	timeout, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	for {
		select {
		case <-timeout.Done():
			return "", func() {}, errors.New("timeout while trying to acquire lease on the mixin feed")
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
			unlock := func() {
				shx.RunE("az", "storage", "blob", "lease", "release", "-c", ContainerName, "-b", mixinFeedBlob, "--lease-id", leaseId)
			}
			return leaseId, unlock, errors.Wrapf(err, "error parsing lease id %s as a json string", leaseJson)
		}
	}
}
