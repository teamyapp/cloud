package security

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/golang-jwt/jwt"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type JWTAuthority struct {
	dataCollector telemetry.DataCollector
	signingKey    []byte
}

func (j JWTAuthority) GenerateToken(ct context.Context, payload interface{}) (string, error) {
	payloadMap := make(map[string]interface{})
	jsonBuf, _ := json.Marshal(payload)
	err := json.Unmarshal(jsonBuf, &payloadMap)
	if err != nil {
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloadMap))
	return token.SignedString(j.signingKey)
}

func (j JWTAuthority) DecodeToken(ct context.Context, jwtToken string, output interface{}) error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	if !token.Valid {
		err = errors.New("token is invalid")
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return j.parseJWTClaims(ct, token.Claims, output)
}

func (j JWTAuthority) DecodeUnverifiedToken(ct context.Context, jwtToken string, output interface{}) error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &claims)
	if err != nil {
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return j.parseJWTClaims(ct, claims, output)
}

func (j JWTAuthority) parseJWTClaims(ct context.Context, claims jwt.Claims, output interface{}) error {
	buf, err := json.Marshal(claims)
	if err != nil {
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return json.Unmarshal(buf, output)
}

func NewJWTAuthority(dataCollector telemetry.DataCollector, signingKey string) JWTAuthority {
	return JWTAuthority{
		dataCollector: dataCollector,
		signingKey:    []byte(signingKey),
	}
}
