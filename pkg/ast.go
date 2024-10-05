package pkg

import (
	// "go/ast"
	"go/token"
)

// Node represents a node in the AST
type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// File represents an entire Go source file
type File struct {
	Package    token.Pos
	Name       *Ident
	Decls      []Decl
	Scope      *Scope
	Imports    []*ImportSpec
	Unresolved []*Ident
	Comments   []*CommentGroup
}

// Ident represents an identifier
type Ident struct {
	NamePos token.Pos
	Name    string
	Obj     *Object
}

// Expr represents an expression
type Expr interface {
	Node
	exprNode()
}

// Stmt represents a statement
type Stmt interface {
	Node
	stmtNode()
}

// Decl represents a declaration
type Decl interface {
	Node
	declNode()
}

// Scope represents a lexical scope
type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

// Object represents a declared constant, type, variable, or function
type Object struct {
	Kind ObjKind
	Name string
	Decl interface{}
	Data interface{}
	Type interface{}
}

// ObjKind represents the kind of object (const, type, var, func)
type ObjKind int

const (
	Bad ObjKind = iota
	Pkg
	Con
	Typ
	Var
	Fun
)

// ImportSpec represents an import declaration
type ImportSpec struct {
	Doc     *CommentGroup
	Name    *Ident
	Path    *BasicLit
	Comment *CommentGroup
	EndPos  token.Pos
}

// CommentGroup represents a sequence of comments
type CommentGroup struct {
	List []*Comment
}

// Comment represents a single comment
type Comment struct {
	Slash token.Pos
	Text  string
}

// BasicLit represents a literal of basic type
type BasicLit struct {
	ValuePos token.Pos
	Kind     token.Token
	Value    string
}
