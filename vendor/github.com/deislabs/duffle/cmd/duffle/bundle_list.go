package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/repo"
)

// NamedRepositoryList is a list of bundle references.
// Implements a sorter on Name.
type NamedRepositoryList []*NamedRepository

// Len returns the length.
func (bl NamedRepositoryList) Len() int { return len(bl) }

// Swap swaps the position of two items in the versions slice.
func (bl NamedRepositoryList) Swap(i, j int) { bl[i], bl[j] = bl[j], bl[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (bl NamedRepositoryList) Less(a, b int) bool {
	return strings.Compare(bl[a].Name(), bl[b].Name()) < 1
}

// NamedRepository is a reference to a repository.
type NamedRepository struct {
	name   string
	tag    string
	digest string
	signed bool
}

// Name returns the full name.
func (n *NamedRepository) String() string {
	return n.name + ":" + n.tag
}

// Name returns the name.
func (n *NamedRepository) Name() string {
	return n.name
}

// Tag returns the tag.
func (n *NamedRepository) Tag() string {
	return n.tag
}

// Digest returns the digest.
func (n *NamedRepository) Digest() string {
	return n.digest
}

// IsSigned determines whether or not the bundle is signed.
func (n *NamedRepository) IsSigned() bool {
	return n.signed
}

func newBundleListCmd(w io.Writer) *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "lists bundles pulled or built and stored locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := home.Home(homePath())
			references, err := searchLocal(home)
			if err != nil {
				return err
			}
			sort.Sort(references)
			if short {
				for _, ref := range references {
					fmt.Println(ref.Name())
				}
				return nil
			}

			table := uitable.New()
			table.AddRow("NAME", "VERSION", "DIGEST", "SIGNED?")
			for _, ref := range references {
				table.AddRow(ref.Name(), ref.Tag(), ref.Digest(), ref.IsSigned())
			}
			fmt.Fprintln(w, table)

			return nil
		},
	}
	cmd.Flags().BoolVarP(&short, "short", "s", false, "output shorter listing format")

	return cmd
}

func searchLocal(home home.Home) (NamedRepositoryList, error) {
	references := NamedRepositoryList{}

	index, err := repo.LoadIndex(home.Repositories())
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %v", home.Repositories(), err)
	}

	for repo, tagList := range index {
		for tag, digest := range tagList {
			isSigned := true
			_, err := loadBundle(filepath.Join(home.Bundles(), digest), true)
			if err == ErrNotSigned {
				isSigned = false
			}
			references = append(references, &NamedRepository{
				repo,
				tag,
				digest,
				isSigned,
			})
		}
	}

	return references, nil
}
