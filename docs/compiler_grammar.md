# AMI Compiler Grammar (EBNF)

This document outlines a minimal EBNF for the current parser scaffold. The language is evolving; this serves as a
guide for tests and future extensions.

Notation: terminals in quotes, `/* comments */`, `// line comments` ignored by scanner.

Program = PackageDecl? { ImportDecl | TopDecl } ;

PackageDecl = "package" identifier ':' Version ;
Version = /* SemVer */ ('v'? digit {digit} '.' digit {digit} '.' digit {digit} [ '-' {alpha|digit|'.'|'-'} ] [ '+' {alpha|digit|'.'|'-'} ]) ;

ImportDecl = "import" ( StringLit | ImportBlock | ImportAlias | ImportUnquoted ) ;
ImportAlias = identifier StringLit ;
ImportUnquoted = ModulePath [ Constraint ] ;
ImportBlock = "(" { (identifier? StringLit) | (ModulePath [ Constraint ]) } ")" ;
Constraint = ">=" Version ;
ModulePath = identifier { ( "/" identifier ) | ( "." identifier ) | ( "-" identifier ) } ;

TopDecl = FuncDecl | PipelineDecl ;

FuncDecl = "func" identifier ParamList [ ResultList ] Block ;
ParamList = "(" { /* tokens until matching ')' */ } ")" ;
ResultList = "(" { /* optional tuple */ } ")" ;
Block = "{" { Stmt } "}" ;
// Mutability is expressed per-assignment via '*' on the LHS or via mutate(...). There are no Rust-like mut blocks in AMI.
Stmt = /* imperative subset; assignments require '*' on LHS; mutate(expr) allowed */ ;

PipelineDecl = "pipeline" identifier "{" NodeChain "}" ;
NodeChain = NodeCall { ( "." | "->" ) NodeCall } ;
NodeCall = identifier "(" [ ArgList | AttrList ] ")" ;
ArgList = /* comma-separated expressions, nesting allowed */ ;
AttrList = Attr { "," Attr } ;
Attr = identifier '=' Expr ;

identifier = letter { letter | digit | '_' } ;
StringLit = '"' { character | escape } '"' ;

This grammar is intentionally permissive in `ParamList`, `ResultList`, and `ArgList` to allow the parser to advance and
collect errors without requiring a complete type system.

Note on `edge.*` constructs:
- Within node argument lists, authors may specify declarative edge specifications such as `edge.FIFO(...)`, `edge.LIFO(...)`, or `edge.Pipeline(...)`.
- These are compiler-generated artifacts, not runtime function calls. The compiler recognizes them and emits high-performance queue/bridge implementations during code generation.
- See `docs/edges.md` for semantics and performance considerations.

Inline function literals
- Minimal literal accepted in attribute positions (e.g., `worker=func(...) { ... }`) and in expressions:

  FuncLit = "func" ParamList [ ResultList | Type ] Block ;

  The parser tolerates omitted explicit result types and focuses on capturing the body for later analysis.

Generic-like calls (scaffold)
- Tolerates type arguments on call sites for scaffolding the type system:

  GenericCall = identifier '<' Type { ',' Type } '>' '(' ArgList? ')' ;

Docx-aligned usage clarifications
- Transform nodes are declared with attributes: `in=edge.FIFO|edge.LIFO(...)`, `worker=<func|factory>`, optional `minWorkers/maxWorkers`, `onError`, and output `type`.
- Collect uses `in=edge.MultiPath(inputs=[ ... ])` where the first input is the default upstream edge, and subsequent items may include `edge.Pipeline(name=Upstream, ...)`. Merge behavior is expressed via `merge=Sort(...)`.

`edge.MultiPath` (shape)
- Syntax used in `Collect` attributes:

  EdgeMultiPath = 'edge.MultiPath' '(' 'inputs' '=' '[' Input { ',' Input } ']' [ ',' 'merge' '=' MergeSpec ] ')' ;
  Input = EdgeFIFO | EdgeLIFO | EdgePipeline ;
  EdgeFIFO = 'edge.FIFO' '(' AttrList ')' ;
  EdgeLIFO = 'edge.LIFO' '(' AttrList ')' ;
  EdgePipeline = 'edge.Pipeline' '(' AttrList ')' ;

  MergeSpec is a list of merge.* attribute calls (e.g., `merge.Sort(...)`, `merge.Stable()`). See `docs/merge.md`.

Mutability rules (subset)
- Default is immutable: assignments must be explicitly marked using `*` on the LHS, e.g., `*x = y`.
- `mutate(expr)` may wrap side-effectful expressions to signal mutation.

Container types (subset)
- Map: `map<K,V>` where `K` and `V` are types. In this phase:
  - Exactly two type arguments required.
  - Key type `K` cannot be a pointer, slice, map, set, or another generic type.
