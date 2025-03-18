package lazamar

import (
	"context"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/carlmjohnson/requests"
	lib "github.com/vic/nix-versions/packages/versions"
)

func Versions(packageName string, channel string) ([]lib.Version, error) {
	var (
		body   string
		result []lib.Version
	)
	err := requests.
		URL("https://lazamar.co.uk/nix-versions/").
		Param("channel", channel).
		Param("package", packageName).
		ToString(&body).
		Fetch(context.Background())
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	list := htmlquery.Find(doc, "/html/body/section/table/tbody/tr/td/a/@href")
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
