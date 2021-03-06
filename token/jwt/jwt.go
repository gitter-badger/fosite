// Package JWT is able to generate and validate json web tokens.
// Follows https://tools.ietf.org/html/draft-ietf-oauth-json-web-token-32
package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/dgrijalva/jwt-go.v2"
)

// Enigma is responsible for generating and validating challenges.
type RS256JWTStrategy struct {
	PrivateKey *rsa.PrivateKey
}

// Generate generates a new authorize code or returns an error. set secret
func (j *RS256JWTStrategy) Generate(claims Mapper, header Mapper) (string, string, error) {
	if header == nil || claims == nil {
		return "", "", errors.New("Either claims or header is nil.")
	}

	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = claims.ToMap()
	token.Header = assign(token.Header, header.ToMap())

	var sig, sstr string
	var err error
	if sstr, err = token.SigningString(); err != nil {
		return "", "", errors.Wrap(err, "")
	}

	if sig, err = token.Method.Sign(sstr, j.PrivateKey); err != nil {
		return "", "", errors.Wrap(err, "")
	}

	return fmt.Sprintf("%s.%s", sstr, sig), sig, nil
}

// Validate : Validates a token and returns its signature or an error if the token is not valid.
func (j *RS256JWTStrategy) Validate(token string) (string, error) {
	if _, err := j.Decode(token); err != nil {
		return "", errors.Wrap(err, "")
	}

	return j.GetSignature(token)
}

func (j *RS256JWTStrategy) Decode(token string) (*jwt.Token, error) {
	// Parse the token.
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return &j.PrivateKey.PublicKey, nil
	})

	if err != nil {
		return nil, errors.Errorf("Couldn't parse token: %v", err)
	} else if !parsedToken.Valid {
		return nil, errors.Errorf("Token is invalid")
	}

	return parsedToken, err
}

func (j *RS256JWTStrategy) GetSignature(token string) (string, error) {
	split := strings.Split(token, ".")
	if len(split) != 3 {
		return "", errors.New("Header, body and signature must all be set")
	}
	return split[2], nil
}

func (c *RS256JWTStrategy) Hash(in []byte) ([]byte, error) {
	// SigningMethodRS256
	hash := sha256.New()
	_, err := hash.Write(in)
	if err != nil {
		return []byte{}, errors.Wrap(err, "")
	}
	return hash.Sum([]byte{}), nil
}

func (c *RS256JWTStrategy) GetSigningMethodLength() int {
	return jwt.SigningMethodRS256.Hash.Size()
}

func assign(a, b map[string]interface{}) map[string]interface{} {
	for k, w := range b {
		if _, ok := a[k]; ok {
			continue
		}
		a[k] = w
	}
	return a
}
