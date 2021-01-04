package templatestests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTestInterfaceWithRateLimit_F(t *testing.T) {
	impl := &testImpl{r1: "1", r2: "2"}

	wrapped := NewTestInterfaceWithRateLimit(impl, 3, 10)

	go func() {
		for i := 0; i < 10; i++ {
			r1, r2, err := wrapped.F(context.Background(), "a1")
			assert.NoError(t, err)
			assert.Equal(t, "1", r1)
			assert.Equal(t, "2", r2)
		}
	}()

	<-time.After(150 * time.Millisecond)

	counter := atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 4, counter) //3 burst request + 1 requests after tick

	<-time.After(200 * time.Millisecond)

	counter = atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 6, counter) //after bust we should receiv 1 request per 100 milliseconds
}

func TestTestInterfaceWithRateLimit_F_HugeRPS(t *testing.T) {
	impl := &testImpl{r1: "1", r2: "2", delay: 100 * time.Millisecond}

	wrapped := NewTestInterfaceWithRateLimit(impl, 3, 1000)

	for i := 0; i < 10; i++ {
		go func() {
			r1, r2, err := wrapped.F(context.Background(), "a1")
			assert.NoError(t, err)
			assert.Equal(t, "1", r1)
			assert.Equal(t, "2", r2)

		}()
	}

	<-time.After(10 * time.Millisecond)

	counter := atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 3, counter) // the first burst

	<-time.After(100 * time.Millisecond)

	counter = atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 6, counter) // the second burst

	<-time.After(100 * time.Millisecond)

	counter = atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 9, counter) // the third burst

	<-time.After(100 * time.Millisecond)

	counter = atomic.LoadUint64(&impl.callCounter)
	assert.EqualValues(t, 10, counter) // the 10th call
}
