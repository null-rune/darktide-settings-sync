package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// TokenType represents lexical token types
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenBoolean
	TokenNil
	TokenAssign    // = or :
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenComma     // ,
	TokenSemicolon // ;
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// Tokenizer transforms SJSON / Darktide config content into a stream of tokens
type Tokenizer struct {
	input []rune
	pos   int
	line  int
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input: []rune(input),
		pos:   0,
		line:  1,
	}
}

func (t *Tokenizer) NextToken() Token {
	for {
		t.skipWhitespace()
		if t.pos >= len(t.input) {
			return Token{Type: TokenEOF, Line: t.line}
		}

		ch := t.input[t.pos]

		// Lua-style comments: -- ...
		if ch == '-' && t.peek(1) == '-' {
			for t.pos < len(t.input) && t.input[t.pos] != '\n' {
				t.pos++
			}
			continue
		}

		// C-style comments: // ...
		if ch == '/' && t.peek(1) == '/' {
			for t.pos < len(t.input) && t.input[t.pos] != '\n' {
				t.pos++
			}
			continue
		}

		// Punctuation
		switch ch {
		case '=', ':':
			t.pos++
			return Token{Type: TokenAssign, Value: "=", Line: t.line}
		case '{':
			t.pos++
			return Token{Type: TokenLBrace, Value: "{", Line: t.line}
		case '}':
			t.pos++
			return Token{Type: TokenRBrace, Value: "}", Line: t.line}
		case '[':
			t.pos++
			return Token{Type: TokenLBracket, Value: "[", Line: t.line}
		case ']':
			t.pos++
			return Token{Type: TokenRBracket, Value: "]", Line: t.line}
		case ',':
			t.pos++
			return Token{Type: TokenComma, Value: ",", Line: t.line}
		case ';':
			t.pos++
			return Token{Type: TokenSemicolon, Value: ";", Line: t.line}
		case '"', '\'':
			return t.readString(ch)
		}

		// Numbers: can start with digit or minus followed by digit/dot
		if unicode.IsDigit(ch) || (ch == '-' && (unicode.IsDigit(t.peek(1)) || t.peek(1) == '.')) {
			return t.readNumber()
		}

		// Identifiers or Booleans or Nil
		if unicode.IsLetter(ch) || ch == '_' {
			return t.readIdentifier()
		}

		// Fallback for unknown character
		t.pos++
		return Token{Type: TokenIdent, Value: string(ch), Line: t.line}
	}
}

func (t *Tokenizer) peek(offset int) rune {
	idx := t.pos + offset
	if idx >= len(t.input) {
		return 0
	}
	return t.input[idx]
}

func (t *Tokenizer) skipWhitespace() {
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == '\n' {
			t.line++
			t.pos++
		} else if unicode.IsSpace(ch) {
			t.pos++
		} else {
			break
		}
	}
}

func (t *Tokenizer) readString(quote rune) Token {
	t.pos++ // skip open quote
	startLine := t.line
	var sb strings.Builder

	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if ch == '\n' {
			t.line++
		}
		if ch == '\\' && t.pos+1 < len(t.input) {
			next := t.input[t.pos+1]
			switch next {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case '\'':
				sb.WriteRune('\'')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(next)
			}
			t.pos += 2
			continue
		}
		if ch == quote {
			t.pos++ // skip close quote
			return Token{Type: TokenString, Value: sb.String(), Line: startLine}
		}
		sb.WriteRune(ch)
		t.pos++
	}
	return Token{Type: TokenString, Value: sb.String(), Line: startLine}
}

func (t *Tokenizer) readNumber() Token {
	start := t.pos
	startLine := t.line
	if t.input[t.pos] == '-' {
		t.pos++
	}
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if unicode.IsDigit(ch) || ch == '.' || ch == 'e' || ch == 'E' || ch == '+' || ch == '-' {
			t.pos++
		} else {
			break
		}
	}
	return Token{Type: TokenNumber, Value: string(t.input[start:t.pos]), Line: startLine}
}

func (t *Tokenizer) readIdentifier() Token {
	start := t.pos
	startLine := t.line
	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			t.pos++
		} else {
			break
		}
	}
	val := string(t.input[start:t.pos])
	switch val {
	case "true", "false":
		return Token{Type: TokenBoolean, Value: val, Line: startLine}
	case "nil", "null":
		return Token{Type: TokenNil, Value: val, Line: startLine}
	default:
		return Token{Type: TokenIdent, Value: val, Line: startLine}
	}
}

