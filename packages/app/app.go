package app

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	find "github.com/vic/nix-versions/packages/find"
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
				Name: "lazamar",
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
				Name:  "nixhub",
				Value: true,
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
				Value: false,
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

	opts := find.Opts{
		Exact:      ctx.Bool("exact"),
		Constraint: ctx.String("constraint"),
		Limit:      ctx.Int("limit"),
		Sort:       ctx.Bool("sort"),
		Reverse:    ctx.Bool("reverse"),
		Lazamar:    ctx.Bool("lazamar"),
		Channel:    ctx.String("channel"),
	}

	versions, err = find.FindVersionsAll(opts, ctx.Args().Slice())
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
