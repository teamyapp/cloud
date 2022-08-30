package security

import (
	"encoding/json"
	"errors"

	"github.com/golang-jwt/jwt"
	"github.com/teamyapp/cloud/libs/obs"
)

type JWTAuthority struct {
	dataCollector obs.DataCollector
	signingKey    []byte
}

func (j JWTAuthority) GenerateToken(payload interface{}) (string, error) {
	payloadMap := make(map[string]interface{})
	jsonBuf, _ := json.Marshal(payload)
	err := json.Unmarshal(jsonBuf, &payloadMap)
	if err != nil {
		j.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloadMap))
	return token.SignedString(j.signingKey)
}

func (j JWTAuthority) DecodeToken(jwtToken string, output interface{}) error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		j.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	if !token.Valid {
		err = errors.New("token is invalid")
		j.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return j.parseJWTClaims(token.Claims, output)
}

func (j JWTAuthority) DecodeUnverifiedToken(jwtToken string, output interface{}) error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &claims)
	if err != nil {
		j.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return j.parseJWTClaims(claims, output)
}

func (j JWTAuthority) parseJWTClaims(claims jwt.Claims, output interface{}) error {
	buf, err := json.Marshal(claims)
	if err != nil {
		j.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return json.Unmarshal(buf, output)
}

func NewJWTAuthority(dataCollector obs.DataCollector, signingKey string) JWTAuthority {
	return JWTAuthority{
		dataCollector: dataCollector,
		signingKey:    []byte(signingKey),
	}
}
