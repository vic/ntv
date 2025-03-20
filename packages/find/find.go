package find

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	nixsearch "github.com/peterldowns/nix-search-cli/pkg/nixsearch"
	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/nixhub"
	lib "github.com/vic/nix-versions/packages/versions"
)

type Opts struct {
	Exact          bool
	Constraint     string
	Limit          int
	Sort           bool
	Reverse        bool
	Lazamar        bool
	Channel        string
	ignoreFetchErr bool
}

func FindPackagesWithQuery(ctx Opts, search string) ([]string, error) {
	query := nixsearch.Query{
		MaxResults: 10,
		Channel:    "unstable",
		Search:     &nixsearch.MatchSearch{Search: search},
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
		names = append(names, pkg.AttrName)
	}
	if len(names) < 1 {
		return nil, fmt.Errorf("No packages found matching `%s`.", search)
	}
	return names, nil
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

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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
		pkgAttr = name[:strings.Index(name, "@")]
		constraint = name[strings.Index(name, "@")+1:]
		if isFile(constraint) {
			bytes, err := os.ReadFile(constraint)
			if err != nil {
				return nil, err
			}
			constraint = strings.TrimSpace(string(bytes))
		}
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
			Exact:          ctx.Exact,
			Constraint:     ctx.Constraint,
			Limit:          ctx.Limit,
			Sort:           ctx.Sort,
			Reverse:        ctx.Reverse,
			Lazamar:        ctx.Lazamar,
			Channel:        ctx.Channel,
			ignoreFetchErr: true,
		}
		return FindVersionsAll(opts, pkgs)
	}
	if strings.HasPrefix(pkgAttr, "~") {
		attrs, err := FindPackagesWithQuery(ctx, pkgAttr[1:])
		return searchAgain(attrs, err)
	}
	if strings.HasPrefix(pkgAttr, "bin/") {
		attrs, err := FindPackagesWithProgram(ctx, pkgAttr[4:])
		return searchAgain(attrs, err)
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

func FindVersionsAll(ctx Opts, names []string) ([]lib.Version, error) {
	var versions []lib.Version
	for _, name := range names {
		vers, err := FindVersions(ctx, name)
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
