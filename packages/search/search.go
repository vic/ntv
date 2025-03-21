package search

import (
	"slices"
	"strings"

	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/marshalling"
	"github.com/vic/nix-versions/packages/nixhub"
	"github.com/vic/nix-versions/packages/nixsearch"
	lib "github.com/vic/nix-versions/packages/versions"
)

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
