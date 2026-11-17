// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package common

import (
	"reflect"
	"testing"
)

func TestParseElementIDCaseInsensitivePrefix(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  ElementID
		err   bool
	}{
		{name: "canonical", value: "SPDXRef-file1", want: "file1"},
		{name: "lowercase", value: "spdxref-file1", want: "file1"},
		{name: "mixed case", value: "SpdxRef-file1", want: "file1"},
		{name: "missing prefix", value: "file1", err: true},
		{name: "empty suffix", value: "SPDXRef-", err: true},
		{name: "lowercase empty suffix", value: "spdxref-", err: true},
		{name: "colon not allowed", value: "SPDXRef-file1:", err: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseElementID(tc.value)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error for %q, got %q", tc.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.value, err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseDocElementIDCaseInsensitivePrefix(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  DocElementID
		err   bool
	}{
		{
			name:  "canonical",
			value: "DocumentRef-doc2:SPDXRef-file2",
			want:  DocElementID{DocumentRefID: "doc2", ElementRefID: "file2"},
		},
		{
			name:  "lowercase prefixes",
			value: "documentref-doc2:spdxref-file2",
			want:  DocElementID{DocumentRefID: "doc2", ElementRefID: "file2"},
		},
		{
			name:  "mixed case prefixes",
			value: "Documentref-doc2:SpdxRef-file2",
			want:  DocElementID{DocumentRefID: "doc2", ElementRefID: "file2"},
		},
		{
			name:  "plain element lowercase",
			value: "spdxref-file1",
			want:  DocElementID{ElementRefID: "file1"},
		},
		{name: "document ref without colon", value: "DocumentRef-doc2", err: true},
		{name: "missing spdxref", value: "DocumentRef-doc2:file2", err: true},
		{name: "no prefix", value: "file1", err: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDocElementID(tc.value)
			if tc.err {
				if err == nil {
					t.Fatalf("expected error for %q, got %#v", tc.value, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.value, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseDocElementSpecial(t *testing.T) {
	permitted := []string{"NONE", "NOASSERTION"}

	got, err := ParseDocElementSpecial("NONE", permitted)
	if err != nil || got.SpecialID != "NONE" {
		t.Fatalf("expected NONE special id, got %#v err %v", got, err)
	}

	got, err = ParseDocElementSpecial("spdxref-file1", permitted)
	if err != nil || got.ElementRefID != "file1" {
		t.Fatalf("expected fallback element parse, got %#v err %v", got, err)
	}

	if _, err := ParseDocElementSpecial("BOGUS", permitted); err == nil {
		t.Fatal("expected an error for a value that is neither special nor a valid ID")
	}
}

func TestTrimDocumentRefPrefixFold(t *testing.T) {
	if rest, ok := TrimDocumentRefPrefix("documentref-abc"); !ok || rest != "abc" {
		t.Fatalf("got (%q, %v), want (abc, true)", rest, ok)
	}
	if _, ok := TrimDocumentRefPrefix("abc"); ok {
		t.Fatal("expected prefix detection to fail for a value without the prefix")
	}
}

func TestElementIDUnmarshalCaseInsensitivePrefix(t *testing.T) {
	for _, raw := range []string{`"SPDXRef-x"`, `"spdxref-x"`, `"SpDxReF-x"`} {
		var id ElementID
		if err := id.UnmarshalJSON([]byte(raw)); err != nil {
			t.Fatalf("unexpected error for %s: %v", raw, err)
		}
		if id != "x" {
			t.Fatalf("got %q for %s, want x", id, raw)
		}
	}
}
