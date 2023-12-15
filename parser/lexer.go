package parser

import (
	"fmt"

	"github.com/bzick/tokenizer"
)

const (
	TokenInlineComment tokenizer.TokenKey = iota + 1
	TokenBlockComment
	TokenDirective
	TokenFragmentOpen
	TokenFragmentClose
	TokenParenOpen
	TokenParenClose
	TokenComparisonOperator
	TokenAssignmentOperator
	TokenDotOperator
	TokenComma
)

const (
	tmNone = iota
	tmAssignment
	tmIdentifier
	tmArguments
	tmType
	tmReturns
)

type Lexer struct {
	tknzr   *tokenizer.Tokenizer
	Lexemes []Lexeme
}

func (l *Lexer) Tokenize(source []byte) {
	stream := l.tknzr.ParseBytes(source)
	defer stream.Close()

	l.Lexemes = make([]Lexeme, 0)

	previousLexeme := func() *Lexeme {
		if len(l.Lexemes) > 1 {
			return &l.Lexemes[len(l.Lexemes)-1]
		}
		return nil
	}

	addLexeme := func(typ LexemeType, tkns ...*tokenizer.Token) {
		l.Lexemes = append(l.Lexemes, NewLexeme(typ, tkns...))
	}

	var cur *tokenizer.Token
	var next *tokenizer.Token
	line := -1
	posLineStart := 0
	mode := tmNone
	depth := 0
	isDecl := false

	for stream.IsValid() {
		cur = stream.CurrentToken()
		next = stream.NextToken()

		if line > 0 && cur.Line() > line {
			// Line has changed, we care again about stuff we didn't?
			addLexeme(LexemeTypeEOL)
			mode = tmNone
			posLineStart = cur.Offset()
		}
		line = cur.Line()

		switch cur.Key() {
		case TokenInlineComment, TokenBlockComment:
			addLexeme(LexemeTypeComment, cur)
		case TokenDirective:
			// We only care about additionals on the same line
			if cur.Line() == next.Line() {
				// Check if the next value is right next to it
				if next.Key() == tokenizer.TokenKeyword && len(next.Indent()) == 0 {
					addLexeme(LexemeTypeDirective, cur, next)
					stream.GoNext()
				}
			}
		case tokenizer.TokenKeyword:
			// We only care about additional on the same line
			if cur.Line() == next.Line() {
				// Check if we have assignment next
				if next.Key() == TokenAssignmentOperator {
					if mode == tmIdentifier {
						// We already where working on an identifier
						previousLexeme().AddToken(cur)
						addLexeme(LexemeTypeAssignmentOperator, next)
					} else {
						// This token is an identifier by itself
						addLexeme(LexemeTypeIdentifier, cur)
						addLexeme(LexemeTypeAssignmentOperator, next)
					}

					stream.GoNext()
				} else if next.Key() == TokenDotOperator && len(next.Indent()) == 0 {
					// Identifier going into something else
					if mode == tmIdentifier {
						previousLexeme().AddToken(cur)
					} else {
						addLexeme(LexemeTypeIdentifier, cur, next)
					}

					mode = tmIdentifier
					stream.GoNext()
				} else if next.Key() == TokenParenOpen && len(next.Indent()) == 0 {
					// If the previous was an identifier and there is no spacing, it is continuation
					if len(l.Lexemes) > 1 && len(cur.Indent()) == 0 && previousLexeme().typ == LexemeTypeIdentifier {
						previousLexeme().AddToken(cur)
						previousLexeme().typ = LexemeTypeFunction
					} else {
						// Function of some sorts
						if cur.Offset()-posLineStart == 0 {
							// First item on root line means declaration
							addLexeme(LexemeTypeFunctionDeclaration, cur)
							isDecl = true
						} else {
							addLexeme(LexemeTypeFunction, cur)
							isDecl = false
						}
					}

					mode = tmArguments
					depth++
					stream.GoNext()
				} else if mode == tmArguments {
					// Currently looking for arguments this should be identifier
					addLexeme(LexemeTypeArgument, cur)

					if next.Key() == tokenizer.TokenKeyword {
						// Next will be a type
						addLexeme(LexemeTypeArgumentType, next)
						stream.GoNext()
					} else if next.Key() == TokenComma {
						// Continuing arguments, consume this one
						stream.GoNext()
					}
				} else if mode == tmReturns {
					// In return mode with more than one keyword
					if next.Key() == TokenComma || next.Key() == TokenParenClose {
						// No, just a type by itself
						addLexeme(LexemeTypeReturnType, cur)

						if next.Key() == TokenComma {
							stream.GoNext()
						}
					} else {
						addLexeme(LexemeTypeReturnIdentifier, cur)
					}
				} else if next.Key() == tokenizer.TokenKeyword {
					// Type definition starting
					addLexeme(LexemeTypeDefinition, cur)
					addLexeme(LexemeTypeDefinitionClass, next)
					stream.GoNext()
				} else {
					// FALLTHROUGH CASE
					addLexeme(LexemeTypeInvalid, cur)
				}
			} else if mode == tmReturns {
				// Return mode with nothing more on the line
				addLexeme(LexemeTypeReturnType, cur)
			} else {
				// FALLTHROUGH CASE FOR LAST OF LINE
				addLexeme(LexemeTypeInvalid, cur)
			}
		case TokenParenClose:
			if mode == tmArguments {
				if isDecl {
					// It was a declaration, we could be doing return statements next
					if (next.Key() == tokenizer.TokenKeyword || next.Key() == TokenParenOpen) && next.Line() == cur.Line() {
						// Will be doing return statement
						mode = tmReturns
					} else {
						mode = tmNone
					}
				} else {
					depth--
					if depth == 0 {
						mode = tmNone
						isDecl = false
					}
				}
			} else {
				addLexeme(LexemeTypeInvalid, cur)
			}
		default:
			if mode == tmArguments {
				addLexeme(LexemeTypeArgument, cur)
			} else {
				addLexeme(LexemeTypeInvalid, cur)
			}
		}
		stream.GoNext()
	}
}

