package configadapter

import (
	"context"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var simpleManifestDigest = "707825deb7c96acca7dd5de3868bc96547ae2a73714ef85904efde9e25c7abd1"

func TestConfig_GenerateStamp(t *testing.T) {
	// Do not run this test in parallel
	// Still need to figure out what is introducing flakey-ness
	testcases := []struct {
		name         string
		preserveTags bool
	}{
		{name: "not preserving tags", preserveTags: false},
		{name: "preserving tags", preserveTags: true},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewTestConfig(t)
			c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			installedMixins := []mixin.Metadata{
				{Name: "exec", VersionInfo: pkgmgmt.VersionInfo{Version: "v1.2.3"}},
			}

			a := NewManifestConverter(c.Config, m, nil, installedMixins, tc.preserveTags)
			stamp, err := a.GenerateStamp(ctx, tc.preserveTags)
			require.NoError(t, err, "DigestManifest failed")
			assert.Equal(t, simpleManifestDigest, stamp.ManifestDigest)
			assert.Equal(t, map[string]MixinRecord{"exec": {Name: "exec", Version: "v1.2.3"}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
			assert.Equal(t, pkg.Version, stamp.Version)
			assert.Equal(t, pkg.Commit, stamp.Commit)
			assert.Equal(t, tc.preserveTags, stamp.PreserveTags)

			gotManifestContentsB, err := stamp.DecodeManifest()
			require.NoError(t, err, "DecodeManifest failed")
			wantManifestContentsB, err := c.FileSystem.ReadFile(config.Name)
			require.NoError(t, err, "could not read %s", config.Name)
			assert.Equal(t, string(wantManifestContentsB), string(gotManifestContentsB), "Stamp.EncodedManifest was not popluated and decoded properly")
		})
	}
}

func TestConfig_LoadStamp(t *testing.T) {
	t.Parallel()

	bun := cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: map[string]interface{}{
				"manifestDigest": "somedigest",
				"manifest":       "abc123",
				"mixins": map[string]interface{}{
					"exec": struct{}{},
				},
				"preserveTags": true,
			},
		},
	})

	stamp, err := LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, "somedigest", stamp.ManifestDigest)
	assert.Equal(t, map[string]MixinRecord{"exec": {}}, stamp.Mixins, "Stamp.Mixins was not populated properly")
	assert.Equal(t, "abc123", stamp.EncodedManifest)
	assert.Equal(t, true, stamp.PreserveTags)
}

func TestConfig_LoadStamp_Invalid(t *testing.T) {
	t.Parallel()

	bun := cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			config.CustomPorterKey: []string{
				"somedigest",
			},
		},
	})

	stamp, err := LoadStamp(bun)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal the porter stamp")
	assert.Equal(t, Stamp{}, stamp)
}

func TestStamp_DecodeManifest(t *testing.T) {
	t.Parallel()

	t.Run("manifest populated", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		s := Stamp{
			EncodedManifest: "bmFtZTogaGVsbG8=", // name: hello
		}

		data, err := s.DecodeManifest()
		require.NoError(t, err, "DecodeManifest failed")

		m, err := manifest.UnmarshalManifest(c.TestContext.Context, data)
		require.NoError(t, err, "UnmarshalManifest failed")

		require.NotNil(t, m, "expected manifest to be populated")
		assert.Equal(t, "hello", m.Name, "expected the manifest name to be populated")
	})

	t.Run("manifest empty", func(t *testing.T) {
		t.Parallel()

		s := Stamp{}

		data, err := s.DecodeManifest()
		require.EqualError(t, err, "no Porter manifest was embedded in the bundle")

		assert.Nil(t, data, "No manifest data should be returned")
	})

	t.Run("manifest invalid", func(t *testing.T) {
		t.Parallel()

		s := Stamp{
			EncodedManifest: "name: hello", // this should be base64 encoded
		}

		data, err := s.DecodeManifest()
		require.Error(t, err, "DecodeManifest should fail for invalid data")

		assert.Contains(t, err.Error(), "could not base64 decode the manifest in the stamp")
		assert.Nil(t, data, "No manifest data should be returned")
	})

}

func TestConfig_DigestManifest(t *testing.T) {
	// Do not run in parallel, it modifies global state
	defer func() { pkg.Version = "" }()

	t.Run("updated version", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err, "could not load manifest")

		a := NewManifestConverter(c.Config, m, nil, nil, false)
		digest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")

		pkg.Version = "foo"
		defer func() { pkg.Version = "" }()
		newDigest, err := a.DigestManifest()
		require.NoError(t, err, "DigestManifest failed")
		assert.NotEqual(t, newDigest, digest, "expected the digest to be different due to the updated pkg version")
	})
}

func TestConfig_GenerateStamp_IncludeVersion(t *testing.T) {
	// Do not run this test in parallel
	// Still need to figure out what is introducing flakey-ness

	pkg.Version = "v1.2.3"
	pkg.Commit = "abc123"
	defer func() {
		pkg.Version = ""
		pkg.Commit = ""
	}()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)
	stamp, err := a.GenerateStamp(ctx, false)
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, "v1.2.3", stamp.Version)
	assert.Equal(t, "abc123", stamp.Commit)
}

