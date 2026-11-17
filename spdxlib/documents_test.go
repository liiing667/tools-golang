// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package spdxlib

import (
	"testing"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

func TestValidDocumentPassesValidation(t *testing.T) {
	// set up document and some packages and relationships
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages: []*spdx.Package{
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
			{PackageName: "pkg2", PackageSPDXIdentifier: "p2"},
			{PackageName: "pkg3", PackageSPDXIdentifier: "p3"},
			{PackageName: "pkg4", PackageSPDXIdentifier: "p4"},
			{PackageName: "pkg5", PackageSPDXIdentifier: "p5"},
		},
		Files: []*spdx.File{
			{FileName: "file1", FileSPDXIdentifier: "f1"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: "DESCRIBES",
			},
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p5"),
				Relationship: "DESCRIBES",
			},
			// inverse relationship -- should also get detected
			{
				RefA:         common.MakeDocElementID("", "p4"),
				RefB:         common.MakeDocElementID("", "DOCUMENT"),
				Relationship: "DESCRIBED_BY",
			},
			// different relationship
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "p2"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	err := ValidateDocument(doc)
	if err != nil {
		t.Fatalf("expected nil error, got: %s", err.Error())
	}
}

func TestInvalidDocumentFailsValidation(t *testing.T) {
	// set up document and some packages and relationships
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages: []*spdx.Package{
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
			{PackageName: "pkg2", PackageSPDXIdentifier: "p2"},
			{PackageName: "pkg3", PackageSPDXIdentifier: "p3"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: "DESCRIBES",
			},
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p5"),
				Relationship: "DESCRIBES",
			},
			// invalid ID p99
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "p99"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	err := ValidateDocument(doc)
	if err == nil {
		t.Fatalf("expected non-nil error, got nil")
	}

	doc = &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},

		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "p99"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	err = ValidateDocument(doc)
	if err == nil {
		t.Fatalf("expected non-nil error, got nil")
	}
}

func TestDocumentWithSpecialRelationshipPassesValidation(t *testing.T) {
	// per SPDX 2.3 section 11.1, NONE and NOASSERTION are permitted on the
	// right-hand side of a relationship, and should not be checked against
	// the set of element identifiers defined in the document.
	for _, specialID := range []string{"NONE", "NOASSERTION"} {
		doc := &spdx.Document{
			SPDXVersion:    spdx.Version,
			DataLicense:    spdx.DataLicense,
			SPDXIdentifier: common.ElementID("DOCUMENT"),
			CreationInfo:   &spdx.CreationInfo{},
			Packages: []*spdx.Package{
				{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
			},
			Relationships: []*spdx.Relationship{
				{
					RefA:         common.MakeDocElementID("", "DOCUMENT"),
					RefB:         common.MakeDocElementID("", "p1"),
					Relationship: "DESCRIBES",
				},
				// special value on the right-hand side; valid per the spec
				{
					RefA:         common.MakeDocElementID("", "p1"),
					RefB:         common.MakeDocElementSpecial(specialID),
					Relationship: "CONTAINS",
				},
			},
		}

		if err := ValidateDocument(doc); err != nil {
			t.Fatalf("expected nil error for RefB %s, got: %s", specialID, err.Error())
		}
	}
}

func TestDocumentWithSpecialRelationshipFailsValidation(t *testing.T) {
	newDoc := func(refA, refB common.DocElementID) *spdx.Document {
		return &spdx.Document{
			SPDXVersion:    spdx.Version,
			DataLicense:    spdx.DataLicense,
			SPDXIdentifier: common.ElementID("DOCUMENT"),
			CreationInfo:   &spdx.CreationInfo{},
			Packages: []*spdx.Package{
				{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
			},
			Relationships: []*spdx.Relationship{
				{RefA: refA, RefB: refB, Relationship: "CONTAINS"},
			},
		}
	}

	tests := map[string]*spdx.Document{
		// a special value is not permitted on the left-hand side
		"NONE on left side": newDoc(
			common.MakeDocElementSpecial("NONE"),
			common.MakeDocElementID("", "p1"),
		),
		"NOASSERTION on left side": newDoc(
			common.MakeDocElementSpecial("NOASSERTION"),
			common.MakeDocElementID("", "p1"),
		),
		// only NONE and NOASSERTION are valid special values
		"unknown special value on right side": newDoc(
			common.MakeDocElementID("", "p1"),
			common.MakeDocElementSpecial("SOMETHINGELSE"),
		),
	}

	for name, doc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateDocument(doc); err == nil {
				t.Fatalf("expected non-nil error, got nil")
			}
		})
	}
}

