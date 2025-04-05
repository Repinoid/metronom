package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func runner(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if fu, ok := node.(*ast.FuncDecl); ok { // fu - объявленная в пакете функция
				if file.Name.Name == "main" && fu.Name.Name == "main" {
					//	Если функция main в пакете main ищем внутри
					ast.Inspect(file, func(nodeFuncMain ast.Node) bool {
						if y, ok := nodeFuncMain.(*ast.CallExpr); ok {
							// ast.CallExpr представляет вызов функции или метода
							if s, ok := y.Fun.(*ast.SelectorExpr); ok {
								// SelectorExpr - глобальные функции типа os.Exit(1), fmt.Println. Локальные - *Ident
								if s.Sel.Name == "Exit" {
									pass.Reportf(s.Pos(), "restricted function %s prefix %s in %s package %s function", s.Sel.Name, s.X, file.Name.Name, fu.Name.Name)
								}
							}
						}
						return true
					})
				}
			}

			return true
		})
	}
	return nil, nil
}