// AST definitions

type Node interface {
	Serialize(indent int) string
	Clone() Node
}

type StringNode struct {
	Value string
}

func (n *StringNode) Serialize(indent int) string {
	escaped := strings.ReplaceAll(n.Value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	return fmt.Sprintf("%q", n.Value)
}

func (n *StringNode) Clone() Node {
	return &StringNode{Value: n.Value}
}

type NumberNode struct {
	Value string
}

func (n *NumberNode) Serialize(indent int) string {
	return n.Value
}

func (n *NumberNode) Clone() Node {
	return &NumberNode{Value: n.Value}
}

type BoolNode struct {
	Value bool
}

func (n *BoolNode) Serialize(indent int) string {
	if n.Value {
		return "true"
	}
	return "false"
}

func (n *BoolNode) Clone() Node {
	return &BoolNode{Value: n.Value}
}

type NilNode struct{}

func (n *NilNode) Serialize(indent int) string {
	return "nil"
}

func (n *NilNode) Clone() Node {
	return &NilNode{}
}

type ArrayNode struct {
	Elements []Node
}

func (n *ArrayNode) Serialize(indent int) string {
	if len(n.Elements) == 0 {
		return "[\n" + strings.Repeat("\t", indent) + "]"
	}
	var sb strings.Builder
	sb.WriteString("[\n")
	for _, elem := range n.Elements {
		sb.WriteString(strings.Repeat("\t", indent+1))
		sb.WriteString(elem.Serialize(indent + 1))
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("\t", indent))
	sb.WriteString("]")
	return sb.String()
}

func (n *ArrayNode) Clone() Node {
	cloned := &ArrayNode{
		Elements: make([]Node, len(n.Elements)),
	}
	for i, elem := range n.Elements {
		cloned.Elements[i] = elem.Clone()
	}
	return cloned
}

type ObjectEntry struct {
	Key       string
	KeyQuoted bool
	Value     Node
}

type ObjectNode struct {
	Entries []ObjectEntry
}

func (n *ObjectNode) Serialize(indent int) string {
	if len(n.Entries) == 0 {
		return "{\n" + strings.Repeat("\t", indent) + "}"
	}
	var sb strings.Builder
	sb.WriteString("{\n")
	for _, entry := range n.Entries {
		sb.WriteString(strings.Repeat("\t", indent+1))
		keyStr := formatKey(entry.Key, entry.KeyQuoted)
		sb.WriteString(keyStr)
		sb.WriteString(" = ")
		sb.WriteString(entry.Value.Serialize(indent + 1))
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("\t", indent))
	sb.WriteString("}")
	return sb.String()
}

func (n *ObjectNode) Clone() Node {
	cloned := &ObjectNode{
		Entries: make([]ObjectEntry, len(n.Entries)),
	}
	for i, entry := range n.Entries {
		cloned.Entries[i] = ObjectEntry{
			Key:       entry.Key,
			KeyQuoted: entry.KeyQuoted,
			Value:     entry.Value.Clone(),
		}
	}
	return cloned
}

func formatKey(key string, forceQuote bool) string {
	if forceQuote || needsQuoting(key) {
		return fmt.Sprintf("%q", key)
	}
	return key
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return true
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return true
			}
		}
	}
	// Reserve words
	if s == "true" || s == "false" || s == "nil" || s == "null" {
		return true
	}
	return false
}

// ConfigFile represents the root settings document
type ConfigFile struct {
	Entries []ObjectEntry
}

