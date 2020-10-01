package system

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const ProcessKillSig = syscall.SIGKILL
const ProcessStopSig = syscall.SIGINT

type process struct {
	name string
	cmd  *exec.Cmd

	Created time.Time
	Stopped time.Time
	Out     io.ReadCloser
	Err     io.ReadCloser
}

func NewProcess(name, target string, params []string) *process {
	process := &process{name: name}
	process.cmd = exec.Command(target, params...)

	var err error
	process.Out, err = process.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	process.Err, err = process.cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	return process
}

func (p process) GetPid() int {
	if p.Finished() {
		log.Error().Msgf("[P][%s] RuntimeError: process is not running", p.name)
		return 0
	}

	return p.cmd.Process.Pid
}

func (p *process) Start(ttl time.Duration) error {
	log.Info().Msgf("[P][%s] starting...", p.name)

	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.Created = time.Now()

	delayed := time.After(ttl)
	for p.cmd.Process == nil {
		select {
		case <-delayed:
			log.Error().Msgf("[P][%s] process starting.. reached timeout after %s", p.name, ttl.String())
		default:
			time.Sleep(time.Second * 1)
		}
	}

	log.Info().Msgf("[P][%s] PID: %d", p.name, p.GetPid())

	return nil
}

func (p process) Stop(finished chan error, ttl time.Duration) error {
	if p.Finished() {
		return nil
	}

	log.Info().Msgf("[P][%s] waiting %s for process to finish", p.name, ttl.String())

	if err := p.cmd.Process.Signal(ProcessStopSig); err != nil {
		log.Warn().Err(err).Msgf("[P][%s] %s", p.name, err.Error())
	}

	// waiting ttl duration to end process, or kill it
	select {
	case err := <-finished:
		log.Info().Err(err).Msgf("[P][%s] stopped", p.name)
		return err
	case <-time.After(ttl):
		return p.kill(ProcessKillSig)
	}
}

func (p process) Wait() error {
	return p.cmd.Wait()
}

func (p process) Finished() bool {
	return p.cmd == nil || (p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited())
}

func (p *process) kill(sig syscall.Signal) error {
	if p.Finished() {
		log.Info().Msgf("[P][%s] nothing to kill", p.name)
		return nil
	}

	log.Info().Msgf("[P][%s] killing..", p.name)

	var killErr string
	if err := p.cmd.Process.Signal(sig); err != nil {
		killErr = fmt.Sprintf("[P][%s] failed to kill PID [%d]: %s", p.name, p.GetPid(), err)
	} else {
		killErr = fmt.Sprintf("[P][%s] killed by timeout", p.name)
		p.Stopped = time.Now()
	}

	return errors.New(killErr)
}
