package paxos

import (
	"log"
	"time"
)

type proposer struct {
	// server id
	id int
	// the largest round number the server has seen
	round int
	// proposal number = (round number, serverID)
	number int
	// proposal value
	value     string
	acceptors map[int]bool
	net       network
}

func newProposer(id int, value string, net network, acceptors ...int) *proposer {
	p := &proposer{id: id, round: 0, value: value, net: net, acceptors: make(map[int]bool, len(acceptors))}
	for _, acceptor := range acceptors {
		p.acceptors[acceptor] = false
	}
	return p
}

func (p *proposer) run() {
	var ok bool
	var msg message
	for !p.majorityReached() {
		if !ok {
			prepareMsg := p.prepare()
			for i := range prepareMsg {
				p.net.send(prepareMsg[i])
			}
		}

		msg, ok = p.net.recv(time.Second)
		if !ok {
			continue
		}

		switch msg.tp {
		case Promise:
			p.handlePromise(msg)
		default:
			panic("UnSupport message.")
		}
	}

	proposeMsg := p.accept()
	for _, msg := range proposeMsg {
		p.net.send(msg)
	}
}

// Phase 1. (a) A proposer selects a proposal number n
// and sends a prepare request with number n to a majority of acceptors.
func (p *proposer) prepare() []message {
	p.round++
	p.number = p.proposalNumber()
	msg := make([]message, p.majority())
	i := 0

	for to := range p.acceptors {
		msg[i] = message{
			tp:     Prepare,
			from:   p.id,
			to:     to,
			number: p.number,
		}
		i++
		if i == p.majority() {
			break
		}
	}
	return msg
}

func (p *proposer) handlePromise(reply message) {
	log.Printf("proposer: %d received a promise %+v", p.id, reply)
	p.acceptors[reply.from] = true
	if reply.number > 0 {
		p.number = reply.number
		p.value = reply.value
	}
}

// Phase 2. (a) If the proposer receives a response to its prepare requests (numbered n)
// from a majority of acceptors, then it sends an accept request to each of those acceptors
// for a proposal numbered n with a value v, where v is the value of the highest-numbered proposal
// among the responses, or is any value if the responses reported no proposals.
func (p *proposer) accept() []message {
	msg := make([]message, p.majority())
	i := 0
	for to, ok := range p.acceptors {
		if ok {
			msg[i] = message{
				tp:     Propose,
				from:   p.id,
				to:     to,
				number: p.number,
				value:  p.value,
			}
			i++
		}

		if i == p.majority() {
			break
		}
	}
	return msg
}

func (p *proposer) majority() int {
	return len(p.acceptors) / 2 + 1
}

func (p *proposer) majorityReached() bool {
	count := 0
	for _, ok := range p.acceptors {
		if ok {
			count++
		}
	}
	if count >= p.majority() {
		return true
	}
	return false
}

// proposal number = (round number, serverID)
func (p *proposer) proposalNumber() int {
	return p.round<< 16 | p.id
}
