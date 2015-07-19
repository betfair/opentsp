package tsdb

type joined struct {
	input []<-chan *Point
}

// Join returns a time series that combines data points from the
// given channels.
func Join(a, b <-chan *Point) Series {
	return &joined{
		input: []<-chan *Point{a, b},
	}
}

func (j *joined) Next() *Point {
	select {
	case p := <-j.input[0]:
		return p
	case p := <-j.input[1]:
		return p
	}
}
