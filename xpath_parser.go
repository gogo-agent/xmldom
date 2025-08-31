package xmldom

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ===========================================================================
// XPath Lexer Implementation
// ===========================================================================

// XPathTokenType represents the different types of tokens in XPath expressions
type XPathTokenType int

const (
	// Literals and identifiers
	TokenName        XPathTokenType = iota // element names, attribute names, function names
	TokenString                            // string literals like "hello"
	TokenNumber                            // numeric literals like 123 or 45.67
	TokenAxis                              // axis specifiers like "child::", "descendant::"

	// Operators
	TokenSlash        // /
	TokenDoubleSlash  // //
	TokenDot          // .
	TokenDoubleDot    // ..
	TokenAt           // @
	TokenPipe         // |
	TokenPlus         // +
	TokenMinus        // -
	TokenStar         // *
	TokenMod          // mod
	TokenDiv          // div
	TokenAnd          // and
	TokenOr           // or
	TokenEq           // =
	TokenNeq          // !=
	TokenLt           // <
	TokenLte          // <=
	TokenGt           // >
	TokenGte          // >=

	// Delimiters
	TokenLeftParen    // (
	TokenRightParen   // )
	TokenLeftBracket  // [
	TokenRightBracket // ]
	TokenComma        // ,
	TokenColon        // :
	TokenDoubleColon  // ::

	// Special
	TokenEOF    // End of input
	TokenError  // Error token
)

// XPathToken represents a single token in an XPath expression
type XPathToken struct {
	Type     XPathTokenType
	Value    string
	Position int // Position in original string
}

// String returns a string representation of the token for debugging
func (t XPathToken) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return fmt.Sprintf("ERROR(%s)", t.Value)
	default:
		return fmt.Sprintf("%s(%s)", tokenTypeName(t.Type), t.Value)
	}
}

// tokenTypeName returns a human-readable name for token types
func tokenTypeName(t XPathTokenType) string {
	switch t {
	case TokenName:
		return "NAME"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenAxis:
		return "AXIS"
	case TokenSlash:
		return "SLASH"
	case TokenDoubleSlash:
		return "DOUBLE_SLASH"
	case TokenDot:
		return "DOT"
	case TokenDoubleDot:
		return "DOUBLE_DOT"
	case TokenAt:
		return "AT"
	case TokenPipe:
		return "PIPE"
	case TokenPlus:
		return "PLUS"
	case TokenMinus:
		return "MINUS"
	case TokenStar:
		return "STAR"
	case TokenMod:
		return "MOD"
	case TokenDiv:
		return "DIV"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenEq:
		return "EQ"
	case TokenNeq:
		return "NEQ"
	case TokenLt:
		return "LT"
	case TokenLte:
		return "LTE"
	case TokenGt:
		return "GT"
	case TokenGte:
		return "GTE"
	case TokenLeftParen:
		return "LEFT_PAREN"
	case TokenRightParen:
		return "RIGHT_PAREN"
	case TokenLeftBracket:
		return "LEFT_BRACKET"
	case TokenRightBracket:
		return "RIGHT_BRACKET"
	case TokenComma:
		return "COMMA"
	case TokenColon:
		return "COLON"
	case TokenDoubleColon:
		return "DOUBLE_COLON"
	default:
		return "UNKNOWN"
	}
}

// XPathLexer tokenizes XPath expressions
type XPathLexer struct {
	input    string
	position int  // Current position in input
	start    int  // Start position of current token
	width    int  // Width of last rune read
	tokens   chan XPathToken // Channel of scanned tokens
}

// NewXPathLexer creates a new XPath lexer
func NewXPathLexer(input string) *XPathLexer {
	l := &XPathLexer{
		input:  input,
		tokens: make(chan XPathToken),
	}
	go l.run() // Start lexing in goroutine
	return l
}

// NextToken returns the next token from the lexer
func (l *XPathLexer) NextToken() XPathToken {
	return <-l.tokens
}

// run is the main lexer loop
func (l *XPathLexer) run() {
	for {
		if l.lexText() {
			break
		}
	}
	close(l.tokens)
}

