# AMI Grammar (EBNF, Scaffold)

This document captures the current implemented subset of the AMI language grammar used by the frontend (token → scanner
→ parser → AST). It is intentionally conservative and will evolve with the spec.

Notation: EBNF with `?` for optional and `*` for repetition.

- file = packageDecl, importDecl*, topDecl*.
- packageDecl = "package", ident.
- importDecl = "import", ( ident | string ), constraint? \n
  | "import", "(", ( ( ident | string ), constraint? )* ")".
- constraint = ">=", version.
- version = "v", number, ".", number, ".", number, ("-", (ident | number), (".", (ident | number))* )?.

- topDecl = funcDecl | pipelineDecl | errorBlock.

- funcDecl = "func", ident, typeParams?, paramList, resultList?, block.
- typeParams = "<", typeParam, (",", typeParam)*, ">".
- typeParam = ident, (ident)? . // optional constraint ident, e.g., any
- paramList = "(", (param (",", param)*)? ")".
- param = ident, (ident)? . // name and optional type ident
- resultList = "(", ident, (",", ident)* ")".
- block = "{", stmt*, "}".

- stmt = varDecl ";"?
       | assign ";"?
       | deferStmt ";"?
       | returnStmt ";"?
       | exprStmt ";"?
       .
- varDecl = "var", ident, (ident)?, ("=", expr)? .
- assign = ident, "=", expr .
- deferStmt = "defer", callExpr .
- returnStmt = "return", (expr (",", expr)*)? .
- exprStmt = expr .

- expr = conditionalExpr | callExpr | basicLit | ident | containerLit | binaryExpr .
- conditionalExpr = expr, "?", expr, ":", expr . // right-associative, lowest precedence
- callExpr = dottedIdent, "(", (expr (",", expr)*)? ")".
- dottedIdent = ident, (".", ident)* .
- basicLit = number | string .
- containerLit = sliceLit | setLit | mapLit .
- sliceLit = "slice", "<", ident, ">", "{", (expr (",", expr)*)? "}" .
- setLit = "set", "<", ident, ">", "{", (expr (",", expr)*)? "}" .
- mapLit = "map", "<", ident, ",", ident, ">", "{", (expr, ":", expr) (",", expr, ":", expr)* )? "}" .
- binaryExpr = expr, op, expr .
- op ∈ { "+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "&&", "||" } with standard precedence.

- pipelineDecl = "pipeline", ident, "(", ")", "{", ( errorBlock | step | edge )* "}" .
- errorBlock = "error", block .
- step = ident, "(", (expr (",", expr)*)? ")", attrs? .
- attrs = dottedIdent, ( "(", (expr (",", expr)*)? ")" )?, ( ",", dottedIdent, ( "(", (expr (",", expr)*)? ")" )? )* .
- edge = ident, "->", ident ";"? .

- ident, number, string are scanner‑provided tokens with the usual definitions.
