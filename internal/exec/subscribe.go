package exec

import (
	"context"
	"encoding/json"
	"reflect"
	"bytes"

	"github.com/neelance/graphql-go/internal/exec/resolvable"
	"github.com/neelance/graphql-go/internal/exec/selected"
	"github.com/neelance/graphql-go/internal/query"
)

func (r *Request) Subscribe(ctx context.Context, s *resolvable.Schema, op *query.Operation) (<-chan json.RawMessage, error) {
	// TODO: check if op is subscription!
	// TODO: handle traceCtx 

	// consider to set capacity the same as resolver. result.Cap()
	ch := make(chan json.RawMessage, 1)
	go func() {
		defer r.handlePanic(ctx)

		sels := selected.ApplyOperation(&r.Request, s, op)
		var fields []*fieldToExec
		collectFieldsToResolve(sels, s.Resolver, &fields, make(map[string]*fieldToExec))

		f := fields[0]

		var in []reflect.Value
		if f.field.HasContext {
			in = append(in, reflect.ValueOf(ctx))
		}
		if f.field.ArgsPacker != nil {
			in = append(in, f.field.PackedArgs)
		}
		callOut := f.resolver.Method(f.field.MethodIndex).Call(in)
		result := callOut[0]

		// TODO: check error callOut[1]

		for {
			obj, ok := result.Recv()
			if !ok {
				close(ch)
				return
			}
			var out bytes.Buffer
			r.execSelectionSet(ctx, f.sels, f.field.Type, &pathSegment{nil,f.field.Alias}, obj, &out)
			ch <- json.RawMessage(out.Bytes())
		}
	}()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return ch, nil

}
