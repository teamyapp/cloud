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
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return "", internalErr
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payloadMap))
	signedStr, err := token.SignedString(j.signingKey)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return "", internalErr
	}

	return signedStr, nil
}

func (j JWTAuthority) DecodeToken(ct context.Context, jwtToken string, output interface{}) *errs.Error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.signingKey, nil
	})
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	if !token.Valid {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return j.parseJWTClaims(ct, token.Claims, output)
}

func (j JWTAuthority) DecodeUnverifiedToken(ct context.Context, jwtToken string, output interface{}) *errs.Error {
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(jwtToken, &claims)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return j.parseJWTClaims(ct, claims, output)
}

func (j JWTAuthority) parseJWTClaims(ct context.Context, claims jwt.Claims, output interface{}) *errs.Error {
	buf, err := json.Marshal(claims)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Serialization,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	err = json.Unmarshal(buf, output)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		j.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewJWTAuthority(dataCollector telemetry.DataCollector, signingKey string) JWTAuthority {
	return JWTAuthority{
		dataCollector: dataCollector,
		signingKey:    []byte(signingKey),
	}
}
