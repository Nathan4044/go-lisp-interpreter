package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store map[string]Symbol
	count int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

func (st *SymbolTable) Define(s string) Symbol {
	sym := Symbol{
		Name:  s,
		Scope: GlobalScope,
		Index: st.count,
	}

	st.store[s] = sym
	st.count++

	return sym
}

func (st *SymbolTable) Resolve(s string) (Symbol, bool) {
	sym, ok := st.store[s]

	return sym, ok
}
