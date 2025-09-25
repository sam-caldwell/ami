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
Block = "{" { Stmt | MutBlock } "}" ;
// No Rust-like mut blocks in AMI; mutability is expressed per-assignment or via mutate(...).
Stmt = /* imperative subset evolves; assignment `=` tokens are only permitted within MutBlock in this phase */ ;

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

Mutability rules (subset)
- Default is immutable: assignments must be explicitly marked using `*` on the LHS, e.g., `*x = y`.
- `mutate(expr)` may wrap side-effectful expressions to signal mutation.

Container types (subset)
- Map: `map<K,V>` where `K` and `V` are types. In this phase:
  - Exactly two type arguments required.
  - Key type `K` cannot be a pointer, slice, map, set, or another generic type.
