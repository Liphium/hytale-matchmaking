package testing_util

import (
	"encoding/json"
	"log"
	"runtime/debug"
	"testing"
)

// Unmarshal interface to struct
func UnmarshalToStruct[T any](data any) T {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal")
	}

	//fmt.Println(string(jsonBytes))

	var res T

	err = json.Unmarshal(jsonBytes, &res)
	if err != nil {
		panic("failed to unmarshal")
	}
	return res
}

// Marshaling for json without error handling cause testing (will just panic instead)
func Marshal(t *testing.T, data any) []byte {
	done, err := json.Marshal(data)
	if err != nil {
		t.Fatal("couldn't marshal:", err)
	}
	return done
}

// Unmarshalling for json without error handling cause testing (will just panic instead)
func Unmarshal(t *testing.T, data []byte, v any) {
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal("couldn't unmarshal:", err)
	}
}

// Helper function for printing errors during testing (to make the test code shorter)
func PrintError(t *testing.T, kind string, err error) {
	if err != nil {
		if t != nil {
			debug.PrintStack()
			t.Fatalf("%s: %v", kind, err)
		} else {
			log.Fatalf("%s: %v", kind, err)
		}
	}
}
