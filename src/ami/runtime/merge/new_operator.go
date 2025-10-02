package merge

func NewOperator(p Plan) *Operator {
    return &Operator{plan:p, parts: map[string]*partition{}, rr: make([]string, 0)}
}

