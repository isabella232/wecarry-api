package models

import (
	"context"
	"sync"
	"testing"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"

	"github.com/gobuffalo/suite"
)

type ModelSuite struct {
	*suite.Model
}

func Test_ModelSuite(t *testing.T) {
	model := suite.NewModel()

	as := &ModelSuite{
		Model: model,
	}
	suite.Run(t, as)
}

func createFixture(ms *ModelSuite, f interface{}) {
	err := ms.DB.Create(f)
	if err != nil {
		ms.T().Errorf("error creating %T fixture, %s", f, err)
		ms.T().FailNow()
	}
}

type buffaloTestCtx struct {
	buffalo.DefaultContext
	data *sync.Map
}

// Set a value onto the Context. Any value set onto the Context
// will be automatically available in templates.
func (b *buffaloTestCtx) Set(key string, value interface{}) {
	b.data.Store(key, value)
}

// Value that has previously stored on the context.
func (b *buffaloTestCtx) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		if v, ok := b.data.Load(k); ok {
			return v
		}
	}
	return b.Context.Value(key)
}

func testContext(tx *pop.Connection) context.Context {
	data := &sync.Map{}

	ctx := &buffaloTestCtx{
		data: data,
	}

	ctx.Set("tx", tx)

	return ctx
}
