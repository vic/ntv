package find

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/nixhub"
	lib "github.com/vic/nix-versions/packages/versions"
)

type Opts struct {
	Exact      bool
	Constraint string
	Limit      int
	Sort       bool
	Reverse    bool
	Lazamar    bool
	Channel    string
}

func FindPackagesWithProgram(ctx Opts, program string) ([]string, error) {
	query := nixsearch.Query{
		MaxResults: 10,
		Channel:    "unstable",
		Program:    &nixsearch.MatchProgram{Program: program},
	}
	client, err := nixsearch.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	packages, err := client.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, pkg := range packages {
		if ctx.Exact {
			if slices.Contains(pkg.Programs, program) {
				names = append(names, pkg.AttrName)
			}
		} else {
			names = append(names, pkg.AttrName)
		}
	}
	if len(names) < 1 {
		return nil, fmt.Errorf("No packages found providing program `bin/%s`.\nTry using `--exact=false` option to match on any part of the program name.", program)
	}
	return names, nil
}

func FindVersions(ctx Opts, name string) ([]lib.Version, error) {
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
		constraint = name[strings.Index(name, "@")+1:]
		pkgAttr = name[:strings.Index(name, "@")]
	}
	if strings.HasPrefix(pkgAttr, "bin/") {
		attrs, err := FindPackagesWithProgram(ctx, pkgAttr[4:])
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
		return FindVersionsAll(ctx, pkgs)
	}
	if strings.Contains(constraint, "latest") {
		constraint = strings.Replace(constraint, "latest", "", 1)
		limit = 1
		sort = true
		reverse = false
	}
	if ctx.Lazamar {
		versions, err = lazamar.Versions(pkgAttr, ctx.Channel)
	} else {
		versions, err = nixhub.Versions(pkgAttr)
	}
	if err != nil {
		return nil, err
	}
	if ctx.Exact {
		versions = lib.Exact(versions, pkgAttr)
	}
	if constraint != "" {
		versions, err = lib.ConstraintBy(versions, constraint)
	}
	if sort {
		lib.SortByVersion(versions)
	}
	if reverse {
		slices.Reverse(versions)
	}
	if err != nil {
		return nil, err
	}
	if limit != 0 {
		versions = lib.Limit(versions, limit)
	}
	return versions, nil
}

func FindVersionsAll(ctx Opts, names []string) ([]lib.Version, error) {
	var versions []lib.Version
	for _, name := range names {
		vers, err := FindVersions(ctx, name)
		if err != nil {
			return nil, err
		}
		versions = append(versions, vers...)
	}
	return versions, nil
}
