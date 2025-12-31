package pool_test

import (
	"testing"

	"github.com/Pklerik/urlshortener/internal/model"
	"github.com/Pklerik/urlshortener/internal/model/pool"
)

func TestNew(t *testing.T) {

	gotLD := pool.New[*model.LinkData]()
	// TODO: update the condition below to compare got with tt.want.
	if gotLD == nil {
		t.Errorf("New() = nil")
	}
	gotR := pool.New[*model.Request]()
	if gotR == nil {
		t.Errorf("New() = nil")
	}
}

func TestPool_Get(t *testing.T) {

	t.Run("model.LinkData", func(t *testing.T) {
		// TODO: construct the receiver type.
		var p = pool.New[*model.LinkData]()

		got := p.Get()
		// TODO: update the condition below to compare got with tt.want.
		if got == nil {
			t.Errorf("Get() = %v", got)
		}
	})

	t.Run("model.Reques", func(t *testing.T) {
		// TODO: construct the receiver type.
		var p = pool.New[*model.Request]()

		got := p.Get()
		// TODO: update the condition below to compare got with tt.want.
		if got == nil {
			t.Errorf("Get() = %v", got)
		}
	})
}

func TestPool_Put(t *testing.T) {

	t.Run("model.LinkData", func(t *testing.T) {
		// TODO: construct the receiver type.
		var (
			p    = pool.New[*model.LinkData]()
			data = model.LinkData{}
		)

		p.Put(&data)
	})

	t.Run("model.Reques", func(t *testing.T) {
		// TODO: construct the receiver type.
		var (
			p    = pool.New[*model.Request]()
			data = model.Request{}
		)

		p.Put(&data)

	})
}
