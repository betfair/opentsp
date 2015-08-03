package filter

import (
	"log"

	"opentsp.org/internal/tsdb"
)

type series struct {
	series tsdb.Series
	filter *Filter
}

func (s series) Next() *tsdb.Point {
	for {
		point := s.series.Next()
		pass, err := s.filter.Eval(point)
		if err != nil {
			log.Printf("tsdb: filter error: %v", err)
			point.Free()
			continue
		}
		if !pass {
			point.Free()
			continue
		}
		return point
	}
}

// Series returns a filtered version of the given time series.
func Series(rules []Rule, in tsdb.Series) tsdb.Series {
	filter, err := New(rules...)
	if err != nil {
		log.Printf("tsdb: error creating filter: %v", err)
		return emptySeries{}
	}
	return series{in, filter}
}

type emptySeries struct{}

func (_ emptySeries) Next() *tsdb.Point { select {} }
