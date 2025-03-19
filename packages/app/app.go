package app

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/peterldowns/nix-search-cli/pkg/nixsearch"
	"github.com/urfave/cli/v2"
	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/nixhub"
	lib "github.com/vic/nix-versions/packages/versions"
)

//go:embed HELP
var AppHelpTemplate string

//go:embed VERSION
var AppVersion string

//go:embed REVISION
var AppRevision string

func App() cli.App {
	cli.AppHelpTemplate = AppHelpTemplate
	return cli.App{
		Action: mainAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "version",
			},
			&cli.BoolFlag{
				Name:  "lazamar",
				Value: true,
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("nixhub", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "channel",
				Value: "nixpkgs-unstable",
				Action: func(ctx *cli.Context, b string) error {
					ctx.Set("nixhub", "false")
					return nil
				},
			},
			&cli.BoolFlag{
				Name: "nixhub",
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("lazamar", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name: "json",
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("text", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "text",
				Value: true,
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("json", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:  "sort",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "reverse",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "exact",
				Value: true,
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: 0,
			},
			&cli.StringFlag{
				Name: "constraint",
			},
		},
	}
}

func findPackagesWithProgram(ctx *cli.Context, program string) ([]string, error) {
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
		if ctx.Bool("exact") {
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

func findVersions(ctx *cli.Context, name string) ([]lib.Version, error) {
	var (
		err              error
		versions         []lib.Version
		constraint       = ctx.String("constraint")
		limit            = ctx.Int("limit")
		sort             = ctx.Bool("sort")
		reverse          = ctx.Bool("reverse")
		pkgAttr          = name
		constraintInName = strings.Contains(name, "@")
	)
	if constraintInName {
		constraint = name[strings.Index(name, "@")+1:]
		pkgAttr = name[:strings.Index(name, "@")]
	}
	if strings.HasPrefix(pkgAttr, "bin/") {
		attrs, err := findPackagesWithProgram(ctx, pkgAttr[4:])
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
		return findVersionsAll(ctx, pkgs)
	}
	if strings.Contains(constraint, "latest") {
		constraint = strings.Replace(constraint, "latest", "", 1)
		limit = 1
		sort = true
		reverse = false
	}
	if ctx.Bool("lazamar") {
		versions, err = lazamar.Versions(pkgAttr, ctx.String("channel"))
	} else {
		versions, err = nixhub.Versions(pkgAttr)
	}
	if err != nil {
		return nil, err
	}
	if ctx.Bool("exact") {
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

func findVersionsAll(ctx *cli.Context, names []string) ([]lib.Version, error) {
	var versions []lib.Version
	for _, name := range names {
		vers, err := findVersions(ctx, name)
		if err != nil {
			return nil, err
		}
		versions = append(versions, vers...)
	}
	return versions, nil
}

func mainAction(ctx *cli.Context) error {
	if ctx.Bool("version") {
		fmt.Print(strings.TrimSpace(AppVersion))
		revision := strings.TrimSpace(AppRevision)
		if revision != "" {
			fmt.Printf(" (%s)", revision)
		}
		fmt.Println()
		return nil
	}

	if ctx.Args().Len() < 1 {
		cli.ShowAppHelpAndExit(ctx, 1)
		return nil
	}
	var (
		versions []lib.Version
		err      error
		str      string
	)

	versions, err = findVersionsAll(ctx, ctx.Args().Slice())
	if err != nil {
		return err
	}

	if ctx.Bool("json") {
		str, err = lib.VersionsJson(versions)
		if err != nil {
			return err
		}
	} else {
		str = lib.VersionsTable(versions)
	}

	fmt.Println(str)
	return nil
}
