// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package filter implements a message filter that may modify or block points.
package filter

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"regexp/syntax"
	"strconv"
)

var Debug *log.Logger

// Point represents a filterable data point.
type Point interface {
	// Metric returns the current value of the Metric field.
	Metric() []byte

	// Tag returns the current value of the tag identified by the given key.
	// Tag returns nil if the tag is absent.
	Tag(key []byte) []byte

	// SetMetric sets the Metric field, expanding ${N} escapes.
	SetMetric([]byte) error

	// SetTags sets value of each tag identified by the provided key, value pairs.
	SetTags(keyval ...[]byte) error
}

// Filter filters time series data points.
type Filter struct {
	rules    []*rule
	submatch submatch
}

// New creates a new Filter based on the given rules. An error is returned if
// the provided ruleset is invalid.
func New(rules ...Rule) (*Filter, error) {
	if err := validateRules(rules...); err != nil {
		return nil, err
	}
	filter := new(Filter)
	for _, r := range rules {
		filter.rules = append(filter.rules, newRule(r))
	}
	return filter, nil
}

// Eval evaluates the given point. It returns true if the point is accepted.
// The original point data may be lost due to calls to SetMetric and SetTags.
// If the original version is required, make a copy before calling Eval.
func (f *Filter) Eval(point Point) (bool, error) {
	if Debug != nil {
		Debug.Print("evaluate ", point)
	}
	submatch := &f.submatch
	for _, rule := range f.rules {
		submatch.Reset()
		if !rule.Match(point, submatch) {
			continue
		}
		if err := rule.Rewrite(point, submatch); err != nil {
			if Debug != nil {
				Debug.Printf("   rewriteError %v, rule=%v", point, rule)
			}
			return false, err
		}
		if rule.Block != nil {
			if Debug != nil {
				outcome := "pass"
				if *rule.Block {
					outcome = "block"
				}
				Debug.Printf("   %v %v, rule=%v", outcome, point, rule)
			}
			return !*rule.Block, nil
		}
	}
	if Debug != nil {
		Debug.Printf("    pass %v", point)
	}
	return true, nil
}

// rule is an execution engine for a Rule.
type rule struct {
	Rule
	MatchMetric *regexp.Regexp
	MatchTag    []tagMatch
	SetMetric   []byte
	SetTag      [][]byte
	Block       *bool
	// Scratch bufs to avoid allocs.
	metricBuf [128]byte
	tagsBuf   [8][]byte
	tagvBuf   [128]byte
}

func newRule(config Rule) *rule {
	r := &rule{
		Rule: config,
	}
	if len(config.Match) > 0 && config.Match[0] != "" {
		r.MatchMetric = regexp.MustCompile(config.Match[0])
	}
	for i := 1; i < len(config.Match); i += 2 {
		tagk := config.Match[i]
		tagv := config.Match[i+1]
		r.MatchTag = append(r.MatchTag, tagMatch{
			Key:   []byte(tagk),
			Value: regexp.MustCompile(tagv),
		})
	}
	if config.Block != nil {
		r.Block = config.Block
	}
	if len(config.Set) > 0 && config.Set[0] != "" {
		r.SetMetric = []byte(config.Set[0])
	}
	for i := 1; i < len(config.Set); i += 2 {
		tagk := config.Set[i]
		tagv := config.Set[i+1]
		r.SetTag = append(r.SetTag, []byte(tagk), []byte(tagv))
	}
	return r
}

var emptyString = []byte{}

// Match tests whether an point is matched by the rule.
// If the rule uses submatches, their matches are written to submatch.
func (r *rule) Match(point Point, submatch *submatch) bool {
	// Check for metric match.
	if re := r.MatchMetric; re != nil {
		src := point.Metric()
		match := re.FindSubmatchIndex(src)
		if match == nil {
			return false
		}
		if submatch.IsZero() && re.NumSubexp() > 0 {
			submatch.Set(re, src, match)
		}
	}
	// Check for tags match.
	for _, tm := range r.MatchTag {
		src := point.Tag(tm.Key)
		if src == nil {
			// If a tag with the given key is absent, the regex is
			// passed an empty string.
			src = emptyString
		}
		re := tm.Value
		match := re.FindSubmatchIndex(src)
		if match == nil {
			return false
		}
		if submatch.IsZero() && re.NumSubexp() > 0 {
			submatch.Set(re, src, match)
		}
	}
	return true
}

// Rewrite modifies the given point according to rule's Set* actions.
// Submatch references are expanded using the provided submatch data.
func (r *rule) Rewrite(point Point, submatch *submatch) error {
	if r.SetMetric != nil {
		metric := submatch.Expand(r.metricBuf[:0], r.SetMetric)
		if err := point.SetMetric(metric); err != nil {
			return err
		}
	}
	if r.SetTag != nil {
		tags := r.tagsBuf[:0]
		tagvBuf := r.tagvBuf[:]
		for i, elem := range r.SetTag {
			if i%2 == 0 {
				tags = append(tags, elem)
			} else {
				tagv := submatch.Expand(tagvBuf[:0], elem)
				if len(tagv) <= len(tagvBuf) {
					tagvBuf = tagvBuf[len(tagv):]
				}
				tags = append(tags, tagv)
			}
		}
		if err := point.SetTags(tags...); err != nil {
			return err
		}
	}
	return nil
}

// submatch represents a set of regexp subexpression matches.
type submatch struct {
	re     *regexp.Regexp
	src    []byte
	match  []int
	srcBuf [128]byte
}

