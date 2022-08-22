// Time duration similar to ISO 8601
package duration

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
	"unicode"
)

const (
	periodSymbol rune = 'P'
	yearSymbol        = 'Y'
	monthSymbol       = 'M'
	weekSymbol        = 'W'
	daySymbol         = 'D'
	timeSymbol        = 'T'
	hourSymbol        = 'H'
	minuteSymbol      = 'M'
	secondSymbol      = 'S'
)

var periodSymbolOrder = []rune{
	yearSymbol,
	monthSymbol,
	weekSymbol,
	daySymbol,
}

var timeSymbolOrder = []rune{
	hourSymbol,
	minuteSymbol,
	secondSymbol,
}

var periodSymbolIndices = map[rune]int{
	yearSymbol:  0,
	monthSymbol: 1,
	weekSymbol:  2,
	daySymbol:   3,
}

var timeSymbolIndices = map[rune]int{
	hourSymbol:   0,
	minuteSymbol: 1,
	secondSymbol: 2,
}

const (
	secondInNanos = 1_000_000_000
	minuteInNanos = 60 * secondInNanos
	hourInNanos   = 60 * minuteInNanos
	dayInNanos    = 24 * hourInNanos
	weekInNanos   = 7 * dayInNanos
	monthInNanos  = 30 * dayInNanos
	yearInNanos   = 365 * dayInNanos
)

var periodSymbolInNanos = map[rune]int64{
	yearSymbol:  yearInNanos,
	monthSymbol: monthInNanos,
	weekSymbol:  weekInNanos,
	daySymbol:   dayInNanos,
}

var timeSymbolInNanos = map[rune]int64{
	hourSymbol:   hourInNanos,
	minuteSymbol: minuteInNanos,
	secondSymbol: secondInNanos,
}

func Parse(input string) (time.Duration, error) {
	if len(input) == 0 || input[0] != uint8(periodSymbol) {
		return 0, fmt.Errorf("duration must start with 'P': duration=%v", input)
	}

	input = input[1:]
	seenSymbols := map[rune]int{}
	visitTimeSection := false
	num := 0
	var hasPeriod bool
	var hasTime bool
	var totalNanoSeconds int64

	for index, currRune := range input {
		if unicode.IsDigit(currRune) {
			digit := (int)(currRune - '0')
			num = num*10 + digit
			continue
		}

		if currRune == timeSymbol {
			seenSymbols = map[rune]int{}
			visitTimeSection = true
			continue
		}

		if visitTimeSection {
			err := validateSymbol(timeSymbolOrder, timeSymbolIndices, seenSymbols, currRune, index)
			if err != nil {
				log.Println(err)
				return 0, err
			}

			totalNanoSeconds += timeSymbolInNanos[currRune] * int64(num)
			hasTime = true
		} else {
			err := validateSymbol(periodSymbolOrder, periodSymbolIndices, seenSymbols, currRune, index)
			if err != nil {
				log.Println(err)
				return 0, err
			}

			totalNanoSeconds += periodSymbolInNanos[currRune] * int64(num)
			hasPeriod = true
		}

		seenSymbols[currRune] = index
		num = 0
	}

	if !hasPeriod && !hasTime {
		return 0, fmt.Errorf("must has either period or time")
	}

	if visitTimeSection && !hasTime {
		return 0, fmt.Errorf("must remove ending T or have non empty time section")
	}

	return time.Duration(totalNanoSeconds), nil
}

func Format(duration time.Duration) string {
	nanos := int64(duration)
	if nanos == 0 || nanos < secondInNanos {
		return "PT0S"
	}

	years := nanos / yearInNanos
	nanos %= yearInNanos
	months := nanos / monthInNanos
	nanos %= monthInNanos
	weeks := nanos / weekInNanos
	nanos %= weekInNanos
	days := nanos / dayInNanos
	nanos %= dayInNanos
	hours := nanos / hourInNanos
	nanos %= hourInNanos
	minutes := nanos / minuteInNanos
	nanos %= minuteInNanos
	seconds := nanos / secondInNanos

	var buffer bytes.Buffer
	buffer.WriteRune(periodSymbol)
	tryWriteNumWithUnit(&buffer, years, yearSymbol)
	tryWriteNumWithUnit(&buffer, months, monthSymbol)
	tryWriteNumWithUnit(&buffer, weeks, weekSymbol)
	tryWriteNumWithUnit(&buffer, days, daySymbol)

	if hours > 0 || minutes > 0 || seconds > 0 {
		buffer.WriteRune(timeSymbol)
		tryWriteNumWithUnit(&buffer, hours, hourSymbol)
		tryWriteNumWithUnit(&buffer, minutes, minuteSymbol)
		tryWriteNumWithUnit(&buffer, seconds, secondSymbol)
	}

	return buffer.String()
}

func tryWriteNumWithUnit(buffer *bytes.Buffer, num int64, symbol rune) {
	if num <= 0 {
		return
	}

	buffer.WriteString(strconv.FormatInt(num, 10))
	buffer.WriteRune(symbol)
}

func validateSymbol(symbolOrder []rune, symbolIndices map[rune]int, seenSymbols map[rune]int, currSymbol rune, currIndex int) error {
	symbolIndex, ok := symbolIndices[currSymbol]
	if !ok {
		return fmt.Errorf("unsupported symbol: index=%v, symbol=%c", currIndex, currSymbol)
	}

	for index := symbolIndex; index < len(symbolOrder); index++ {
		symbol := symbolOrder[index]
		seenSymbolIndex, ok := seenSymbols[symbol]
		if ok {
			return fmt.Errorf("%c(%v) already showed up before %c(%v)",
				symbol,
				seenSymbolIndex,
				currSymbol,
				currIndex)
		}
	}

	return nil
}
