package token

import "time"

// interface for managing tokens
type TokenMaker interface {
	// creates new token for a specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)

	// verifies if a token is valid
	VerifyToken(token string) (*Payload, error)
}
