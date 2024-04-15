package evaluator

import (
	"github.com/SVendittelli/monkey/ast"
	"github.com/SVendittelli/monkey/object"
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

// Recursively walk down the `program` AST and find calls to macros, quote their
// arguments, extend their environments with the args, evaluate the result and
// finally modify the AST to replace the macro call with the quoted AST node.
func ExpandMacros(program ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(program, func(node ast.Node) ast.Node {
		callExpression, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(callExpression, env)
		if !ok {
			return node
		}

		args := quoteArgs(callExpression)
		evalEnv := extendMacroEnv(macro, args)

		evaluated := Eval(macro.Body, evalEnv)

		quote, ok := evaluated.(*object.Quote)
		if !ok {
			panic("we only support returning AST-nodes from macros")
		}

		return quote.Node
	})
}

// Check if a call expression is a macro.
func isMacroCall(
	exp *ast.CallExpression,
	env *object.Environment,
) (*object.Macro, bool) {
	identifier, ok := exp.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(identifier.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

// Take the arguments from a call and turn them into `Quotes`.
func quoteArgs(exp *ast.CallExpression) []*object.Quote {
	args := []*object.Quote{}

	for _, a := range exp.Arguments {
		args = append(args, &object.Quote{Node: a})
	}

	return args
}

// Extend the macro's environment with the arguments of the call bound to the
// parameter names of the macro literal.
func extendMacroEnv(
	macro *object.Macro,
	args []*object.Quote,
) *object.Environment {
	extended := object.NewEnclosedEnvironment(macro.Env)

	for paramIdx, param := range macro.Parameters {
		extended.Set(param.Value, args[paramIdx])
	}

	return extended
}
