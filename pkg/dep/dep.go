package dep

// These imports turn transitive dependencies into direct dependencies
// so that we can control then using "constraint" in our dep manifest.
// This is important so that consumers of our porter packages automatically
// get the same vetted list of package versions that we use.
// "override" entries in the dep manifest are not used by consumers of
// porter as a library.

import (
	_ "github.com/containerd/containerd"
	_ "github.com/docker/compose-on-kubernetes"
	_ "github.com/docker/go-metrics"
	_ "k8s.io/apimachinery/pkg/version"
	_ "k8s.io/client-go/pkg/version"
)
