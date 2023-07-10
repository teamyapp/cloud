package service

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type UniqueStringGen struct {
	uniqueNumGen *UniqueNumberGen
	stringLen    int
	alphabet     []rune
}

func (u UniqueStringGen) GenerateUniqueString(ct context.Context) (string, *errs.Error) {
	currNum, err := u.uniqueNumGen.GenerateUniqueNumber(ct)
	if err != nil {
		return "", err
	}

	return u.toString(currNum), nil
}

func (u UniqueStringGen) toString(num uint64) string {
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

func NewUniqueStringGen(
	name string,
	stringLen int,
	alphabet string,
	uniqueNumGenRegistry *UniqueNumberGenRegistry,
) (UniqueStringGen, *errs.Error) {
	numNum, err := uniqueNumGenRegistry.GetUniqueNumberGen(name)
	if err != nil {
		return UniqueStringGen{}, err
	}

	return UniqueStringGen{
		uniqueNumGen: numNum,
		stringLen:    stringLen,
		alphabet:     []rune(alphabet),
	}, nil
}
