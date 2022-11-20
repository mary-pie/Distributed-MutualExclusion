package coordinator

import (
	"log"
	"mutualexclusion-project/alg/tokenBased/coordinator/utils"
	"time"

	"github.com/DistributedClocks/GoVector/govec/vclock"
)

type RequestPayload struct {
	SenderId  string
	Timestamp vclock.VClock
}

type Token string

type TokenBack struct {
	SenderId string
	T        Token
}

type MutualExclusion struct {
	state *utils.State
}

const (
	token = "token123456789"
)

/*
Handler che gestisce le richieste di accesso alla CS:
se eleggibile e token free, send TOKEN
altrimenti, append nella coda
*/
func (me *MutualExclusion) AccessCS(req RequestPayload, res *Token) error {
	log.Println("New request to access critical section: ", req)
	s := me.state

	if s.CheckEligible(req.Timestamp, req.SenderId) && s.GetIsTokenFree() {
		log.Println("Sending token to ", req.SenderId)
		s.SetStateToken(false)
		time.Sleep(time.Duration(7) * time.Second)
		*res = token
	} else {
		s.AppendRequest(utils.PendingRequest{NodeId: req.SenderId, Timestamp: req.Timestamp})
		log.Println("Request not eligible or token not free. Added to queue: ", s.GetQueue())
	}
	return nil
}

/*
Handler che gestisce la ricezione del token
*/
func (me *MutualExclusion) ResendToken(token_back TokenBack, res *string) error {
	log.Println("Token resend back ", token_back)
	s := me.state
	log.Println("Queue: ", s.GetQueue())

	s.IncreaseV(token_back.SenderId) //incremento il contatore delle richieste per token_back.SenderId
	log.Println("Updated V: ", s.GetV())
	s.SetStateToken(true) //token free

	//cerco l'eventuale prossima richiesta eleggibile
	if len(s.GetQueue()) != 0 {
		id := s.SearchEligibleReq()
		if id != "" {
			sendToken(id, s)
		}
	} else {
		log.Println("No eligible request. Waiting for new requests..")
	}
	return nil
}