// lexText scans for the next token
func (l *XPathLexer) lexText() bool {
	for {
		r := l.next()
		if r == eof {
			l.emit(TokenEOF)
			return true
		}

		switch {
		case unicode.IsSpace(r):
			l.ignore() // Skip whitespace
		case r == '/':
			return l.lexSlash()
		case r == '.':
			return l.lexDot()
		case r == '@':
			l.emit(TokenAt)
		case r == '|':
			l.emit(TokenPipe)
		case r == '+':
			l.emit(TokenPlus)
		case r == '-':
			l.emit(TokenMinus)
		case r == '*':
			l.emit(TokenStar)
		case r == '=':
			l.emit(TokenEq)
		case r == '!':
			return l.lexBangEquals()
		case r == '<':
			return l.lexLessEquals()
		case r == '>':
			return l.lexGreaterEquals()
		case r == '(':
			l.emit(TokenLeftParen)
		case r == ')':
			l.emit(TokenRightParen)
		case r == '[':
			l.emit(TokenLeftBracket)
		case r == ']':
			l.emit(TokenRightBracket)
		case r == ',':
			l.emit(TokenComma)
		case r == ':':
			return l.lexColon()
		case r == '"' || r == '\'':
			return l.lexString(r)
		case unicode.IsDigit(r):
			return l.lexNumber()
		case isNameStartChar(r):
			return l.lexName()
		default:
			l.errorf("unexpected character: %c", r)
			return true
		}
	}
}

// lexSlash handles / and // operators
func (l *XPathLexer) lexSlash() bool {
	if l.peek() == '/' {
		l.next() // consume second /
		l.emit(TokenDoubleSlash)
	} else {
		l.emit(TokenSlash)
	}
	return false
}

// lexDot handles . and .. operators
func (l *XPathLexer) lexDot() bool {
	if l.peek() == '.' {
		l.next() // consume second .
		l.emit(TokenDoubleDot)
	} else {
		l.emit(TokenDot)
	}
	return false
}

// lexBangEquals handles != operator
func (l *XPathLexer) lexBangEquals() bool {
	if l.peek() == '=' {
		l.next() // consume =
		l.emit(TokenNeq)
	} else {
		l.errorf("unexpected character after !: %c", l.peek())
		return true
	}
	return false
}

// lexLessEquals handles < and <= operators
func (l *XPathLexer) lexLessEquals() bool {
	if l.peek() == '=' {
		l.next() // consume =
		l.emit(TokenLte)
	} else {
		l.emit(TokenLt)
	}
	return false
}

// lexGreaterEquals handles > and >= operators
func (l *XPathLexer) lexGreaterEquals() bool {
	if l.peek() == '=' {
		l.next() // consume =
		l.emit(TokenGte)
	} else {
		l.emit(TokenGt)
	}
	return false
}

// lexColon handles : and :: operators
func (l *XPathLexer) lexColon() bool {
	if l.peek() == ':' {
		l.next() // consume second :
		l.emit(TokenDoubleColon)
	} else {
		l.emit(TokenColon)
	}
	return false
}

// lexString handles quoted string literals
func (l *XPathLexer) lexString(quote rune) bool {
	for {
		r := l.next()
		if r == eof {
			l.errorf("unterminated string")
			return true
		}
		if r == quote {
			// Don't include the quotes in the token value
			value := l.input[l.start+1 : l.position-1]
			l.emitValue(TokenString, value)
			return false
		}
	}
}

// lexNumber handles numeric literals
func (l *XPathLexer) lexNumber() bool {
	// Consume all digits
	for unicode.IsDigit(l.peek()) {
		l.next()
	}

	// Check for decimal point
	if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
		l.next() // consume .
		for unicode.IsDigit(l.peek()) {
			l.next()
		}
	}

	l.emit(TokenNumber)
	return false
}

