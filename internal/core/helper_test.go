package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-rod/rod/lib/cdp"
	"github.com/stretchr/testify/assert"
)

func TestIsNavigationRace(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "target navigated",
			err: &cdp.Error{
				Code:    -32000,
				Message: "Inspected target navigated or closed",
			},
			want: true,
		},
		{
			name: "execution context destroyed",
			err:  cdp.ErrCtxDestroyed,
			want: true,
		},
		{
			name: "wrapped navigation error",
			err: fmt.Errorf("query document: %w", &cdp.Error{
				Code:    -32000,
				Message: "Inspected target navigated or closed",
			}),
			want: true,
		},
		{
			name: "different protocol error",
			err:  cdp.ErrObjNotFound,
			want: false,
		},
		{
			name: "ordinary error",
			err:  errors.New("failed"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isNavigationRace(tt.err))
		})
	}
}
