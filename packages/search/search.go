package search

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/vic/ntv/packages/lazamar"
	"github.com/vic/ntv/packages/nix"
	"github.com/vic/ntv/packages/nixhub"
	ss "github.com/vic/ntv/packages/search_spec"
	lib "github.com/vic/ntv/packages/versions"
)

type PackageSearchSpecs ss.PackageSearchSpecs
type PackageSearchSpec ss.PackageSearchSpec

type PackageSearchResult struct {
	FromSearch  *PackageSearchSpec
	Versions    []*lib.Version
	Constrained []*lib.Version
	Selected    *lib.Version
}

type PackageSearchResults []*PackageSearchResult

func (s *PackageSearchSpec) Search() (*PackageSearchResult, error) {
	var (
		result   *PackageSearchResult
		versions []*lib.Version
		err      error
	)

	if s.VersionsBackend.CurrentNixpkgs != nil {
		installable := "nixpkgs#" + *s.Query
		pv, err := nix.InstallablePackageVersion(installable)
		if err != nil {
			return nil, err
		}
		one := lib.Version{
			Name:      pv.PackageName,
			Version:   pv.Version,
			Attribute: *s.Query,
			Flake:     "nixpkgs",
			Revision:  "",
		}
		versions = []*lib.Version{&one}
	}

	if s.VersionsBackend.FlakeInstallable != nil {
		installable := string(*s.VersionsBackend.FlakeInstallable)
		pv, err := nix.InstallablePackageVersion(installable)
		if err != nil {
			return nil, err
		}
		var (
			attribute = "default"
			flake     = installable
		)
		if strings.Contains(installable, "#") {
			idx := strings.LastIndex(installable, "#")
			flake = installable[:idx]
			attribute = installable[idx+1:]
		}
		one := lib.Version{
			Name:      pv.PackageName,
			Version:   pv.Version,
			Attribute: attribute,
			Flake:     flake,
			Revision:  "",
		}
		versions = []*lib.Version{&one}
	}

	if s.VersionsBackend.NixHub != nil {
		if versions, err = nixhub.Search(*s.Query); err != nil {
			return nil, err
		}
	}

	if s.VersionsBackend.LazamarChannel != nil {
		if versions, err = lazamar.Search(*s.Query, string(*s.VersionsBackend.LazamarChannel)); err != nil {
			return nil, err
		}
	}

	lib.SortByVersion(versions)

	result = &PackageSearchResult{
		FromSearch:  s,
		Versions:    versions,
		Constrained: []*lib.Version{},
	}

	if s.VersionConstraint != nil {
		result.Constrained, err = lib.ConstraintBy(versions, *s.VersionConstraint)
		if err != nil {
			return nil, err
		}
	}

	if len(result.Constrained) > 0 {
		result.Selected = result.Constrained[len(result.Constrained)-1]
	} else if len(versions) > 0 {
		result.Selected = versions[len(versions)-1]
	}

	return result, nil
}

func (ss PackageSearchSpecs) Search() (PackageSearchResults, error) {
	group, _ := errgroup.WithContext(context.Background())
	results := make([]*PackageSearchResult, len(ss))
	for i, s := range ss {
		i, s := i, (*PackageSearchSpec)(s)
		group.Go(func() error {
			result, err := s.Search()
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r PackageSearchResults) EnsureOneSelected() error {
	for _, result := range r {
		if result.Selected == nil {
			return fmt.Errorf("no versions found for %s", *result.FromSearch.Spec)
		}
	}
	return nil
}

func (r PackageSearchResults) EnsureUniquePackageNames() error {
	var seen = make(map[string]int)
	var failed = false
	for _, result := range r {
		seen[result.Selected.Name]++
		if seen[result.Selected.Name] > 1 {
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("expected at most one version per package, but got %v - try using @latest or a more specific version constraint", seen)
	}
	return nil
}

func (r PackageSearchResults) Size() int {
	var size = 0
	for _, result := range r {
		size += len(result.Versions)
	}
	return size
}

func (r PackageSearchResult) FlakeUrl() string {
	var url = r.Selected.Flake
	if len(r.Selected.Revision) > 0 {
		url = fmt.Sprintf("%s/%s", r.Selected.Flake, r.Selected.Revision)
	}
	return url
}

func (r PackageSearchResult) Installable() string {
	return fmt.Sprintf("%s#%s", r.FlakeUrl(), r.Selected.Attribute)
}
