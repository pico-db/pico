package server

import (
	"context"
	"log"
	"time"

	"github.com/panjf2000/ants/v2"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

type server struct {
	tp *ants.Pool
}

func NewServer() Server {
	return &server{}
}

func (s *server) Start() error {
	log.Printf("initializing thread pool")
	tp, err := ants.NewPool(
		100,
		ants.WithLogger(log.Default()),
		ants.WithPanicHandler(func(i interface{}) {
			log.Fatalf("panic caught inside thread: %v", i)
		}),
	)
	if err != nil {
		log.Printf("unable to initialize thread pool: %s", err.Error())
		return err
	}
	s.tp = tp
	return nil
}

func (s *server) Stop(ctx context.Context) error {
	log.Println("releasing thread pool")
	dl, ok := ctx.Deadline()
	if !ok {
		s.tp.Release()
	} else {
		t := time.Until(dl)
		err := s.tp.ReleaseTimeout(t)
		if err != nil {
			log.Fatalf("unable to release thread pool: %s", err.Error())
			if err == ants.ErrTimeout {
				ctx.Done()
			}
			return err
		}
	}
	return nil
}
