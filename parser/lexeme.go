package parser

import (
	"fmt"
	"math"

	"github.com/bzick/tokenizer"
	"github.com/chris-pikul/vole/utils"
)

type LexemeType byte

const (
	LexemeTypeInvalid LexemeType = iota
	LexemeTypeEOL
	LexemeTypeComment
	LexemeTypeDirective
	LexemeTypeIdentifier
	LexemeTypeAssignmentOperator
	LexemeTypeType
	LexemeTypeDefinition
	LexemeTypeDefinitionClass
	LexemeTypeFunctionDeclaration
	LexemeTypeFunction
	LexemeTypeArgument
	LexemeTypeArgumentType
	LexemeTypeReturnIdentifier
	LexemeTypeReturnType
)

type Lexeme struct {
	typ     LexemeType
	content []byte
	offset  uint64
	line    uint
	length  uint
}

func (lex *Lexeme) AddToken(tkn *tokenizer.Token) {
	lex.content = append(lex.content, tkn.Value()...)
	lex.offset = utils.Min(lex.offset, uint64(tkn.Offset()))
	lex.line = utils.Min(lex.line, uint(tkn.Line()))
	lex.length += uint(len(tkn.Indent()) + len(tkn.Value()))
}

func (l Lexeme) DebugPrint() {
	switch l.typ {
	case LexemeTypeEOL:
		fmt.Print("[/]")
	case LexemeTypeComment:
		fmt.Printf("[c %s]", l.content)
	case LexemeTypeDirective:
		fmt.Printf("[D %s]", l.content)
	case LexemeTypeIdentifier:
		fmt.Printf("[I %s]", l.content)
	case LexemeTypeAssignmentOperator:
		fmt.Printf("[O %s]", l.content)
	case LexemeTypeType:
		fmt.Printf("[t %s]", l.content)
	case LexemeTypeDefinition:
		fmt.Printf("[T %s]", l.content)
	case LexemeTypeDefinitionClass:
		fmt.Printf("[Tc %s]", l.content)
	case LexemeTypeFunctionDeclaration:
		fmt.Printf("[F %s]", l.content)
	case LexemeTypeFunction:
		fmt.Printf("[f %s]", l.content)
	case LexemeTypeArgument:
		fmt.Printf("[fa %s]", l.content)
	case LexemeTypeArgumentType:
		fmt.Printf("[Fat %s]", l.content)
	case LexemeTypeReturnIdentifier:
		fmt.Printf("[Fr %s]", l.content)
	case LexemeTypeReturnType:
		fmt.Printf("[Frt %s]", l.content)
	default:
		fmt.Printf("[X %s]", l.content)
	}
}

func NewLexeme(typ LexemeType, tokens ...*tokenizer.Token) Lexeme {
	lex := Lexeme{typ: typ}

	if len(tokens) > 0 {
		lex.offset = math.MaxUint64
		for _, tkn := range tokens {
			lex.AddToken(tkn)
		}
	}

	return lex
}
