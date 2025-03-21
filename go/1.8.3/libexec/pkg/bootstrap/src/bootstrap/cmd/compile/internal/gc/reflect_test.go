// Do not edit. Bootstrap copy of /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/reflect_test.go

//line /tmp/go-20170524-43289-mangl1/go/src/cmd/compile/internal/gc/reflect_test.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"reflect"
	"sort"
	"testing"
)

func TestSortingByMethodNameAndPackagePath(t *testing.T) {
	data := []*Sig{
		&Sig{name: "b", pkg: &Pkg{Path: "abc"}},
		&Sig{name: "b", pkg: nil},
		&Sig{name: "c", pkg: nil},
		&Sig{name: "c", pkg: &Pkg{Path: "uvw"}},
		&Sig{name: "c", pkg: nil},
		&Sig{name: "b", pkg: &Pkg{Path: "xyz"}},
		&Sig{name: "a", pkg: &Pkg{Path: "abc"}},
		&Sig{name: "b", pkg: nil},
	}
	want := []*Sig{
		&Sig{name: "a", pkg: &Pkg{Path: "abc"}},
		&Sig{name: "b", pkg: nil},
		&Sig{name: "b", pkg: nil},
		&Sig{name: "b", pkg: &Pkg{Path: "abc"}},
		&Sig{name: "b", pkg: &Pkg{Path: "xyz"}},
		&Sig{name: "c", pkg: nil},
		&Sig{name: "c", pkg: nil},
		&Sig{name: "c", pkg: &Pkg{Path: "uvw"}},
	}
	if len(data) != len(want) {
		t.Fatal("want and data must match")
	}
	if reflect.DeepEqual(data, want) {
		t.Fatal("data must be shuffled")
	}
	sort.Sort(byMethodNameAndPackagePath(data))
	if !reflect.DeepEqual(data, want) {
		t.Logf("want: %#v", want)
		t.Logf("data: %#v", data)
		t.Errorf("sorting failed")
	}

}
