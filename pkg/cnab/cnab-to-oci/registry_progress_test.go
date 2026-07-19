package cnabtooci

import (
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/cnabio/cnab-to-oci/remotes"
	ocischemav1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func desc(size int64, done bool, children ...remotes.DescriptorProgressSnapshot) remotes.DescriptorProgressSnapshot {
	return remotes.DescriptorProgressSnapshot{
		Descriptor: ocischemav1.Descriptor{Size: size},
		Done:       done,
		Children:   children,
	}
}

func TestFlattenProgress(t *testing.T) {
	snapshot := remotes.ProgressSnapshot{
		Roots: []remotes.DescriptorProgressSnapshot{
			desc(10, true, desc(3, true), desc(4, false)),
			desc(20, false),
		},
	}

	flat := flattenProgress(snapshot)
	require.Len(t, flat, 4)
}

func TestFormatProgressLine(t *testing.T) {
	testcases := []struct {
		name         string
		snapshot     remotes.ProgressSnapshot
		expectedLine string
		expectedDone int
		expectedTot  int
	}{
		{
			name:         "empty",
			snapshot:     remotes.ProgressSnapshot{},
			expectedLine: "  layers: 0/0 copied (0 B/0 B)",
		},
		{
			name: "partial, nested",
			snapshot: remotes.ProgressSnapshot{
				Roots: []remotes.DescriptorProgressSnapshot{
					desc(10, true, desc(5, true), desc(5, false)),
				},
			},
			expectedLine: "  layers: 2/3 copied (15 B/20 B)",
			expectedDone: 2,
			expectedTot:  3,
		},
		{
			name: "all done",
			snapshot: remotes.ProgressSnapshot{
				Roots: []remotes.DescriptorProgressSnapshot{
					desc(1000, true),
					desc(2000, true),
				},
			},
			expectedLine: "  layers: 2/2 copied (3.0 kB/3.0 kB)",
			expectedDone: 2,
			expectedTot:  2,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			line, done, total := formatProgressLine(tt.snapshot)
			assert.Equal(t, tt.expectedLine, line)
			assert.Equal(t, tt.expectedDone, done)
			assert.Equal(t, tt.expectedTot, total)
		})
	}
}

func TestDisplayEvent_NonTerminal(t *testing.T) {
	tc := portercontext.NewTestContext(t)
	r := NewRegistry(tc.Context)

	r.displayEvent(remotes.FixupEvent{
		EventType:   remotes.FixupEventTypeCopyImageStart,
		SourceImage: "test/image:latest",
	})

	// A progress event with no change in done-count should not print anything.
	snapshot := remotes.ProgressSnapshot{
		Roots: []remotes.DescriptorProgressSnapshot{desc(10, false)},
	}
	r.displayEvent(remotes.FixupEvent{EventType: remotes.FixupEventTypeProgress, Progress: snapshot})
	r.displayEvent(remotes.FixupEvent{EventType: remotes.FixupEventTypeProgress, Progress: snapshot})

	// Progress that advances the done-count should print exactly one new line.
	snapshot.Roots[0].Done = true
	r.displayEvent(remotes.FixupEvent{EventType: remotes.FixupEventTypeProgress, Progress: snapshot})

	r.displayEvent(remotes.FixupEvent{
		EventType:   remotes.FixupEventTypeCopyImageEnd,
		SourceImage: "test/image:latest",
	})

	output := tc.GetOutput()
	assert.Contains(t, output, "Starting to copy image test/image:latest...")
	assert.Contains(t, output, "layers: 1/1 copied (10 B/10 B)")
	assert.Contains(t, output, "Completed image test/image:latest copy")
	// Only one progress line should have been printed, since the first two
	// progress events reported no change in the done-count.
	assert.Equal(t, 1, strings.Count(output, "layers:"))
}
