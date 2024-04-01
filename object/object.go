// Define the Object interface, and the object types that
// utilise it.
package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"lisp/ast"
	"lisp/code"
	"strings"
)

const (
	INTEGER_OBJ           = "INTEGER"
	FLOAT_OBJ             = "FLOAT"
	STRING_OBJ            = "STRING"
	FUNCTION_OBJ          = "FUNCTION"
	LIST_OBJ              = "LIST"
	DICT_OBJ              = "DICT"
	BOOLEAN_OBJ           = "BOOL"
	LAMBDA_OBJ            = "LAMBDA"
	NULL_OBJ              = "NULL"
	ERROR_OBJ             = "ERROR"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
)

// The Function type is the definition of a builtin function.
type Function func(env *Environment, args ...Object) Object

type ObjectType string

// Object defines a basic interface that specific types will implement.
type Object interface {
	Type() ObjectType // Return one of the predefined object type strings.
	Inspect() string  // Return a string representation of the Object.
}

// Integer is an Object that holds an integer value.
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

// Return the integer value as a string.
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

// Float is an Object that holds a float value.
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType {
	return FLOAT_OBJ
}

// Return the float value as a string.
func (f *Float) Inspect() string {
	return fmt.Sprintf("%f", f.Value)
}

// String is an Object that holds a string value.
type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}

// Return the String Object's inner value.
func (s *String) Inspect() string {
	return s.Value
}

// The List Object wraps an Object slice.
type List struct {
	Values []Object
}

func (l *List) Type() ObjectType {
	return LIST_OBJ
}

// Returns a string representation of the items contained
// within the List.
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

// FunctionObject wraps a builtin function, along with the associated
// name.
//
// This provides the ability to associate builtin functions as
// a lisp Object.
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

// Boolean is an Object that holds a bool value.
type BooleanObject struct {
	Value bool
}

func (b *BooleanObject) Type() ObjectType {
	return BOOLEAN_OBJ
}

// Return a string representation of the bool value.
func (b *BooleanObject) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// An Object type that represents no return value.
type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

func (n *Null) Inspect() string {
	return "null"
}

// The Lambda type stores user defined lambda functions.
type LambdaObject struct {
	Args []string         // The Arguments passed to the function.
	Env  *Environment     // The Environment in which the lambda was defined, allowing for closures.
	Body []ast.Expression // The SExpressions defined by the user, which are evaluated when the lambda is called.
}

func (l *LambdaObject) Type() ObjectType {
	return LAMBDA_OBJ
}

// Return a string representation of the defined lambda.
// Essentially recreating the source code that defined the lambda.
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

// An object type used to store returned errors for when evaluation
// goes wrong.
type ErrorObject struct {
	Error string
}

func (e *ErrorObject) Type() ObjectType {
	return ERROR_OBJ
}

func (e *ErrorObject) Inspect() string {
	return fmt.Sprintf("ERROR: %s", e.Error)
}

// The HashKey Object stores a hashed value of a Hashable Object so
// that it can be used as a key in a Dictionary.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Used to establish if an Object can be used as a key in a Dictionary.
type Hashable interface {
	HashKey() HashKey
}

// Create a HashKey object that represents an Integer.
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: INTEGER_OBJ, Value: uint64(i.Value)}
}

// Create a HashKey object that represents a Boolean.
func (b *BooleanObject) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: STRING_OBJ, Value: value}
}

// Create a HashKey object that represents a String
// by converting the string Value to a uint64.
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: STRING_OBJ, Value: h.Sum64()}
}

// The DictPair type represents both the key and value
// to be stored in a Dictionary.
type DictPair struct {
	Key   Object
	Value Object
}

// The Dictionary type wraps a map of HashKey to DictPair
// where the HashKey is created from the Key from the DictPair.
type Dictionary struct {
	Values map[HashKey]DictPair
}

func (d *Dictionary) Type() ObjectType {
	return DICT_OBJ
}

// Create a string representation of a Dictionary by
// concatenating the string representations of its DictPairs.
func (d *Dictionary) Inspect() string {
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

// CompiledLambda is an object that holds compiled instructions.
type CompiledLambda struct {
	Instructions   code.Instructions
	LocalsCount    int
	ParameterCount int
}

func (cl *CompiledLambda) Type() ObjectType {
	return COMPILED_FUNCTION_OBJ
}

// Provide a string representation of the compiled lambda, providing its
// pointer as a means to distinguish between inspected lambdas.
func (cl *CompiledLambda) Inspect() string {
	return fmt.Sprintf("CompiledLambda[%p]", cl)
}
