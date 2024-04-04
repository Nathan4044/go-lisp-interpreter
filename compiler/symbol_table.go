package compiler

// The scope which the Symbol is defined for.
type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
	FreeScope    SymbolScope = "FREE"
)

// Symbol is an instance of a defined identifier.
type Symbol struct {
	Name  string      // the defined identifier
	Scope SymbolScope // the scope in which it was defined
	Index int         // the index in memory that holds the associated value
}

// SymbolTable holds a map of identifier strings to their symbol definitions.
type SymbolTable struct {
	store       map[string]Symbol
	count       int
	outer       *SymbolTable
	FreeSymbols []Symbol
}

// Create a new empty SymbolTable.
func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{
		store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
	}

	return st
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	return st
}

// Define a symbol within the SymbolTable associated with the given identifier.
func (st *SymbolTable) Define(s string) Symbol {
	sym, ok := st.store[s]

	if ok {
		return sym
	}

	sym = Symbol{
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

		if ok && sym.Scope != BuiltinScope && sym.Scope != GlobalScope {
			free := st.defineFree(sym)

			sym = free
		}
	}

	return
}

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
