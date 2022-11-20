package utils

import "sort"

/*
Priority Queue è una lista ordinata in senso crescente applicando la relazione di ordine totale
*/

type PriorityQueue struct {
	pendingRequests []Request
}

type Request struct {
	Sender    string
	Timestamp int
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{make([]Request, 0)}
}

func (pq *PriorityQueue) GetHead() Request {
	return pq.pendingRequests[0]
}

/*
Inserimento nella coda mantenendo la relazione di ordine totale
*/
func (pq *PriorityQueue) Enqueue(req Request) {
	//ricerca del più piccolo index i per cui f(i) è true,
	i := sort.Search(len(pq.pendingRequests), func(i int) bool {
		if pq.pendingRequests[i].Timestamp == req.Timestamp {
			return pq.pendingRequests[i].Sender > req.Sender
		}
		return pq.pendingRequests[i].Timestamp >= req.Timestamp
	})

	pq.pendingRequests = append(pq.pendingRequests, Request{})
	copy(pq.pendingRequests[i+1:], pq.pendingRequests[i:])
	pq.pendingRequests[i] = req
}

/*
Eliminazione della richiesta in posizione i
*/
func (pq *PriorityQueue) Delete(index int) {
	pq.pendingRequests = append(pq.pendingRequests[:index], pq.pendingRequests[index+1:]...)
}
