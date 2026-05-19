package entity

// Challenge is a practice-facing challenge projection used by scoring flows.
type Challenge struct {
	ID         int64
	Points     int
	Difficulty string
}
