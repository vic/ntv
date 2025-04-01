package search

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/vic/ntv/packages/backends/lazamar"
	"github.com/vic/ntv/packages/backends/nix_packages_com"
	"github.com/vic/ntv/packages/backends/nixhub"
	"github.com/vic/ntv/packages/backends/nixsearch"
	"github.com/vic/ntv/packages/nix"
	ss "github.com/vic/ntv/packages/search_spec"
	"github.com/vic/ntv/packages/versions"
	lib "github.com/vic/ntv/packages/versions"
)

type PackageSearchSpecs ss.PackageSearchSpecs
type PackageSearchSpec ss.PackageSearchSpec

type PackageSearchResult struct {
	FromSearch  *PackageSearchSpec
	Versions    []*lib.Version
	Constrained []*lib.Version
	Selected    *lib.Version
	Package     *nixsearch.Package
}

type PackageSearchResults []*PackageSearchResult

func (s *PackageSearchSpec) findNixpkgs() ([]nixsearch.Package, error) {
	var (
		pkgs []nixsearch.Package
		err  error
	)
	if strings.HasPrefix(*s.Query, "bin/") {
		program := strings.TrimPrefix(*s.Query, "bin/")
		pkgs, err = nixsearch.FindPackagesWithProgram(10, program)
		if err != nil {
			return nil, err
		}
		if len(pkgs) < 1 {
			return nil, fmt.Errorf("no packages found providing program `bin/%s`. try using `bin/*%s*`", program, program)
		}
	} else {
		pkgs, err = nixsearch.FindPackagesWithAttr(10, *s.Query)
		if err != nil {
			return nil, err
		}
		if len(pkgs) < 1 {
			return nil, fmt.Errorf("no packages found for attribute-path `%s`. try using `*%s*`", *s.Query, *s.Query)
		}
	}
	return pkgs, nil
}

func (s *PackageSearchSpec) isNotNixpkgs() bool {
	return s.VersionsBackend != nil && s.VersionsBackend.FlakeInstallable != nil
}

func (s *PackageSearchSpec) Search() ([]*PackageSearchResult, error) {
	if s.isNotNixpkgs() {
		res, err := s.searchVersions(nil)
		if err != nil {
			return nil, err
		}
		return []*PackageSearchResult{res}, nil
	}

	pkgs, err := s.findNixpkgs()
	if err != nil {
		return nil, err
	}
	group, _ := errgroup.WithContext(context.Background())
	acc := make([]*PackageSearchResult, len(pkgs))
	for i, pkg := range pkgs {
		i, pkg := i, pkg
		group.Go(func() error {
			res, err := s.searchVersions(&pkg)
			if err != nil {
				return err
			}
			acc[i] = res
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *PackageSearchSpec) searchVersions(pkg *nixsearch.Package) (*PackageSearchResult, error) {
	var (
		versions []*lib.Version
		result   *PackageSearchResult
		err      error
	)

	if s.VersionsBackend.CurrentNixpkgs != nil {
		installable := "nixpkgs#" + pkg.AttrName
		pv, err := nix.InstallablePackageVersion(installable)
		if err != nil {
			return nil, err
		}
		one := lib.Version{
			Name:      pv.PackageName,
			Version:   pv.Version,
			Attribute: pkg.AttrName,
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

	if s.VersionsBackend.NixPackagesCom != nil {
		if versions, err = nix_packages_com.Search(pkg.AttrName); err != nil {
			return nil, err
		}
	}

	if s.VersionsBackend.NixHub != nil {
		if versions, err = nixhub.Search(pkg.AttrName); err != nil {
			return nil, err
		}
	}

	if s.VersionsBackend.LazamarChannel != nil {
		if versions, err = lazamar.Search(pkg.AttrName, string(*s.VersionsBackend.LazamarChannel)); err != nil {
			return nil, err
		}
	}

	lib.SortByVersion(versions)

	result = &PackageSearchResult{
		FromSearch:  s,
		Versions:    versions,
		Constrained: []*lib.Version{},
		Package:     pkg,
	}

	var constraint = ""
	if s.VersionConstraint != nil {
		constraint = *s.VersionConstraint
	}

	result.Constrained, err = lib.ConstraintBy(versions, constraint)
	if err != nil {
		return nil, err
	}

	if len(result.Constrained) > 0 {
		result.Selected = result.Constrained[len(result.Constrained)-1]
	} else {
		result.Selected = nil
	}

	return result, nil
}

func (ss PackageSearchSpecs) Search() (PackageSearchResults, error) {
	group, _ := errgroup.WithContext(context.Background())
	results := make([][]*PackageSearchResult, len(ss))
	for i, s := range ss {
		i, s := i, (*PackageSearchSpec)(s)
		group.Go(func() error {
			res, err := s.Search()
			if err != nil {
				return err
			}
			results[i] = res
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	var res []*PackageSearchResult
	for _, r := range results {
		res = append(res, r...)
	}
	return res, nil
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

func (r PackageSearchResult) FlakeUrl(v *versions.Version) string {
	var url = v.Flake
	if len(v.Revision) > 0 {
		rev := v.Revision
		if len(rev) == 40 {
			rev = rev[:7]
		}
		url = fmt.Sprintf("%s/%s", v.Flake, rev)
	}
	return url
}

func (r PackageSearchResult) Installable(v *versions.Version) string {
	outSelectors := ""
	if r.FromSearch.OutputSelectors != nil {
		outSelectors = "^" + strings.Join(r.FromSearch.OutputSelectors, ",")
	}
	return fmt.Sprintf("%s#%s%s", r.FlakeUrl(v), v.Attribute, outSelectors)
}
