package evaluator

import (
	"github.com/SVendittelli/monkey-interpreter/ast"
	"github.com/SVendittelli/monkey-interpreter/object"
)

// DefineMacros does two thinds: finding macro definitions in and removing them
// from the AST. Of note is the that we only allow top-level macro defintions.
// We don't walk down the `Statements` and check the child nodes for more.
// TODO allow nested macro definitions
func DefineMacros(program *ast.Program, env *object.Environment) {
	definitions := []int{}

	for i, statement := range program.Statements {
		if isMacroDefinition(statement) {
			addMacro(statement, env)
			definitions = append(definitions, i)
		}
	}

	for i := len(definitions) - 1; i >= 0; i = i - 1 {
		definitionIndex := definitions[i]
		program.Statements = append(
			program.Statements[:definitionIndex],
			program.Statements[definitionIndex+1:]...,
		)
	}
}

// Check if a statement is a macro definition i.e. an `*ast.LetStatement` that
// binds a `MacroLiteral` to a name.
// Consider:
//
//	let myMacro = macro(x) { x };
//	let anotherNameForMyMacro = myMacro;
//
// isMacroDefinition won’t recognize the second let statement as a valid macro
// defintion.
func isMacroDefinition(node ast.Statement) bool {
	letStatement, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = letStatement.Value.(*ast.MacroLiteral)
	return ok
}

// Bind a macro to the environment. Note we assume `stmt` is a valid macro
// literal as we use `isMacroDefinition` to check before the `addMacro`
// invocation in `DefineMacros`.
func addMacro(stmt ast.Statement, env *object.Environment) {
	letStatement, _ := stmt.(*ast.LetStatement)
	macroLiteral, _ := letStatement.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLiteral.Parameters,
		Env:        env,
		Body:       macroLiteral.Body,
	}

	env.Set(letStatement.Name.Value, macro)
}
