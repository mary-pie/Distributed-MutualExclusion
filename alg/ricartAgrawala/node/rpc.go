package node

import (
	"log"
	"mutualexclusion-project/alg/ricartAgrawala/node/utils"
)

type RequestPayload struct {
	SenderId  string
	Timestamp int
}

type Reply struct {
	SenderId  string
	Timestamp int
}

type MutualExclusion struct {
	state *utils.State
}

/*
Handler che gestisce le richieste di accesso alla CS:
se Status=CS oppure (se Status=Requesting {LastReq, id} < {req.Timestamp, req.SenderId}), accodo la richiesta
altrimenti invio REPLY
*/
func (me *MutualExclusion) AccessCS(req RequestPayload, res *Reply) error {
	log.Println("New request to access critical section: ", req)
	s := me.state

	s.UpdateClock(req.Timestamp) //aggiornamento clock
	if s.GetStatus() == "CS" || (s.GetStatus() == "Requesting" && compareRequest(req, s.GetLastReq(), s.GetHostname())) {
		s.AddRequest(utils.Request{Sender: req.SenderId, Timestamp: req.Timestamp})
		log.Println("Added to queue ", s.GetQueue())
		*res = Reply{} //invio Reply vuoto
	} else {
		s.IncreaseClock() //incremento clock prima di invio REPLY
		*res = Reply{SenderId: s.GetHostname(), Timestamp: s.GetClock()}
		log.Println("send REPLY ", *res)
	}
	return nil
}

/*
Handler che gestisce la ricezione di msg REPLY:
se #replies = n-1, entro nella sezione critica
*/
func (me *MutualExclusion) ReceiveReply(rep Reply, res *string) error {
	s := me.state
	me.state.IncreaseReplies() //aggiorno il # di reply ricevuti
	log.Println("New REPLY received ", rep, ". Total replies currently received", s.GetReplies())
	s.UpdateClock(rep.Timestamp) //aggiornamento clock

	if me.state.GetReplies() == len(s.GetMembers())-1 {
		s.SetStatus(utils.CS)
		enterCS()
		exit(s)
	}
	return nil
}
