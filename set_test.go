// Copyright 2020 YourBase Inc.
// SPDX-License-Identifier: BSD-3-Clause

package ini

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNilFileSet(t *testing.T) {
	fset := (FileSet)(nil)
	if got := fset.Get("foo", "bar"); got != "" {
		t.Errorf("Get(...) = %q; want empty", got)
	}
	if got := fset.Find("foo", "bar"); len(got) > 0 {
		t.Errorf("Find(...) = %q; want empty", got)
	}
	if got := fset.Sections(); len(got) > 0 {
		t.Errorf("Sections(...) = %q; want empty", got)
	}
	if fset.HasSections() {
		t.Error("HasSections() = true; want false")
	}
	if got := fset.Section("foo"); len(got) > 0 {
		t.Errorf("Section(...) = %q; want empty", got)
	}
}

func TestFileSetAccess(t *testing.T) {
	tests := []struct {
		name           string
		sources        []string
		section        string
		key            string
		wantGet        string
		wantValue      *Value
		wantFind       []string
		wantFindValues []*Value
	}{
		{
			name:    "ExistsInFirst",
			sources: []string{"FOO=bar\n", "BAZ=quux\n"},
			section: "",
			key:     "FOO",
			wantGet: "bar",
			wantValue: &Value{
				Value:    "bar",
				Filename: "ExistsInFirst_0",
				Line:     1,
			},
			wantFind: []string{"bar"},
			wantFindValues: []*Value{
				{
					Value:    "bar",
					Filename: "ExistsInFirst_0",
					Line:     1,
				},
			},
		},
		{
			name:    "ExistsInSecond",
			sources: []string{"FOO=bar\n", "BAZ=quux\n"},
			section: "",
			key:     "BAZ",
			wantGet: "quux",
			wantValue: &Value{
				Value:    "quux",
				Filename: "ExistsInSecond_1",
				Line:     1,
			},
			wantFind: []string{"quux"},
			wantFindValues: []*Value{
				{
					Value:    "quux",
					Filename: "ExistsInSecond_1",
					Line:     1,
				},
			},
		},
		{
			name:           "DoesNotExist",
			sources:        []string{"FOO=bar\n", "BAZ=quux\n"},
			section:        "",
			key:            "bork",
			wantGet:        "",
			wantValue:      nil,
			wantFind:       []string{},
			wantFindValues: nil,
		},
		{
			name:    "MultipleValues",
			sources: []string{"FOO=bar\n", "FOO=baz\n"},
			section: "",
			key:     "FOO",
			wantGet: "bar",
			wantValue: &Value{
				Value:    "bar",
				Filename: "MultipleValues_0",
				Line:     1,
			},
			wantFind: []string{"baz", "bar"},
			wantFindValues: []*Value{
				{
					Value:    "baz",
					Filename: "MultipleValues_1",
					Line:     1,
				},
				{
					Value:    "bar",
					Filename: "MultipleValues_0",
					Line:     1,
				},
			},
		},
		{
			name: "Section",
			sources: []string{
				"[foo]\n" +
					"bar=baz\n" +
					"[xyzzy]\n" +
					"bork=bork\n",
				"[foo]\n" +
					"something=else\n",
			},
			section: "foo",
			key:     "bar",
			wantGet: "baz",
			wantValue: &Value{
				Value:    "baz",
				Filename: "Section_0",
				Line:     2,
			},
			wantFind: []string{"baz"},
			wantFindValues: []*Value{
				{
					Value:    "baz",
					Filename: "Section_0",
					Line:     2,
				},
			},
		},
	}
	t.Run("Get", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var fset FileSet
				for i, src := range test.sources {
					f, err := Parse(strings.NewReader(src), &ParseOptions{
						Name: fmt.Sprintf("%s_%d", test.name, i),
					})
					if err != nil {
						t.Fatal(err)
					}
					fset = append(fset, f)
				}
				if got := fset.Get(test.section, test.key); got != test.wantGet {
					t.Errorf("fset.Get(%q, %q) = %q; want %q", test.section, test.key, got, test.wantGet)
				}
				if got := fset.Section(test.section).Get(test.key); got != test.wantGet {
					t.Errorf("fset.Section(%q).Get(%q) = %q; want %q", test.section, test.key, got, test.wantGet)
				}
			})
		}
	})
	t.Run("Value", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var fset FileSet
				for i, src := range test.sources {
					f, err := Parse(strings.NewReader(src), &ParseOptions{
						Name: fmt.Sprintf("%s_%d", test.name, i),
					})
					if err != nil {
						t.Fatal(err)
					}
					fset = append(fset, f)
				}
				got := fset.Value(test.section, test.key)
				if diff := cmp.Diff(test.wantValue, got); diff != "" {
					t.Errorf("fset.Value(%q, %q) (-want +got):\n%s", test.section, test.key, diff)
				}
			})
		}
	})
	t.Run("Find", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var fset FileSet
				for i, src := range test.sources {
					f, err := Parse(strings.NewReader(src), &ParseOptions{
						Name: fmt.Sprintf("%s_%d", test.name, i),
					})
					if err != nil {
						t.Fatal(err)
					}
					fset = append(fset, f)
				}
				got := fset.Find(test.section, test.key)
				if diff := cmp.Diff(test.wantFind, got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("fset.Find(%q, %q) (-want +got):\n%s", test.section, test.key, diff)
				}
				got = fset.Section(test.section)[test.key]
				if diff := cmp.Diff(test.wantFind, got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("fset.Section(%q)[%q] (-want +got):\n%s", test.section, test.key, diff)
				}
			})
		}
	})
	t.Run("FindValues", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var fset FileSet
				for i, src := range test.sources {
					f, err := Parse(strings.NewReader(src), &ParseOptions{
						Name: fmt.Sprintf("%s_%d", test.name, i),
					})
					if err != nil {
						t.Fatal(err)
					}
					fset = append(fset, f)
				}
				got := fset.FindValues(test.section, test.key)
				if diff := cmp.Diff(test.wantFindValues, got, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("fset.FindValues(%q, %q) (-want +got):\n%s", test.section, test.key, diff)
				}
			})
		}
	})
}

