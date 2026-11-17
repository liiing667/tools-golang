// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package reader

import (
	"fmt"
	"strings"

	"github.com/spdx/tools-golang/spdx/v2/common"
)

// used to extract key / value from embedded substrings
// returns subkey, subvalue, nil if no error, or "", "", error otherwise
func extractSubs(value string) (string, string, error) {
	// parse the value to see if it's a valid subvalue format
	sp := strings.SplitN(value, ":", 2)
	if len(sp) == 1 {
		return "", "", fmt.Errorf("invalid subvalue format for %s (no colon found)", value)
	}

	subkey := strings.TrimSpace(sp[0])
	subvalue := strings.TrimSpace(sp[1])

	return subkey, subvalue, nil
}

// used to extract DocumentRef and SPDXRef values from an SPDX Identifier
// which can point either to this document or to a different one
func extractDocElementID(value string) (common.DocElementID, error) {
	return common.ParseDocElementID(value)
}

// used to extract SPDXRef values from an SPDX Identifier, OR "special" strings
// from a specified set of permitted values. The primary use case for this is
// the right-hand side of Relationships, where beginning in SPDX 2.2 the values
// "NONE" and "NOASSERTION" are permitted. If the value does not match one of
// the specified permitted values, it will fall back to the ordinary
// DocElementID extractor.
func extractDocElementSpecial(value string, permittedSpecial []string) (common.DocElementID, error) {
	return common.ParseDocElementSpecial(value, permittedSpecial)
}

// used to extract SPDXRef values only from an SPDX Identifier which can point
// to this document only. Use extractDocElementID for parsing IDs that can
// refer either to this document or to a different one.
func extractElementID(value string) (common.ElementID, error) {
	return common.ParseElementID(value)
}
