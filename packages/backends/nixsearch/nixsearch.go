package nixsearch

import (
	"context"

	lib "github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

// nixos-search elastic indexes.
// https://github.com/NixOS/nixos-search/blob/main/flake-info/src/elastic.rs
type Package = lib.Package

func FindPackagesWithAttr(maxRes int, search string) ([]lib.Package, error) {
	query := lib.Query{
		MaxResults:  maxRes,
		Channel:     "unstable",
		QueryString: &lib.MatchQueryString{QueryString: "package_attr_name: " + search},
	}
	client, err := lib.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	return client.Search(context.Background(), query)
}

func FindPackagesWithProgram(maxRes int, program string) ([]lib.Package, error) {
	query := lib.Query{
		MaxResults:  maxRes,
		Channel:     "unstable",
		QueryString: &lib.MatchQueryString{QueryString: "package_programs: " + program},
	}
	client, err := lib.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	return client.Search(context.Background(), query)
}
