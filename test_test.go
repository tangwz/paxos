package paxos

import (
	"testing"
	"time"
)

func TestSingleProposer(t *testing.T) {
	// 1, 2, 3 are acceptors
	// 1001 is a proposer
	nt := newNetwork(1, 2, 3, 1001, 2001)

	acceptors := make([]*acceptor, 0)
	for i := 1; i <= 3; i++ {
		acceptors = append(acceptors, newAcceptor(i, nt.nodeNetwork(i), 2001))
	}

	for _, a := range acceptors {
		go a.run()
	}

	p := newProposer(1001, "hello world", nt.nodeNetwork(1001), 1, 2, 3)
	go p.run()

	l := newLearner(2001, nt.nodeNetwork(2001), 1, 2, 3)
	value := l.learn()
	if value != "hello world" {
		t.Errorf("value = %s, excepted %s", value, "hello world")
	}
}

func TestTwoProposers(t *testing.T) {
	// 1, 2, 3 are acceptors
	// 1001,1002 is a proposer
	nt := newNetwork(1, 2, 3, 1001, 1002, 2001)

	as := make([]*acceptor, 0)
	for i := 1; i <= 3; i++ {
		as = append(as, newAcceptor(i, nt.nodeNetwork(i), 2001))
	}

	for _, a := range as {
		go a.run()
	}

	p1 := newProposer(1001, "hello world", nt.nodeNetwork(1001), 1, 2, 3)
	go p1.run()

	time.Sleep(time.Millisecond)
	p2 := newProposer(1002, "bad day", nt.nodeNetwork(1002), 1, 2, 3)
	go p2.run()

	l := newLearner(2001, nt.nodeNetwork(2001), 1, 2, 3)
	value := l.learn()
	if value != "hello world" {
		t.Errorf("value = %s, want %s", value, "hello world")
	}
}
