// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"fmt"

	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

// ValidateDocument returns an error if the Document is found to be invalid, or nil if the Document is valid.
// It verifies that every element referenced by a Relationship can be resolved
// in a single element index made up of the Document's SPDXRef-DOCUMENT,
// Packages, unpackaged Files and Snippets, or, for elements prefixed with
// DocumentRef-, in an External Document Reference declared by the Document.
// NONE and NOASSERTION keep their SPDX 2.3 section 11.1 semantics: they are
// permitted only on the right-hand side of a relationship.
func ValidateDocument(doc *spdx.Document) error {
	// index every element defined in the present document for quick lookups
	validElementIDs := make(map[common.ElementID]bool)
	for _, docPackage := range doc.Packages {
		if docPackage == nil {
			continue
		}
		validElementIDs[docPackage.PackageSPDXIdentifier] = true
	}

	for _, unpackagedFile := range doc.Files {
		if unpackagedFile == nil {
			continue
		}
		validElementIDs[unpackagedFile.FileSPDXIdentifier] = true
	}

	for _, snippet := range doc.Snippets {
		validElementIDs[snippet.SnippetSPDXIdentifier] = true
	}

	// add the Document element ID
	validElementIDs[doc.SPDXIdentifier] = true

	// index the External Document References, without the DocumentRef- prefix,
	// so DocumentRef- endpoints can be matched against a declaration
	externalDocumentRefs := make(map[common.DocumentID]bool)
	for _, externalDocumentRef := range doc.ExternalDocumentReferences {
		externalDocumentRefs[externalDocumentRef.DocumentRefID] = true
	}

	for _, relationship := range doc.Relationships {
		if relationship == nil {
			continue
		}

		// The left side must always resolve to a defined element. NONE and
		// NOASSERTION are only permitted on the right side (SPDX 2.3 section
		// 11.1), so a special value here is invalid.
		if err := validateRelationshipEndpoint(relationship.RefA, validElementIDs, externalDocumentRefs, false); err != nil {
			return err
		}

		// The right side may be a defined element or one of the permitted
		// special values NONE or NOASSERTION.
		if err := validateRelationshipEndpoint(relationship.RefB, validElementIDs, externalDocumentRefs, true); err != nil {
			return err
		}
	}

	return nil
}

// validateRelationshipEndpoint resolves a relationship endpoint against the
// document's element index. Endpoints carrying a DocumentRef- portion must
// match a declared External Document Reference; the referenced element itself
// lives in that external document and therefore cannot be checked against the
// local element index. When allowSpecial is true, the special values NONE and
// NOASSERTION are accepted without further resolution.
func validateRelationshipEndpoint(elementID common.DocElementID, validElementIDs map[common.ElementID]bool, externalDocumentRefs map[common.DocumentID]bool, allowSpecial bool) error {
	if elementID.SpecialID != "" {
		if allowSpecial && (elementID.SpecialID == "NONE" || elementID.SpecialID == "NOASSERTION") {
			return nil
		}
		if allowSpecial {
			return fmt.Errorf("%s is not a valid relationship reference", elementID.SpecialID)
		}
		return fmt.Errorf("%s is not allowed on the left side of a relationship", elementID.SpecialID)
	}

	if elementID.DocumentRefID != "" {
		if !externalDocumentRefs[elementID.DocumentRefID] {
			return fmt.Errorf("%s used in relationship but no such external document reference exists", common.RenderDocElementID(elementID))
		}
		return nil
	}

	if !validElementIDs[elementID.ElementRefID] {
		return fmt.Errorf("%s used in relationship but no such element exists", common.RenderElementID(elementID.ElementRefID))
	}

	return nil
}
