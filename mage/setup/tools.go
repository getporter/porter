package setup

import (
	"os/exec"
	"runtime"

	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/pkg/archive"
	"github.com/carolynvs/magex/pkg/downloads"
	"github.com/magefile/mage/mg"
)

func EnsureProtobufTools() {
	//TODO: add more tools
	// https://github.com/bufbuild/buf/releases protoc-gen-buf-breaking, protoc-gen-buf-lint
	// protoc https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-x86_64.zip``
	//TODO: add more tools
	// https://github.com/bufbuild/buf/releases protoc-gen-buf-breaking, protoc-gen-buf-lint
	// protoc https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-x86_64.zip``
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

// IsCommandInPath determines if a command can be called based on the current PATH.
func IsCommandInPath(cmd string) (bool, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func EnsureBufBuild() {
	mg.Deps(EnsureProtobufTools)
	if ok, _ := pkg.IsCommandAvailable("buf", "--version", "1.11.0"); ok {
		return
	}

	target := "buf-{{.GOOS}}-{{.GOARCH}}{{.EXT}}"
	if runtime.GOOS == "windows" {
		target = "buf-{{.GOOS}}-{{.GOARCH}}.exe"
	}

	opts := archive.DownloadArchiveOptions{
		DownloadOptions: downloads.DownloadOptions{
			UrlTemplate: "https://github.com/bufbuild/buf/releases/download/v{{.VERSION}}/buf-{{.GOOS}}-{{.GOARCH}}{{.EXT}}",
			Name:        "buf",
			Version:     "1.11.0",
			OsReplacement: map[string]string{
				"darwin":  "Darwin",
				"linux":   "Linux",
				"windows": "Windows",
			},
			ArchReplacement: map[string]string{
				"amd64": "x86_64",
			},
			// empty hook to override default archive extraction
			Hook: func(archivePath string) (string, error) { return archivePath, nil },
		},
		ArchiveExtensions: map[string]string{
			"windows": ".exe",
		},
		TargetFileTemplate: target,
	}

	err := archive.DownloadToGopathBin(opts)
	mgx.Must(err)
}

func EnsureGRPCurl() {
	if ok, _ := IsCommandInPath("grpcurl"); ok {
		return
	}

	target := "grpcurl{{.EXT}}"
	if runtime.GOOS == "windows" {
		target = "grpcurl.exe"
	}

	opts := archive.DownloadArchiveOptions{
		DownloadOptions: downloads.DownloadOptions{
			UrlTemplate: "https://github.com/fullstorydev/grpcurl/releases/download/v{{.VERSION}}/grpcurl_{{.VERSION}}_{{.GOOS}}_{{.GOARCH}}{{.EXT}}",
			Name:        "grpcurl",
			Version:     "1.8.7",
			OsReplacement: map[string]string{
				"darwin": "osx",
			},
			ArchReplacement: map[string]string{
				"amd64": "x86_64",
			},
		},
		ArchiveExtensions: map[string]string{
			"linux":   ".tar.gz",
			"darwin":  ".tar.gz",
			"windows": ".zip",
		},
		TargetFileTemplate: target,
	}
	err := archive.DownloadToGopathBin(opts)
	mgx.Must(err)
}
