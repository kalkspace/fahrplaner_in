package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type VoteAPI struct {
	votes map[string]map[string]struct{}
	mutex sync.Mutex
}

func main() {
	api := VoteAPI{
		votes: make(map[string]map[string]struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/vote", api.vote)
	mux.HandleFunc("/votes", api.voteList)
	http.ListenAndServe(":8080", mux)
}

type VoteParams struct {
	ContentID string `json:"content_id"`
	UserID    string `json:"user_id"`
}

type VoteResponse struct {
	CurrentVotes uint
}

func (a *VoteAPI) vote(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var params VoteParams
	err := json.NewDecoder(req.Body).Decode(&params)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if params.ContentID == "" || params.UserID == "" {
		http.Error(rw, "Missing field", http.StatusBadRequest)
		return
	}
	a.mutex.Lock()
	defer a.mutex.Unlock()
	contentVotes, ok := a.votes[params.ContentID]
	if !ok {
		a.votes[params.ContentID] = make(map[string]struct{})
		contentVotes = a.votes[params.ContentID]
	}
	contentVotes[params.UserID] = struct{}{}
}

type VoteListItem struct {
	ContentID string `json:"content_id"`
	VoteCount uint   `json:"vote_count"`
}

func (a *VoteAPI) voteList(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	list := make([]VoteListItem, 0, len(a.votes))
	for cid, votes := range a.votes {
		list = append(list, VoteListItem{
			ContentID: cid,
			VoteCount: uint(len(votes)),
		})
	}
	json.NewEncoder(rw).Encode(list)
}
