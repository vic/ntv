package lazamar

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/carlmjohnson/requests"
	lib "github.com/vic/nix-versions/packages/versions"
)

func Versions(name string, channel string) ([]lib.Version, error) {
	var (
		body   string
		result []lib.Version
	)
	err := requests.
		URL("https://lazamar.co.uk/nix-versions/").
		Param("channel", channel).
		Param("package", name).
		ToString(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error fetching versions from lazamar.co.uk for `%s`: %v\nPerhaps the package is not available on nixpkgs under the `%s` name.\nTry using `~%s` as argument or use https://search.nixos.org/packages?query=%s to find the proper attribute name.", name, err, name, name, name)
	}
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	list := htmlquery.Find(doc, "/html/body/section/table/tbody/tr/td/a/@href")
	if len(list) == 0 {
		return nil, fmt.Errorf("No versions found on lazamar.co.uk for `%s`.\nPerhaps the package is not available on nixpkgs under the `%s` name.\nTry using `~%s` as argument or use https://search.nixos.org/packages?query=%s to find the proper attribute name.", name, name, name, name)
	}
	for _, link := range list {
		href, err := url.Parse(htmlquery.InnerText(link))
		if err != nil {
			return nil, err
		}
		query := href.Query()
		version := lib.Version{
			Attribute: query.Get("keyName"),
			Version:   query.Get("version"),
			Revision:  query.Get("revision"),
		}
		result = append(result, version)
	}
	return result, nil
}
