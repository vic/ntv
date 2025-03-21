package marshalling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	lib "github.com/vic/nix-versions/packages/versions"
)

type VersionAndAttribute struct {
	Version   string `json:"version"`
	Attribute string `json:"attribute"`
}

func Flake(versions []lib.Version) (string, error) {
	namesToAttrs := make(map[string]VersionAndAttribute)
	for _, version := range versions {
		namesToAttrs[version.Name] = VersionAndAttribute{
			Version:   version.Version,
			Attribute: version.Attribute,
		}
	}
	jsonBytes, err := json.MarshalIndent(namesToAttrs, "    ", "  ")
	if err != nil {
		return "", err
	}
	json := string(jsonBytes)

	var buff bytes.Buffer
	buff.WriteString("{\n")
	buff.WriteString("  inputs.nix-versions.url = \"github:vic/nix-versions\";\n")
	buff.WriteString("  inputs.nix-versions.inputs.nixpkgs.follows = \"nixpkgs\";\n")
	for _, version := range versions {
		name := version.Name
		emptyRevs := []string{"master", "main", "HEAD", ""}
		var url string
		if slices.Contains(emptyRevs, version.Revision) {
			url = version.Flake
		} else {
			url = fmt.Sprintf("%s/%s", version.Flake, version.Revision)
		}
		buff.WriteString(fmt.Sprintf("  inputs.\"%s\".url = \"%s\";\n", name, url))
	}
	buff.WriteString("  outputs = inputs@{nixpkgs, self, ...}: inputs.nix-versions.lib.mkFlake {\n")
	buff.WriteString("    inherit inputs;\n")
	buff.WriteString("    flakeModule = ./flakeModule.nix;\n")
	buff.WriteString(fmt.Sprintf("    nix-versions = builtins.fromJSON ''%s'';\n", json))
	buff.WriteString("  };\n")
	buff.WriteString("}\n")
	return buff.String(), nil
}