// A null entry in packages, files or relationships reaches ValidateDocument.
// The JSON reader already tolerates one (see json/test_fixtures/
// spdx2_3_null_package.json), so a document that parses can still arrive here
// with a nil element and used to dereference it.
func TestDocumentWithNilEntriesPassesValidation(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages: []*spdx.Package{
			nil,
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
		},
		Files: []*spdx.File{
			nil,
			{FileName: "file1", FileSPDXIdentifier: "f1"},
		},
		Relationships: []*spdx.Relationship{
			nil,
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: "DESCRIBES",
			},
		},
	}

	if err := ValidateDocument(doc); err != nil {
		t.Fatalf("expected nil error, got: %s", err.Error())
	}
}

// The nil entries must not hide a genuinely dangling reference either.
func TestDocumentWithNilEntriesStillCatchesBadReference(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages:       []*spdx.Package{nil},
		Relationships: []*spdx.Relationship{
			nil,
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("", "p404"),
				Relationship: "DESCRIBES",
			},
		},
	}

	if err := ValidateDocument(doc); err == nil {
		t.Fatal("expected an error for a relationship pointing at a missing element")
	}
}

// Snippets are elements defined in the document, so relationships referencing
// them must resolve through the same element index as packages and files.
func TestDocumentWithSnippetRelationshipPassesValidation(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Files: []*spdx.File{
			{FileName: "file1", FileSPDXIdentifier: "f1"},
		},
		Snippets: []spdx.Snippet{
			{SnippetSPDXIdentifier: "s1", SnippetFromFileSPDXIdentifier: "f1"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "s1"),
				RefB:         common.MakeDocElementSpecial("NOASSERTION"),
				Relationship: "GENERATED_FROM",
			},
			{
				RefA:         common.MakeDocElementID("", "f1"),
				RefB:         common.MakeDocElementID("", "s1"),
				Relationship: "CONTAINS",
			},
		},
	}

	if err := ValidateDocument(doc); err != nil {
		t.Fatalf("expected nil error, got: %s", err.Error())
	}
}

// A relationship endpoint that looks like a snippet ID but is never defined
// must still be rejected, rather than being treated as a trusted reference.
func TestDocumentWithDanglingSnippetReferenceFailsValidation(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages: []*spdx.Package{
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "s404"),
				Relationship: "CONTAINS",
			},
		},
	}

	if err := ValidateDocument(doc); err == nil {
		t.Fatal("expected an error for a relationship pointing at an undefined snippet")
	}
}

// A DocumentRef- endpoint is valid only when the External Document Reference
// is declared; the external element itself cannot be resolved locally.
func TestDocumentWithExternalDocumentReferencePassesValidation(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		ExternalDocumentReferences: []spdx.ExternalDocumentRef{
			{
				DocumentRefID: "external-doc",
				URI:           "http://example.org/external.spdx",
				Checksum:      common.Checksum{Algorithm: common.SHA1, Value: "d6a770ba38583ed4bb4525bd96e50461655d2759"},
			},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("external-doc", "external-pkg"),
				Relationship: "COPY_OF",
			},
			// DocumentRef- endpoints are also legitimate on the left side.
			{
				RefA:         common.MakeDocElementID("external-doc", "external-pkg"),
				RefB:         common.MakeDocElementSpecial("NONE"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	if err := ValidateDocument(doc); err != nil {
		t.Fatalf("expected nil error, got: %s", err.Error())
	}
}

// A DocumentRef- that no External Document Reference declares is a dangling
// reference and must make the document fail validation.
func TestDocumentWithUndeclaredDocumentRefFailsValidation(t *testing.T) {
	tests := map[string]struct {
		refA common.DocElementID
		refB common.DocElementID
	}{
		"undeclared DocumentRef on right side": {
			refA: common.MakeDocElementID("", "DOCUMENT"),
			refB: common.MakeDocElementID("missing-doc", "external-pkg"),
		},
		"undeclared DocumentRef on left side": {
			refA: common.MakeDocElementID("missing-doc", "external-pkg"),
			refB: common.MakeDocElementSpecial("NONE"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			doc := &spdx.Document{
				SPDXVersion:    spdx.Version,
				DataLicense:    spdx.DataLicense,
				SPDXIdentifier: common.ElementID("DOCUMENT"),
				CreationInfo:   &spdx.CreationInfo{},
				Relationships: []*spdx.Relationship{
					{RefA: tc.refA, RefB: tc.refB, Relationship: "COPY_OF"},
				},
			}

			if err := ValidateDocument(doc); err == nil {
				t.Fatal("expected an error for an undeclared DocumentRef-")
			}
		})
	}
}

// A DocumentRef- that is declared but points at the wrong reference must not
// be silently accepted: only the declared DocumentRef- ID is resolvable.
func TestDocumentWithMismatchedDocumentRefFailsValidation(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		ExternalDocumentReferences: []spdx.ExternalDocumentRef{
			{DocumentRefID: "declared-doc"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.MakeDocElementID("", "DOCUMENT"),
				RefB:         common.MakeDocElementID("other-doc", "external-pkg"),
				Relationship: "COPY_OF",
			},
		},
	}

	if err := ValidateDocument(doc); err == nil {
		t.Fatal("expected an error for a DocumentRef- that does not match a declaration")
	}
}
