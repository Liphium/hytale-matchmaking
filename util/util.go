package util

import "os"

func GetCredential() string {
	return os.Getenv("CREDENTIAL")
}
