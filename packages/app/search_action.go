package app

import (
	_ "embed"
	"fmt"
	"os"

	find "github.com/vic/nix-versions/packages/find"
	lib "github.com/vic/nix-versions/packages/versions"
)

func (ctx *SearchArgs) ParseAndRun(args []string) error {
	err := ctx.Parse(args)
	if err != nil {
		return err
	}
	return ctx.SearchAction()
}

func (ctx *SearchArgs) SearchAction() error {
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
