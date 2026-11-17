// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package common

import (
	"encoding/json"
)

// Extensions holds fields that appear in an SPDX document but are not modeled
// by the current structs. Unknown (including custom or future-version)
// extension fields are retained here instead of being silently discarded, so
// that a document round-trips without losing data.
//
// The keys are the original JSON/YAML property names and the values are the
// decoded payloads. The nil value means no extension fields were present.
type Extensions map[string]json.RawMessage

// CaptureExtensions unmarshals b into a raw map and returns every property
// whose name is not listed in knownFields. The returned map is nil when there
// are no unknown properties, keeping documents without extensions unchanged.
func CaptureExtensions(b []byte, knownFields map[string]struct{}) (Extensions, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	var extensions Extensions
	for name, payload := range raw {
		if _, known := knownFields[name]; known {
			continue
		}
		if extensions == nil {
			extensions = Extensions{}
		}
		extensions[name] = payload
	}

	return extensions, nil
}

// MergeExtensions returns the JSON encoding of model overlaid with the
// retained extension fields. Modeled properties take precedence on key
// collisions. It is used by the custom document marshalers so that unknown
// fields are emitted as top-level (inline) properties rather than being nested
// under an "extensions" object.
func MergeExtensions(model interface{}, extensions Extensions) ([]byte, error) {
	if len(extensions) == 0 {
		return json.Marshal(model)
	}

	encoded, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}

	var merged map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &merged); err != nil {
		return nil, err
	}

	for name, payload := range extensions {
		if _, exists := merged[name]; exists {
			continue
		}
		merged[name] = payload
	}

	return json.Marshal(merged)
}
