package service

import (
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// All states for servers
const (
	ServerStateAvailable = "available"
	ServerStateReady     = "ready"
	ServerStateFull      = "full"
)

const RecommendedRenewInterval = 20 * time.Second
const ServerTTL = 60 * time.Second

type ServerInfo struct {
	Mutex   *sync.Mutex
	TokenId int // Also used
	State   string
	IP      string
	Port    int
}

var serverCache *ristretto.Cache[int, *ServerInfo]

func init() {
	var err error
	serverCache, err = ristretto.NewCache(&ristretto.Config[int, *ServerInfo]{
		MaxCost:     10_000,      // Maximum 10.000 stored items
		NumCounters: 10_000 * 10, // 10x what we want to store
		BufferItems: 64,          // Read description of field

		OnEvict: func(item *ristretto.Item[*ServerInfo]) {
			MarkTokenAsUnused(item.Value.TokenId)
		},
	})
	if err != nil {
		log.Fatalln("couldn't create cache:", err)
	}
}

func CreateServer(id int, ip string, port int, game string) bool {
	overwritten := serverCache.SetWithTTL(id, &ServerInfo{
		Mutex:   &sync.Mutex{},
		TokenId: id,
		IP:      ip,
		Port:    port,
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

func SetServerState(id int, state string) {
	info, valid := serverCache.Get(id)
	if !valid {
		return
	}

	info.Mutex.Lock()
	defer info.Mutex.Unlock()
	info.State = state
}
