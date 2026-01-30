package main

import (
	"context"
)

type Service struct {
	ctx context.Context // Doctor should warn about this
	DB  string
}

func NewService(ctx context.Context) *Service {
	return &Service{ctx: ctx, DB: "postgres"}
}
