package paxos

import (
	"log"
	"time"
)

type network interface {
	send(m message)
	recv(timeout time.Duration) (message, bool)
}

type Network struct {
	queue map[int]chan message
}

func newNetwork(nodes ...int) *Network {
	pn := &Network{
		queue: make(map[int]chan message, 0),
	}

	for _, a := range nodes {
		pn.queue[a] = make(chan message, 1024)
	}
	return pn
}

func (net *Network) nodeNetwork(id int) *nodeNetwork {
	return &nodeNetwork{id: id, Network: net}
}

func (net *Network) send(m message) {
	log.Printf("net: send %+v", m)
	net.queue[m.to] <- m
}

func (net *Network) recvFrom(from int, timeout time.Duration) (message, bool) {
	select {
	case m := <-net.queue[from]:
		log.Printf("net: recv %+v", m)
		return m, true
	case <-time.After(timeout):
		return message{}, false
	}
}

type nodeNetwork struct {
	id int
	*Network
}

func (n *nodeNetwork) send(m message) {
	n.Network.send(m)
}

func (n *nodeNetwork) recv(timeout time.Duration) (message, bool) {
	return n.recvFrom(n.id, timeout)
}
