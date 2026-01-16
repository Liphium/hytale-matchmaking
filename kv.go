package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"
	"time"
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

type TokenInfo struct {
	Mutex     *sync.Mutex
	Id        int
	Used      bool
	ExpiresAt time.Time
	Token     Token
}

var tokenMap = &sync.Map{}

var tokenCounter int = 0
var tokenCounterMutex = &sync.Mutex{}

func loadTokens() {
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
		tokenMap.Store(i, &TokenInfo{
			Id:    i,
			Used:  false,
			Mutex: &sync.Mutex{},
			Token: token,
		})
	}
	tokenCounter = len(tokens) + 1
}

func getTokenForServer() (*TokenInfo, bool) {
	var foundToken *TokenInfo = nil

	tokenMap.Range(func(key, value any) bool {
		token := value.(*TokenInfo)
		token.Mutex.Lock()
		defer token.Mutex.Unlock()

		// Use the token when expired or not used currently
		if !token.Used || time.Now().After(token.ExpiresAt) {
			token.Used = true
			token.ExpiresAt = time.Now().Add(TokenTTL)
			foundToken = token
			return false
		}

		return true
	})

	return foundToken, foundToken != nil
}

func refreshToken(id int) {
	obj, valid := tokenMap.Load(id)
	if !valid {
		return
	}
	token := obj.(*TokenInfo)
	token.ExpiresAt = time.Now().Add(TokenTTL)
}

func addToken(token Token) {
	tokenCounterMutex.Lock()
	defer tokenCounterMutex.Unlock()

	tokenMap.Store(tokenCounter, &TokenInfo{
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
	tokenMap.Range(func(key, value any) bool {
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