// lexName handles names (identifiers, keywords, axis names)
func (l *XPathLexer) lexName() bool {
	// Consume name characters (including hyphens for axis names, but excluding colons)
	for {
		r := l.peek()
		if !isNameChar(r) && r != '-' {
			break
		}
		// Don't consume colons as part of the name - they're handled separately
		if r == ':' {
			break
		}
		l.next()
	}

	name := l.input[l.start:l.position]

	// Check if it's an axis followed by ::
	if l.peek() == ':' && l.peekNext() == ':' && isAxis(name) {
		// Emit just the axis name with ::
		l.next() // consume first :
		l.next() // consume second :
		l.emitValue(TokenAxis, name+"::") 
		return false
	}

	// Check for keywords
	switch name {
	case "mod":
		l.emit(TokenMod)
	case "div":
		l.emit(TokenDiv)
	case "and":
		l.emit(TokenAnd)
	case "or":
		l.emit(TokenOr)
	default:
		l.emit(TokenName)
	}
	return false
}

// Helper methods for lexer state

const eof = -1

// next returns the next rune in the input
func (l *XPathLexer) next() rune {
	if l.position >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.position:])
	l.width = w
	l.position += l.width
	return r
}

// peek returns the next rune without advancing position
func (l *XPathLexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// peekNext returns the rune after the next rune without advancing position
func (l *XPathLexer) peekNext() rune {
	next := l.next()
	nextNext := l.next()
	l.backup()
	l.backup()
	_ = next // avoid unused variable
	return nextNext
}

// backup moves position back by one rune
func (l *XPathLexer) backup() {
	l.position -= l.width
}

// emit passes a token back to the client
func (l *XPathLexer) emit(t XPathTokenType) {
	l.tokens <- XPathToken{
		Type:     t,
		Value:    l.input[l.start:l.position],
		Position: l.start,
	}
	l.start = l.position
}

// emitValue passes a token with custom value back to the client
func (l *XPathLexer) emitValue(t XPathTokenType, value string) {
	l.tokens <- XPathToken{
		Type:     t,
		Value:    value,
		Position: l.start,
	}
	l.start = l.position
}

// ignore discards the current token
func (l *XPathLexer) ignore() {
	l.start = l.position
}

// errorf emits an error token
func (l *XPathLexer) errorf(format string, args ...interface{}) {
	l.tokens <- XPathToken{
		Type:     TokenError,
		Value:    fmt.Sprintf(format, args...),
		Position: l.start,
	}
	l.start = l.position
}

// Character classification helpers
// Note: isNameStartChar and isNameChar are already defined in core.go

// isAxis checks if a name is a valid XPath axis
func isAxis(name string) bool {
	switch name {
	case "child", "descendant", "parent", "ancestor",
		"following-sibling", "preceding-sibling", "following", "preceding",
		"attribute", "namespace", "self", "descendant-or-self", "ancestor-or-self":
		return true
	default:
		return false
	}
}

// ===========================================================================
// XPath Parser Implementation
// ===========================================================================

// XPathParser parses XPath expressions into AST
type XPathParser struct {
	lexer    *XPathLexer
	current  XPathToken
	previous XPathToken
	peek     *XPathToken // For lookahead
}

// Parse parses an XPath expression string into an AST
func (p *XPathParser) Parse(expression string) (XPathNode, error) {
	if expression == "" {
		return nil, NewXPathException("INVALID_EXPRESSION_ERR", "Empty expression")
	}

	p.lexer = NewXPathLexer(expression)
	p.advance() // Get first token

	node, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	// Ensure we've consumed all tokens
	if p.current.Type != TokenEOF {
		return nil, p.error(fmt.Sprintf("unexpected token: %s", p.current))
	}

	return node, nil
}

// advance moves to the next token
func (p *XPathParser) advance() {
	p.previous = p.current
	if p.peek != nil {
		p.current = *p.peek
		p.peek = nil
	} else {
		p.current = p.lexer.NextToken()
	}

	// Don't panic on error tokens, let them be handled by the parser
}

// check returns true if current token is of given type
func (p *XPathParser) check(tokenType XPathTokenType) bool {
	return p.current.Type == tokenType
}

// match checks if current token matches any of the given types and advances if so
func (p *XPathParser) match(types ...XPathTokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

// consume ensures current token is of expected type and advances
func (p *XPathParser) consume(tokenType XPathTokenType, message string) error {
	if p.check(tokenType) {
		p.advance()
		return nil
	}
	return p.error(message)
}

// error creates a parsing error
func (p *XPathParser) error(message string) *XPathError {
	return NewXPathError(XPathErrorTypeSyntax, 
		fmt.Sprintf("Parse error at position %d: %s", p.current.Position, message), 
		p.current.Position)
}

// XPath Grammar Implementation (Recursive Descent)

// parseOrExpr: AndExpr ( 'or' AndExpr )*
func (p *XPathParser) parseOrExpr() (XPathNode, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenOr) {
		operator := p.previous
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &xpathBinaryOpNode{
			operator: XPathOperatorOr,
			left:     left,
			right:    right,
		}
		_ = operator // avoid unused variable warning
	}

	return left, nil
}

// parseAndExpr: EqualityExpr ( 'and' EqualityExpr )*
func (p *XPathParser) parseAndExpr() (XPathNode, error) {
	left, err := p.parseEqualityExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenAnd) {
		operator := p.previous
		right, err := p.parseEqualityExpr()
		if err != nil {
			return nil, err
		}
		left = &xpathBinaryOpNode{
			operator: XPathOperatorAnd,
			left:     left,
			right:    right,
		}
		_ = operator // avoid unused variable warning
	}

	return left, nil
}

