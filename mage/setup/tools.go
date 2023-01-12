package setup

import (
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
)

func EnsureProtobufTools() {
	mgx.Must(pkg.EnsurePackageWith(pkg.EnsurePackageOptions{
		Name:           "google.golang.org/protobuf/cmd/protoc-gen-go",
		DefaultVersion: "v1.28",
		VersionCommand: "--version",
	}))
	mgx.Must(pkg.EnsurePackageWith(pkg.EnsurePackageOptions{
		Name:           "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
		DefaultVersion: "v1.2",
		VersionCommand: "--version",
	}))
}
