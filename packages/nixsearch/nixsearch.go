package nixsearch

import (
	"context"
	"fmt"
	"slices"

	lib "github.com/peterldowns/nix-search-cli/pkg/nixsearch"
)

func FindPackagesWithQuery(maxRes int, search string) ([]string, error) {
	var limit int
	if maxRes < 0 {
		limit = max(maxRes*-1, 10)
	} else {
		limit = max(maxRes, 10)
	}
	query := lib.Query{
		MaxResults: limit,
		Channel:    "unstable",
		Search:     &lib.MatchSearch{Search: search},
	}
	client, err := lib.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	packages, err := client.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.AttrName)
	}
	if len(names) < 1 {
		return nil, fmt.Errorf("No packages found matching `%s`.", search)
	}
	return names, nil
}

func FindPackagesWithProgram(maxRes int, exact bool, program string) ([]string, error) {
	var limit int
	if maxRes < 0 {
		limit = max(maxRes*-1, 10)
	} else {
		limit = max(maxRes, 10)
	}
	query := lib.Query{
		MaxResults: limit,
		Channel:    "unstable",
		Program:    &lib.MatchProgram{Program: program},
	}
	client, err := lib.NewElasticSearchClient()
	if err != nil {
		return nil, err
	}
	packages, err := client.Search(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, pkg := range packages {
		if exact {
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
