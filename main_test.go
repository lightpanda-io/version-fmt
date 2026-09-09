// Copyright 2026 Lightpanda (Selecy SAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestLastVersions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		prev    string
		version string
		want    []string
	}{
		{
			name:    "no previous entry",
			prev:    "",
			version: "1.0.0-nightly.3+c",
			want:    []string{},
		},
		{
			name:    "first previous version",
			prev:    `{"version":"1.0.0-nightly.2+b"}`,
			version: "1.0.0-nightly.3+c",
			want:    []string{"1.0.0-nightly.2+b"},
		},
		{
			name:    "push at the head",
			prev:    `{"version":"1.0.0-nightly.2+b","last":["1.0.0-nightly.1+a"]}`,
			version: "1.0.0-nightly.3+c",
			want:    []string{"1.0.0-nightly.2+b", "1.0.0-nightly.1+a"},
		},
		{
			name:    "drop the oldest",
			prev:    `{"version":"6","last":["5","4","3","2","1"]}`,
			version: "7",
			want:    []string{"6", "5", "4", "3", "2"},
		},
		{
			name:    "same version twice",
			prev:    `{"version":"3","last":["2","1"]}`,
			version: "3",
			want:    []string{"2", "1"},
		},
		{
			name:    "version already in the list",
			prev:    `{"version":"3","last":["2","1"]}`,
			version: "1",
			want:    []string{"3", "2"},
		},
		{
			name:    "duplicate in the list",
			prev:    `{"version":"3","last":["2","2","1"]}`,
			version: "4",
			want:    []string{"3", "2", "1"},
		},
		{
			name:    "empty previous version",
			prev:    `{"last":["2","1"]}`,
			version: "3",
			want:    []string{"2", "1"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			last, err := lastVersions(json.RawMessage(tc.prev), tc.version)
			if err != nil {
				t.Fatalf("lastVersions: %v", err)
			}
			if !slices.Equal(last, tc.want) {
				t.Errorf("got %v, want %v", last, tc.want)
			}
		})
	}
}

func TestLastVersionsBadEntry(t *testing.T) {
	if _, err := lastVersions(json.RawMessage(`{"version":42}`), "1"); err == nil {
		t.Error("want an error on a bad previous entry")
	}
}
