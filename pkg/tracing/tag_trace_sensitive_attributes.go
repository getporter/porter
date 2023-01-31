//go:build traceSensitiveAttributes

package tracing

func init() {
	// This file should only be included in the build when we run `go build -tags traceSensitiveAttributes`
	// and it helps a dev create a local build of Porter that will included sensitive data in traced attributes
	traceSensitiveAttributes = true
}
