package versions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// TODO: Rename to Installable
// TODO: Produce as few different revisions as possible.
//
//	Each nixpkgs checkout is about 40M
//	--optimize=true
type Version struct {
	Name      string `json:"name"`
	Attribute string `json:"attr_path"`
	Version   string `json:"version"`
	Flake     string `json:"flake"`
	Revision  string `json:"revision"`
}

type ByVersion []Version

func (a ByVersion) Len() int      { return len(a) }
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool {
	x, err := semver.NewVersion(a[i].Version)
	if err != nil {
		return false
	}
	y, err := semver.NewVersion(a[j].Version)
	if err != nil {
		return false
	}
	return x.LessThan(y)
}

func SortByVersion(versions []Version) {
	sort.Sort(ByVersion(versions))
}

func ConstraintBy(versions []Version, constraint string) ([]Version, error) {
	constraint = strings.Replace(constraint, "latest", "", 1)
	if strings.TrimSpace(constraint) == "" {
		constraint = "*"
	}
	cond, err := semver.NewConstraint(constraint)
	if err != nil {
		return nil, fmt.Errorf("could not create constraint from string `%s`:\n%v\nSee https://github.com/Masterminds/semver?tab=readme-ov-file#basic-comparisons", constraint, err)
	}
	var res []Version
	for _, ver := range versions {
		v, err := semver.NewVersion(ver.Version)
		if err != nil {
			continue
		}
		if cond.Check(v) {
			res = append(res, ver)
		}
	}
	return res, nil
}
