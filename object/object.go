package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"lisp/ast"
	"strings"
)

const (
	INTEGER_OBJ  = "INTEGER"
    FLOAT_OBJ    = "FLOAT"
	STRING_OBJ   = "STRING"
	FUNCTION_OBJ = "FUNCTION"
	LIST_OBJ     = "LIST"
	DICT_OBJ     = "DICT"
	BOOLEAN_OBJ  = "BOOL"
	LAMBDA_OBJ   = "LAMBDA"
	NULL_OBJ     = "NULL"
	ERROR_OBJ    = "ERROR"
)

type Function func(env *Environment, args ...ast.Expression) Object

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType {
	return FLOAT_OBJ
}

func (f *Float) Inspect() string {
	return fmt.Sprintf("%f", f.Value)
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}

func (s *String) Inspect() string {
	return s.Value
}

type List struct {
	Values []Object
}

func (l *List) Type() ObjectType {
	return LIST_OBJ
}

func (l *List) Inspect() string {
	var result bytes.Buffer

	result.WriteString("(")

	items := []string{}

	for _, val := range l.Values {
		items = append(items, val.Inspect())
	}

	result.WriteString(strings.Join(items, " "))
	result.WriteString(")")

	return result.String()
}

type FunctionObject struct {
	Fn   Function
	Name string
}

func (f *FunctionObject) Type() ObjectType {
	return FUNCTION_OBJ
}

func (f *FunctionObject) Inspect() string {
	return f.Name
}

type BooleanObject struct {
	Value bool
}

func (b *BooleanObject) Type() ObjectType {
	return BOOLEAN_OBJ
}

func (b *BooleanObject) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

func (n *Null) Inspect() string {
	return "null"
}

type LambdaObject struct {
	Args []string
	Env  *Environment
	Body []ast.Expression
}

func (l *LambdaObject) Type() ObjectType {
	return LAMBDA_OBJ
}

func (l *LambdaObject) Inspect() string {
	var result bytes.Buffer

	result.WriteString("(lambda (")
	result.WriteString(
		strings.Join(l.Args, " "),
	)
	result.WriteString(") ")

	body := []string{}

	for _, exp := range l.Body {
		body = append(body, exp.String())
	}
	result.WriteString(
		strings.Join(body, " "),
	)
	result.WriteString(")")

	return result.String()
}

type ErrorObject struct {
	Error string
}

func (e *ErrorObject) Type() ObjectType {
	return ERROR_OBJ
}

func (e *ErrorObject) Inspect() string {
	return fmt.Sprintf("ERROR: %s", e.Error)
}

type HashKey struct {
    Type ObjectType
    Value uint64
}

type Hashable interface {
    HashKey() HashKey
}

func (i *Integer) HashKey() HashKey {
    return HashKey{ Type: INTEGER_OBJ, Value: uint64(i.Value) }
}

func (b *BooleanObject) HashKey() HashKey {
    var value uint64

    if b.Value {
        value = 1
    } else {
        value = 0
    }

    return HashKey{ Type: STRING_OBJ, Value: value }
}

func (s *String) HashKey() HashKey {
    h := fnv.New64a()
    h.Write([]byte(s.Value))

    return HashKey{ Type: STRING_OBJ, Value: h.Sum64() }
}

type DictPair struct {
    Key Object
    Value Object
}

type Dict struct {
    Values map[HashKey]DictPair
}

func (d *Dict) Type() ObjectType {
    return DICT_OBJ
}

func (d *Dict) Inspect() string {
    var result bytes.Buffer

    result.WriteString("{")

    items := []string{}

    for _, pair := range d.Values {
        items = append(
            items, 
            fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()),
        )
    }

    result.WriteString(strings.Join(items, ", "))
    result.WriteString("}")

    return result.String()
}
