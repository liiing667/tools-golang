// SPDX-License-Identifier: Apache-2.0 OR GPL-2.0-or-later

package common

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCaptureExtensionsKeepsOnlyUnknownFields(t *testing.T) {
	payload := []byte(`{"known": 1, "x-a": "two", "x-b": [1, 2]}`)
	known := map[string]struct{}{"known": {}}

	got, err := CaptureExtensions(payload, known)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := got["known"]; ok {
		t.Fatal("known field must not be captured as an extension")
	}
	if len(got) != 2 {
		t.Fatalf("expected two extension fields, got %#v", got)
	}
	if string(got["x-a"]) != `"two"` {
		t.Fatalf("unexpected x-a payload: %s", string(got["x-a"]))
	}
}

func TestCaptureExtensionsNilWhenNothingUnknown(t *testing.T) {
	payload := []byte(`{"known": 1}`)
	known := map[string]struct{}{"known": {}}

	got, err := CaptureExtensions(payload, known)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil extensions, got %#v", got)
	}
}

func TestMergeExtensionsInlinesUnknownFields(t *testing.T) {
	type model struct {
		Known string `json:"known"`
	}
	extensions := Extensions{
		"x-a": json.RawMessage(`"two"`),
		// modeled keys in extensions must not overwrite the model's value
		"known": json.RawMessage(`"shadow"`),
	}

	out, err := MergeExtensions(model{Known: "real"}, extensions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if string(got["x-a"]) != `"two"` {
		t.Fatalf("extension not inlined: %s", string(out))
	}
	if string(got["known"]) != `"real"` {
		t.Fatalf("modeled value must take precedence: %s", string(out))
	}
}

func TestMergeExtensionsWithoutExtensionsMatchesJSON(t *testing.T) {
	type model struct {
		Known string `json:"known"`
	}
	want, _ := json.Marshal(model{Known: "v"})

	got, err := MergeExtensions(model{Known: "v"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %s, want %s", string(got), string(want))
	}
}
