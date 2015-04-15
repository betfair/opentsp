// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"strings"
	"unicode"
)

// Clean inserts underscore in place of any character that is not storable in
// the database. Must be applied for any externally-provided metric name,
// tag key, or tag value.
//
// Note that Clean being a surjective function may cause data conflicts.
func Clean(s string) string {
	return strings.Map(toMarshal, s)
}

func toMarshal(r rune) rune {
	if r > unicode.MaxASCII {
		goto bad
	}
	if !unicode.IsPrint(r) {
		goto bad
	}
	if unicode.IsSpace(r) {
		goto bad
	}
	if isQuery(r) {
		goto bad
	}
	return r

bad:
	return '_'
}

func isQuery(r rune) bool {
	switch r {
	default:
		return false

	case '{', '}', '=', ',', '|', '*':
		return true
	}
}
