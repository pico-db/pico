package api

import "context"

// Database clients interact with the database
// through a RESTful API interface.
//
// This starts a new HTTP server service that accepts and perform actions on the database
func New() *Service {
	return &Service{}
}

type Service struct {
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}
