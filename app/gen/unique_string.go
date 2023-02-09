package gen

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type UniqueString struct {
	dataCollector telemetry.DataCollector
	uniqueNumGen  *UniqueNumber
	stringLen     int
	alphabet      []rune
}

func (u UniqueString) GenerateUniqueString(ct context.Context) (string, *errs.Error) {
	currNum, err := u.uniqueNumGen.GenerateUniqueNumber(ct)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	return u.toString(currNum), nil
}

func (u UniqueString) toString(num uint64) string {
	base := uint64(len(u.alphabet))
	resultRunes := make([]rune, u.stringLen)
	for strRuneIndex := 0; strRuneIndex < u.stringLen; strRuneIndex++ {
		alphabetRuneIndex := num % base
		num /= base
		alphabetRune := u.alphabet[alphabetRuneIndex]
		resultRunes[strRuneIndex] = alphabetRune
	}

	return string(resultRunes)
}

func NewUniqueString(
	dataCollector telemetry.DataCollector,
	name string,
	stringLen int,
	alphabet string,
	uniqueNumFactory UniqueNumberFactory,
) (UniqueString, *errs.Error) {
	numNum, err := uniqueNumFactory.MakeUniqueNumber(name)
	if err != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return UniqueString{}, err
	}

	return UniqueString{
		dataCollector: dataCollector,
		uniqueNumGen:  numNum,
		stringLen:     stringLen,
		alphabet:      []rune(alphabet),
	}, nil
}
