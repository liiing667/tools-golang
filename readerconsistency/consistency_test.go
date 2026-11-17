// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

// Package readerconsistency verifies that the JSON, YAML and tag-value readers
// normalize identifiers and validation errors identically, so a CI gate sees
// the same result regardless of which serialization of an SBOM it consumes.
package readerconsistency

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdxlib"
	"github.com/spdx/tools-golang/tagvalue"
	"github.com/spdx/tools-golang/yaml"
)

// Every document below describes the same logical SBOM. The package and
// relationship identifiers deliberately use mixed casing for their mandatory
// SPDXRef-/DocumentRef- prefixes, and each document also carries an unknown
// extension field that must be retained.
const jsonDocument = `{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "spdxref-DOCUMENT",
  "name": "doc",
  "documentNamespace": "https://example.com/ns",
  "x-custom-extension": {"kept": true},
  "creationInfo": {"creators": ["Tool: t"], "created": "2024-01-01T00:00:00Z"},
  "packages": [
    {"SPDXID": "SPDXRef-PkgA", "name": "a", "downloadLocation": "NOASSERTION"},
    {"SPDXID": "spdxref-PkgB", "name": "b", "downloadLocation": "NOASSERTION"}
  ],
  "relationships": [
    {"spdxElementId": "SPDXRef-DOCUMENT", "relationshipType": "DESCRIBES", "relatedSpdxElement": "SPDXRef-PkgA"},
    {"spdxElementId": "spdxref-PkgA", "relationshipType": "DEPENDS_ON", "relatedSpdxElement": "SPDXRef-Missing"},
    {"spdxElementId": "SpdxRef-PkgB", "relationshipType": "CONTAINS", "relatedSpdxElement": "NOASSERTION"}
  ]
}`

const yamlDocument = `
spdxVersion: SPDX-2.3
dataLicense: CC0-1.0
SPDXID: spdxref-DOCUMENT
name: doc
documentNamespace: https://example.com/ns
x-custom-extension:
  kept: true
creationInfo:
  creators:
    - "Tool: t"
  created: "2024-01-01T00:00:00Z"
packages:
  - SPDXID: SPDXRef-PkgA
    name: a
    downloadLocation: NOASSERTION
  - SPDXID: spdxref-PkgB
    name: b
    downloadLocation: NOASSERTION
relationships:
  - spdxElementId: SPDXRef-DOCUMENT
    relationshipType: DESCRIBES
    relatedSpdxElement: SPDXRef-PkgA
  - spdxElementId: spdxref-PkgA
    relationshipType: DEPENDS_ON
    relatedSpdxElement: SPDXRef-Missing
  - spdxElementId: SpdxRef-PkgB
    relationshipType: CONTAINS
    relatedSpdxElement: NOASSERTION
`

const tvDocument = `SPDXVersion: SPDX-2.3
DataLicense: CC0-1.0
SPDXID: spdxref-DOCUMENT
DocumentName: doc
DocumentNamespace: https://example.com/ns
Creator: Tool: t
Created: 2024-01-01T00:00:00Z
PackageName: a
SPDXID: SPDXRef-PkgA
PackageDownloadLocation: NOASSERTION
PackageName: b
SPDXID: spdxref-PkgB
PackageDownloadLocation: NOASSERTION
Relationship: SPDXRef-DOCUMENT DESCRIBES SPDXRef-PkgA
Relationship: spdxref-PkgA DEPENDS_ON SPDXRef-Missing
Relationship: SpdxRef-PkgB CONTAINS NOASSERTION
`

func readAll(t *testing.T) map[string]*spdx.Document {
	t.Helper()

	jsonDoc, err := json.Read(strings.NewReader(jsonDocument))
	if err != nil {
		t.Fatalf("json read error: %v", err)
	}

	yamlDoc, err := yaml.Read(strings.NewReader(yamlDocument))
	if err != nil {
		t.Fatalf("yaml read error: %v", err)
	}

	tvDoc, err := tagvalue.Read(strings.NewReader(tvDocument))
	if err != nil {
		t.Fatalf("tag-value read error: %v", err)
	}

	return map[string]*spdx.Document{
		"json":      jsonDoc,
		"yaml":      yamlDoc,
		"tag-value": tvDoc,
	}
}

func TestReadersNormalizeIdentifierCasing(t *testing.T) {
	docs := readAll(t)

	for name, doc := range docs {
		if got, want := doc.SPDXIdentifier, spdx.ElementID("DOCUMENT"); got != want {
			t.Errorf("%s: document id = %q, want %q", name, got, want)
		}
		if got, want := doc.Packages[1].PackageSPDXIdentifier, spdx.ElementID("PkgB"); got != want {
			t.Errorf("%s: second package id = %q, want %q", name, got, want)
		}
	}
}

func TestReadersProduceIdenticalValidationErrors(t *testing.T) {
	docs := readAll(t)

	results := make(map[string]spdxlib.ValidationErrors, len(docs))
	for name, doc := range docs {
		results[name] = spdxlib.ValidateDocumentErrors(doc)
	}

	base := results["json"]
	if len(base) != 1 {
		t.Fatalf("expected exactly one validation error, got %#v", base)
	}
	if base[0].Kind != spdxlib.ValidationErrorMissingElement {
		t.Fatalf("expected a missing-element error, got %#v", base[0])
	}
	if base[0].Reference != "SPDXRef-Missing" {
		t.Fatalf("unexpected reference %q", base[0].Reference)
	}

	for name := range results {
		if !reflect.DeepEqual(results[name], base) {
			t.Errorf("%s validation errors differ from json:\n got %#v\nwant %#v", name, results[name], base)
		}
	}
}

func TestReadersRetainUnknownExtensionFields(t *testing.T) {
	docs := readAll(t)

	// tag-value has no generic key/value extension mechanism; only the
	// structured formats can carry an arbitrary unknown property.
	for _, name := range []string{"json", "yaml"} {
		doc := docs[name]
		payload, ok := doc.CustomFields["x-custom-extension"]
		if !ok {
			t.Fatalf("%s: extension field was not retained: %#v", name, doc.CustomFields)
		}
		if !strings.Contains(string(payload), "kept") {
			t.Fatalf("%s: extension payload was not preserved: %s", name, string(payload))
		}
	}
}