func (c *ConfigFile) Serialize() string {
	var sb strings.Builder
	for _, entry := range c.Entries {
		keyStr := formatKey(entry.Key, entry.KeyQuoted)
		sb.WriteString(keyStr)
		sb.WriteString(" = ")
		sb.WriteString(entry.Value.Serialize(0))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (c *ConfigFile) Clone() *ConfigFile {
	cloned := &ConfigFile{
		Entries: make([]ObjectEntry, len(c.Entries)),
	}
	for i, entry := range c.Entries {
		cloned.Entries[i] = ObjectEntry{
			Key:       entry.Key,
			KeyQuoted: entry.KeyQuoted,
			Value:     entry.Value.Clone(),
		}
	}
	return cloned
}

// Parser
type Parser struct {
	tokens []Token
	pos    int
}

func ParseConfigFile(content string) (*ConfigFile, error) {
	tok := NewTokenizer(content)
	var tokens []Token
	for {
		t := tok.NextToken()
		if t.Type == TokenEOF {
			break
		}
		tokens = append(tokens, t)
	}

	p := &Parser{tokens: tokens, pos: 0}
	cfg := &ConfigFile{}

	for p.pos < len(p.tokens) {
		p.skipSeparators()
		if p.pos >= len(p.tokens) {
			break
		}

		keyToken := p.tokens[p.pos]
		if keyToken.Type != TokenIdent && keyToken.Type != TokenString {
			return nil, fmt.Errorf("line %d: expected setting key identifier or string, got %v (%q)",
				keyToken.Line, keyToken.Type, keyToken.Value)
		}
		p.pos++

		p.skipSeparators()
		if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenAssign {
			return nil, fmt.Errorf("line %d: expected '=' after key %q", keyToken.Line, keyToken.Value)
		}
		p.pos++ // consume '='

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		cfg.Entries = append(cfg.Entries, ObjectEntry{
			Key:       keyToken.Value,
			KeyQuoted: keyToken.Type == TokenString,
			Value:     val,
		})

		p.skipSeparators()
	}

	return cfg, nil
}

func (p *Parser) skipSeparators() {
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == TokenComma || t.Type == TokenSemicolon {
			p.pos++
		} else {
			break
		}
	}
}

func (p *Parser) parseValue() (Node, error) {
	p.skipSeparators()
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of file while reading value")
	}

	cur := p.tokens[p.pos]
	switch cur.Type {
	case TokenLBrace:
		return p.parseObject()
	case TokenLBracket:
		return p.parseArray()
	case TokenString:
		p.pos++
		return &StringNode{Value: cur.Value}, nil
	case TokenNumber:
		p.pos++
		return &NumberNode{Value: cur.Value}, nil
	case TokenBoolean:
		p.pos++
		b, _ := strconv.ParseBool(cur.Value)
		return &BoolNode{Value: b}, nil
	case TokenNil:
		p.pos++
		return &NilNode{}, nil
	default:
		// Could be an unquoted string / ident as value (e.g. high, custom, en, etc.)
		p.pos++
		return &StringNode{Value: cur.Value}, nil
	}
}

func (p *Parser) parseObject() (*ObjectNode, error) {
	startLine := p.tokens[p.pos].Line
	p.pos++ // consume '{'
	obj := &ObjectNode{}

	for p.pos < len(p.tokens) {
		p.skipSeparators()
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("line %d: unclosed '{'", startLine)
		}

		if p.tokens[p.pos].Type == TokenRBrace {
			p.pos++ // consume '}'
			return obj, nil
		}

		keyToken := p.tokens[p.pos]
		if keyToken.Type != TokenIdent && keyToken.Type != TokenString {
			return nil, fmt.Errorf("line %d: expected object key, got %v (%q)",
				keyToken.Line, keyToken.Type, keyToken.Value)
		}
		p.pos++

		p.skipSeparators()
		if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != TokenAssign {
			return nil, fmt.Errorf("line %d: expected '=' after key %q", keyToken.Line, keyToken.Value)
		}
		p.pos++ // consume '='

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		obj.Entries = append(obj.Entries, ObjectEntry{
			Key:       keyToken.Value,
			KeyQuoted: keyToken.Type == TokenString,
			Value:     val,
		})
	}

	return nil, fmt.Errorf("line %d: unclosed '{'", startLine)
}

func (p *Parser) parseArray() (*ArrayNode, error) {
	startLine := p.tokens[p.pos].Line
	p.pos++ // consume '['
	arr := &ArrayNode{}

	for p.pos < len(p.tokens) {
		p.skipSeparators()
		if p.pos >= len(p.tokens) {
			return nil, fmt.Errorf("line %d: unclosed '['", startLine)
		}

		if p.tokens[p.pos].Type == TokenRBracket {
			p.pos++ // consume ']'
			return arr, nil
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr.Elements = append(arr.Elements, val)
	}

	return nil, fmt.Errorf("line %d: unclosed '['", startLine)
}
