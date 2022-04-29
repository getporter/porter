package setup

import (
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
)

func EnsureProtobufTools() {
	mgx.Must(pkg.EnsurePackage("google.golang.org/protobuf/cmd/protoc-gen-go", "v1.28", "--version"))
	mgx.Must(pkg.EnsurePackage("google.golang.org/grpc/cmd/protoc-gen-go-grpc", "v1.2", "--version"))
}
