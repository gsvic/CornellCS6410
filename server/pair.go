package server

// A Pair consists of an int and an int64 timestamp
type Pair struct {
	data int
	ts int64
}

func CreatePair(value int, ts int64) *Pair{
	return &Pair{data: value, ts: ts}
}
// GetData returns the data
func (pair *Pair) GetData() int {
	return pair.data
}

// GetTs returns the timestamp
func (pair *Pair) GetTs() int64 {
	return pair.ts
}

func (pair *Pair) SetData(data int) *Pair {
	pair.data = data

	return pair
}

func (pair *Pair) SetTs(ts int64) *Pair {
	pair.ts = ts

	return pair
}