func TestMixinRecord_Sort(t *testing.T) {
	records := MixinRecords{
		{Name: "helm", Version: "0.1.13"},
		{Name: "helm", Version: "v0.1.2"},
		{Name: "testmixin", Version: "1.2.3"},
		{Name: "exec", Version: "2.1.0"},
		// These won't parse as valid semver, so just sort them by the string representation instead
		{
			Name:    "az",
			Version: "invalid-version2",
		},
		{
			Name:    "az",
			Version: "invalid-version1",
		},
	}

	sort.Sort(records)

	wantRecords := MixinRecords{
		{
			Name:    "az",
			Version: "invalid-version1",
		},
		{
			Name:    "az",
			Version: "invalid-version2",
		},
		{
			Name:    "exec",
			Version: "2.1.0",
		},
		{
			Name:    "helm",
			Version: "v0.1.2",
		},
		{
			Name:    "helm",
			Version: "0.1.13",
		},
		{
			Name:    "testmixin",
			Version: "1.2.3",
		},
	}

	assert.Equal(t, wantRecords, records)
}

func TestConfig_DigestManifest_FileContentChanges(t *testing.T) {
	// Test that file content changes trigger a different digest
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Get initial digest
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Add a file to the bundle directory
	testFile := filepath.Join(c.TestContext.Getwd(), "test-script.sh")
	err = c.FileSystem.WriteFile(testFile, []byte("#!/bin/bash\necho 'hello'"), pkg.FileModeWritable)
	require.NoError(t, err, "could not write test file")

	// Get digest after adding file
	digest2, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.NotEqual(t, digest1, digest2, "expected digest to change after adding file")

	// Modify the file content
	err = c.FileSystem.WriteFile(testFile, []byte("#!/bin/bash\necho 'goodbye'"), pkg.FileModeWritable)
	require.NoError(t, err, "could not modify test file")

	// Get digest after modifying file
	digest3, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.NotEqual(t, digest2, digest3, "expected digest to change after modifying file")
	assert.NotEqual(t, digest1, digest3, "expected digest to be different from original")
}

func TestConfig_DigestManifest_FilePermissionChanges(t *testing.T) {
	// Only check executable permissions on Linux where bundles run
	if runtime.GOOS != "linux" {
		t.Skip("Skipping executable permission test on non-Linux platform")
	}

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Create a non-executable script
	testFile := filepath.Join(c.TestContext.Getwd(), "test-script.sh")
	err = c.FileSystem.WriteFile(testFile, []byte("#!/bin/bash\necho 'hello'"), pkg.FileModeWritable)
	require.NoError(t, err, "could not write test file")

	// Get digest with non-executable file
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Make the file executable
	err = c.FileSystem.Chmod(testFile, pkg.FileModeExecutable)
	require.NoError(t, err, "could not chmod test file")

	// Get digest after making file executable
	digest2, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.NotEqual(t, digest1, digest2, "expected digest to change after making file executable")

	// Make the file non-executable again
	err = c.FileSystem.Chmod(testFile, pkg.FileModeWritable)
	require.NoError(t, err, "could not chmod test file")

	// Get digest after making file non-executable
	digest3, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, digest1, digest3, "expected digest to return to original after removing executable permission")
	assert.NotEqual(t, digest2, digest3, "expected digest to differ from executable version")
}

func TestConfig_DigestManifest_FileDeletion(t *testing.T) {
	// Test that deleting files triggers a different digest
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Add multiple files
	file1 := filepath.Join(c.TestContext.Getwd(), "file1.txt")
	file2 := filepath.Join(c.TestContext.Getwd(), "file2.txt")
	err = c.FileSystem.WriteFile(file1, []byte("content 1"), pkg.FileModeWritable)
	require.NoError(t, err, "could not write file1")
	err = c.FileSystem.WriteFile(file2, []byte("content 2"), pkg.FileModeWritable)
	require.NoError(t, err, "could not write file2")

	// Get digest with both files
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Delete one file
	err = c.FileSystem.Remove(file1)
	require.NoError(t, err, "could not delete file1")

	// Get digest after deleting file1
	digest2, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.NotEqual(t, digest1, digest2, "expected digest to change after deleting file")

	// Delete the other file
	err = c.FileSystem.Remove(file2)
	require.NoError(t, err, "could not delete file2")

	// Get digest after deleting file2
	digest3, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.NotEqual(t, digest2, digest3, "expected digest to change after deleting second file")
	assert.NotEqual(t, digest1, digest3, "expected digest to be different from original")
}

