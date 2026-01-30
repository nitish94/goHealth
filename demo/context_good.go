package main

import (
	"context"
)

type ServiceGood struct {
	// Good: No context in struct
	name string
}

func (s *ServiceGood) DoWork(ctx context.Context) {
	// Good: Pass context as parameter
	_ = ctx
}