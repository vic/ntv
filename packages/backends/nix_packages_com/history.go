package nix_packages_com

// Backend for https://history.nix-packages.com/

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/requests"
	lib "github.com/vic/ntv/packages/versions"
)

type version struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func Search(name string) ([]*lib.Version, error) {
	var (
		body   []version
		result []*lib.Version
	)
	err := requests.
		URL("https://api.history.nix-packages.com/packages/" + name).
		Method("GET").
		Accept("application/json").
		ToJSON(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error fetching versions from history.nix-packages.com for `%s`: %v\nPerhaps the package is not available on history.nix-packages.com under the `%s` name.\nTry using `*%s*` as argument or use https://history.nix-packages.com/search?search=%s to find the proper attribute name", name, err, name, name, name)
	}
	for _, v := range body {
		version := lib.Version{
			Name:      v.Name,
			Attribute: v.Name,
			Version:   v.Version,
			Revision:  v.Revision,
			Flake:     "nixpkgs",
		}
		result = append(result, &version)
	}
	return result, nil
}
