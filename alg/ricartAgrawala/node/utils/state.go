package utils

import (
	"github.com/hashicorp/memberlist"
)

type State struct {
	info      NodeInfo
	ml        *memberlist.Memberlist
	queue     []Request
	clock     int
	num_reply int
	status    string
	last_req  int
}

type Request struct {
	Sender    string
	Timestamp int
}

type NodeInfo struct {
	hostname string
	portRCP  int
}

const (
	NCS        = "NCS"
	CS         = "CS"
	Requesting = "Requesting"
)

func NewState(host string, port int, group *memberlist.Memberlist) *State {
	return &State{info: NodeInfo{hostname: host, portRCP: port}, ml: group, queue: make([]Request, 0), clock: 0, num_reply: 0, status: NCS, last_req: 0}
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

func (s *State) GetReplies() int {
	return s.num_reply
}
func (s *State) GetQueue() []Request {
	return s.queue
}

func (s *State) GetStatus() string {
	return s.status
}

func (s *State) GetLastReq() int {
	return s.last_req
}

func (s *State) SetClock(new_clock int) {
	s.clock = new_clock
}

func (s *State) SetLastReq(new int) {
	s.last_req = new
}

func (s *State) SetStatus(new_status string) {
	s.status = new_status
}

func (s *State) IncreaseReplies() {
	s.num_reply += 1
}

func (s *State) IncreaseClock() {
	s.clock += 1
}

func (s *State) AddRequest(req Request) {
	s.queue = append(s.queue, req)
}

/*
Funzione di reset dello stato
*/
func (s *State) ResetState() {
	s.queue = make([]Request, 0) //svuoto la coda
	s.num_reply = 0              //azzero replies
	s.SetStatus(NCS)             //status a NCS
}

/*
Funziona di aggiornamento del clock alla ricezione di un msg:
clock = max{req.Timestamp, clock} + 1
*/
func (s *State) UpdateClock(timestamp int) {
	if timestamp > s.clock {
		s.SetClock(timestamp)
	}
	s.IncreaseClock()
}
