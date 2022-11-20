package utils

import (
	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/hashicorp/memberlist"
)

type State struct {
	info  NodeInfo
	ml    *memberlist.Memberlist
	clock vclock.VClock // map di richieste già servite:  key = nodeId, value = num richieste servite per nodeId
}

type PendingRequest struct {
	NodeId    string
	Timestamp vclock.VClock
}

type NodeInfo struct {
	Hostname string
	PortRCP  int
}

func NewState(hostname string, port int, list *memberlist.Memberlist) *State {
	return &State{info: NodeInfo{Hostname: hostname, PortRCP: port}, ml: list, clock: vclock.New()}
}

func (state *State) GetMembers() []*memberlist.Node {
	return state.ml.Members()
}

func (state *State) GetHostname() string {
	return state.info.Hostname
}

func (state *State) GetClock() map[string]uint64 {
	return state.clock
}

func (state *State) GetPort() int {
	return state.info.PortRCP
}

/*
inizializzazione del vettore che tiene conto delle richieste servite per ogni nodo del cluster
*/
func (state *State) SetUpClock() {
	for _, m := range state.ml.Members() {
		host := m.Name
		if host != "coordinator" {
			state.clock[host] = 0
		}
	}
}

func (state *State) IncreaseClock() {
	state.clock[state.info.Hostname] += 1
}

func (state *State) UpdateClock(id string, new_value uint64) {
	state.clock[id] = new_value
}
