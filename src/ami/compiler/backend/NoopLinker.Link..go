package backend

func (NoopLinker) Link(opts LinkOptions) (LinkProducts, error) { return LinkProducts{}, nil }
