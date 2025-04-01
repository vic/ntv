package new

import (
	"fmt"

	"github.com/vic/ntv/packages/flake"
	"github.com/vic/ntv/packages/search"
	"github.com/vic/ntv/packages/search_spec"
)

func (a *InitArgs) Run() error {
	f := flake.New()

	if a.NtvFlake != "" {
		f.Flake.OverrideInput("ntv", a.NtvFlake)
	}

	specs, err := search_spec.ParseSearchSpecs(a.rest, a.versionsBackend)
	if err != nil {
		return err
	}

	res, err := search.PackageSearchSpecs(specs).Search()
	if err != nil {
		return err
	}

	code, err := FlakeCode(f, res)
	if err != nil {
		return err
	}

	fmt.Println(code)
	return nil
}

func FlakeCode(f *flake.Context, res search.PackageSearchResults) (string, error) {
	if err := res.EnsureOneSelected(); err != nil {
		return "", err
	}
	if err := res.EnsureUniquePackageNames(); err != nil {
		return "", err
	}

	for _, r := range res {
		f.AddTool(r)
	}

	return f.Render(true)
}
