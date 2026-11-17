// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"reflect"
	"testing"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

func TestValidateDocumentErrorsIsStableAndSorted(t *testing.T) {
	doc := &spdx.Document{
		SPDXVersion:    spdx.Version,
		DataLicense:    spdx.DataLicense,
		SPDXIdentifier: common.ElementID("DOCUMENT"),
		CreationInfo:   &spdx.CreationInfo{},
		Packages: []*spdx.Package{
			{PackageName: "pkg1", PackageSPDXIdentifier: "p1"},
		},
		Relationships: []*spdx.Relationship{
			// declared in reverse order on purpose: the result must still be
			// sorted by the stable error keys rather than discovery order
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "z-missing"),
				Relationship: "DEPENDS_ON",
			},
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementID("", "a-missing"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	errs := ValidateDocumentErrors(doc)
	want := ValidationErrors{
		{
			Element:   "a-missing",
			Reference: "SPDXRef-a-missing",
			Kind:      ValidationErrorMissingElement,
			Message:   "SPDXRef-a-missing used in relationship but no such package exists",
		},
		{
			Element:   "z-missing",
			Reference: "SPDXRef-z-missing",
			Kind:      ValidationErrorMissingElement,
			Message:   "SPDXRef-z-missing used in relationship but no such package exists",
		},
	}

	if !reflect.DeepEqual(errs, want) {
		t.Fatalf("unexpected error set:\n got: %#v\nwant: %#v", errs, want)
	}

	// ValidateDocument must surface the same content through the error interface
	if err := ValidateDocument(doc); err == nil {
		t.Fatal("expected a non-nil error from ValidateDocument")
	} else if err.Error() != errs.Error() {
		t.Fatalf("error text mismatch:\n got: %s\nwant: %s", err.Error(), errs.Error())
	}
}

func TestValidationErrorsDeduplicate(t *testing.T) {
	errs := ValidationErrors{
		{Element: "p9", Reference: "SPDXRef-p9", Kind: ValidationErrorMissingElement, Message: "m"},
		{Element: "p1", Reference: "SPDXRef-p1", Kind: ValidationErrorMissingElement, Message: "m"},
		{Element: "p9", Reference: "SPDXRef-p9", Kind: ValidationErrorMissingElement, Message: "m"},
	}

	got := errs.Deduplicate()
	want := ValidationErrors{
		{Element: "p1", Reference: "SPDXRef-p1", Kind: ValidationErrorMissingElement, Message: "m"},
		{Element: "p9", Reference: "SPDXRef-p9", Kind: ValidationErrorMissingElement, Message: "m"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("deduplication mismatch:\n got: %#v\nwant: %#v", got, want)
	}

	if !got.Has(ValidationErrorMissingElement) {
		t.Fatal("expected Has to report the missing-element kind")
	}
	if got.Has(ValidationErrorInvalidSpecialID) {
		t.Fatal("did not expect an invalid-special-id kind")
	}
}

func TestValidationErrorKindsForSpecialIDs(t *testing.T) {
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
				RefA:         common.MakeDocElementSpecial("NONE"),
				RefB:         common.MakeDocElementID("", "p1"),
				Relationship: "DEPENDS_ON",
			},
			{
				RefA:         common.MakeDocElementID("", "p1"),
				RefB:         common.MakeDocElementSpecial("SOMETHINGELSE"),
				Relationship: "DEPENDS_ON",
			},
		},
	}

	errs := ValidateDocumentErrors(doc)
	kinds := map[ValidationErrorKind]bool{}
	for _, e := range errs {
		kinds[e.Kind] = true
	}
	if !kinds[ValidationErrorSpecialIDOnLeft] {
		t.Errorf("expected a %s error, got %#v", ValidationErrorSpecialIDOnLeft, errs)
	}
	if !kinds[ValidationErrorInvalidSpecialID] {
		t.Errorf("expected a %s error, got %#v", ValidationErrorInvalidSpecialID, errs)
	}
}
