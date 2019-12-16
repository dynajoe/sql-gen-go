package generator

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Bindvar types supported by Rebind, BindMap and BindStruct.
const (
	UNKNOWN = iota
	QUESTION
	DOLLAR
	NAMED
	AT
)

var allowedBindRunes = []*unicode.RangeTable{unicode.Letter, unicode.Digit}

type parseNamedState int

const (
	parseStateConsumingIdent parseNamedState = iota
	parseStateQuery
	parseStateQuotedIdent
	parseStateStringConstant
	parseStateLineComment
	parseStateBlockComment
	parseStateSkipThenTransition
	parseStateDollarQuoteLiteral
)

type parseNamedContext struct {
	state parseNamedState
	data  map[string]interface{}
}

const (
	colon        = ':'
	backSlash    = '\\'
	forwardSlash = '/'
	singleQuote  = '\''
	dash         = '-'
	star         = '*'
	newLine      = '\n'
	dollarSign   = '$'
	doubleQuote  = '"'
)

// Query parses a sql query and rewrites it to include bind parameters
func Query(qs []byte, bindType int, combineDuplicate bool) (string, []string, error) {
	var result strings.Builder
	var params []string
	paramLookup := make(map[string]int)

	addParam := func(paramName string) {
		var paramIndex int

		// Question bindtype is not compatible with combining duplicates
		if combineDuplicate && bindType != QUESTION {
			if i, ok := paramLookup[paramName]; ok {
				paramIndex = i
			} else {
				params = append(params, paramName)
				paramIndex = len(params)
				paramLookup[paramName] = paramIndex
			}
		} else {
			params = append(params, paramName)
			paramIndex = len(params)
		}

		switch bindType {
		// oracle only supports named type bind vars even for positional
		case NAMED:
			result.WriteByte(':')
			result.WriteString(paramName)
		case QUESTION, UNKNOWN:
			result.WriteByte('?')
		case DOLLAR:
			result.WriteByte('$')
			result.WriteString(strconv.Itoa(paramIndex))
		case AT:
			result.WriteString("@p")
			result.WriteString(strconv.Itoa(paramIndex))
		}
	}

	isRuneStartOfIdent := func(r rune) bool {
		return unicode.In(r, unicode.Letter) || r == '_'
	}

	isRunePartOfIdent := func(r rune) bool {
		return isRuneStartOfIdent(r) || unicode.In(r, allowedBindRunes...) || r == '_' || r == '.'
	}

	ctx := parseNamedContext{state: parseStateQuery}

	setState := func(s parseNamedState, d map[string]interface{}) {
		ctx.data = d
		ctx.state = s
	}

	var previousRune rune
	maxIndex := len(qs)

	for byteIndex := 0; byteIndex < maxIndex; {
		currentRune, runeWidth := utf8.DecodeRune(qs[byteIndex:])
		nextRuneByteIndex := byteIndex + runeWidth

		nextRune := utf8.RuneError
		if nextRuneByteIndex < maxIndex {
			nextRune, _ = utf8.DecodeRune(qs[nextRuneByteIndex:])
		}

		writeCurrentRune := true
		switch ctx.state {
		case parseStateQuery:
			if currentRune == colon && previousRune != colon && isRuneStartOfIdent(nextRune) {
				// :foo
				writeCurrentRune = false
				setState(parseStateConsumingIdent, map[string]interface{}{
					"ident": &strings.Builder{},
				})
			} else if currentRune == singleQuote && previousRune != backSlash {
				// \'
				setState(parseStateStringConstant, nil)
			} else if currentRune == dash && nextRune == dash {
				// -- single line comment
				setState(parseStateLineComment, nil)
			} else if currentRune == forwardSlash && nextRune == star {
				// /*
				setState(parseStateSkipThenTransition, map[string]interface{}{
					"state": parseStateBlockComment,
					"data": map[string]interface{}{
						"depth": 1,
					},
				})
			} else if currentRune == dollarSign && previousRune == dollarSign {
				// $$
				setState(parseStateDollarQuoteLiteral, nil)
			} else if currentRune == doubleQuote {
				// "foo"."bar"
				setState(parseStateQuotedIdent, nil)
			}
		case parseStateConsumingIdent:
			if isRunePartOfIdent(currentRune) {
				ctx.data["ident"].(*strings.Builder).WriteRune(currentRune)
				writeCurrentRune = false
			} else {
				addParam(ctx.data["ident"].(*strings.Builder).String())
				setState(parseStateQuery, nil)
			}
		case parseStateBlockComment:
			if previousRune == star && currentRune == forwardSlash {
				newDepth := ctx.data["depth"].(int) - 1
				if newDepth == 0 {
					setState(parseStateQuery, nil)
				} else {
					ctx.data["depth"] = newDepth
				}
			}
		case parseStateLineComment:
			if currentRune == newLine {
				setState(parseStateQuery, nil)
			}
		case parseStateStringConstant:
			if currentRune == singleQuote && previousRune != backSlash {
				setState(parseStateQuery, nil)
			}
		case parseStateDollarQuoteLiteral:
			if currentRune == dollarSign && previousRune != dollarSign {
				setState(parseStateQuery, nil)
			}
		case parseStateQuotedIdent:
			if currentRune == doubleQuote {
				setState(parseStateQuery, nil)
			}
		case parseStateSkipThenTransition:
			setState(ctx.data["state"].(parseNamedState), ctx.data["data"].(map[string]interface{}))
		default:
			setState(parseStateQuery, nil)
		}

		if writeCurrentRune {
			result.WriteRune(currentRune)
		}

		previousRune = currentRune
		byteIndex = nextRuneByteIndex
	}

	// If parsing left off while consuming an ident, add that ident to params
	if ctx.state == parseStateConsumingIdent {
		addParam(ctx.data["ident"].(*strings.Builder).String())
	}

	return result.String(), params, nil
}
