**Conditional Operator**

- Purpose: C-style ternary conditional expression for concise branching.
- Syntax: `cond ? thenExpr : elseExpr`
- Precedence: Lower than `||`, higher than assignment; right-associative.
- Types: Both branches should yield compatible types; when uncertain, the compiler treats the result as `any`.

Examples
- `x = flag ? 1 : 0`
- `msg := (n > 1) ? "items" : "item"`
- `a = (a == 1) ? b : 2`  // assigns `b` to `a` when `a==1`, else `2`.

Notes
- Parentheses around `cond` or branches are optional but can improve clarity.
- Nested conditionals associate to the right: `a ? b : c ? d : e` parses as `a ? b : (c ? d : e)`.
- Evaluation: In assignment statements, only the selected branch runs (no side effects from the unselected branch). Other expression positions currently lower to a `select` in the debug backend; avoid side effects in branches there until full short‑circuit lowering is applied uniformly.
