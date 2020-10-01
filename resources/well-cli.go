package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"time"
)

func main() {
	flag.Parse()
	err := well.LogConfig{}.Apply()
	if err != nil {
		log.ErrorExit(err)
	}

	well.Go(func(ctx context.Context) error {
		err := doSomething(ctx, 10)

		if err != nil {
			// non-nil error will be passed to Cancel
			// by the framework.
			return err
		}

		select {
		case <-ctx.Done():
			fmt.Println("Exiting")
			return ctx.Err()
		default:
			// on success, nil should be returned.
			return doOtherthing()
		}
	})

	well.Go(func(ctx context.Context) error {
		err := doSomething(ctx, 20)

		if err != nil {
			// non-nil error will be passed to Cancel
			// by the framework.
			return err
		}

		select {
		case <-ctx.Done():
			fmt.Println("Exiting")
			return ctx.Err()
		default:
			// on success, nil should be returned.
			return doOtherthing()
		}
	})

	// Stop declares no Go calls will be made from this point.
	// Calling Stop is optional if Cancel is guaranteed to be called
	// at some point.
	well.Stop()

	fmt.Println("reached well.Wait()")

	// Wait waits for all goroutines started by Go to complete,
	// or one of such goroutine returns non-nil error.
	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}

func doSomething(ctx context.Context, limit int) error {
	for i := limit; i > 0; i-- {
		if err := log.Info("Tick!", map[string]interface{}{"count": i}); err != nil {
			return err
		}
		time.Sleep(time.Second)

		select {
		case <-ctx.Done():
			log.Info("Sick.. exiting!", map[string]interface{}{})
			return nil
		default:
		}
	}

	return nil
}

func doOtherthing() error {
	time.Sleep(time.Second * 3)

	if err := log.Info("KABOOOOOM!", map[string]interface{}{}); err != nil {
		return err
	}

	return nil
}
