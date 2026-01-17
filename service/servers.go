package service

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

// All states for servers
const (
	StateAvailable = "available"
	StateLobby     = "lobby"
	StateIngame    = "ingame"
	StateEnd       = "end"
)

const TokenFileName = "tokens.json"
const RecommendedRenewInterval = 20 * time.Second
const TokenTTL = 60 * time.Second

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Account      string `json:"account"`
	UUID         string `json:"uuid"`
}

type ServerInfo struct {
	Mutex     *sync.Mutex
	Id        int
	Used      bool
	State     string
	Game      string
	ExpiresAt time.Time
	Token     Token
}

var serverMap = &sync.Map{}

var tokenCounter int = 0
var tokenCounterMutex = &sync.Mutex{}

func LoadTokens() {
	tokensFile := path.Join(os.Getenv("TOKEN_FILE_LOCATION"), TokenFileName)
	content, err := os.ReadFile(tokensFile)
	if err != nil {
		bytes, err := json.Marshal([]Token{})
		if err != nil {
			log.Fatal("Couldn't write empty tokens file (marshal):", err)
		}

		if err := os.WriteFile(tokensFile, bytes, 0644); err != nil {
			log.Fatal("Couldn't write empty tokens file:", err)
		}
		content = bytes
	}

	var tokens []Token
	if err := json.Unmarshal(content, &tokens); err != nil {
		log.Fatal("Couldn't parse tokens file:", err)
	}

	// Fill the map with all of the tokens in the file
	for i, token := range tokens {
		serverMap.Store(i, &ServerInfo{
			Id:    i,
			Used:  false,
			Mutex: &sync.Mutex{},
			Token: token,
		})
	}
	tokenCounter = len(tokens) + 1
}

func CreateNewServer(game string) (*ServerInfo, bool) {
	var foundServer *ServerInfo = nil

	// Find a token that's not currently used
	serverMap.Range(func(key, value any) bool {
		info := value.(*ServerInfo)
		info.Mutex.Lock()
		defer info.Mutex.Unlock()

		// Use the token when expired or not used currently
		if !info.Used || time.Now().After(info.ExpiresAt) {
			info.Used = true
			info.ExpiresAt = time.Now().Add(TokenTTL)
			info.State = StateAvailable
			info.Game = game
			foundServer = info
			return false
		}

		return true
	})

	return foundServer, foundServer != nil
}

func RefreshServer(id int) {
	obj, valid := serverMap.Load(id)
	if !valid {
		return
	}
	info := obj.(*ServerInfo)

	info.Mutex.Lock()
	defer info.Mutex.Unlock()
	info.ExpiresAt = time.Now().Add(TokenTTL)
}

func SetServerState(id int, state string) {
	obj, valid := serverMap.Load(id)
	if !valid {
		return
	}
	info := obj.(*ServerInfo)

	info.Mutex.Lock()
	defer info.Mutex.Unlock()
	info.State = state
}

func ReplaceAccessTokenForServer(id int, accessToken string) {
	obj, ok := serverMap.Load(id)
	if !ok {
		return
	}
	info := obj.(*ServerInfo)

	info.Mutex.Lock()
	token := info.Token
	token.AccessToken = accessToken
	info.Token = token
	serverMap.Store(id, info)
	info.Mutex.Unlock()

	saveToTokens()
}

func AddToken(token Token) {
	tokenCounterMutex.Lock()
	defer tokenCounterMutex.Unlock()

	serverMap.Store(tokenCounter, &ServerInfo{
		Id:    tokenCounter,
		Used:  false,
		Mutex: &sync.Mutex{},
		Token: token,
	})
	tokenCounter++

	saveToTokens()
}

// Always lock the token counter mutex before
func saveToTokens() {
	foundTokens := []Token{}
	serverMap.Range(func(key, value any) bool {
		token := value.(*ServerInfo)

		token.Mutex.Lock()
		defer token.Mutex.Unlock()

		foundTokens = append(foundTokens, token.Token)
		return true
	})

	tokensFile := path.Join(os.Getenv("TOKEN_FILE_LOCATION"), TokenFileName)

	bytes, err := json.Marshal(foundTokens)
	if err != nil {
		log.Fatal("Couldn't write tokens file (marshal):", err)
	}
	if err := os.WriteFile(tokensFile, bytes, 0644); err != nil {
		log.Fatal("Couldn't write tokens file:", err)
	}
}
