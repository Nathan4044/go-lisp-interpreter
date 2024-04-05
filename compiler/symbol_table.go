package compiler

// The scope which the Symbol is defined for.
type SymbolScope string

const (
	// For Symbols defined at the highest level of the program, can be accessed
	// directly at any depth so long as they aren't shadowed by a variable in an
	// enclosed scope.
	GlobalScope SymbolScope = "GLOBAL"
	// For Symbols defined within a lambda, accessed only within that lambda's
	// scope, passed to inner scopes as a free variable.
	LocalScope SymbolScope = "LOCAL"
	// For Symbols associated with builtin functions. Accessed at any scope
	// unless shadowed by another variable in an enclosed scope.
	BuiltinScope SymbolScope = "BUILTIN"
	// For Symbols passed into a closure from an enclosing scope below the
	// global scope. Enabled closures to capture variables needed from their
	// defining scope.
	FreeScope SymbolScope = "FREE"
	// Special scope reserved for the name of the function that the current
	// scope is associated with. Used to enable recursive closure calls by
	// resolving to itself and emitting a specific instruction.
	FunctionScope SymbolScope = "FUNCTION"
)

// Symbol is an instance of a defined identifier.
type Symbol struct {
	Name  string      // the defined identifier
	Scope SymbolScope // the scope in which it was defined
	Index int         // the index in memory that holds the associated value
}

// SymbolTable holds a map of identifier strings to their symbol definitions.
type SymbolTable struct {
	store       map[string]Symbol // maps a string to its associated Symbol
	count       int               // the number of Symbols in the store
	outer       *SymbolTable      // address of enclosing SymbolTable
	FreeSymbols []Symbol          // tracks variables required from enclosing scope
}

// Create a new empty SymbolTable.
func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{
		store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
	}

	return st
}

// Create a new empty SymbolTable with an associated outer scope.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	return st
}

// Define a symbol within the SymbolTable associated with the given identifier.
func (st *SymbolTable) Define(s string) Symbol {
	sym := Symbol{
		Name:  s,
		Index: st.count,
	}

	if st.outer == nil {
		sym.Scope = GlobalScope
	} else {
		sym.Scope = LocalScope
	}

	st.store[s] = sym
	st.count++

	return sym
}

// Define a symbol within the SymbolTable associated with the provided builtin
// function name.
func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	sym := Symbol{
		Name:  name,
		Index: index,
		Scope: BuiltinScope,
	}

	st.store[name] = sym

	return sym
}

// Retrieve the Symbol associated with the given identifier.
func (st *SymbolTable) Resolve(s string) (sym Symbol, ok bool) {
	sym, ok = st.store[s]

	if !ok && st.outer != nil {
		sym, ok = st.outer.Resolve(s)

		// If Symbol resolves to scope enclosing the current one, and that
		// Symbol is not global or builtin, define it as a free Symbol in the
		// current scope.
		if ok && sym.Scope != BuiltinScope && sym.Scope != GlobalScope {
			free := st.defineFree(sym)

			sym = free
		}
	}

	return
}

// Define the provided symbol in the current scope as a free Symbol, and keep
// track of the Symbols defined this way from the enclosing scope's SymbolTable.
func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	free := Symbol{
		Name:  original.Name,
		Scope: FreeScope,
		Index: len(st.FreeSymbols) - 1,
	}

	st.store[original.Name] = free

	return free
}

// Define the name of the function associated with the current scope. This is
// used to identify when a function calls itself recursively.
func (st *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Scope: FunctionScope,
		Index: 0,
	}

	st.store[name] = symbol

	return symbol
}
