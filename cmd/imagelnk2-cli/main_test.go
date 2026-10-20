package main

import (
	"context"
	"errors"
	"imagelnk2/internal/core"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProcessor struct {
	errors []error
	calls  int
}

func (p *stubProcessor) GetImageURLs(context.Context, string) (*core.Result, error) {
	err := p.errors[p.calls]
	p.calls++
	if err != nil {
		return nil, err
	}
	return &core.Result{Title: "result"}, nil
}

func TestGetWithRetrySucceedsAfterRetryableError(t *testing.T) {
	p := &stubProcessor{errors: []error{context.DeadlineExceeded, nil}}

	result, err := getWithRetry(context.Background(), p, "https://example.com")

	require.NoError(t, err)
	assert.Equal(t, "result", result.Title)
	assert.Equal(t, 2, p.calls)
}

func TestGetWithRetryStopsAfterNonRetryableError(t *testing.T) {
	expected := errors.New("unsupported URL")
	p := &stubProcessor{errors: []error{expected}}

	result, err := getWithRetry(context.Background(), p, "invalid")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, expected)
	assert.Equal(t, 1, p.calls)
}

func TestGetWithRetryStopsAfterMaximumAttempts(t *testing.T) {
	p := &stubProcessor{errors: []error{
		context.DeadlineExceeded,
		context.DeadlineExceeded,
		context.DeadlineExceeded,
		context.DeadlineExceeded,
		context.DeadlineExceeded,
	}}

	result, err := getWithRetry(context.Background(), p, "https://example.com")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, maxAttempts, p.calls)
}