// Reset clears submatch data.
func (s *submatch) Reset() {
	s.re = nil
}

// IsZero checks if any submatch data is present.
func (s *submatch) IsZero() bool {
	return s.re == nil
}

// Set provides submatch data.
func (s *submatch) Set(re *regexp.Regexp, src []byte, match []int) {
	s.re = re
	s.src = append(s.srcBuf[:0], src...)
	s.match = match
}

// Expand appends to dst a an execution of the provided template, and returns
// the resulting slice. The template's ${N} actions are expanded using the
// submatch data provided to Set.
func (s *submatch) Expand(dst []byte, template []byte) []byte {
	if s.IsZero() {
		return append(dst, template...)
	}
	return s.re.Expand(dst, template, s.src, s.match)
}

type tagMatch struct {
	Key   []byte
	Value *regexp.Regexp
}

func validateRules(rules ...Rule) error {
	if len(rules) == 0 {
		return errors.New("no rules defined")
	}
	for _, r := range rules {
		if err := r.validate(); err != nil {
			return fmt.Errorf("%v, rule=%v", err, r)
		}
	}
	return nil
}

// Rule represents configuration of a single rule.
type Rule struct {
	Match []string `json:",omitempty"`
	Set   []string `json:",omitempty"`
	Block *bool    `json:",omitempty"`
}

var submatchRE = regexp.MustCompile(`\${[0-9]+}`)

func (r Rule) isBlock() bool {
	return r.Block != nil && *r.Block;
}

func (r Rule) validate() error {
	if r.Block == nil {
		noop := false
		switch {
		case len(r.Set) == 3 && r.Set[0] == "" && r.Set[2] == "":
			noop = true
		case len(r.Set) == 2 && r.Set[0] == "":
			noop = true
		case len(r.Set) == 1 && r.Set[0] == "":
			noop = true
		case len(r.Set) == 0:
			noop = true
		}
		if noop {
			return errors.New("rule is a no-op, either Set or Block must be used")
		}
	}

	if r.isBlock() && len(r.Set) > 0 {
		return fmt.Errorf("Set and Block=true used together")
	}

	if len(r.Match) > 1 && (len(r.Match)-1)%2 != 0 {
		return fmt.Errorf("Match array has %d fields, expect %d or %d", len(r.Match),
			len(r.Match)-1, len(r.Match)+1)
	}

	nsubs := 0
	maxCap := -1
	for i, s := range r.Match {
		re, err := syntax.Parse(s, syntax.Perl)
		if err != nil {
			return err
		}
		maxCap = re.MaxCap()
		if maxCap > 0 {
			if (i-1)%2 == 0 {
				return errors.New("regex submatch used in tag name context")
			}
			nsubs++
		}
	}

	if nsubs > 1 {
		return fmt.Errorf("too many regexes with subexpressions: want 0 or 1")
	}

	if len(r.Set) > 0 && r.Set[0] != "" {
		var err error
		submatchRE.ReplaceAllStringFunc(r.Set[0], func(s string) string {
			if err == nil {
				err = validateSubmatch(s, maxCap)
			}
			return ""
		})
		if err != nil {
			return fmt.Errorf("invalid SetMetric: %v", err)
		}
	}

	if len(r.Set) > 1 {
		setTag := r.Set[1:]

		if len(setTag) > 0 && len(setTag)%2 != 0 {
			return fmt.Errorf("SetTag array has %d elements, expect %d or %d", len(setTag),
				len(setTag)-1, len(setTag)+1)
		}

		for i := 0; i < len(setTag); i += 2 {
			tagk := setTag[i]
			tagv := setTag[i+1]
			var err error
			submatchRE.ReplaceAllStringFunc(tagk, func(s string) string {
				if err == nil {
					err = fmt.Errorf("invalid tag name: submatch expansion used")
				}
				return ""
			})
			submatchRE.ReplaceAllStringFunc(tagv, func(s string) string {
				if err == nil {
					err = validateSubmatch(s, maxCap)
				}
				return ""
			})
			if err != nil {
				return fmt.Errorf("invalid SetTag: %v", err)
			}
		}
	}

	return nil
}

// validateSubmatch validates the submatch string in the "${N}" format.
// N is known not to exceed maxCap.
func validateSubmatch(s string, maxCap int) error {
	if len(s) < 4 {
		return fmt.Errorf("submatch string too short: %q", s)
	}
	if !(s[0] == '$' && s[1] == '{' && s[len(s)-1] == '}') {
		return fmt.Errorf("submatch string invalid: %q", s)
	}
	indexText := s[2 : len(s)-1]
	index, err := strconv.ParseInt(indexText, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid submatch number: %v", err)
	}
	switch {
	case int(index) < 1:
		return fmt.Errorf("submatch number out of range: want at least ${1}")
	case int(index) > maxCap:
		want := fmt.Sprintf("at most ${%d}", maxCap)
		if maxCap == 0 {
			want = "none"
		}
		return fmt.Errorf("submatch number out of range: want %s", want)
	case int(index) > 9:
		return fmt.Errorf("submatch number out of range: want at most ${9}")
	}
	return nil
}

func (r Rule) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "{")
	sep := ""
	if r.Match != nil {
		fmt.Fprintf(buf, "%sMatch:%q", sep, r.Match)
		sep = " "
	}
	if r.Set != nil {
		fmt.Fprintf(buf, "%sSet:%q", sep, r.Set)
		sep = " "
	}
	if r.Block != nil {
		fmt.Fprintf(buf, "%sBlock:%v", sep, *r.Block)
		sep = " "
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}
