package marshalling

import (
	"encoding/json"

	lib "github.com/vic/nix-versions/packages/versions"
)

func VersionsJson(versions []lib.Version) (string, error) {
	if len(versions) == 0 {
		return "", nil
	}
	var obj any
	if len(versions) == 1 {
		obj = versions[0]
	} else {
		obj = versions
	}
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
