package graphql

import (
	"context"
	"encoding/json"

	"github.com/neelance/graphql-go/internal/common"
	"github.com/neelance/graphql-go/internal/exec/resolvable"
	"github.com/neelance/graphql-go/internal/exec/selected"
	"github.com/neelance/graphql-go/internal/query"

	"github.com/neelance/graphql-go/internal/validation"
	"github.com/neelance/graphql-go/internal/exec"
	"github.com/neelance/graphql-go/introspection"
)

func (s *Schema) Subscribe(ctx context.Context, queryString string, operationName string, variables map[string]interface{}) (<-chan json.RawMessage,error) {
		return s.subscribe(ctx,queryString,operationName,variables,s.res)
}

func (s *Schema) subscribe(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, res *resolvable.Schema) (<-chan json.RawMessage, error) {
	doc, qErr := query.Parse(queryString)
	if qErr != nil {
		return nil, qErr
	}

	errs := validation.Validate(s.schema, doc)
	if len(errs) != 0 {
		return nil, errs[0]
	}

	op, err := getOperation(doc, operationName)
	if err != nil {
		return nil, err
	}

	r := &exec.Request{
		Request: selected.Request{
			Doc:    doc,
			Vars:   variables,
			Schema: s.schema,
		},
		Limiter: make(chan struct{}, s.maxParallelism),
		Tracer:  s.tracer,
		Logger:  s.logger,
	}

	varTypes := make(map[string]*introspection.Type)

	for _, v := range op.Vars {
		t, err := common.ResolveType(v.Type, s.schema.Resolve)
		if err != nil {
			return nil, err //&Response{Errors: []*errors.QueryError{err}}
		}
		varTypes[v.Name.Name] = introspection.WrapType(t)
	}
	traceCtx, finish := s.tracer.TraceQuery(ctx, queryString, operationName, variables, varTypes)

	// XXX: maybe here is good place for checking subscription operation
	// TODO: check errors
	ch, _  := r.Subscribe(traceCtx, res, op)
	finish(errs)
	return ch,nil
}

