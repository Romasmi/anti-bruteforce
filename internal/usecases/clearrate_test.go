package usecases_test

import (
	"context"
	"testing"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClearRateUsecase(t *testing.T) {
	uc := &usecases.ClearRateUsecase{}
	ctx := context.Background()

	t.Run("valid login", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeLogin, Value: "user123"})
		require.NoError(t, err)
	})

	t.Run("valid ip", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeIP, Value: "10.0.0.1"})
		require.NoError(t, err)
	})

	t.Run("unspecified type", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeUnspecified, Value: "user123"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("empty value", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeLogin, Value: ""})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid ip format", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeIP, Value: "999.999.999.999"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Type: usecases.IdentifierTypeIP, Value: "::1"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("wrong request type", func(t *testing.T) {
		_, err := uc.Do(ctx, 42)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})
}
