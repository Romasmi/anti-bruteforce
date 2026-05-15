package usecases_test

import (
	"testing"

	"github.com/Romasmi/anti-bruteforce/internal/usecases"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClearRateUsecase(t *testing.T) {
	uc := &usecases.ClearRateUsecase{}
	ctx := t.Context()

	t.Run("valid login", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "user123"})
		require.NoError(t, err)
	})

	t.Run("valid ip", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "10.0.0.1"})
		require.NoError(t, err)
	})

	t.Run("valid both", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "user123", IP: "10.0.0.1"})
		require.NoError(t, err)
	})

	t.Run("empty both", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{Login: "", IP: ""})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid ip format", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "999.999.999.999"})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		_, err := uc.Do(ctx, &usecases.ClearRateInput{IP: "::1"})
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
