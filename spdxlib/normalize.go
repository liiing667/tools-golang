// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"github.com/spdx/tools-golang/spdx"
	"github.com/spdx/tools-golang/spdx/v2/common"
)

// NormalizeDocument applies the canonical, reader-independent normalization to
// an in-memory document:
//
//   - Element identifiers are routed through the shared common ID parser so the
//     casing of the mandatory SPDXRef-/DocumentRef- prefixes is canonicalized.
//
// It is safe to call on partially populated documents and nil entries are
// skipped, matching the tolerance of the individual readers.
//
// The relative order of the document's collections is intentionally preserved
// (matching the input document). Deterministic, comparable ordering is applied
// to the validation error set itself, where stable ordering is actually
// required, rather than mutating the parsed document.
func NormalizeDocument(doc *spdx.Document) {
	if doc == nil {
		return
	}

	doc.SPDXIdentifier = normalizeElementID(doc.SPDXIdentifier)

	for _, docPackage := range doc.Packages {
		if docPackage == nil {
			continue
		}
		docPackage.PackageSPDXIdentifier = normalizeElementID(docPackage.PackageSPDXIdentifier)
	}

	for _, file := range doc.Files {
		if file == nil {
			continue
		}
		file.FileSPDXIdentifier = normalizeElementID(file.FileSPDXIdentifier)
	}

	for _, relationship := range doc.Relationships {
		if relationship == nil {
			continue
		}
		relationship.RefA = normalizeDocElementID(relationship.RefA)
		relationship.RefB = normalizeDocElementID(relationship.RefB)
	}

	for _, annotation := range doc.Annotations {
		if annotation == nil {
			continue
		}
		annotation.AnnotationSPDXIdentifier = normalizeDocElementID(annotation.AnnotationSPDXIdentifier)
	}
}

// normalizeDocElementID canonicalizes a DocElementID by re-parsing its rendered
// form through the shared common parser. Special values (NONE / NOASSERTION)
// and empty identifiers are returned unchanged.
func normalizeDocElementID(id common.DocElementID) common.DocElementID {
	if id.SpecialID != "" {
		return id
	}
	if id.ElementRefID == "" && id.DocumentRefID == "" {
		return id
	}

	normalized, err := common.ParseDocElementID(common.RenderDocElementID(id))
	if err != nil {
		return id
	}
	return normalized
}
