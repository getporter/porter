//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/tests/tester"
	"github.com/carolynvs/magex/shx"
	"github.com/docker/distribution/reference"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestCosign(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-cosign.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))

	setupCosign(t, testr)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--sign-bundle", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
		pc.Env("COSIGN_PASSWORD='test'")
	})
	require.NoError(t, err, "Publish failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := resolveInvocationImageDigest(t, output, "sign")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", ref.String(), "--insecure-registry", "--force")
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", ref.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func TestCosignFromArchive(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-cosign.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))

	setupCosign(t, testr)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
		pc.Env("COSIGN_PASSWORD='test'")
	})
	require.NoError(t, err, "Publish failed")

	tmpDir, err := os.MkdirTemp("", "signBundle")
	require.NoError(t, err, "Error creating temporary directory")
	defer func() {
		os.RemoveAll(tmpDir)
	}()
	archivePath := filepath.Join(tmpDir, "signBundle.tgz")
	_, output = testr.RequirePorter("archive", archivePath, "--insecure-registry", "--reference", ref.String())
	ref, err = cnab.ParseOCIReference(fmt.Sprintf("%s/sign-from-archive:v1.0.0", reg.String()))
	require.NoError(t, err, "error parsing OCI reference")

	_, output, err = testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--sign-bundle", "--insecure-registry", "--archive", archivePath, "-r", ref.String())
		pc.Env("COSIGN_PASSWORD='test'")
	})
	fmt.Println(output)
	require.NoError(t, err, "Publish archive failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := getInvocationImageDigest(t, output, "sign-from-archive")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", ref.String(), "--insecure-registry", "--force")
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", ref.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func TestCosignCopyBundle(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-cosign.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))
	secondReg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: true})
	defer secondReg.Close()
	copiedRef := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", secondReg.String()))

	setupCosign(t, testr)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
		pc.Env("COSIGN_PASSWORD='test'")
	})
	require.NoError(t, err, "Publish failed")

	_, output, err = testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("copy", "--insecure-registry", "--sign-bundle", "--source", ref.String(), "--destination", copiedRef.String())
		pc.Env("COSIGN_PASSWORD='test'")
	})
	fmt.Println(output)
	require.NoError(t, err, "Copy failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := getInvocationImageDigest(t, output, "sign")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", copiedRef.String(), "--insecure-registry", "--force")
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", copiedRef.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func setupCosign(t *testing.T, testr tester.Tester) {
	cmd := shx.Command("cosign", "generate-key-pair").Env("COSIGN_PASSWORD='test'").In(testr.PorterHomeDir)
	err := cmd.RunE()
	require.NoError(t, err, "Generate cosign key pair failed")
}

func TestNotation(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-notation.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))

	setupNotation(t, testr)
	defer cleanupNotation(t)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--sign-bundle", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
	})
	require.NoError(t, err, "Publish failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := resolveInvocationImageDigest(t, output, "sign")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", ref.String(), "--insecure-registry", "--force")
	fmt.Println(output)
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", ref.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func TestNotationFromArchive(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-notation.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))

	setupNotation(t, testr)
	defer cleanupNotation(t)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
	})
	require.NoError(t, err, "Publish failed")

	tmpDir, err := os.MkdirTemp("", "signBundle")
	require.NoError(t, err, "Error creating temporary directory")
	defer func() {
		os.RemoveAll(tmpDir)
	}()
	archivePath := filepath.Join(tmpDir, "signBundle.tgz")
	_, output = testr.RequirePorter("archive", archivePath, "--insecure-registry", "--reference", ref.String())
	ref, err = cnab.ParseOCIReference(fmt.Sprintf("%s/sign-from-archive:v1.0.0", reg.String()))
	require.NoError(t, err, "error parsing OCI reference")

	_, output, err = testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--sign-bundle", "--insecure-registry", "--archive", archivePath, "-r", ref.String())
	})
	fmt.Println(output)
	require.NoError(t, err, "Publish archive failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := getInvocationImageDigest(t, output, "sign-from-archive")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", ref.String(), "--insecure-registry", "--force")
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", ref.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func TestNotationCopyBundle(t *testing.T) {
	testr, err := tester.NewTestWithConfig(t, "tests/integration/testdata/signing/config/config-notation.yaml")
	require.NoError(t, err, "tester.NewTest failed")
	defer testr.Close()
	reg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer reg.Close()
	ref := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", reg.String()))
	secondReg := testr.StartTestRegistry(tester.TestRegistryOptions{UseTLS: false})
	defer secondReg.Close()
	copiedRef := cnab.MustParseOCIReference(fmt.Sprintf("%s/sign:v1.0.0", secondReg.String()))

	setupNotation(t, testr)
	defer cleanupNotation(t)
	_, output, err := testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("publish", "--insecure-registry", "-f", "testdata/bundles/signing/porter.yaml", "-r", ref.String())
	})
	require.NoError(t, err, "Publish failed")

	_, output, err = testr.RunPorterWith(func(pc *shx.PreparedCommand) {
		pc.Args("copy", "--insecure-registry", "--sign-bundle", "--source", ref.String(), "--destination", copiedRef.String())
	})
	require.NoError(t, err, "Copy failed")

	ref = toRefWithDigest(t, ref)
	invocationImageRef := getInvocationImageDigest(t, output, "sign")

	_, output = testr.RequirePorter("install", "--verify-bundle", "--reference", copiedRef.String(), "--insecure-registry", "--force")
	require.Contains(t, output, fmt.Sprintf("bundle signature verified for %s", copiedRef.String()))
	require.Contains(t, output, fmt.Sprintf("invocation image signature verified for %s", invocationImageRef.String()))
}

