package gen

import (
	"github.com/teamyapp/cloud/libs/obs"
)

type UniqueString struct {
	dataCollector obs.DataCollector
	uniqueNumGen  *UniqueNumber
	stringLen     int
	alphabet      []rune
}

func (u UniqueString) GenerateUniqueString() (string, error) {
	currNum, err := u.uniqueNumGen.GenerateUniqueNumber()
	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	dataCollector obs.DataCollector,
	name string,
	stringLen int,
	alphabet string,
	uniqueNumFactory UniqueNumberFactory,
) (UniqueString, error) {
	numNum, err := uniqueNumFactory.MakeUniqueNumber(name)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return UniqueString{}, err
	}

	return UniqueString{
		dataCollector: dataCollector,
		uniqueNumGen:  numNum,
		stringLen:     stringLen,
		alphabet:      []rune(alphabet),
	}, nil
}
