//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/tests"
	"get.porter.sh/porter/tests/tester"
)

func TestSyft(t *testing.T) {
	testr, err := tester.NewTestWithConfig(
		t,
		"tests/integration/testdata/sbom/config/config-syft.yaml",
	)
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()
	imageName := "syft"
	imageTag := "v1.2.3"
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/%s:%s", reg.String(), imageName, imageTag))

	sbomFilePath := filepath.Join(testr.TestDir, "sbom.json")

	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args(
			"publish",
			"--insecure-registry",
			"-f",
			"testdata/bundles/sbom/porter.yaml",
			"-r",
			ref.String(),
			"--sbom-file",
			sbomFilePath,
		)
	})
	require.NoError(t, err, "Publish failed")

	// Confirm that publish picks up the version override
	// Use an insecure registry to validate that we can publish to one
	tests.RequireOutputContains(
		t,
		output,
		fmt.Sprintf("Bundle %s/%s:%s pushed successfully", reg, imageName, imageTag),
	)

	sbomContent, err := os.ReadFile(sbomFilePath)
	require.NoError(t, err, "error reading the sbom file %s", sbomFilePath)
	tests.RequireOutputContains(
		t,
		string(sbomContent),
		`"spdxVersion":"SPDX-`,
		"no SPDX version found in sbom content",
	)
	tests.RequireOutputContains(
		t,
		string(sbomContent),
		fmt.Sprintf(`/%s"`, imageName),
		"no image tag found in sbom content",
	)
	tests.RequireOutputContains(
		t,
		string(sbomContent),
		fmt.Sprintf(`"%s"`, imageTag),
		"no version found in the SBOM content",
	)
}
