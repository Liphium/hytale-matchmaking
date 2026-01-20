package service

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"
)

const TokenFileName = "tokens.json"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Account      string `json:"account"`
	UUID         string `json:"uuid"`
}

type TokenInfo struct {
	Mutex *sync.Mutex
	Id    int
	Used  bool
	Token Token
}

var tokensMap = &sync.Map{}

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
		tokensMap.Store(i, &TokenInfo{
			Id:    i,
			Used:  false,
			Mutex: &sync.Mutex{},
			Token: token,
		})
	}
	tokenCounter = len(tokens) + 1
}

func GetFreeToken() (*TokenInfo, bool) {
	var foundServer *TokenInfo = nil

	// Find a token that's not currently used
	tokensMap.Range(func(key, value any) bool {
		info := value.(*TokenInfo)
		info.Mutex.Lock()
		defer info.Mutex.Unlock()

		// Use the token when expired or not used currently
		if !info.Used {
			info.Used = true
			foundServer = info
			return false
		}

		return true
	})

	return foundServer, foundServer != nil
}

func ReplaceAccessToken(id int, accessToken string) {
	obj, ok := tokensMap.Load(id)
	if !ok {
		return
	}
	info := obj.(*TokenInfo)

	info.Mutex.Lock()
	token := info.Token
	token.AccessToken = accessToken
	info.Token = token
	tokensMap.Store(id, info)
	info.Mutex.Unlock()

	saveToTokens()
}

func AddToken(token Token) {
	tokenCounterMutex.Lock()
	defer tokenCounterMutex.Unlock()

	tokensMap.Store(tokenCounter, &TokenInfo{
		Id:    tokenCounter,
		Used:  false,
		Mutex: &sync.Mutex{},
		Token: token,
	})
	tokenCounter++

	saveToTokens()
}

func MarkTokenAsUnused(token int) {
	if obj, ok := tokensMap.Load(token); ok {
		info := obj.(*TokenInfo)

		info.Mutex.Lock()
		defer info.Mutex.Unlock()
		info.Used = false
	}
}

// Always lock the token counter mutex before
func saveToTokens() {
	foundTokens := []Token{}
	tokensMap.Range(func(key, value any) bool {
		token := value.(*TokenInfo)

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
