package tsdb

type joined []Chan

// Join returns a time series that combines data points from the
// given channels.
func Join(a, b Chan) Series {
	return joined{a, b}
}

func (ch joined) Next() *Point {
	select {
	case p := <-ch[0]:
		return p
	case p := <-ch[1]:
		return p
	}
}
