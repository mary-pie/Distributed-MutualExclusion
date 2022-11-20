package utils

import (
	"net"

	"github.com/DistributedClocks/GoVector/govec/vclock"
	"github.com/hashicorp/memberlist"
	"golang.org/x/exp/slices"
)

type State struct {
	info        NodeInfo
	ml          *memberlist.Memberlist
	queue       []PendingRequest  // lista di PendingRequest, dove le richieste vengono aggiunte in coda
	v           map[string]uint64 // map di richieste già servite:  key = nodeId, value = num richieste servite per nodeId
	isTokenFree bool
}

type PendingRequest struct {
	NodeId    string
	Timestamp vclock.VClock
}

type NodeInfo struct {
	hostname string
	portRCP  int
}

func NewState(host string, port int, list *memberlist.Memberlist) *State {
	return &State{info: NodeInfo{hostname: host, portRCP: port}, ml: list, queue: make([]PendingRequest, 0), v: map[string]uint64{}, isTokenFree: true}
}

func (state *State) GetMembers() []*memberlist.Node {
	return state.ml.Members()
}

func (state *State) GetQueue() []PendingRequest {
	return state.queue
}

func (state *State) GetIsTokenFree() bool {
	return state.isTokenFree
}

func (state *State) GetHostname() string {
	return state.info.hostname
}

func (state *State) GetV() map[string]uint64 {
	return state.v
}

/*
inizializzazione del vettore che tiene conto delle richieste servite per ogni nodo del cluster
*/
func (state *State) SetUpV() {
	for _, m := range state.ml.Members() {
		host := m.Name
		if host != "coordinator" {
			state.v[host] = 0
		}
	}
}
func (state *State) IncreaseV(id string) {
	state.v[id] += 1
}

func (state *State) AppendRequest(r PendingRequest) {
	state.queue = append(state.queue, r)
}

func (state *State) SetStateToken(b bool) {
	state.isTokenFree = b

}

/*
Elimina l'elemento in posizione i dalla coda
*/
func (state *State) RemoveIndex(index int) {
	state.queue = append(state.queue[:index], state.queue[index+1:]...)
}

/*
Funzione di ricerca dell'indirizzo ip dall'id (hostaname) nel cluster
Output: index dell'elemento nella lista dei membri
*/
func (state *State) SearchMember(host string) net.IP {
	indexNode := slices.IndexFunc(state.ml.Members(), func(m *memberlist.Node) bool { return m.Name == host })

	return state.ml.Members()[indexNode].Addr
}

/*
Funzione di ricerca della prima richiesta eleggibile
*/
func (state *State) SearchEligibleReq() string {
	//prendo dalla coda la prossima richiesta e verifico se è eleggibile
	for i, req := range state.GetQueue() {
		if state.CheckEligible(req.Timestamp, req.NodeId) {
			state.RemoveIndex(i)
			return req.NodeId
		}
	}
	return ""
}

/*
Controllo se la richiesta da parte di id_sender è eleggibile:
se per ogni i != id_sender, timestamp[i] <= V[i]
*/
func (state *State) CheckEligible(timestamp_req vclock.VClock, id_sender string) bool {
	v := state.v
	for id, clock := range timestamp_req {
		if id != id_sender && clock > v[id] {
			return false
		}
	}
	//verifico se id_sender ha inviato un'altra richiesta precedente che ancora non è stata processata
	return timestamp_req[id_sender] <= v[id_sender]+1
}