// parseEqualityExpr: RelationalExpr ( ('=' | '!=') RelationalExpr )*
func (p *XPathParser) parseEqualityExpr() (XPathNode, error) {
	left, err := p.parseRelationalExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenEq, TokenNeq) {
		operator := p.previous
		right, err := p.parseRelationalExpr()
		if err != nil {
			return nil, err
		}

		var op XPathOperator
		switch operator.Type {
		case TokenEq:
			op = XPathOperatorEq
		case TokenNeq:
			op = XPathOperatorNeq
		}

		left = &xpathBinaryOpNode{
			operator: op,
			left:     left,
			right:    right,
		}
	}

	return left, nil
}

// parseRelationalExpr: AdditiveExpr ( ('<' | '<=' | '>' | '>=') AdditiveExpr )*
func (p *XPathParser) parseRelationalExpr() (XPathNode, error) {
	left, err := p.parseAdditiveExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenLt, TokenLte, TokenGt, TokenGte) {
		operator := p.previous
		right, err := p.parseAdditiveExpr()
		if err != nil {
			return nil, err
		}

		var op XPathOperator
		switch operator.Type {
		case TokenLt:
			op = XPathOperatorLt
		case TokenLte:
			op = XPathOperatorLte
		case TokenGt:
			op = XPathOperatorGt
		case TokenGte:
			op = XPathOperatorGte
		}

		left = &xpathBinaryOpNode{
			operator: op,
			left:     left,
			right:    right,
		}
	}

	return left, nil
}

// parseAdditiveExpr: MultiplicativeExpr ( ('+' | '-') MultiplicativeExpr )*
func (p *XPathParser) parseAdditiveExpr() (XPathNode, error) {
	left, err := p.parseMultiplicativeExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenPlus, TokenMinus) {
		operator := p.previous
		right, err := p.parseMultiplicativeExpr()
		if err != nil {
			return nil, err
		}

		var op XPathOperator
		switch operator.Type {
		case TokenPlus:
			op = XPathOperatorPlus
		case TokenMinus:
			op = XPathOperatorMinus
		}

		left = &xpathBinaryOpNode{
			operator: op,
			left:     left,
			right:    right,
		}
	}

	return left, nil
}

// parseMultiplicativeExpr: UnaryExpr ( ('*' | 'div' | 'mod') UnaryExpr )*
func (p *XPathParser) parseMultiplicativeExpr() (XPathNode, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenStar, TokenDiv, TokenMod) {
		operator := p.previous
		right, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}

		var op XPathOperator
		switch operator.Type {
		case TokenStar:
			op = XPathOperatorMultiply
		case TokenDiv:
			op = XPathOperatorDiv
		case TokenMod:
			op = XPathOperatorMod
		}

		left = &xpathBinaryOpNode{
			operator: op,
			left:     left,
			right:    right,
		}
	}

	return left, nil
}

