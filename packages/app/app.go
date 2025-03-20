package app

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	find "github.com/vic/nix-versions/packages/find"
	lib "github.com/vic/nix-versions/packages/versions"
)

//go:embed HELP
var AppHelp string

//go:embed VERSION
var AppVersion string

//go:embed REVISION
var AppRevision string

type CliArgs struct {
	OnHelp     func()       `long:"help" short:"h"`
	OnVersion  func()       `long:"version"`
	OnChannel  func(string) `long:"channel"`
	OnLazamar  func()       `long:"lazamar"`
	OnNixHub   func()       `long:"nixhub"`
	OnJson     func()       `long:"json"`
	OnText     func()       `long:"text"`
	Lazamar    bool
	Channel    string
	Json       bool
	Sort       bool   `long:"sort"`
	Reverse    bool   `long:"reverse"`
	Exact      bool   `long:"exact"`
	Limit      int    `long:"limit"`
	Constraint string `long:"constraint"`
	Names      []string
}

func ParseCliArgs(args []string) (CliArgs, error) {
	var cliArgs = CliArgs{
		Channel: "nixpkgs-unstable",
		Sort:    true,
	}
	cliArgs.OnHelp = func() {
		fmt.Println(AppHelp)
		os.Exit(0)
	}
	cliArgs.OnVersion = func() {
		fmt.Print(strings.TrimSpace(AppVersion))
		revision := strings.TrimSpace(AppRevision)
		if revision != "" {
			fmt.Printf(" (%s)", revision)
		}
		fmt.Println()
		os.Exit(0)
	}
	cliArgs.OnLazamar = func() {
		cliArgs.Lazamar = true
	}
	cliArgs.OnNixHub = func() {
		cliArgs.Lazamar = false
	}
	cliArgs.OnChannel = func(s string) {
		cliArgs.Channel = s
		cliArgs.Lazamar = true
	}
	cliArgs.OnJson = func() {
		cliArgs.Json = true
	}
	cliArgs.OnText = func() {
		cliArgs.Json = false
	}
	parser := flags.NewParser(&cliArgs, flags.AllowBoolValues)
	names, err := parser.ParseArgs(args)
	cliArgs.Names = names
	return cliArgs, err
}

func MainAction(ctx CliArgs) error {
	if len(ctx.Names) < 1 {
		fmt.Println(AppHelp)
		os.Exit(1)
		return nil
	}
	var (
		versions []lib.Version
		err      error
		str      string
	)

	opts := find.Opts{
		Exact:      ctx.Exact,
		Constraint: ctx.Constraint,
		Limit:      ctx.Limit,
		Sort:       ctx.Sort,
		Reverse:    ctx.Reverse,
		Lazamar:    ctx.Lazamar,
		Channel:    ctx.Channel,
	}

	versions, err = find.FindVersionsAll(opts, ctx.Names)
	if err != nil {
		return err
	}

	if ctx.Json {
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
