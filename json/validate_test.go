// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package json

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
	"github.com/spdx/tools-golang/spdxlib"
)

func validatableDocument() *spdx.Document {
	return &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		DocumentName:   "test-doc",
		CreationInfo:   &spdx.CreationInfo{},
		ExternalDocumentReferences: []spdx.ExternalDocumentRef{
			{
				DocumentRefID: "external-doc",
				URI:           "http://example.org/external.spdx",
				Checksum:      common.Checksum{Algorithm: common.SHA1, Value: "d6a770ba38583ed4bb4525bd96e50461655d2759"},
			},
		},
		Packages: []*spdx.Package{
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
		},
		Files: []*spdx.File{
			{FileName: "file1", FileSPDXIdentifier: "f1"},
		},
		Snippets: []spdx.Snippet{
			{SnippetSPDXIdentifier: "s1", SnippetFromFileSPDXIdentifier: "f1"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: common.TypeRelationshipDescribe,
			},
			{
				RefA:         common.MakeDocElementID("", "f1"),
				RefB:         common.MakeDocElementID("", "s1"),
				Relationship: common.TypeRelationshipContains,
			},
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("external-doc", "external-pkg"),
				Relationship: common.TypeRelationshipCopyOf,
			},
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementSpecial("NOASSERTION"),
				Relationship: common.TypeRelationshipDependsOn,
			},
		},
	}
}

// A document whose relationship endpoints cover packages, files, snippets,
// a declared DocumentRef- and the NOASSERTION special value must round-trip
// through JSON and still pass validation.
func TestReadValidatesCompleteElementIndex(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := Write(validatableDocument(), buf); err != nil {
		t.Fatalf("error writing document: %v", err)
	}

	parsedDoc, err := Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("error reading document: %v", err)
	}

	if err := spdxlib.ValidateDocument(parsedDoc); err != nil {
		t.Fatalf("expected valid document, got: %s", err.Error())
	}
}

// A dangling reference that survives the JSON round-trip, whether a local
// element or an undeclared DocumentRef-, must be rejected by validation.
func TestReadRejectsDanglingRelationships(t *testing.T) {
	tests := map[string]func(*spdx.Document){
		"undefined local element": func(doc *spdx.Document) {
			doc.Relationships = append(doc.Relationships, &spdx.Relationship{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "missing"),
				Relationship: common.TypeRelationshipContains,
			})
		},
		"undeclared external document": func(doc *spdx.Document) {
			doc.Relationships = append(doc.Relationships, &spdx.Relationship{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("missing-doc", "external-pkg"),
				Relationship: common.TypeRelationshipCopyOf,
			})
		},
		"NONE on left side": func(doc *spdx.Document) {
			doc.Relationships = append(doc.Relationships, &spdx.Relationship{
				RefA:         common.MakeDocElementSpecial("NONE"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: common.TypeRelationshipContains,
			})
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			doc := validatableDocument()
			mutate(doc)

			buf := new(bytes.Buffer)
			if err := Write(doc, buf); err != nil {
				t.Fatalf("error writing document: %v", err)
			}

			parsedDoc, err := Read(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("error reading document: %v", err)
			}

			if err := spdxlib.ValidateDocument(parsedDoc); err == nil {
				t.Fatal("expected validation to reject the dangling relationship")
			} else if !strings.Contains(err.Error(), "relationship") {
				t.Fatalf("expected a relationship error, got: %s", err.Error())
			}
		})
	}
}
