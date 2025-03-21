package app

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
	find "github.com/vic/nix-versions/packages/find"
	lib "github.com/vic/nix-versions/packages/versions"
)

//go:embed HELP
var AppHelp string

//go:embed VERSION
var AppVersion string

//go:embed REVISION
var AppRevision string

type OutputType int

const (
	Text OutputType = iota
	Json
	Installable
	Flake
)

type CliArgs struct {
	OnHelp        func()       `long:"help" short:"h"`
	OnVersion     func()       `long:"version"`
	OnChannel     func(string) `long:"channel"`
	OnLazamar     func()       `long:"lazamar"`
	OnNixHub      func()       `long:"nixhub"`
	OnJson        func()       `long:"json"`
	OnText        func()       `long:"text"`
	OnInstallable func()       `long:"installable"`
	OnFlake       func()       `long:"flake"`
	OnWrite       func(string) `long:"out" short:"o"`
	Lazamar       bool
	Channel       string
	OutType       OutputType
	WriteTo       string
	Color         bool   `long:"color"`
	One           bool   `long:"assert-one"`
	Sort          bool   `long:"sort"`
	Reverse       bool   `long:"reverse"`
	Exact         bool   `long:"exact"`
	Limit         int    `long:"limit"`
	Constraint    string `long:"constraint"`
	Names         []string
}

func ParseCliArgs(args []string) (CliArgs, error) {
	var cliArgs = CliArgs{
		Channel: "nixpkgs-unstable",
		Sort:    true,
		Color:   isatty.IsTerminal(os.Stdout.Fd()),
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
		cliArgs.OutType = Json
	}
	cliArgs.OnText = func() {
		cliArgs.OutType = Text
	}
	cliArgs.OnInstallable = func() {
		cliArgs.OutType = Installable
	}
	cliArgs.OnFlake = func() {
		cliArgs.OutType = Flake
	}
	cliArgs.OnWrite = func(s string) {
		cliArgs.One = true
		cliArgs.WriteTo = s
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
		One:        ctx.One,
		Exact:      ctx.Exact,
		Constraint: ctx.Constraint,
		Limit:      ctx.Limit,
		Sort:       ctx.Sort,
		Reverse:    ctx.Reverse,
		Lazamar:    ctx.Lazamar,
		Channel:    ctx.Channel,
	}

	versions, err = find.FindReadingVersionsAll(opts, ctx.Names)
	if err != nil {
		return err
	}

	if ctx.OutType == Json {
		str, err = lib.VersionsJson(versions)
		if err != nil {
			return err
		}
	} else if ctx.OutType == Installable {
		str = lib.Installables(versions)
	} else if ctx.OutType == Flake {
		str, err = lib.Flake(versions)
		if err != nil {
			return err
		}
	} else {
		str = lib.VersionsTable(versions, ctx.Color)
	}

	if ctx.One {
		var seen = make(map[string]int)
		var anyFailed = false
		for _, v := range versions {
			seen[v.Name]++
			if seen[v.Name] > 1 {
				anyFailed = true
			}
		}
		if anyFailed {
			str = lib.VersionsTable(versions, ctx.Color)
			fmt.Fprint(os.Stderr, str, "\n")
			return fmt.Errorf("Assertion failure. Expected at most one version per package.\nBut got %v.\nTry using @latest or a more specific version constraint.", seen)
		}
	}

	if ctx.WriteTo != "" && ctx.WriteTo != "-" {
		err := os.WriteFile(ctx.WriteTo, []byte(str), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	fmt.Println(str)
	return nil
}