// parseUnaryExpr: ('-')? UnionExpr
func (p *XPathParser) parseUnaryExpr() (XPathNode, error) {
	if p.match(TokenMinus) {
		expr, err := p.parseUnionExpr()
		if err != nil {
			return nil, err
		}
		return &xpathUnaryOpNode{
			operator: XPathOperatorUnaryMinus,
			operand:  expr,
		}, nil
	}

	return p.parseUnionExpr()
}

// parseUnionExpr: PathExpr ( '|' PathExpr )*
func (p *XPathParser) parseUnionExpr() (XPathNode, error) {
	left, err := p.parsePathExpr()
	if err != nil {
		return nil, err
	}

	for p.match(TokenPipe) {
		right, err := p.parsePathExpr()
		if err != nil {
			return nil, err
		}
		left = &xpathBinaryOpNode{
			operator: XPathOperatorUnion,
			left:     left,
			right:    right,
		}
	}

	return left, nil
}

// parsePathExpr: LocationPath | FilterExpr (('/' | '//') RelativeLocationPath)?
func (p *XPathParser) parsePathExpr() (XPathNode, error) {
	// Check if this is a function call first (NAME followed by '(')
	if p.check(TokenName) {
		// Look ahead to see if this is a function call
		nextToken := p.lexer.NextToken()
		if nextToken.Type == TokenLeftParen {
			// This is a function call, put the paren back and parse as primary expr
			p.peek = &nextToken
			return p.parsePrimaryExpr()
		} else {
			// Not a function call, put the token back and parse as location path
			p.peek = &nextToken
			return p.parseLocationPath()
		}
	}

	// Try to parse as location path first
	if p.check(TokenSlash) || p.check(TokenDoubleSlash) || p.check(TokenDot) || p.check(TokenDoubleDot) || 
	   p.check(TokenAt) || p.check(TokenAxis) {
		return p.parseLocationPath()
	}

	// Otherwise, parse as primary expression (number, string, literals, etc.)
	return p.parsePrimaryExpr()
}

// parseLocationPath: RelativeLocationPath | AbsoluteLocationPath
func (p *XPathParser) parseLocationPath() (XPathNode, error) {
	var steps []XPathNode

	// Check for absolute path
	if p.match(TokenSlash) {
		// Root step
		steps = append(steps, &xpathRootNode{})
		
		// Check if there are more steps
		if p.isLocationStep() {
			relativeSteps, err := p.parseRelativeLocationPath()
			if err != nil {
				return nil, err
			}
			if pathNode, ok := relativeSteps.(*xpathPathNode); ok {
				steps = append(steps, pathNode.steps...)
			} else {
				steps = append(steps, relativeSteps)
			}
		}
	} else if p.match(TokenDoubleSlash) {
		// Descendant-or-self root step
		steps = append(steps, &xpathRootNode{})
		steps = append(steps, &xpathAxisNode{
			axis:     XPathAxisDescendantOrSelf,
			nodeTest: &xpathNodeTest{name: "*", nodeType: "node()"},
		})
		
		// Parse remaining steps
		relativeSteps, err := p.parseRelativeLocationPath()
		if err != nil {
			return nil, err
		}
		if pathNode, ok := relativeSteps.(*xpathPathNode); ok {
			steps = append(steps, pathNode.steps...)
		} else {
			steps = append(steps, relativeSteps)
		}
	} else {
		// Relative path
		return p.parseRelativeLocationPath()
	}

	return &xpathPathNode{steps: steps}, nil
}

