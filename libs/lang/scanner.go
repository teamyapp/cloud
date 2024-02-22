package lang

import (
	"fmt"
	"strconv"
	"unicode"
)

type Scanner struct {
	source        []rune
	line          int
	nextColumn    int
	nextRuneIndex int
}

func (s *Scanner) ScanTokens(source string) ([]Token, []Err) {
	s.source = []rune(source)
	tokens := make([]Token, 0)
	errs := make([]Err, 0)
	for s.nextRuneIndex < len(s.source) {
		token, err := s.scanToken()
		if err != nil {
			errs = append(errs, *err)
			continue
		}

		if token != nil {
			tokens = append(tokens, *token)
		}
	}

	tokens = append(tokens, Token{
		Type:   EOFTokenType,
		Line:   s.line,
		Column: s.nextColumn,
	})
	return tokens, errs
}

func (s *Scanner) scanToken() (*Token, *Err) {
	startColumn := s.nextColumn
	r := s.readRune()
	switch r {
	case '(':
		return &Token{
			Type:   LeftParenthesisTokenType,
			Lexeme: "(",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case ')':
		return &Token{
			Type:   RightParenthesisTokenType,
			Lexeme: ")",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '{':
		return &Token{
			Type:   LeftBraceTokenType,
			Lexeme: "{",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '}':
		return &Token{
			Type:   RightBraceTokenType,
			Lexeme: "}",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '[':
		return &Token{
			Type:   LeftBracketTokenType,
			Lexeme: "[",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case ']':
		return &Token{
			Type:   RightBracketTokenType,
			Lexeme: "]",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case ',':
		return &Token{
			Type:   CommaTokenType,
			Lexeme: ",",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case ';':
		return &Token{
			Type:   SemicolonTokenType,
			Lexeme: ";",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case ':':
		return &Token{
			Type:   ColonTokenType,
			Lexeme: ":",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '?':
		return &Token{
			Type:   QuestionMarkTokenType,
			Lexeme: "?",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '.':
		return &Token{
			Type:   DotTokenType,
			Lexeme: ".",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '+':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   AddAssignTokenType,
				Lexeme: "+=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   AddTokenType,
			Lexeme: "+",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '-':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   SubtractAssignTokenType,
				Lexeme: "-=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   MinusTokenType,
			Lexeme: "-",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '*':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   MultiplyAssignTokenType,
				Lexeme: "*=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   StarTokenType,
			Lexeme: "*",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '/':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '=':
				s.readRune()
				return &Token{
					Type:   DivideAssignTokenType,
					Lexeme: "/=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			case '/':
				return s.scanSingleLineComment(startColumn), nil
			}
		}

		return &Token{
			Type:   DivideTokenType,
			Lexeme: "/",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '%':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   ModuloAssignTokenType,
				Lexeme: "%=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   ModuloTokenType,
			Lexeme: "%",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '=':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   LogicalEqualTokenType,
				Lexeme: "==",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   AssignTokenType,
			Lexeme: "=",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '!':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   LogicalNotEqualTokenType,
				Lexeme: "!=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   LogicalNotTokenType,
			Lexeme: "!",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '>':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '=':
				s.readRune()
				return &Token{
					Type:   GreaterThanOrEqualTokenType,
					Lexeme: ">=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			case '>':
				s.readRune()
				return &Token{
					Type:   BitwiseRightShiftTokenType,
					Lexeme: ">>",
					Line:   s.line,
					Column: startColumn,
				}, nil
			}
		}

		return &Token{
			Type:   GreaterThanTokenType,
			Lexeme: ">",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '<':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '=':
				s.readRune()
				return &Token{
					Type:   LessThanOrEqualTokenType,
					Lexeme: "<=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			case '<':
				s.readRune()
				return &Token{
					Type:   BitwiseLeftShiftTokenType,
					Lexeme: "<<",
					Line:   s.line,
					Column: startColumn,
				}, nil
			}
		}

		return &Token{
			Type:   LessThanTokenType,
			Lexeme: "<",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '&':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '&':
				s.readRune()
				return &Token{
					Type:   LogicalAndTokenType,
					Lexeme: "&&",
					Line:   s.line,
					Column: startColumn,
				}, nil
			case '=':
				s.readRune()
				return &Token{
					Type:   BitwiseAndAssignTokenType,
					Lexeme: "&=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			}
		}

		return &Token{
			Type:   BitwiseAndTokenType,
			Lexeme: "&",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '|':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '|':
				s.readRune()
				return &Token{
					Type:   LogicalOrTokenType,
					Lexeme: "||",
					Line:   s.line,
					Column: startColumn,
				}, nil
			case '=':
				s.readRune()
				return &Token{
					Type:   BitwiseOrAssignTokenType,
					Lexeme: "|=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			}
		}

		return &Token{
			Type:   BitwiseOrTokenType,
			Lexeme: "|",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '^':
		nextRune, ok := s.peekRune()
		if ok {
			switch nextRune {
			case '=':
				s.readRune()
				return &Token{
					Type:   BitwiseXorAssignTokenType,
					Lexeme: "^=",
					Line:   s.line,
					Column: startColumn,
				}, nil
			}
		}

		return &Token{
			Type:   BitwiseXorTokenType,
			Lexeme: "^",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '~':
		nextRune, ok := s.peekRune()
		if ok && nextRune == '=' {
			s.readRune()
			return &Token{
				Type:   BitwiseNotAssignTokenType,
				Lexeme: "~=",
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   BitwiseNotTokenType,
			Lexeme: "~",
			Line:   s.line,
			Column: startColumn,
		}, nil
	case '\n':
		s.line++
		s.nextColumn = 1
		return nil, nil
	case '\r':
		return nil, nil
	case '\t':
		return nil, nil
	case ' ':
		s.scanWhitespaces(startColumn)
		return nil, nil
	case '"':
		return s.scanString(startColumn)
	}

	if unicode.IsDigit(r) {
		return s.scanNumber(startColumn)
	}

	if unicode.IsLetter(r) || r == '_' {
		identifier := s.scanIdentifier()
		switch identifier {
		case "true":
			return &Token{
				Type:   BoolTokenType,
				Lexeme: identifier,
				Value:  true,
				Line:   s.line,
				Column: startColumn,
			}, nil
		case "false":
			return &Token{
				Type:   BoolTokenType,
				Lexeme: identifier,
				Value:  false,
				Line:   s.line,
				Column: startColumn,
			}, nil
		case "nil":
			return &Token{
				Type:   NilTokenType,
				Lexeme: identifier,
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		if tokenType, ok := keywords[identifier]; ok {
			return &Token{
				Type:   tokenType,
				Lexeme: identifier,
				Line:   s.line,
				Column: startColumn,
			}, nil
		}

		return &Token{
			Type:   IdentifierTokenType,
			Lexeme: identifier,
			Value:  identifier,
			Line:   s.line,
			Column: startColumn,
		}, nil
	}

	return nil, &Err{
		Message: fmt.Sprintf("Unexpected rune: %v", string(r)),
		Line:    s.line,
		Column:  startColumn,
	}
}

func (s *Scanner) readRune() rune {
	c := s.source[s.nextRuneIndex]
	s.nextRuneIndex++
	s.nextColumn++
	return c
}

func (s *Scanner) peekRune() (rune, bool) {
	if s.nextRuneIndex >= len(s.source) {
		return 0, false
	}

	return s.source[s.nextRuneIndex], true
}

func (s *Scanner) scanSingleLineComment(startColumn int) *Token {
	startIndex := s.nextRuneIndex
	for {
		r, ok := s.peekRune()
		if !ok || r == '\n' {
			break
		}

		s.readRune()
	}

	lexeme := string(s.source[startIndex-2 : s.nextRuneIndex])
	return &Token{
		Type:   CommentTokenType,
		Lexeme: lexeme,
		Value:  s.source[startIndex:s.nextRuneIndex],
		Line:   s.line,
		Column: startColumn,
	}
}

func (s *Scanner) scanWhitespaces(startColumn int) *Token {
	startIndex := s.nextRuneIndex - 1
	for {
		r, ok := s.peekRune()
		if !ok || r != ' ' {
			break
		}

		s.readRune()
	}

	whitespaces := string(s.source[startIndex:s.nextRuneIndex])
	return &Token{
		Type:   WhitespaceTokenType,
		Lexeme: whitespaces,
		Value:  whitespaces,
		Line:   s.line,
		Column: startColumn,
	}
}

func (s *Scanner) scanString(startColumn int) (*Token, *Err) {
	startIndex := s.nextRuneIndex - 1
	for {
		r, ok := s.peekRune()
		if !ok {
			return nil, &Err{
				Message: "Unterminated string",
				Line:    s.line,
				Column:  startColumn,
			}
		}

		s.readRune()
		if r == '"' {
			break
		}
	}

	lexeme := string(s.source[startIndex:s.nextRuneIndex])
	return &Token{
		Type:   StringTokenType,
		Lexeme: lexeme,
		Value:  lexeme[1 : len(lexeme)-1],
		Line:   s.line,
		Column: startColumn,
	}, nil
}

func (s *Scanner) scanNumber(startColumn int) (*Token, *Err) {
	startIndex := s.nextRuneIndex - 1
	for {
		r, ok := s.peekRune()
		if !ok {
			break
		}

		if !unicode.IsDigit(r) {
			break
		}

		s.readRune()
	}

	r, ok := s.peekRune()
	if !ok || r != '.' {
		lexeme := string(s.source[startIndex:s.nextRuneIndex])
		num, _ := strconv.ParseInt(lexeme, 10, 64)
		return &Token{
			Type:   IntTokenType,
			Lexeme: lexeme,
			Value:  num,
			Line:   s.line,
			Column: startColumn,
		}, nil
	}

	s.readRune()
	foundDigit := false
	for {
		r, ok := s.peekRune()
		if !ok {
			break
		}

		if !unicode.IsDigit(r) {
			break
		}

		s.readRune()
		foundDigit = true
	}

	if !foundDigit {
		return nil, &Err{
			Message: "Invalid decimal number",
			Line:    s.line,
			Column:  startColumn,
		}
	}

	lexeme := string(s.source[startIndex:s.nextRuneIndex])
	num, _ := strconv.ParseFloat(lexeme, 64)
	return &Token{
		Type:   DecimalTokenType,
		Lexeme: lexeme,
		Value:  num,
		Line:   s.line,
		Column: startColumn,
	}, nil
}

func (s *Scanner) scanIdentifier() string {
	startIndex := s.nextRuneIndex - 1
	for {
		r, ok := s.peekRune()
		if !ok {
			break
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}

		s.readRune()
	}

	return string(s.source[startIndex:s.nextRuneIndex])
}

func NewScanner() *Scanner {
	return &Scanner{
		line:          1,
		nextColumn:    1,
		nextRuneIndex: 0,
	}
}
