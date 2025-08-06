package peer2peer

import (
	"errors"
	"sync"
)

// progress bar
type ProgressBar struct {
	Max            int64
	CurrentPercent float64
	CurrentBytes   int64
	lock           sync.Mutex
}

// New64 returns a new ProgressBar
// with the specified maximum
func New64(max int64) *ProgressBar {
	b := ProgressBar{
		Max:            max,
		CurrentPercent: 0,
		CurrentBytes:   0,
		lock:           sync.Mutex{},
	}
	return &b
}

// GetMax returns the max of a bar
func (p *ProgressBar) GetMax() int64 {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.Max
}

// GetCurrentBytes returns the current bytes
func (p *ProgressBar) GetCurrentBytes() int64 {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.CurrentBytes
}

// Add will add the specified amount to the progressbar
func (p *ProgressBar) Add(num int) error {
	return p.Add64(int64(num))
}

// Add64 will add the specified amount to the progressbar
func (p *ProgressBar) Add64(num int64) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.Max <= 0 {
		return errors.New("max must be greater than 0")
	}

	p.CurrentBytes += num
	p.CurrentPercent = float64(p.CurrentBytes) / float64(p.Max)
	return nil
}

// Write implement io.Writer
func (p *ProgressBar) Write(b []byte) (n int, err error) {
	n = len(b)
	err = p.Add(n)
	return
}

// Read implement io.Reader
func (p *ProgressBar) Read(b []byte) (n int, err error) {
	n = len(b)
	err = p.Add(n)
	return
}
