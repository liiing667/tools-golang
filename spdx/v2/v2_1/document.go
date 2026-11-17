// Package spdx contains the struct definition for an SPDX Document
// and its constituent parts.
// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package v2_1

import (
	"encoding/json"

	"github.com/spdx/tools-golang/spdx/v2/common"
)

const Version = "SPDX-2.1"
const DataLicense = "CC0-1.0"

// ExternalDocumentRef is a reference to an external SPDX document
// as defined in section 2.6 for version 2.1 of the spec.
type ExternalDocumentRef struct {
	// DocumentRefID is the ID string defined in the start of the
	// reference. It should _not_ contain the "DocumentRef-" part
	// of the mandatory ID string.
	DocumentRefID common.DocumentID `json:"externalDocumentId"`

	// URI is the URI defined for the external document
	URI string `json:"spdxDocument"`

	// Checksum is the actual hash data
	Checksum common.Checksum `json:"checksum"`
}

// Document is an SPDX Document for version 2.1 of the spec.
// See https://spdx.org/sites/cpstandard/files/pages/files/spdxversion2.1.pdf
type Document struct {
	// 2.1: SPDX Version; should be in the format "SPDX-2.1"
	// Cardinality: mandatory, one
	SPDXVersion string `json:"spdxVersion"`

	// 2.2: Data License; should be "CC0-1.0"
	// Cardinality: mandatory, one
	DataLicense string `json:"dataLicense"`

	// 2.3: SPDX Identifier; should be "DOCUMENT" to represent
	//      mandatory identifier of SPDXRef-DOCUMENT
	// Cardinality: mandatory, one
	SPDXIdentifier common.ElementID `json:"SPDXID"`

	// 2.4: Document Name
	// Cardinality: mandatory, one
	DocumentName string `json:"name"`

	// 2.5: Document Namespace
	// Cardinality: mandatory, one
	DocumentNamespace string `json:"documentNamespace"`

	// 2.6: External Document References
	// Cardinality: optional, one or many
	ExternalDocumentReferences []ExternalDocumentRef `json:"externalDocumentRefs,omitempty"`

	// 2.11: Document Comment
	// Cardinality: optional, one
	DocumentComment string `json:"comment,omitempty"`

	CreationInfo  *CreationInfo   `json:"creationInfo"`
	Packages      []*Package      `json:"packages,omitempty"`
	Files         []*File         `json:"files,omitempty"`
	OtherLicenses []*OtherLicense `json:"hasExtractedLicensingInfos,omitempty"`
	Relationships []*Relationship `json:"relationships,omitempty"`
	Annotations   []*Annotation   `json:"annotations,omitempty"`
	Snippets      []Snippet       `json:"snippets,omitempty"`

	// DEPRECATED in version 2.0 of spec
	Reviews []*Review `json:"-"`

	// CustomFields retains unknown (custom or future-version) top-level fields
	// instead of discarding them. It is populated while unmarshalling and is
	// emitted inline when marshalling; it never appears as a named property.
	CustomFields common.Extensions `json:"-" yaml:"-"`
}

// knownDocumentJSONFields lists every top-level property modeled by the
// Document struct. Anything else encountered while parsing is preserved in
// Extensions.
var knownDocumentJSONFields = map[string]struct{}{
	"spdxVersion":                {},
	"dataLicense":                {},
	"SPDXID":                     {},
	"name":                       {},
	"documentNamespace":          {},
	"externalDocumentRefs":       {},
	"comment":                    {},
	"creationInfo":               {},
	"packages":                   {},
	"files":                      {},
	"hasExtractedLicensingInfos": {},
	"relationships":              {},
	"annotations":                {},
	"snippets":                   {},
}

// UnmarshalJSON decodes the modeled document fields and retains any unknown
// top-level fields in Extensions rather than discarding them.
func (d *Document) UnmarshalJSON(b []byte) error {
	type doc Document
	var d2 doc
	if err := json.Unmarshal(b, &d2); err != nil {
		return err
	}

	extensions, err := common.CaptureExtensions(b, knownDocumentJSONFields)
	if err != nil {
		return err
	}

	*d = Document(d2)
	d.CustomFields = extensions
	return nil
}

// MarshalJSON emits the modeled document fields with any retained extension
// fields merged in inline, so unknown properties survive a round-trip without
// being nested under an extra wrapper property.
func (d *Document) MarshalJSON() ([]byte, error) {
	type doc Document
	return common.MergeExtensions(doc(*d), d.CustomFields)
}

var (
	_ json.Unmarshaler = (*Document)(nil)
	_ json.Marshaler   = (*Document)(nil)
)
