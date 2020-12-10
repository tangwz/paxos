package paxos

import (
	"log"
	"time"
)

type learner struct {
	id        int
	acceptors map[int]message

	net network
}

func newLearner(id int, net network, acceptors ...int) *learner {
	l := &learner{id: id, net: net, acceptors: make(map[int]message)}
	for _, a := range acceptors {
		l.acceptors[a] = message{tp: Accept}
	}
	return l
}

func (l *learner) learn() string {
	for {
		msg, ok := l.net.recv(time.Hour)
		if !ok {
			continue
		}
		if msg.tp != Accept {
			log.Panicf("learner: %d receivd unexpected message %+v", l.id, msg)
		}
		l.handleAccepted(msg)
		accept, ok := l.chosen()
		if !ok {
			continue
		}
		return accept.value
	}
}

func (l *learner) handleAccepted(args message) {
	a := l.acceptors[args.from]
	if a.number < args.number {
		log.Printf("learner: %d received a new accepted proposal %+v", l.id, args)
		l.acceptors[args.from] = args
	}
}

func (l *learner) majority() int {
	return len(l.acceptors)/2 + 1
}

// To learn that a value has been chosen, a learner must find out that
// a proposal has been accepted by a majority of acceptors.
func (l *learner) chosen() (message, bool) {
	acceptCounts := make(map[int]int)
	acceptMsg := make(map[int]message)

	for _, accepted := range l.acceptors {
		if accepted.number != 0 {
			acceptCounts[accepted.number]++
			acceptMsg[accepted.number] = accepted
		}
	}

	for n, count := range acceptCounts {
		if count >= l.majority() {
			return acceptMsg[n], true
		}
	}
	return message{}, false
}
