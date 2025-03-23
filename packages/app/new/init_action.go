package new

import (
	"fmt"
	"os/exec"

	"github.com/vic/ntv/packages/flake"
	"github.com/vic/ntv/packages/search"
	"github.com/vic/ntv/packages/search_spec"
)

func (a *InitArgs) Run() error {
	f := flake.New()

	if a.NVFlake != "" {
		f.Flake.OverrideInput("ntv", a.NVFlake)
	}

	err := a.addPackages(f)
	if err != nil {
		return err
	}

	code, err := f.Render()
	if err != nil {
		return err
	}
	fmt.Println(code)

	return nil
}

func (a *InitArgs) addPackages(f *flake.Context) error {
	specs, err := search_spec.ParseSearchSpecs(a.rest, a.LazamarChannel)
	if err != nil {
		return err
	}

	res, err := search.PackageSearchSpecs(specs).Search()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%v: %s", err, string(ee.Stderr))
		}
		return err
	}
	if err = res.EnsureOneSelected(); err != nil {
		return err
	}
	if err = res.EnsureUniquePackageNames(); err != nil {
		return err
	}

	for _, r := range res {
		f.AddTool(r)
	}

	return nil
}
