package nixhub

import (
	"context"

	"github.com/carlmjohnson/requests"
	lib "github.com/vic/nix-versions/packages/versions"
)

type platform struct {
	AttributePath string `json:"attribute_path"`
	CommitHash    string `json:"commit_hash"`
}

type release struct {
	Version   string     `json:"version"`
	Platforms []platform `json:"platforms"`
}

type response struct {
	Releases []release `json:"releases"`
}

func Versions(name string) ([]lib.Version, error) {
	var (
		body   response
		result []lib.Version
	)
	err := requests.
		URL("https://search.devbox.sh/v2/pkg").
		Method("GET").
		Param("name", name).
		Accept("application/json").
		ToJSON(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, err
	}
	for _, release := range body.Releases {
		platform := release.Platforms[len(release.Platforms)-1]
		version := lib.Version{
			Attribute: platform.AttributePath,
			Version:   release.Version,
			Revision:  platform.CommitHash,
		}
		result = append(result, version)
	}
	return result, nil
}
