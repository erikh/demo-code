package submitter

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	fmt.Println("installing stub binary for testing submissions")
	out, err := exec.Command("go", "install", "-v", "./testdata/testbin").CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestSupervisor(t *testing.T) {
	cfg := &Config{
		MaxProcessors: 10,
		DebugMode:     true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	h, err := NewHandler(ctx, cfg, func(io.ReadCloser) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	runConfig := map[string]interface{}{"configuration": "is cool"}

	for i := 0; i < 200; i++ {
		if err := h.launchCommand([]string{"testbin"}, runConfig); err != nil {
			t.Fatal(err)
		}
	}

	if err := h.waitForQuiet(time.Second); err != nil {
		t.Fatal(err)
	}

	h.processMutex.Lock()
	if len(h.processes) != 0 {
		t.Fatal("not all processes exited")
	}
	h.processMutex.Unlock()

	wg := &sync.WaitGroup{}
	errChan := make(chan error, 1)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if err := h.launchCommand([]string{"testbin", "-sleep", "1"}, runConfig); err != nil {
				errChan <- err
			}
		}(wg)
	}

	wg.Wait()

	if err := h.waitForQuiet(time.Minute); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}

	h.processMutex.Lock()
	if len(h.processes) != 0 {
		t.Fatal("not all processes exited")
	}
	h.processMutex.Unlock()
}

func TestSupervisorHang(t *testing.T) {
	cfg := &Config{
		MaxProcessors: 10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	h, err := NewHandler(ctx, cfg, func(io.ReadCloser) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	runConfig := map[string]interface{}{"this test hangs": "yep"}

	for i := 0; i < 10; i++ {
		if err := h.launchCommand([]string{"testbin", "-hang"}, runConfig); err != nil {
			t.Fatal(err)
		}
	}

	errChan := make(chan error, 2) // in the error case, a nil will be sent after the error
	go func() {
		defer func() { errChan <- nil }()
		if err := h.launchCommand([]string{"testbin"}, runConfig); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}

	if err := h.waitForQuiet(time.Second); err == nil {
		t.Fatal("did not timeout waiting")
	}

	cancel()

	if err := h.waitForQuiet(time.Second); err == nil {
		t.Fatal("did not report context canceled")
	}

	time.Sleep(100 * time.Millisecond) // wait for processes to die

	if err := exec.Command("pgrep", "testbin").Run(); err == nil {
		t.Fatal("processes were left behind")
	}
}
