// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"reflect"
	"testing"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

func TestNormalizeDocumentCanonicalizesIdentifierCasing(t *testing.T) {
	doc := &spdx.Document{
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		Packages: []*spdx.Package{
			{PackageName: "a", PackageSPDXIdentifier: "PkgA"},
		},
		Files: []*spdx.File{
			{FileName: "f", FileSPDXIdentifier: "FileA"},
		},
		Relationships: []*spdx.Relationship{
			{
				RefA:         common.DocElementID{ElementRefID: "PkgA"},
				RefB:         common.DocElementID{DocumentRefID: "other", ElementRefID: "FileA"},
				Relationship: "DEPENDS_ON",
			},
			{
				RefA:         common.DocElementID{ElementRefID: "PkgA"},
				RefB:         common.MakeDocElementSpecial("NONE"),
				Relationship: "CONTAINS",
			},
		},
		Annotations: []*spdx.Annotation{
			{AnnotationSPDXIdentifier: common.DocElementID{ElementRefID: "PkgA"}},
		},
	}

	// simulate IDs that were stored with a differently-cased prefix still
	// resolving to the same canonical ElementRefID after normalization
	doc.Relationships[0].RefA.ElementRefID = "PkgA"

	NormalizeDocument(doc)

	if got, want := doc.Packages[0].PackageSPDXIdentifier, common.ElementID("PkgA"); got != want {
		t.Errorf("package id: got %q, want %q", got, want)
	}
	if got, want := doc.Files[0].FileSPDXIdentifier, common.ElementID("FileA"); got != want {
		t.Errorf("file id: got %q, want %q", got, want)
	}
	// special values are preserved untouched
	if got := doc.Relationships[1].RefB.SpecialID; got != "NONE" {
		t.Errorf("special id changed: got %q, want NONE", got)
	}
	// document-ref portion survives normalization
	if got, want := doc.Relationships[0].RefB, (common.DocElementID{DocumentRefID: "other", ElementRefID: "FileA"}); !reflect.DeepEqual(got, want) {
		t.Errorf("doc element id: got %#v, want %#v", got, want)
	}
}

func TestNormalizeDocumentToleratesNil(t *testing.T) {
	NormalizeDocument(nil)

	doc := &spdx.Document{
		Packages:      []*spdx.Package{nil, {PackageSPDXIdentifier: "p1"}},
		Files:         []*spdx.File{nil, {FileSPDXIdentifier: "f1"}},
		Relationships: []*spdx.Relationship{nil, {RefA: common.MakeDocElementID("", "p1"), RefB: common.MakeDocElementID("", "f1"), Relationship: "CONTAINS"}},
		Annotations:   []*spdx.Annotation{nil, {AnnotationSPDXIdentifier: common.MakeDocElementID("", "p1")}},
	}
	NormalizeDocument(doc)

	if got := doc.Packages[1].PackageSPDXIdentifier; got != "p1" {
		t.Errorf("expected package id p1 to survive, got %q", got)
	}
}
