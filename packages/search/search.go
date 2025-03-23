package search

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/marshalling"
	"github.com/vic/nix-versions/packages/nix"
	"github.com/vic/nix-versions/packages/nixhub"
	"github.com/vic/nix-versions/packages/nixsearch"
	ss "github.com/vic/nix-versions/packages/search_spec"
	lib "github.com/vic/nix-versions/packages/versions"
)

type PackageSearchSpecs ss.PackageSearchSpecs
type PackageSearchSpec ss.PackageSearchSpec

type PackageSearchResult struct {
	FromSearch *PackageSearchSpec
	Versions   *[]lib.Version
	Selected   *lib.Version
}

type PackageSearchResults []*PackageSearchResult

func (s *PackageSearchSpec) Search() (*PackageSearchResult, error) {
	var (
		result   *PackageSearchResult
		versions []lib.Version
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
		result := PackageSearchResult{
			FromSearch: s,
			Versions:   &[]lib.Version{one},
		}
		return &result, nil
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
		versions = []lib.Version{{
			Name:      pv.PackageName,
			Version:   pv.Version,
			Attribute: attribute,
			Flake:     flake,
			Revision:  "",
		}}
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

	if s.VersionConstraint != nil {
		versions, err = lib.ConstraintBy(versions, *s.VersionConstraint)
		if err != nil {
			return nil, err
		}
	}

	lib.SortByVersion(versions)

	result = &PackageSearchResult{
		FromSearch: s,
		Versions:   &versions,
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

func (r PackageSearchResults) SelectLatest() error {
	for _, result := range r {
		if len(*result.Versions) == 0 {
			return fmt.Errorf("no versions found for %s", *result.FromSearch.Spec)
		}
		result.Selected = &(*result.Versions)[len(*result.Versions)-1]
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

// ---- WILL REMOVE FROM HERE ----

type Opts struct {
	One            bool
	Exact          bool
	Constraint     string
	Limit          int
	Sort           bool
	Reverse        bool
	Lazamar        bool
	Channel        string
	ignoreFetchErr bool
}

func findVersions(ctx Opts, name string) ([]lib.Version, error) {
	if marshalling.IsInstallable(name) {
		return []lib.Version{marshalling.FromInstallableStr(name)}, nil
	}

	var (
		err              error
		versions         []lib.Version
		constraint       = ctx.Constraint
		limit            = ctx.Limit
		sort             = ctx.Sort
		reverse          = ctx.Reverse
		pkgAttr          = name
		constraintInName = strings.Contains(name, "@")
	)
	if constraintInName {
		pkgAttr = name[:strings.Index(name, "@")]
		constraint = name[strings.Index(name, "@")+1:]
	}
	searchAgain := func(attrs []string, err error) ([]lib.Version, error) {
		if err != nil {
			return nil, err
		}
		var pkgs []string
		if constraintInName {
			for _, attr := range attrs {
				pkgs = append(pkgs, attr+"@"+constraint)
			}
		} else {
			pkgs = attrs
		}
		opts := Opts{
			One:            ctx.One,
			Exact:          ctx.Exact,
			Constraint:     ctx.Constraint,
			Limit:          ctx.Limit,
			Sort:           ctx.Sort,
			Reverse:        ctx.Reverse,
			Lazamar:        ctx.Lazamar,
			Channel:        ctx.Channel,
			ignoreFetchErr: true,
		}
		return findVersionsAll(opts, pkgs)
	}
	if strings.HasPrefix(pkgAttr, "~") {
		attrs, err := nixsearch.FindPackagesWithQuery(ctx.Limit, pkgAttr[1:])
		return searchAgain(attrs, err)
	}
	if strings.HasPrefix(pkgAttr, "bin/") {
		attrs, err := nixsearch.FindPackagesWithProgram(ctx.Limit, ctx.Exact, pkgAttr[4:])
		return searchAgain(attrs, err)
	}
	if strings.Contains(constraint, "latest") {
		constraint = strings.Replace(constraint, "latest", "", 1)
		limit = 1
		sort = true
		reverse = false
	}
	if ctx.Lazamar {
		versions, err = lazamar.Search(pkgAttr, ctx.Channel)
	} else {
		versions, err = nixhub.Search(pkgAttr)
	}
	if err != nil {
		if ctx.ignoreFetchErr {
			return versions, nil
		}
		return nil, err
	}
	if ctx.Exact {
		versions = lib.Exact(versions, pkgAttr)
	}
	if constraint != "" {
		versions, err = lib.ConstraintBy(versions, constraint)
		if err != nil {
			return nil, err
		}
	}
	if sort {
		lib.SortByVersion(versions)
	}
	if reverse {
		slices.Reverse(versions)
	}
	if limit != 0 {
		versions = lib.Limit(versions, limit)
	}
	return versions, nil
}

func findVersionsAll(ctx Opts, names []string) ([]lib.Version, error) {
	var versions []lib.Version

	for _, name := range names {
		vers, err := findVersions(ctx, name)
		if err != nil {
			return nil, err
		}
		versions = append(versions, vers...)
	}

	if ctx.Sort {
		lib.SortByVersion(versions)
	}
	if ctx.Reverse {
		slices.Reverse(versions)
	}
	if ctx.Limit != 0 {
		versions = lib.Limit(versions, ctx.Limit)
	}

	return versions, nil
}

func FindReadingVersionsAll(ctx Opts, args []string) ([]lib.Version, error) {
	var names []string
	for _, arg := range args {
		namex, err := marshalling.ReadPackagesFromFile(arg)
		if err != nil {
			return nil, err
		}
		names = append(names, namex...)
	}
	return findVersionsAll(ctx, names)
}
