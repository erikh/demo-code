// Package submitter was extracted from some processing code I wrote for a CI.
// It's used to execute processes that ingest stdin and produce stdout to a
// JSON protocol. There is an upper bound on running processes, and slow
// processes are logged. Canceling the context will kill all processes (see
// tests). This is basically managing a bunch of pipe(2) calls to processes.
package submitter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"time"
)

// list of processes. They are iterated in the scheduler routine and appended to in add.
type processList []*process

// a single process, with its creation time alongside it for easy tracking.
// This struct is created in all situations.
type process struct {
	cmd     *exec.Cmd
	created time.Time
}

// Handler is the supervising handle for the submitter
type Handler struct {
	config *Config
	ctx    context.Context

	waitMutex    sync.Mutex // mutex to control whether we can add new processes, affects wait and add calls.
	processMutex sync.Mutex // mutex for anytime the process list is checked or modified
	processes    processList

	// this submits its response back to another processor, but can be rearranged
	// by tests and such.
	readerFunc func(io.ReadCloser) error
}

// Config is a simple configuration struct for programming a few parameters.
type Config struct {
	MaxProcessors uint
	DebugMode     bool // if set, produces stderr to the main process's stderr.
}

// NewHandler returns a new handler.
func NewHandler(ctx context.Context, cfg *Config, reader func(io.ReadCloser) error) (*Handler, error) {
	h := &Handler{
		ctx:        ctx,
		config:     cfg,
		processes:  processList{},
		readerFunc: reader,
	}

	if h.readerFunc == nil {
		h.readerFunc = func(stdout io.ReadCloser) error {
			// all the default does here is print to stdout. It's not used in the
			// tests because it's too noisy.
			io.Copy(os.Stdout, stdout)
			return nil
		}
	}

	go h.supervisor()

	return h, nil
}

func (h *Handler) supervisor() {
	for {
		select {
		case <-h.ctx.Done():
			h.processMutex.Lock()
			defer h.processMutex.Unlock()
			for _, proc := range h.processes {
				proc.cmd.Process.Kill()
			}
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}

		h.processMutex.Lock()
		processes := processList{}

		for _, proc := range h.processes {
			if proc.cmd.ProcessState == nil {
				processes = append(processes, proc)
			}

			if proc.created.Before(time.Now().Add(-time.Minute)) {
				fmt.Printf("slow processor after %v: pid %d", time.Since(proc.created), proc.cmd.Process.Pid)
			}
		}

		h.processes = processes
		h.processMutex.Unlock()
	}
}

// waitForQuiet just waits for things to settle before trying to insert a
// process. it's not honestly very useful when the process list is being
// hammered, but it's used in the tests so there you go.
func (h *Handler) waitForQuiet(d time.Duration) error {
	h.waitMutex.Lock()
	defer h.waitMutex.Unlock()

	after := time.After(d)

	for {
		select {
		case <-after:
			return errors.New("could not settle")
		case <-h.ctx.Done():
			return errors.New("context closed")
		default:
		}

		h.processMutex.Lock()
		if len(h.processes) == 0 {
			h.processMutex.Unlock()
			return nil
		}
		h.processMutex.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

}

func (h *Handler) launchCommand(args []string, sr map[string]interface{}) error {
	h.waitMutex.Lock()
	defer h.waitMutex.Unlock()

	// we only unlock once unless we try the retry loop, at which point we do it deliberately before returning
	defer h.processMutex.Unlock()

	// hammer the lock until we get our process in. this is chaotic but ultimately
	// OK as processors are intended to be short-lived things.
retry:
	h.processMutex.Lock()
	// if we're outside the maxprocessors boundary, just retry again until we get in.
	if uint(len(h.processes)) < h.config.MaxProcessors || h.config.MaxProcessors == 0 {
		var cmd *exec.Cmd

		if len(args) == 1 {
			cmd = exec.CommandContext(h.ctx, args[0])
		} else {
			cmd = exec.CommandContext(h.ctx, args[0], args[1:]...)
		}

		// outpipe returns the json from the process.
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		if h.config.DebugMode {
			// stderr is only printed in debug mode; it is assumed it'll be trapped
			// by systemd/docker/etc for debugging things.
			go io.Copy(os.Stderr, stderr)
		} else {
			go io.Copy(ioutil.Discard, stderr)
		}

		// in pipe accepts our inital JSON payload
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		go h.readerFunc(stdout)

		go func() {
			if err := json.NewEncoder(stdin).Encode(sr); err != nil {
				if h.config.DebugMode {
					fmt.Printf("Could not encode value during submission: %+v: %v\n", sr, err)
				}
			}

			stdin.Close()
		}()

		p := &process{
			cmd:     cmd,
			created: time.Now(),
		}

		// fire and forget the wait status, because we'll poll it later.
		if err := cmd.Start(); err != nil {
			return err
		}

		// it's typically important to call wait no matter what, I'm not sure if
		// golang cares or not to be honest. So we toss it off in a goroutine to
		// exit on its own.
		go cmd.Wait()

		// stow the process struct and return
		h.processes = append(h.processes, p)
	} else {
		h.processMutex.Unlock()
		goto retry
	}

	return nil
}
