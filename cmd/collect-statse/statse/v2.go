// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"bytes"
	"errors"
	"fmt"
)

// commonTagCount is used to avoid heap allocations in most runs.
const commonTagCount = 4

func (ev *Event) parseV2(s []byte) error {
	// Field 1: class name.
	if !bytes.HasPrefix(s, []byte("EVENT|")) {
		return fmt.Errorf("class value invalid")
	}
	s = s[len("EVENT|"):]

	// Field 2: metric name.
	i := bytes.IndexByte(s, '|')
	if i == -1 {
		return fmt.Errorf("missing metric field")
	}
	ev.Metric = string(s[:i])
	if ev.Metric == "" {
		return fmt.Errorf("invalid metric value: empty string")
	}
	s = s[i+1:]

	// Field 3: tags in key=value format
	i = bytes.IndexByte(s, '|')
	if i == -1 {
		return fmt.Errorf("missing tags field")
	}
	tags := s[:i]
	s = s[i+1:]
	if err := validateTags(tags); err != nil {
		return fmt.Errorf("invalid tag: %v", err)
	}
	ev.Tags = string(tags)

	// Field 4: named statistics
	ev.parseStatistics(s, false)

	return nil
}

func validateTags(s []byte) error {
	have := make([][]byte, 0, commonTagCount)
	for len(s) > 0 {
		var tag []byte
		if i := bytes.IndexByte(s, ' '); i >= 0 {
			tag = s[:i]
			s = s[i+1:]
		} else {
			tag = s
			s = s[:0]
		}

		// Check basic syntax.
		i := bytes.IndexByte(tag, '=')
		if i == -1 {
			return errors.New("missing separator character: =")
		}
		tagk := tag[:i]

		// Check if duplicate.
		dup := false
		for _, k := range have {
			if bytes.Equal(tagk, k) {
				dup = true
				break
			}
		}
		if dup {
			return fmt.Errorf("duplicate tag: %s", tagk)
		}
		have = append(have, tagk)
	}
	if len(have) > MaxTagsPerEvent {
		return fmt.Errorf("want at most %d tags", MaxTagsPerEvent)
	}
	return nil
}
