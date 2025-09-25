# AMI Compiler Grammar (EBNF)

This document outlines a minimal EBNF for the current parser scaffold. The language is evolving; this serves as a
guide for tests and future extensions.

Notation: terminals in quotes, `/* comments */`, `// line comments` ignored by scanner.

Program = PackageDecl? { ImportDecl | TopDecl } ;

PackageDecl = "package" identifier ;

ImportDecl = "import" ( StringLit | ImportBlock | ImportAlias ) ;
ImportAlias = identifier StringLit ;
ImportBlock = "(" { (identifier? StringLit) } ")" ;

TopDecl = FuncDecl | PipelineDecl ;

FuncDecl = "func" identifier ParamList [ ResultList ] Block ;
ParamList = "(" { /* tokens until matching ')' */ } ")" ;
ResultList = "(" { /* optional tuple */ } ")" ;
Block = "{" { /* tokens until matching '}' */ } "}" ;

PipelineDecl = "pipeline" identifier "{" NodeChain "}" ;
NodeChain = NodeCall { ( "." | "->" ) NodeCall } ;
NodeCall = identifier "(" [ ArgList ] ")" ;
ArgList = /* comma-separated expressions, nesting allowed */ ;

identifier = letter { letter | digit | '_' } ;
StringLit = '"' { character | escape } '"' ;

This grammar is intentionally permissive in `ParamList`, `ResultList`, and `ArgList` to allow the parser to advance and
collect errors without requiring a complete type system.

Note on `edge.*` constructs:
- Within node argument lists, authors may specify declarative edge specifications such as `edge.FIFO(...)`, `edge.LIFO(...)`, or `edge.Pipeline(...)`.
- These are compiler-generated artifacts, not runtime function calls. The compiler recognizes them and emits high-performance queue/bridge implementations during code generation.
- See `docs/edges.md` for semantics and performance considerations.