func TestConfig_DigestManifest_IgnoresHiddenFiles(t *testing.T) {
	// Test that hidden files and directories are ignored
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Get initial digest
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Add a hidden file
	hiddenFile := filepath.Join(c.TestContext.Getwd(), ".hidden-file")
	err = c.FileSystem.WriteFile(hiddenFile, []byte("secret content"), pkg.FileModeWritable)
	require.NoError(t, err, "could not write hidden file")

	// Get digest after adding hidden file
	digest2, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, digest1, digest2, "expected digest to remain the same after adding hidden file")
}

func TestConfig_DigestManifest_IgnoresSkippedDirectories(t *testing.T) {
	// Test that .cnab, .git, node_modules, etc. are ignored
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Get initial digest
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Add files in directories that should be skipped
	skipDirs := []string{".cnab", ".git", "node_modules", ".porter", "vendor"}
	for _, dir := range skipDirs {
		dirPath := filepath.Join(c.TestContext.Getwd(), dir)
		err = c.FileSystem.MkdirAll(dirPath, pkg.FileModeDirectory)
		require.NoError(t, err, "could not create directory %s", dir)

		filePath := filepath.Join(dirPath, "test-file.txt")
		err = c.FileSystem.WriteFile(filePath, []byte("content in "+dir), pkg.FileModeWritable)
		require.NoError(t, err, "could not write file in %s", dir)
	}

	// Get digest after adding files in skipped directories
	digest2, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")
	assert.Equal(t, digest1, digest2, "expected digest to remain the same after adding files in skipped directories")
}

func TestConfig_DigestManifest_DeterministicOrdering(t *testing.T) {
	// Test that file order doesn't affect digest (files are sorted)
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil, false)

	// Add files in a specific order
	files := []string{"zebra.txt", "apple.txt", "middle.txt"}
	for _, file := range files {
		filePath := filepath.Join(c.TestContext.Getwd(), file)
		err = c.FileSystem.WriteFile(filePath, []byte("content"), pkg.FileModeWritable)
		require.NoError(t, err, "could not write file %s", file)
	}

	// Get digest
	digest1, err := a.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	// Create a new test config and add the same files in different order
	c2 := config.NewTestConfig(t)
	c2.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	m2, err := manifest.LoadManifestFrom(context.Background(), c2.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a2 := NewManifestConverter(c2.Config, m2, nil, nil, false)

	// Add files in reverse order
	for i := len(files) - 1; i >= 0; i-- {
		filePath := filepath.Join(c2.TestContext.Getwd(), files[i])
		err = c2.FileSystem.WriteFile(filePath, []byte("content"), pkg.FileModeWritable)
		require.NoError(t, err, "could not write file %s", files[i])
	}

	// Get digest
	digest2, err := a2.DigestManifest()
	require.NoError(t, err, "DigestManifest failed")

	assert.Equal(t, digest1, digest2, "expected digest to be the same regardless of file creation order")
}

func TestIsExecutable(t *testing.T) {
	// Only check executable permissions on Linux where bundles run
	if runtime.GOOS != "linux" {
		t.Skip("Skipping executable permission test on non-Linux platform")
	}

	c := config.NewTestConfig(t)
	testFile := filepath.Join(c.TestContext.Getwd(), "test-script.sh")

	// Create non-executable file
	err := c.FileSystem.WriteFile(testFile, []byte("#!/bin/bash\necho 'hello'"), pkg.FileModeWritable)
	require.NoError(t, err)

	info, err := c.FileSystem.Stat(testFile)
	require.NoError(t, err)
	assert.False(t, isExecutable(info), "file should not be executable")

	// Make executable
	err = c.FileSystem.Chmod(testFile, pkg.FileModeExecutable)
	require.NoError(t, err)

	info, err = c.FileSystem.Stat(testFile)
	require.NoError(t, err)
	assert.True(t, isExecutable(info), "file should be executable")
}

func TestHashBundleFiles(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err)

		a := NewManifestConverter(c.Config, m, nil, nil, false)
		hashes, err := a.hashBundleFiles(c.TestContext.Getwd())
		require.NoError(t, err)

		// Should only contain porter.yaml
		assert.Equal(t, 1, len(hashes), "expected only porter.yaml to be hashed")
		assert.Contains(t, hashes, "porter.yaml")
	})

	t.Run("with files", func(t *testing.T) {
		c := config.NewTestConfig(t)
		c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err)

		a := NewManifestConverter(c.Config, m, nil, nil, false)

		// Add some files
		err = c.FileSystem.WriteFile(filepath.Join(c.TestContext.Getwd(), "script.sh"), []byte("#!/bin/bash"), pkg.FileModeWritable)
		require.NoError(t, err)

		err = c.FileSystem.WriteFile(filepath.Join(c.TestContext.Getwd(), "README.md"), []byte("# Test"), pkg.FileModeWritable)
		require.NoError(t, err)

		hashes, err := a.hashBundleFiles(c.TestContext.Getwd())
		require.NoError(t, err)

		assert.Equal(t, 3, len(hashes), "expected 3 files to be hashed")
		assert.Contains(t, hashes, "porter.yaml")
		assert.Contains(t, hashes, "script.sh")
		assert.Contains(t, hashes, "README.md")
	})
}
