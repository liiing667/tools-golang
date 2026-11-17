// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"fmt"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

// ValidateDocument returns an error if the Document is found to be invalid, or nil if the Document is valid.
// Currently, this only verifies that all Element IDs mentioned in Relationships exist in the Document as either a
// Package or an UnpackagedFile.
//
// The returned error is a ValidationErrors value: it contains every detected
// problem (rather than failing on the first one), normalized and ordered
// deterministically. Because the identifiers are normalized through the shared
// common ID parser and the error set is sorted and de-duplicated, validating
// the same logical document after reading it as JSON, YAML or tag-value yields
// an identical result.
func ValidateDocument(doc *spdx.Document) error {
	errs := ValidateDocumentErrors(doc)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ValidateDocumentErrors validates the document and returns its full set of
// normalized, stably sorted and de-duplicated validation errors. An empty
// slice means the document is valid.
func ValidateDocumentErrors(doc *spdx.Document) ValidationErrors {
	// cache a map of package IDs for quick lookups. Identifiers are normalized
	// through the shared common parser so that casing differences in the
	// SPDXRef-/DocumentRef- prefixes cannot create false mismatches.
	validElementIDs := make(map[common.ElementID]bool)
	for _, docPackage := range doc.Packages {
		if docPackage == nil {
			continue
		}
		validElementIDs[normalizeElementID(docPackage.PackageSPDXIdentifier)] = true
	}

	for _, unpackagedFile := range doc.Files {
		if unpackagedFile == nil {
			continue
		}
		validElementIDs[normalizeElementID(unpackagedFile.FileSPDXIdentifier)] = true
	}

	// add the Document element ID
	validElementIDs[common.MakeDocElementID("", "DOCUMENT").ElementRefID] = true

	var errs ValidationErrors
	for _, relationship := range doc.Relationships {
		if relationship == nil {
			continue
		}

		// The left side must always resolve to a defined element. NONE and
		// NOASSERTION are only permitted on the right side (SPDX 2.3 section
		// 11.1), so a special value here is invalid.
		if relationship.RefA.SpecialID != "" {
			errs = append(errs, ValidationError{
				Reference: common.RenderDocElementID(relationship.RefA),
				Kind:      ValidationErrorSpecialIDOnLeft,
				Message:   fmt.Sprintf("%s is not allowed on the left side of a relationship", relationship.RefA.SpecialID),
			})
		} else if !validElementIDs[normalizeElementID(relationship.RefA.ElementRefID)] {
			errs = append(errs, missingElementError(relationship.RefA.ElementRefID))
		}

		// The right side may be a defined element or one of the permitted
		// special values NONE or NOASSERTION.
		if relationship.RefB.SpecialID != "" {
			if relationship.RefB.SpecialID != "NONE" && relationship.RefB.SpecialID != "NOASSERTION" {
				errs = append(errs, ValidationError{
					Reference: common.RenderDocElementID(relationship.RefB),
					Kind:      ValidationErrorInvalidSpecialID,
					Message:   fmt.Sprintf("%s is not a valid relationship reference", relationship.RefB.SpecialID),
				})
			}
		} else if !validElementIDs[normalizeElementID(relationship.RefB.ElementRefID)] {
			errs = append(errs, missingElementError(relationship.RefB.ElementRefID))
		}
	}

	return errs.Deduplicate()
}

// normalizeElementID routes an identifier through the shared ID parser so that
// any casing of the SPDXRef- prefix is normalized. Identifiers that are
// already stored without their prefix (the in-memory representation) are
// passed through unchanged.
func normalizeElementID(id common.ElementID) common.ElementID {
	parsed, err := common.ParseElementID(common.RenderElementID(id))
	if err != nil {
		return id
	}
	return parsed
}

// missingElementError builds the canonical error for a relationship that
// references an undefined element.
func missingElementError(id common.ElementID) ValidationError {
	rendered := common.RenderElementID(id)
	return ValidationError{
		Element:   id,
		Reference: rendered,
		Kind:      ValidationErrorMissingElement,
		Message:   fmt.Sprintf("%s used in relationship but no such package exists", rendered),
	}
}
