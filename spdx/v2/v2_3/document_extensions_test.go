// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package v2_3

import (
	"encoding/json"
	"testing"
)

func TestDocumentExtensionsOnlyCapturesUnknown(t *testing.T) {
	in := []byte(`{
		"spdxVersion": "SPDX-2.3",
		"dataLicense": "CC0-1.0",
		"SPDXID": "SPDXRef-DOCUMENT",
		"name": "n",
		"documentNamespace": "ns",
		"documentDescribes": ["SPDXRef-p1"],
		"knownButExtra": "x"
	}`)

	var d Document
	if err := json.Unmarshal(in, &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(d.CustomFields) != 1 {
		t.Fatalf("expected exactly one custom field, got %#v", d.CustomFields)
	}
	if _, ok := d.CustomFields["knownButExtra"]; !ok {
		t.Fatalf("expected knownButExtra to be retained, got %#v", d.CustomFields)
	}
	if _, ok := d.CustomFields["spdxVersion"]; ok {
		t.Fatal("modeled field must not be captured as an extension")
	}

	out, err := json.Marshal(&d)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var round map[string]json.RawMessage
	if err := json.Unmarshal(out, &round); err != nil {
		t.Fatalf("round-trip parse error: %v", err)
	}
	if _, ok := round["knownButExtra"]; !ok {
		t.Fatalf("extension missing from marshalled output: %s", string(out))
	}
	if _, ok := round["CustomFields"]; ok {
		t.Fatal("extensions must be emitted inline, not under a CustomFields key")
	}
}
