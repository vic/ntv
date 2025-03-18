package app

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/urfave/cli/v2"
	"github.com/vic/nix-versions/packages/lazamar"
	"github.com/vic/nix-versions/packages/nixhub"
	lib "github.com/vic/nix-versions/packages/versions"
)

func App() cli.App {
	return cli.App{
		Name:            "nix-versions",
		Usage:           "show available nix packages versions",
		ArgsUsage:       "PKG_ATTRIBUTE_NAME",
		HideHelpCommand: true,
		Action:          mainAction,
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Victor Hugo Borja",
				Email: "vborja@apache.org",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "lazamar",
				Value:    true,
				Category: "NIX VERSIONS BACKEND",
				Usage:    "Use https://lazamar.co.uk/nix-versions as backend",
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("nixhub", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.StringFlag{
				Name:     "channel",
				Value:    "nixpkgs-unstable",
				Category: "NIX VERSIONS BACKEND",
				Usage:    "Nixpkgs channel for lazamar backend. Enables lazamar when set.",
				Action: func(ctx *cli.Context, b string) error {
					ctx.Set("nixhub", "false")
					return nil
				},
			},
			&cli.BoolFlag{
				Name:     "nixhub",
				Category: "NIX VERSIONS BACKEND",
				Usage:    "Use https://www.nixhub.io/ as backend",
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("lazamar", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:     "json",
				Category: "FORMAT",
				Usage:    "Output JSON array of versions",
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("text", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:     "text",
				Category: "FORMAT",
				Usage:    "Output text table of versions",
				Value:    true,
				Action: func(ctx *cli.Context, b bool) error {
					ctx.Set("json", strconv.FormatBool(!b))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:     "sort",
				Category: "FILTERING",
				Usage:    "Sorted by version instead of using backend ordering",
				Value:    true,
			},
			&cli.BoolFlag{
				Name:     "reverse",
				Category: "FILTERING",
				Usage:    "New versions last",
				Value:    false,
			},
			&cli.UintFlag{
				Name:     "limit",
				Category: "FILTERING",
				Usage:    "Limit to a number of results. eg: `10`",
				Value:    0,
			},
			&cli.StringFlag{
				Name:     "constraint",
				Category: "FILTERING",
				Usage:    "Version constraint. eg: `'~1.0'`. See github.com/Masterminds/semver",
			},
		},
	}
}

func mainAction(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		cli.ShowAppHelpAndExit(ctx, 1)
		return nil
	}
	var (
		versions []lib.Version
		err      error
		str      string
	)

	if ctx.Bool("lazamar") {
		versions, err = lazamar.Versions(ctx.Args().First(), ctx.String("channel"))
	} else {
		versions, err = nixhub.Versions(ctx.Args().First())
	}
	if err != nil {
		return err
	}
	if ctx.String("constraint") != "" {
		versions, err = lib.ConstraintBy(versions, ctx.String("constraint"))
	}
	if ctx.Bool("sort") {
		lib.SortByVersion(versions)
	}
	if ctx.Bool("reverse") {
		slices.Reverse(versions)
	}
	if err != nil {
		return err
	}
	if ctx.Uint("limit") > 0 {
		upto := min(len(versions), ctx.Int("limit"))
		versions = versions[:upto]
	}
	if ctx.Bool("json") {
		str, err = lib.VersionsJson(versions)
	} else {
		str = lib.VersionsTable(versions)
	}
	if err != nil {
		return err
	}
	fmt.Println(str)
	return nil
}
