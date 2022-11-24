package node

import (
	"log"
	"mutualexclusion-project/alg/lamport/node/utils"
)

const Ack = "ACK"

type RequestPayload struct {
	SenderId  string
	Timestamp int
}
type ReleasePayload struct {
	SenderId  string
	Timestamp int
}

type AckMsg struct {
	Msg       string
	Timestamp int
}

type MutualExclusion struct {
	state *utils.State
}

/*
Handler che gestisce le richieste di accesso alla CS:
1. accodo la richiesta
2. invio ACK
*/
func (me *MutualExclusion) AccessCS(req RequestPayload, res *AckMsg) error {

	log.Println("New REQUEST message ", req)
	s := me.state

	//aggiornamento clock = max{ack.Timestamp, clock}
	if maxClock(req.Timestamp, s.GetClock()) {
		s.SetClock(req.Timestamp)
	}
	s.IncreaseClock() //clock += 1
	log.Println("Updated Clock: ", s.GetClock())

	s.AddRequest(utils.Request{Sender: req.SenderId, Timestamp: req.Timestamp}) //aggiunta nella coda
	log.Println("Added in queue: ", s.GetQueue())
	s.IncreaseClock()                //incremento clock prima di inviare la risposta
	*res = AckMsg{Ack, s.GetClock()} //invio ack
	log.Println("Sent ACK ", *res)

	return nil
}

/*
Handler che gestisce la ricezione dei messaggi di RELEASE:
eliminino richiesta corrispondente dalla coda
se la sua richiesta è la prossima a dover essere processata e num_acks = n-1 --> accesso alla CS
*/
func (me *MutualExclusion) Release(rel ReleasePayload, res *string) error {

	log.Println("New RELEASE message ", rel)
	s := me.state

	//aggiornamento clock = max{ack.Timestamp, clock}
	if maxClock(rel.Timestamp, s.GetClock()) {
		s.SetClock(rel.Timestamp)
	}
	s.IncreaseClock() //clock += 1
	log.Println("Updated Clock: ", s.GetClock())

	s.DeleteReq(rel.SenderId)
	log.Println("Deleted request from queue ", s.GetQueue())

	if (s.GetNextRequest() == utils.Request{}) {
		log.Println("Empty queue.. ")
	} else if s.GetNextRequest().Sender == s.GetHostname() && s.GetAcks() == len(s.GetMembers())-1 {
		enterCS()
		exit(s)
	}
	return nil
}
