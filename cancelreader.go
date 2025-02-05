package cancelreader

import (
	"fmt"
	"io"
	"sync"
)

// ErrCanceled gets returned when trying to read from a canceled reader.
var ErrCanceled = fmt.Errorf("read canceled")

// CancelReader is a io.Reader whose Read() calls can be canceled without data
// being consumed. The cancelReader has to be closed.
type CancelReader interface {
	io.ReadCloser

	// Cancel cancels ongoing and future reads an returns true if it succeeded.
	Cancel() bool
}

// fallbackCancelReader implements cancelReader but does not actually support
// cancelation during an ongoing Read() call. Thus, Cancel() always returns
// false. However, after calling Cancel(), new Read() calls immediately return
// errCanceled and don't consume any data anymore.
type fallbackCancelReader struct {
	r        io.Reader
	canceled bool
}

// newFallbackCancelReader is a fallback for NewReader that cannot actually
// cancel an ongoing read but will immediately return on future reads if it has
// been canceled.
func newFallbackCancelReader(reader io.Reader) (CancelReader, error) {
	return &fallbackCancelReader{r: reader}, nil
}

func (r *fallbackCancelReader) Read(data []byte) (int, error) {
	if r.canceled {
		return 0, ErrCanceled
	}

	return r.r.Read(data)
}

func (r *fallbackCancelReader) Cancel() bool {
	r.canceled = true

	return false
}

func (r *fallbackCancelReader) Close() error {
	return nil
}

// cancelMixin represents a goroutine-safe cancelation status.
type cancelMixin struct {
	unsafeCanceled bool
	lock           sync.Mutex
}

func (c *cancelMixin) isCanceled() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.unsafeCanceled
}

func (c *cancelMixin) setCanceled() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.unsafeCanceled = true
}
