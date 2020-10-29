package stats

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

func NewKeeper(usernames []string) *Keeper {
	k := &Keeper{
		UserStats: make(map[string]*UserStats),
	}
	for _, name := range usernames {
		k.UserStats[name] = &UserStats{}
	}
	return k
}

type Keeper struct {
	UserStats map[string]*UserStats

	rw sync.RWMutex //lock down our stats for concurrent access
}

type UserStats struct {
	LastSuccessful   int64 `json:"last_successful_session_unix_time"`
	SuccessCounter   int64 `json:"authorized_attempt"`
	LastUnsuccessful int64 `json:"last_unsucessful_session_unix_time"`
	UnsuccessCounter int64 `json:"unauthorized_attempt"`
}

func (me *Keeper) IncrementAuthorizedAttempt(username string) {
	me.rw.Lock()
	defer me.rw.Unlock()

	stats := me.UserStats[username]
	stats.LastSuccessful = time.Now().Unix()
	stats.SuccessCounter++
}

func (me *Keeper) IncrementUnauthorizedAttempt(username string) {
	me.rw.Lock()
	defer me.rw.Unlock()

	stats := me.UserStats[username]
	stats.LastUnsuccessful = time.Now().Unix()
	stats.UnsuccessCounter++
}

func (me *Keeper) GetStats(username string) string {
	me.rw.RLock()
	defer me.rw.RUnlock()

	stats := me.UserStats[username]
	jsonBytes, err := json.Marshal(stats)
	if err != nil {
		log.Printf("Error marshalling user stats: %s", err.Error())
	}
	return string(jsonBytes)
}
