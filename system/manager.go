package system

import (
	"context"
	"fmt"
	"time"

	"github.com/cybozu-go/well"
	"github.com/rs/zerolog/log"
)

type manager struct {
	serviceList []*Service

	outPipe chan string
	errPipe chan string

	timeout time.Duration
}

func NewServiceManager(services []*Service, timeout time.Duration) *manager {
	m := &manager{
		serviceList: services,
		timeout:     timeout,
	}

	bufSize := len(m.serviceList)
	log.Info().Msgf("[M] buffer size: %d", bufSize)

	m.outPipe = make(chan string, bufSize)
	m.errPipe = make(chan string, bufSize)

	return m
}

func (m *manager) Run() error {
	log.Info().Msg("[M] starting services")

	for i := range m.serviceList {
		service := m.serviceList[i]

		well.Go(func(ctx context.Context) error {
			// 10 seconds to start
			return service.Run(ctx, m.outPipe, m.errPipe, 10*time.Second)
		})
	}

	stopStd := make(chan struct{})
	defer close(stopStd)

	go m.proxyOutput(stopStd)

	well.Stop()
	err := well.Wait()
	//stopStd <- struct{}{}

	return err
}

func (m manager) proxyOutput(stopStd chan struct{}) {
	for {
		select {
		case out := <-m.outPipe:
			fmt.Println(out)
		case err := <-m.errPipe:
			fmt.Println(err)
		case <-stopStd:
			log.Info().Msg("[M] stopped services output")
			return
		}
	}
}