// parseRelativeLocationPath: Step ( ('/' | '//') Step )*
func (p *XPathParser) parseRelativeLocationPath() (XPathNode, error) {
	var steps []XPathNode

	// Parse first step
	step, err := p.parseStep()
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	// Parse additional steps
	for p.match(TokenSlash, TokenDoubleSlash) {
		separator := p.previous

		// If it was //, add descendant-or-self step
		if separator.Type == TokenDoubleSlash {
			steps = append(steps, &xpathAxisNode{
				axis:     XPathAxisDescendantOrSelf,
				nodeTest: &xpathNodeTest{name: "*", nodeType: "node()"},
			})
		}

		step, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	if len(steps) == 1 {
		return steps[0], nil
	}

	return &xpathPathNode{steps: steps}, nil
}

// parseStep: AxisSpecifier NodeTest Predicate* | AbbreviatedStep
func (p *XPathParser) parseStep() (XPathNode, error) {
	// Handle abbreviated steps
	if p.match(TokenDot) {
		return &xpathAxisNode{
			axis:     XPathAxisSelf,
			nodeTest: &xpathNodeTest{name: ".", nodeType: "node()"},
		}, nil
	}

	if p.match(TokenDoubleDot) {
		return &xpathAxisNode{
			axis:     XPathAxisParent,
			nodeTest: &xpathNodeTest{name: "..", nodeType: "node()"},
		}, nil
	}

	// Parse axis specifier
	axis := XPathAxisChild // default axis
	if p.check(TokenAxis) {
		axisName := strings.TrimSuffix(p.current.Value, "::")
		axis = parseAxisName(axisName)
		p.advance()
	} else if p.match(TokenAt) {
		axis = XPathAxisAttribute
	}

	// Parse node test
	nodeTest, err := p.parseNodeTest()
	if err != nil {
		return nil, err
	}

	// Adjust node type based on axis
	if axis == XPathAxisAttribute {
		// For attribute axis, the node test should match attribute nodes
		if concreteTest, ok := nodeTest.(*xpathNodeTest); ok {
			if concreteTest.nodeType == "element" {
				concreteTest.nodeType = "attribute"
			}
		}
	}

	// Parse predicates
	var predicates []XPathNode
	for p.check(TokenLeftBracket) {
		predicate, err := p.parsePredicate()
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, predicate)
	}

	return &xpathAxisNode{
		axis:       axis,
		nodeTest:   nodeTest,
		predicates: predicates,
	}, nil
}

// parseNodeTest: NameTest | NodeType '(' ')' | 'processing-instruction' '(' Literal ')'
func (p *XPathParser) parseNodeTest() (XPathNodeTest, error) {
	if p.match(TokenName) {
		name := p.previous.Value

		// Check for function-style node tests like text(), node()
		if p.match(TokenLeftParen) {
			if err := p.consume(TokenRightParen, "expected ')' after node test"); err != nil {
				return nil, err
			}
			return &xpathNodeTest{name: name, nodeType: name}, nil
		}

		// Regular name test
		return &xpathNodeTest{name: name, nodeType: "element"}, nil
	}

	if p.match(TokenStar) {
		return &xpathNodeTest{name: "*", nodeType: "element"}, nil
	}

	return nil, p.error("expected node test")
}

// parsePredicate: '[' PredicateExpr ']'
func (p *XPathParser) parsePredicate() (XPathNode, error) {
	if err := p.consume(TokenLeftBracket, "expected '['"); err != nil {
		return nil, err
	}

	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	if err := p.consume(TokenRightBracket, "expected ']'"); err != nil {
		return nil, err
	}

	return &xpathPredicateNode{expression: expr}, nil
}

// parsePrimaryExpr: VariableReference | '(' Expr ')' | Literal | Number | FunctionCall
func (p *XPathParser) parsePrimaryExpr() (XPathNode, error) {
	// Parenthesized expression
	if p.match(TokenLeftParen) {
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if err := p.consume(TokenRightParen, "expected ')' after expression"); err != nil {
			return nil, err
		}
		return expr, nil
	}

	// String literal
	if p.match(TokenString) {
		return &xpathLiteralNode{
			value: NewXPathStringValue(p.previous.Value),
		}, nil
	}

	// Number literal
	if p.match(TokenNumber) {
		numStr := p.previous.Value
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, p.error("invalid number: " + numStr)
		}
		return &xpathLiteralNode{
			value: NewXPathNumberValue(num),
		}, nil
	}

	// Function call - we should only get here if parsePathExpr determined this is a function call
	if p.check(TokenName) {
		return p.parseFunctionCall()
	}

	return nil, p.error("expected primary expression")
}

