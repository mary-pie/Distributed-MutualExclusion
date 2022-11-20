package node

import (
	"log"
	"mutualexclusion-project/alg/ricartAgrawala/node/utils"
)

type RequestPayload struct {
	SenderId  string
	Timestamp int
}

type Reply string

type MutualExclusion struct {
	state *utils.State
}

/*
Handler che gestisce le richieste di accesso alla CS:
se Status=CS oppure (se Status=Requesting {LastReq, id} < {req.Timestamp, req.SenderId}), accodo la richiesta
altrimenti invio REPLY
aggiorno clock = max{req.Timestamp, clock}
*/
func (me *MutualExclusion) AccessCS(req RequestPayload, res *Reply) error {
	log.Println("New request to access critical section: ", req)
	s := me.state
	if s.GetStatus() == "CS" || (s.GetStatus() == "Requesting" && compareRequest(req, s.GetLastReq(), s.GetHostname())) {
		s.AddRequest(utils.Request{Sender: req.SenderId, Timestamp: req.Timestamp})
		log.Println("Added to queue")
	} else {
		log.Println("send REPLY")
		*res = "REPLY"
	}
	if req.Timestamp > s.GetClock() {
		s.SetClock(req.Timestamp)
	}
	return nil
}

/*
Handler che gestisce la ricezione di msg REPLY:
aggiorno il # di reply ricevuti
se #replies = n-1, entro nella sezione critica, se
*/
func (me *MutualExclusion) Reply(req Reply, res *string) error {
	s := me.state
	me.state.IncreaseReplies()
	log.Println("New REPLY received. Total replies currently received", s.GetReplies())

	if me.state.GetReplies() == len(s.GetMembers())-1 {
		s.SetStatus(utils.CS)
		enterCS()
		exit(s)
	}
	return nil
}
