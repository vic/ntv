package search_spec

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"

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
	NixPackagesCom   *Unit
}

type PackageSearchSpec struct {
	Spec              *string
	Query             *string
	OutputSelectors   []string
	VersionConstraint *string
	VersionsBackend   *VersionsBackend
}

func (b VersionsBackend) String() string {
	if b.CurrentNixpkgs != nil {
		return "system"
	}
	if b.NixHub != nil {
		return "nixhub"
	}
	if b.LazamarChannel != nil {
		return "lazamar:" + string(*b.LazamarChannel)
	}
	if b.NixPackagesCom != nil {
		return "history"
	}
	return "flake"
}

func ParseSearchSpecs(args []string, defaultBackend VersionsBackend) (PackageSearchSpecs, error) {
	group, _ := errgroup.WithContext(context.Background())
	specs := make(PackageSearchSpecs, len(args))
	for i, pkg := range args {
		i, pkg := i, pkg
		group.Go(func() error {
			s, err := newPackageSearchSpec(pkg, defaultBackend)
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
	return !(s.VersionsBackend == nil || (s.VersionsBackend.CurrentNixpkgs == nil &&
		s.VersionsBackend.NixHub == nil &&
		s.VersionsBackend.LazamarChannel == nil &&
		s.VersionsBackend.FlakeInstallable == nil &&
		s.VersionsBackend.NixPackagesCom == nil))
}

func newPackageSearchSpec(spec string, defaultBackend VersionsBackend) (*PackageSearchSpec, error) {
	original_spec := strings.Clone(spec)
	s := &PackageSearchSpec{
		Spec:  &original_spec,
		Query: &spec,
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

	// has output selectors
	if strings.Contains(*s.Query, "^") {
		idx := strings.LastIndex(*s.Query, "^")
		q := (*s.Query)[:idx]
		v := (*s.Query)[idx+1:]
		s.Query = &q
		s.OutputSelectors = strings.Split(v, ",")
	}

	if strings.HasPrefix(*s.Query, "system:") {
		str := strings.TrimPrefix(*s.Query, "system:")
		s.Query = &str
		s.VersionsBackend = &VersionsBackend{CurrentNixpkgs: &Unit{}}
	}

	if strings.HasPrefix(*s.Query, "history:") {
		str := strings.TrimPrefix(*s.Query, "history:")
		s.Query = &str
		s.VersionsBackend = &VersionsBackend{NixPackagesCom: &Unit{}}
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
		if defaultBackend.LazamarChannel != nil {
			channel = string(*defaultBackend.LazamarChannel)
		}
		if strings.Contains(*s.Query, ":") {
			parts := strings.SplitN(*s.Query, ":", 2)
			channel = parts[0]
			s.Query = &parts[1]
		}
		s.VersionsBackend = &VersionsBackend{LazamarChannel: (*LazamarChannel)(&channel)}
	}

	if !s.HasBackend() {
		if strings.HasPrefix(*s.Query, "bin/") || !strings.ContainsAny(*s.Query, "/:#") {
			s.VersionsBackend = &defaultBackend
		} else {
			installable := *s.Query
			s.VersionsBackend = &VersionsBackend{FlakeInstallable: (*FlakeInstallable)(&installable)}
		}
	}

	return s, nil
}

var SimpleAttrRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

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
