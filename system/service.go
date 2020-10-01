package system

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cybozu-go/well"
	"github.com/rs/zerolog/log"
)

type Service struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`

	running *process

	isStarted bool
	isStopped bool
}

func (s *Service) Run(ctx context.Context, out, err chan<- string, ttl time.Duration) error {
	s.running = NewProcess(s.Name, s.Command, s.Args)
	if err := s.running.Start(ttl); err != nil {
		return err
	}

	// listen for STD
	s.scanOutput("%s", &s.running.Out, out)
	s.scanOutput("%s", &s.running.Err, err)

	// process finished on its own
	var finished = make(chan error)
	go func() {
		finished <- s.running.Wait()
		close(finished)
	}()

	// process finished on its own
	select {
	case <-ctx.Done():
		log.Info().Msgf("[S][%s] received stop signal", s.Name)
		return s.running.Stop(finished, ttl)
	case err := <-finished:
		log.Info().Err(err).Msgf("[S][%s] finished, sending cancel to global ctx", s.Name)
		well.Cancel(err)
		return err
	}
}

func (s Service) IsRunning() bool {
	return s.running != nil && !s.running.Finished()
}

func (s Service) scanOutput(format string, src *io.ReadCloser, dst chan<- string) {
	stdScanner := bufio.NewScanner(*src)

	go func() {
		for s.IsRunning() && stdScanner.Scan() {
			logs := stdScanner.Text()
			dst <- fmt.Sprintf("[%s] "+format, s.Name, logs)
		}
	}()
}