// parseFunctionCall: FunctionName '(' ( Argument ( ',' Argument )* )? ')'
func (p *XPathParser) parseFunctionCall() (XPathNode, error) {
	if !p.match(TokenName) {
		return nil, p.error("expected function name")
	}

	functionName := p.previous.Value

	if err := p.consume(TokenLeftParen, "expected '(' after function name"); err != nil {
		return nil, err
	}

	var args []XPathNode

	// Parse arguments
	if !p.check(TokenRightParen) {
		for {
			arg, err := p.parseOrExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)

			if !p.match(TokenComma) {
				break
			}
		}
	}

	if err := p.consume(TokenRightParen, "expected ')' after function arguments"); err != nil {
		return nil, err
	}

	return &xpathFunctionNode{
		name: functionName,
		args: args,
	}, nil
}

// Helper methods

// isLocationStep checks if current token can start a location step
func (p *XPathParser) isLocationStep() bool {
	return p.check(TokenDot) || p.check(TokenDoubleDot) || p.check(TokenAt) ||
		p.check(TokenAxis) || p.check(TokenName) || p.check(TokenStar)
}

// peekNext returns the type of the next token without advancing
func (p *XPathParser) peekNext() XPathTokenType {
	// This is a simplified implementation - in a real parser you'd want to buffer tokens
	// For now, we'll just check common patterns
	return TokenEOF
}

// parseAxisName converts axis name string to XPathAxis enum
func parseAxisName(name string) XPathAxis {
	switch name {
	case "child":
		return XPathAxisChild
	case "descendant":
		return XPathAxisDescendant
	case "parent":
		return XPathAxisParent
	case "ancestor":
		return XPathAxisAncestor
	case "following-sibling":
		return XPathAxisFollowingSibling
	case "preceding-sibling":
		return XPathAxisPrecedingSibling
	case "following":
		return XPathAxisFollowing
	case "preceding":
		return XPathAxisPreceding
	case "attribute":
		return XPathAxisAttribute
	case "namespace":
		return XPathAxisNamespace
	case "self":
		return XPathAxisSelf
	case "descendant-or-self":
		return XPathAxisDescendantOrSelf
	case "ancestor-or-self":
		return XPathAxisAncestorOrSelf
	default:
		return XPathAxisChild
	}
}

// ===========================================================================
// Parser-specific AST Node Types  
// ===========================================================================

// xpathNodeTest represents node tests in steps
type xpathNodeTest struct {
	name     string
	nodeType string
}

func (nt *xpathNodeTest) Matches(node Node, ctx *XPathContext) bool {
	switch nt.nodeType {
	case "node()":
		return true
	case "text()":
		return node.NodeType() == TEXT_NODE
	case "comment()":
		return node.NodeType() == COMMENT_NODE
	case "processing-instruction()":
		return node.NodeType() == PROCESSING_INSTRUCTION_NODE
	case "element":
		if node.NodeType() != ELEMENT_NODE {
			return false
		}
		if nt.name == "*" {
			return true
		}
		return string(node.NodeName()) == nt.name
	case "attribute":
		if node.NodeType() != ATTRIBUTE_NODE {
			return false
		}
		if nt.name == "*" {
			return true
		}
		return string(node.NodeName()) == nt.name
	default:
		return false
	}
}

func (nt *xpathNodeTest) Name() string {
	return nt.name
}

func (nt *xpathNodeTest) IsWildcard() bool {
	return nt.name == "*" || nt.nodeType == "node()"
}

// xpathPredicateNode represents predicate expressions
type xpathPredicateNode struct {
	expression XPathNode
}

func (n xpathPredicateNode) Type() XPathNodeType { return XPathNodeTypePredicate }

func (n xpathPredicateNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	return n.expression.Evaluate(ctx)
}
