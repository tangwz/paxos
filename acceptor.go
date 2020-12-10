package paxos

import (
	"log"
	"time"
)

type acceptor struct {
	// server id
	id int
	// the number of the proposal this server will accept, or 0 if it has never received a Prepare request
	promiseNumber int
	// the number of the last proposal the server has accepted, or 0 if it never accepted any.
	acceptedNumber int
	// the value from the most recent proposal the server has accepted, or null if it has never accepted a proposal
	acceptedValue string

	learners []int
	net      network
}

func newAcceptor(id int, net network, learners ...int) *acceptor {
	return &acceptor{id: id, net: net, acceptedNumber: 0, acceptedValue: "", learners: learners}
}

func (a *acceptor) run() {
	log.Printf("acceptor %d run", a.id)
	for {
		msg, ok := a.net.recv(time.Hour)
		if !ok {
			continue
		}
		switch msg.tp {
		case Prepare:
			promise, ok := a.handlePrepare(msg)
			if ok {
				a.net.send(promise)
			}
		case Propose:
			proposalAccepted := a.handleAccept(msg)
			if proposalAccepted {
				for _, l := range a.learners {
					msg := message{
						tp:     Accept,
						from:   a.id,
						to:     l,
						number: a.acceptedNumber,
						value:  a.acceptedValue,
					}
					a.net.send(msg)
				}
			}
		default:
			log.Panicf("acceptor: %d unexpected message type: %v", a.id, msg.tp)
		}
	}
}

// Phase 1. (b) If an acceptor receives a prepare request with number n greater than that of
// any prepare request to which it has already responded, then it responds to the request
// with a promise not to accept any more proposals numbered less than n and with
// the highest-numbered proposal (if any) that it has accepted.
func (a *acceptor) handlePrepare(args message) (message, bool) {
	if a.promiseNumber >= args.number {
		return message{}, false
	}
	a.promiseNumber = args.number
	msg := message{
		tp:     Promise,
		from:   a.id,
		to:     args.from,
		number: a.acceptedNumber,
		value:  a.acceptedValue,
	}
	return msg, true
}

// Phase 2. (b) If an acceptor receives an accept request for a proposal numbered n,
// it accepts the proposal unless it has already responded to a prepare request
// having a number greater than n.
func (a *acceptor) handleAccept(args message) bool {
	number := args.number
	if number >= a.promiseNumber {
		a.acceptedNumber = number
		a.acceptedValue = args.value
		a.promiseNumber = number
		return true
	}

	return false
}