func setupNotation(t *testing.T, testr tester.Tester) {
	cmd := shx.Command("notation", "cert", "generate-test", "porter-test.org")
	err := cmd.RunE()
	require.NoError(t, err, "Generate notation certificate failed")
	trustPolicy := `
	{
		"version": "1.0",
		"trustPolicies": [
			{
				"name": "porter-test-images",
				"registryScopes": [ "*" ],
				"signatureVerification": {
					"level" : "strict"
				},
				"trustStores": [ "ca:porter-test.org" ],
				"trustedIdentities": [
					"*"
				]
			}
		]
	}`
	trustPolicyPath := filepath.Join(testr.PorterHomeDir, "trustpolicy.json")
	err = os.WriteFile(trustPolicyPath, []byte(trustPolicy), 0644)
	require.NoError(t, err, "Creation of trust policy failed")
	err = shx.Command("notation", "policy", "import", trustPolicyPath).RunE()
	require.NoError(t, err, "importing trust policy failed")
}

func cleanupNotation(t *testing.T) {
	output, err := shx.Command("notation", "key", "ls").Output()
	require.NoError(t, err)
	keyRegex := regexp.MustCompile(`(/.+porter-test\.org\.key)`)
	keyMatches := keyRegex.FindAllStringSubmatch(output, -1)
	require.Len(t, keyMatches, 1)
	crtRegex := regexp.MustCompile(`key\s+(/.+porter-test\.org\.crt)`)
	crtMatches := crtRegex.FindAllStringSubmatch(output, -1)
	require.Len(t, crtMatches, 1)
	err = shx.Command("notation", "key", "delete", "porter-test.org").RunV()
	require.NoError(t, err)
	err = shx.Command("notation", "cert", "delete", "--type", "ca", "--store", "porter-test.org", "porter-test.org.crt", "--yes").RunV()
	require.NoError(t, err)
	err = os.Remove(keyMatches[0][1])
	require.NoError(t, err)
	err = os.Remove(crtMatches[0][1])
	require.NoError(t, err)
}

func toRefWithDigest(t *testing.T, ref cnab.OCIReference) cnab.OCIReference {
	desc, err := crane.Head(ref.String(), crane.Insecure)
	require.NoError(t, err)
	ref.Named = reference.TrimNamed(ref.Named)
	ref, err = ref.WithDigest(digest.Digest(desc.Digest.String()))
	require.NoError(t, err)
	return ref
}

func resolveInvocationImageDigest(t *testing.T, output string, imageName string) cnab.OCIReference {
	r := regexp.MustCompile(fmt.Sprintf(`(?m:^Signing invocation image (localhost:\d+/%s:porter-[0-9a-z]+)\.)`, imageName))
	matches := r.FindAllStringSubmatch(output, -1)
	require.Len(t, matches, 1)
	invocationImageRefString := matches[0][1]
	desc, err := crane.Head(invocationImageRefString, crane.Insecure)
	require.NoError(t, err)
	ref := cnab.MustParseOCIReference(invocationImageRefString)
	ref.Named = reference.TrimNamed(ref.Named)
	ref, err = ref.WithDigest(digest.Digest(desc.Digest.String()))
	require.NoError(t, err)
	return ref
}

func getInvocationImageDigest(t *testing.T, output string, imageName string) cnab.OCIReference {
	r := regexp.MustCompile(fmt.Sprintf(`(?m:^Signing invocation image (localhost:\d+/%s@sha256:[0-9a-z]+)\.)`, imageName))
	matches := r.FindAllStringSubmatch(output, -1)
	require.Len(t, matches, 1)
	invocationImageRefString := matches[0][1]
	return cnab.MustParseOCIReference(invocationImageRefString)
}
