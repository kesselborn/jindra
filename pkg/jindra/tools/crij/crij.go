package crij

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// EnvToJSON converts env variables to json structures:
// foo.bar=baz
// foo.baz=baz
//
// results in:
// {"foo": {"bar": "baz", "baz": "baz" }
func EnvToJSON(prefix string) (string, error) {
	envVars := os.Environ()
	jsonStructure := map[string]interface{}{}

	for _, envVar := range envVars {
		envVarTokens := strings.SplitN(envVar, "=", 2)
		key := envVarTokens[0]
		value := envVarTokens[1]

		// remove prefix if set
		if (strings.Index(key, prefix+".") == 0 || prefix == "") && strings.Contains(key, ".") {
			keyTokens := strings.Split(key, ".")
			if prefix != "" {
				keyTokens = keyTokens[1:]
			}

			cur := jsonStructure
			prev := map[string]interface{}{}
			prevKey := ""
			for i, key := range keyTokens {
				// most inner json element / last element in dot notation
				if i == len(keyTokens)-1 {
					// if value is json, unmarshal it
					var inlinedJSON interface{}
					err := json.Unmarshal([]byte(value), &inlinedJSON)
					if err != nil {
						if key == "" {
							prev[prevKey] = interface{}(value)
						} else {
							cur[key] = interface{}(value)
						}
					} else {
						if key == "" {
							prev[prevKey] = inlinedJSON
						} else {
							cur[key] = inlinedJSON
						}
					}
					break
				}

				// create new key
				if _, ok := cur[key]; !ok {
					cur[key] = interface{}(map[string]interface{}{})
				}

				if next, ok := cur[key].(map[string]interface{}); ok {
					prev = cur
					prevKey = key
					cur = next
				} else {
					return "", fmt.Errorf("error creating json for var %s: not sure whether to create string or map for %s", envVar, keyTokens[0:i])
				}
			}
		}
	}

	b, err := json.MarshalIndent(jsonStructure, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling structure to json: %s", err)
	}

	return string(b), nil
}
