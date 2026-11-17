// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later
package spdxlib

import (
	"sort"
	"strings"

	"github.com/spdx/tools-golang/spdx/v2/common"
)

// ValidationError is a single, normalized problem found while validating a
// document.
//
// The fields are intentionally plain and comparable so that validation results
// produced by different readers (JSON, YAML, tag-value) for the same logical
// document can be compared with reflect.DeepEqual and de-duplicated
// deterministically.
type ValidationError struct {
	// Element is the normalized identifier of the element the problem refers
	// to, rendered without its SPDXRef- prefix (e.g. "DOCUMENT", "p1"). It is
	// empty for problems that are not tied to a single element.
	Element common.ElementID `json:"element,omitempty"`

	// Reference is the full rendered identifier that triggered the problem,
	// including any DocumentRef-/SPDXRef- prefix or special value. It is empty
	// when Element already carries enough context.
	Reference string `json:"reference,omitempty"`

	// Kind identifies the category of the problem using a stable, reader
	// independent vocabulary (see the ValidationErrorKind constants).
	Kind ValidationErrorKind `json:"kind"`

	// Message is the human-readable description of the problem.
	Message string `json:"message"`
}

// ValidationErrorKind is the stable category of a ValidationError. Using a
// fixed vocabulary instead of free-form message text keeps validation
// results comparable across readers and resilient to wording changes.
type ValidationErrorKind string

const (
	// ValidationErrorMissingElement means a relationship references an
	// element that is not defined in the document.
	ValidationErrorMissingElement ValidationErrorKind = "missing-element"

	// ValidationErrorInvalidSpecialID means a special value (other than the
	// permitted NONE / NOASSERTION) was used where an element reference is
	// required.
	ValidationErrorInvalidSpecialID ValidationErrorKind = "invalid-special-id"

	// ValidationErrorSpecialIDOnLeft means NONE or NOASSERTION was used on
	// the left-hand side of a relationship, where it is not permitted.
	ValidationErrorSpecialIDOnLeft ValidationErrorKind = "special-id-on-left"
)

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Message
}

// ValidationErrors is a set of ValidationErrors. It is kept in a stable,
// sorted order so that the same logical document always produces an identical
// result regardless of which reader parsed it or the order relationships were
// encountered in.
type ValidationErrors []ValidationError

// Error implements the error interface, joining the normalized messages in
// stable order.
func (errs ValidationErrors) Error() string {
	messages := make([]string, 0, len(errs))
	for _, e := range errs {
		messages = append(messages, e.Message)
	}
	return strings.Join(messages, "; ")
}

// sort canonicalizes the ordering of the error set. Errors are ordered by
// Kind, Reference, Element and finally Message, which is fully deterministic
// and independent of the order in which problems were discovered.
func (errs ValidationErrors) sort() {
	sort.SliceStable(errs, func(i, j int) bool {
		if errs[i].Kind != errs[j].Kind {
			return errs[i].Kind < errs[j].Kind
		}
		if errs[i].Reference != errs[j].Reference {
			return errs[i].Reference < errs[j].Reference
		}
		if errs[i].Element != errs[j].Element {
			return errs[i].Element < errs[j].Element
		}
		return errs[i].Message < errs[j].Message
	})
}

// Deduplicate removes identical errors (same Kind, Element, Reference and
// Message) and returns the result in stable sorted order.
func (errs ValidationErrors) Deduplicate() ValidationErrors {
	seen := make(map[ValidationError]bool, len(errs))
	unique := make(ValidationErrors, 0, len(errs))
	for _, e := range errs {
		if seen[e] {
			continue
		}
		seen[e] = true
		unique = append(unique, e)
	}
	unique.sort()
	return unique
}

// Has reports whether the error set contains an error of the given kind.
func (errs ValidationErrors) Has(kind ValidationErrorKind) bool {
	for _, e := range errs {
		if e.Kind == kind {
			return true
		}
	}
	return false
}
