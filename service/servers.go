package service

import (
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

const RecommendedRenewInterval = 20 * time.Second
const ServerTTL = 60 * time.Second

type ServerInfo struct {
	Mutex   *sync.RWMutex // Just for the general data on the server (IP, etc.)
	TokenId int           // Also used
	IP      string
	Port    int

	Matches *sync.Map // Match id -> *Match
	Players *sync.Map // Player id -> *PlayerInfo
}

var serverCache *ristretto.Cache[int, *ServerInfo]

func init() {
	var err error
	serverCache, err = ristretto.NewCache(&ristretto.Config[int, *ServerInfo]{
		MaxCost:     10_000,      // Maximum 10.000 stored items
		NumCounters: 10_000 * 10, // 10x what we want to store
		BufferItems: 64,          // Read description of field

		OnEvict: func(item *ristretto.Item[*ServerInfo]) {

			// Cleanup server (in goroutine to make sure it doesn't block anything in ristretto)
			go func() {
				MarkTokenAsUnused(item.Value.TokenId)

				// Delete all players
				item.Value.Players.Range(func(key, value any) bool {
					p := value.(*PlayerInfo)

					p.Mutex.RLock()
					defer p.Mutex.RUnlock()

					deletePlayer(p.Account)
					return true
				})
				item.Value.Players.Clear()

				// Mark all matches as ended (in case they are in some game they will get cleaned and no-one will be able to join)
				item.Value.Matches.Range(func(key, value any) bool {
					m := value.(*Match)

					m.Mutex.Lock()
					defer m.Mutex.Unlock()
					m.State = MatchStateEnd
					return true
				})
				item.Value.Matches.Clear()
			}()
		},
	})
	if err != nil {
		log.Fatalln("couldn't create cache:", err)
	}
}

func CreateServer(id int, ip string, port int, game string) bool {
	overwritten := serverCache.SetWithTTL(id, &ServerInfo{
		Mutex:   &sync.RWMutex{},
		TokenId: id,
		IP:      ip,
		Port:    port,
		Players: &sync.Map{},
		Matches: &sync.Map{},
	}, 1, ServerTTL)

	serverCache.Wait()
	return overwritten
}

func RefreshServer(id int) {
	if item, ok := serverCache.Get(id); ok {
		serverCache.SetWithTTL(id, item, 1, ServerTTL)
		serverCache.Wait()
	}
}

// Get a server's ip and port
func GetServerDetails(id int) (ip string, port int, ok bool) {
	server, ok := serverCache.Get(id)
	if !ok {
		return "", 0, false
	}

	server.Mutex.RLock()
	defer server.Mutex.RUnlock()
	return server.IP, server.Port, true
}
