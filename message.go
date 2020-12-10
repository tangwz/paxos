package paxos

// MsgType represents the type of a paxos phase.
type MsgType uint8

const (
	Prepare MsgType = iota
	Promise
	Propose
	Accept
)

type message struct {
	tp     MsgType
	from   int
	to     int
	number int    // proposal number
	value  string // proposal value
}