func (l Lexer) DebugPrint() {
	fmt.Print("Lexer Output: START ")
	for _, lex := range l.Lexemes {
		lex.DebugPrint()
	}
	fmt.Print(" END \n")
}

func NewLexer() Lexer {
	lexer := Lexer{
		tknzr: tokenizer.New(),
	}

	lexer.tknzr.AllowKeywordSymbols(tokenizer.Underscore, tokenizer.Numbers)
	lexer.tknzr.DefineTokens(TokenParenClose, []string{")"})
	lexer.tknzr.DefineTokens(TokenParenOpen, []string{"("})
	lexer.tknzr.DefineTokens(TokenFragmentOpen, []string{"${"})
	lexer.tknzr.DefineTokens(TokenFragmentClose, []string{"}"})
	lexer.tknzr.DefineStringToken(TokenInlineComment, "//", "\n")
	lexer.tknzr.DefineStringToken(TokenBlockComment, "/*", "*/").AddSpecialStrings(tokenizer.DefaultSpecialString)
	lexer.tknzr.DefineStringToken(tokenizer.TokenString, "\"", "\"").SetEscapeSymbol('\\')
	lexer.tknzr.DefineStringToken(tokenizer.TokenString, "'", "'").SetEscapeSymbol('\\')
	lexer.tknzr.DefineStringToken(tokenizer.TokenString, "`", "`").SetEscapeSymbol('\\')

	lexer.tknzr.DefineTokens(TokenComparisonOperator, []string{"==", "!=", ">=", "<="})
	lexer.tknzr.DefineTokens(TokenAssignmentOperator, []string{"=", ":=", "+=", "-=", "*=", "/=", "^=", "&=", "|=", "<|="})
	lexer.tknzr.DefineTokens(TokenDotOperator, []string{"."})
	lexer.tknzr.DefineTokens(TokenComma, []string{","})
	lexer.tknzr.DefineTokens(TokenDirective, []string{"#"})

	return lexer
}
