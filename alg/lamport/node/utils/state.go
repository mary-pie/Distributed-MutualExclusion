package utils

import (
	"github.com/hashicorp/memberlist"
)

type State struct {
	info     NodeInfo
	ml       *memberlist.Memberlist
	queue    *PriorityQueue
	clock    int
	num_acks int
}

type NodeInfo struct {
	hostname string
	portRCP  int
}

func NewState(host string, port int, group *memberlist.Memberlist) *State {
	return &State{info: NodeInfo{hostname: host, portRCP: port}, ml: group, queue: NewPriorityQueue(), clock: 0, num_acks: 0}
}

func (s *State) GetClock() int {
	return s.clock
}

func (s *State) GetPort() int {
	return s.info.portRCP
}

func (s *State) GetHostname() string {
	return s.info.hostname
}

func (s *State) GetMembers() []*memberlist.Node {
	return s.ml.Members()
}

func (s *State) GetAcks() int {
	return s.num_acks
}
func (s *State) GetQueue() []Request {
	return s.queue.pendingRequests
}

func (s *State) SetClock(new_clock int) {
	s.clock = new_clock
}

func (s *State) IncreaseClock() {
	s.clock += 1
}

func (s *State) IncreaseAcks() {
	s.num_acks += 1
}

func (s *State) AddRequest(req Request) {
	s.queue.Enqueue(req)
}

func (s *State) DeleteReq(sender string) {
	list := s.queue.pendingRequests
	var index int
	//ricerca posizione della richiesta
	for i, req := range list {
		if req.Sender == sender {
			index = i
		}
	}
	s.queue.Delete(index)
}

func (s *State) GetNextRequest() Request {
	if len(s.queue.pendingRequests) == 0 {
		return Request{}
	}
	return s.queue.GetHead()

}