func TestFileSetSet(t *testing.T) {
	tests := []struct {
		name    string
		sources []string
		section string
		key     string
		value   string
		want    []string
	}{
		{
			name:    "AddToEmpty",
			sources: []string{""},
			section: "",
			key:     "foo",
			value:   "bar",
			want:    []string{"foo=bar\n"},
		},
		{
			name:    "AddSectionToEmpty",
			sources: []string{""},
			section: "foo",
			key:     "bar",
			value:   "baz",
			want:    []string{"[foo]\nbar=baz\n"},
		},
		{
			name:    "Overwrite",
			sources: []string{""},
			section: "",
			key:     "foo",
			value:   "xyzzy",
			want:    []string{"foo=xyzzy\n"},
		},
		{
			name:    "DeleteInLaterFiles",
			sources: []string{"", "; Comment 1\nfoo=bar\n; Comment 2\nfoo=baz\n"},
			section: "",
			key:     "foo",
			value:   "quux",
			want:    []string{"foo=quux\n", ""},
		},
		{
			name:    "AddToExistingSection",
			sources: []string{"", "foo=bar\n"},
			section: "",
			key:     "baz",
			value:   "quux",
			want:    []string{"baz=quux\n", "foo=bar\n"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var fset FileSet
			for _, src := range test.sources {
				var f *File
				if src != "" {
					var err error
					f, err = Parse(strings.NewReader(src), nil)
					if err != nil {
						t.Fatal(err)
					}
				}
				fset = append(fset, f)
			}

			fset.Set(test.section, test.key, test.value)

			got := make([]string, len(fset))
			for i, f := range fset {
				text, err := f.MarshalText()
				if err != nil {
					t.Fatal(err)
				}
				got[i] = string(text)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MarshalText (-want +got):\n%s", diff)
			}
		})
	}
}
