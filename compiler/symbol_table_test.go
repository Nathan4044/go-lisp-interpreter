package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	a := global.Define("a")

	if a != expected["a"] {
		t.Errorf("expected=%+v, got=%+v", expected["a"], a)
	}

	b := global.Define("b")

	if b != expected["b"] {
		t.Errorf("expected=%+v, got=%+v", expected["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")

	if c != expected["c"] {
		t.Errorf("expected=%+v, got=%+v", expected["c"], c)
	}

	d := firstLocal.Define("d")

	if d != expected["d"] {
		t.Errorf("expected=%+v, got=%+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")

	if e != expected["e"] {
		t.Errorf("expected=%+v, got=%+v", expected["e"], e)
	}

	f := secondLocal.Define("f")

	if f != expected["f"] {
		t.Errorf("expected=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			table: global,
			expectedSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
			},
		},
		{
			table: firstLocal,
			expectedSymbols: []Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			table: secondLocal,
			expectedSymbols: []Symbol{
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)

			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got %+v", sym.Name, sym, result)
			}
		}
	}
}

func TestDefineAndResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "b", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{
		global, firstLocal, secondLocal,
	} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)

			if !ok {
				t.Errorf("could not resolve symbol: %s", sym.Name)
			}

			if result != sym {
				t.Errorf("resolved wrong symbol: want=%+v got=%+v",
					sym, result)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table              *SymbolTable
		expectedSymbols    []Symbol
		expectedFree       []Symbol
		expectedUnresolved []string
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
			[]string{"g"},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]string{"g"},
		},
	}

	for _, tt := range tests {
		for _, expected := range tt.expectedSymbols {
			sym, ok := tt.table.Resolve(expected.Name)

			if !ok {
				t.Errorf("could not resolve symbol %s", expected.Name)
			}

			if sym != expected {
				t.Errorf("symbol resolved incorrectly: want=%+v got=%+v",
					expected, sym)
			}
		}

		if len(tt.table.FreeSymbols) != len(tt.expectedFree) {
			t.Errorf("wrong number of free symbols")
			continue
		}

		for i, expected := range tt.expectedFree {
			if tt.table.FreeSymbols[i] != expected {
				t.Errorf("incorrect free symbol: want=%+v got=%+v",
					expected, tt.table.FreeSymbols[i])
			}
		}

		for _, name := range tt.expectedUnresolved {
			_, ok := tt.table.Resolve(name)

			if ok {
				t.Errorf("resolved unresolvable symbol %s", name)
			}
		}
	}
}
