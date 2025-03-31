package search_spec

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/vic/ntv/packages/backends/nixsearch"
	"golang.org/x/sync/errgroup"
)

type PackageSearchSpecs []*PackageSearchSpec

type Unit struct{}
type LazamarChannel string
type FlakeInstallable string

// a kind of union type. only one of these should be set.
type VersionsBackend struct {
	CurrentNixpkgs   *Unit
	NixHub           *Unit
	LazamarChannel   *LazamarChannel
	FlakeInstallable *FlakeInstallable
	// TODO: FlakeHub
	// TODO: FlakeURL git tags
}

type PackageSearchSpec struct {
	Spec              *string
	Query             *string
	VersionConstraint *string
	VersionsBackend   *VersionsBackend
}

func ParseSearchSpecs(args []string, defaultToLazamarChannel *string) (PackageSearchSpecs, error) {
	group, _ := errgroup.WithContext(context.Background())
	specs := make(PackageSearchSpecs, len(args))
	for i, pkg := range args {
		i, pkg := i, pkg
		group.Go(func() error {
			s, err := newPackageSearchSpec(pkg, defaultToLazamarChannel)
			if err != nil {
				return err
			}
			specs[i] = s
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return specs, nil
}

func (s *PackageSearchSpec) HasBackend() bool {
	return s.VersionsBackend != nil &&
		(s.VersionsBackend.CurrentNixpkgs != nil ||
			s.VersionsBackend.NixHub != nil ||
			s.VersionsBackend.LazamarChannel != nil ||
			s.VersionsBackend.FlakeInstallable != nil)
}

func newPackageSearchSpec(spec string, defaultToLazamarChannel *string) (*PackageSearchSpec, error) {
	original_spec := strings.Clone(spec)
	s := &PackageSearchSpec{
		Spec:  &original_spec,
		Query: &spec,
	}

	if simpleAttrRegex.MatchString(*s.Query) {
		s.VersionsBackend = &VersionsBackend{CurrentNixpkgs: &Unit{}}
		return s, nil
	}

	// has version constraint
	if strings.Contains(*s.Query, "@") {
		idx := strings.LastIndex(*s.Query, "@")
		q := (*s.Query)[:idx]
		v := (*s.Query)[idx+1:]
		s.Query = &q
		s.VersionConstraint = &v

		if fileExists(*s.VersionConstraint) {
			content, err := readFile(*s.VersionConstraint)
			if err != nil {
				return nil, err
			}
			s.VersionConstraint = &content
		}
	}

	if strings.HasPrefix(*s.Query, "nixhub:") {
		str := strings.TrimPrefix(*s.Query, "nixhub:")
		s.Query = &str
		s.VersionsBackend = &VersionsBackend{NixHub: &Unit{}}
	}

	// lazamar:channel:package or lazamar:package
	if strings.HasPrefix(*s.Query, "lazamar:") {
		str := strings.TrimPrefix(*s.Query, "lazamar:")
		s.Query = &str
		var channel = "nixpkgs-unstable"
		if defaultToLazamarChannel != nil {
			channel = *defaultToLazamarChannel
		}
		if strings.Contains(*s.Query, ":") {
			parts := strings.SplitN(*s.Query, ":", 2)
			channel = parts[0]
			s.Query = &parts[1]
		}
		s.VersionsBackend = &VersionsBackend{LazamarChannel: (*LazamarChannel)(&channel)}
	}

	// search by provided program. eg: bin/pwd
	if strings.HasPrefix(*s.Query, "bin/") {
		str := strings.TrimPrefix(*s.Query, "bin/")
		s.Query = &str
		isExact := !strings.Contains(*s.Query, "*")
		res, err := nixsearch.FindPackagesWithProgram(10, isExact, *s.Query)
		if err != nil {
			return nil, err
		}
		if len(res) == 0 {
			return nil, fmt.Errorf("no packages found providing program: `%s`", *s.Query)
		}
		if len(res) > 1 {
			return nil, fmt.Errorf("multiple packages found providing program: `%s`. Use one of %v instead of `bin/%s`", *s.Query, res, *s.Query)
		}
		s.Query = &res[0]
	}

	// TODO: support search by query: ? (dunno will be useful)

	if !s.HasBackend() {
		if simpleAttrRegex.MatchString(*s.Query) {
			if defaultToLazamarChannel == nil {
				s.VersionsBackend = &VersionsBackend{NixHub: &Unit{}}
			} else {
				channel := *defaultToLazamarChannel
				s.VersionsBackend = &VersionsBackend{LazamarChannel: (*LazamarChannel)(&channel)}
			}
		} else {
			installable := *s.Query
			s.VersionsBackend = &VersionsBackend{FlakeInstallable: (*FlakeInstallable)(&installable)}
		}
	}

	return s, nil
}

var simpleAttrRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	bts, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bts)), nil
}
