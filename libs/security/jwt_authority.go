package security

import (
	"context"
	"encoding/json"

	"github.com/golang-jwt/jwt"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type JWTAuthority struct {
	dataCollector telemetry.DataCollector
	signingKey    []byte
}

func (j JWTAuthority) GenerateToken(ct context.Context, payload interface{}) (string, *errs.Error) {
	payloadMap := make(map[string]interface{})
	jsonBuf, _ := json.Marshal(payload)
	err := json.Unmarshal(jsonBuf, &payloadMap)
	if err != nil {
		return "", errs.NewError(errs.Deserialization, err.Error())
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloadMap))
	signedStr, err := token.SignedString(j.signingKey)
	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	return signedStr, nil
}

func (j JWTAuthority) DecodeToken(ct context.Context, jwtToken string, output interface{}) *errs.Error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	if !token.Valid {
		return errs.NewError(errs.InvalidArgument, err.Error())
	}

	return j.parseJWTClaims(ct, token.Claims, output)
}

func (j JWTAuthority) DecodeUnverifiedToken(ct context.Context, jwtToken string, output interface{}) *errs.Error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &claims)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return j.parseJWTClaims(ct, claims, output)
}

func (j JWTAuthority) parseJWTClaims(ct context.Context, claims jwt.Claims, output interface{}) *errs.Error {
	buf, err := json.Marshal(claims)
	if err != nil {
		return errs.NewError(errs.Serialization, err.Error())
	}

	err = json.Unmarshal(buf, output)
	if err != nil {
		return errs.NewError(errs.Deserialization, err.Error())
	}

	return nil
}

func NewJWTAuthority(dataCollector telemetry.DataCollector, signingKey string) JWTAuthority {
	return JWTAuthority{
		dataCollector: dataCollector,
		signingKey:    []byte(signingKey),
	}
}
