package integration

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Romasmi/anti-bruteforce/pkg/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	client  api.AntiBruteforceClient
	counter atomic.Int64
)

// nextID returns a process-unique integer string, safe for parallel tests.
func nextID() string { return fmt.Sprintf("%d", counter.Add(1)) }

// uniqueIP returns a unique routable-looking IPv4 from the 10.x.x.0/8 range.
func uniqueIP() string {
	n := counter.Add(1)
	return fmt.Sprintf("10.%d.%d.1", (n/254)%254+1, n%254+1)
}

func TestMain(m *testing.M) {
	host := os.Getenv("ANTI_BRUTEFORCE_GRPC_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("ANTI_BRUTEFORCE_GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", host, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "grpc dial:", err)
		os.Exit(1)
	}

	client = api.NewAntiBruteforceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := waitReady(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "service not ready:", err)
		_ = conn.Close()
		os.Exit(1)
	}

	code := m.Run()
	_ = conn.Close()
	os.Exit(code)
}

func waitReady(ctx context.Context) error {
	for {
		_, err := client.Healthcheck(ctx, &api.HealthcheckRequest{})
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service: %w", err)
		case <-time.After(2 * time.Second):
		}
	}
}

// requireInvalidArgument asserts that err is a gRPC InvalidArgument error.
func requireInvalidArgument(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

// --- tests ---

func TestCheckAuth_Validation(t *testing.T) {
	t.Parallel()

	t.Run("empty login", func(t *testing.T) {
		t.Parallel()
		_, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: "", Password: "pass", Ip: "1.0.0.1"})
		requireInvalidArgument(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()
		_, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: "user", Password: "", Ip: "1.0.0.1"})
		requireInvalidArgument(t, err)
	})

	t.Run("empty ip", func(t *testing.T) {
		t.Parallel()
		_, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: "user", Password: "pass", Ip: ""})
		requireInvalidArgument(t, err)
	})

	t.Run("invalid ip format", func(t *testing.T) {
		t.Parallel()
		_, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: "user", Password: "pass", Ip: "not-an-ip"})
		requireInvalidArgument(t, err)
	})

	t.Run("ipv6 rejected", func(t *testing.T) {
		t.Parallel()
		_, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: "user", Password: "pass", Ip: "::1"})
		requireInvalidArgument(t, err)
	})
}

func TestClearRate_Validation(t *testing.T) {
	t.Parallel()
	_, err := client.ClearRate(t.Context(), &api.ClearRateRequest{Login: "", Ip: ""})
	requireInvalidArgument(t, err)
}

func TestCheckAuth_Allowed(t *testing.T) {
	t.Parallel()
	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "allow_" + nextID(), Password: "pass", Ip: uniqueIP(),
	})
	require.NoError(t, err)
	require.True(t, resp.Ok)
}

// TestCheckAuth_LoginRateLimit exhausts the login bucket (capacity=3 in the integration config) and verifies the next request is denied.
func TestCheckAuth_LoginRateLimit(t *testing.T) {
	t.Parallel()
	login := "lrl_" + nextID()
	ip := uniqueIP()

	for i := range 3 {
		resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: login, Password: "pass", Ip: ip})
		require.NoError(t, err)
		require.True(t, resp.Ok, "request %d should be allowed", i+1)
	}

	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: login, Password: "pass", Ip: ip})
	require.NoError(t, err)
	require.False(t, resp.Ok, "request after login bucket exhausted should be denied")
}

// TestCheckAuth_IPRateLimit exhausts the IP bucket (capacity=10 in the integration config).
func TestCheckAuth_IPRateLimit(t *testing.T) {
	t.Parallel()
	ip := uniqueIP()

	for i := range 10 {
		resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
			Login: fmt.Sprintf("iprl_%s_%d", nextID(), i), Password: "pass", Ip: ip,
		})
		require.NoError(t, err)
		require.True(t, resp.Ok, "request %d should be allowed", i+1)
	}

	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "iprl_" + nextID(), Password: "pass", Ip: ip,
	})
	require.NoError(t, err)
	require.False(t, resp.Ok, "request after IP bucket exhausted should be denied")
}

// TestClearRate_Login exhausts the login bucket, resets it, and verifies that
// the next request is allowed again.
func TestClearRate_Login(t *testing.T) {
	t.Parallel()
	login := "crl_" + nextID()
	ip := uniqueIP()

	for i := range 3 {
		resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: login, Password: "pass", Ip: ip})
		require.NoError(t, err)
		require.True(t, resp.Ok, "fill request %d should be allowed", i+1)
	}

	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: login, Password: "pass", Ip: ip})
	require.NoError(t, err)
	require.False(t, resp.Ok, "should be denied after login bucket exhausted")

	_, err = client.ClearRate(t.Context(), &api.ClearRateRequest{Login: login})
	require.NoError(t, err)

	resp, err = client.CheckAuth(t.Context(), &api.CheckAuthRequest{Login: login, Password: "pass", Ip: ip})
	require.NoError(t, err)
	require.True(t, resp.Ok, "should be allowed again after bucket reset")
}

// TestClearRate_IP exhausts the IP bucket, resets it, and verifies recovery.
func TestClearRate_IP(t *testing.T) {
	t.Parallel()
	ip := uniqueIP()

	for i := range 10 {
		resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
			Login: fmt.Sprintf("crip_%s_%d", nextID(), i), Password: "pass", Ip: ip,
		})
		require.NoError(t, err)
		require.True(t, resp.Ok, "fill request %d should be allowed", i+1)
	}

	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "crip_chk_" + nextID(), Password: "pass", Ip: ip,
	})
	require.NoError(t, err)
	require.False(t, resp.Ok, "should be denied after IP bucket exhausted")

	_, err = client.ClearRate(t.Context(), &api.ClearRateRequest{Ip: ip})
	require.NoError(t, err)

	resp, err = client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "crip_fin_" + nextID(), Password: "pass", Ip: ip,
	})
	require.NoError(t, err)
	require.True(t, resp.Ok, "should be allowed again after IP bucket reset")
}

// TestBlacklist_Validation ensures the API rejects invalid inputs.
func TestBlacklist_Validation(t *testing.T) {
	t.Parallel()

	t.Run("empty subnet", func(t *testing.T) {
		t.Parallel()
		_, err := client.AddToBlacklist(t.Context(), &api.IPListRequest{Subnet: ""})
		requireInvalidArgument(t, err)
	})

	t.Run("invalid cidr", func(t *testing.T) {
		t.Parallel()
		_, err := client.AddToBlacklist(t.Context(), &api.IPListRequest{Subnet: "not-a-cidr"})
		requireInvalidArgument(t, err)
	})
}

// TestWhitelist_Validation ensures the API rejects invalid inputs.
func TestWhitelist_Validation(t *testing.T) {
	t.Parallel()

	t.Run("empty subnet", func(t *testing.T) {
		t.Parallel()
		_, err := client.AddToWhitelist(t.Context(), &api.IPListRequest{Subnet: ""})
		requireInvalidArgument(t, err)
	})

	t.Run("invalid cidr", func(t *testing.T) {
		t.Parallel()
		_, err := client.AddToWhitelist(t.Context(), &api.IPListRequest{Subnet: "not-a-cidr"})
		requireInvalidArgument(t, err)
	})
}

// TestBlacklist_BlocksAndUnblocks adds an IP to the blacklist, verifies auth is denied,
// removes it, and verifies auth is allowed again.
func TestBlacklist_BlocksAndUnblocks(t *testing.T) {
	t.Parallel()
	ip := uniqueIP()
	subnet := ip + "/32"

	_, err := client.AddToBlacklist(t.Context(), &api.IPListRequest{Subnet: subnet})
	require.NoError(t, err)

	resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "bl_" + nextID(), Password: "pass", Ip: ip,
	})
	require.NoError(t, err)
	require.False(t, resp.Ok, "blacklisted IP should be denied")

	_, err = client.RemoveFromBlacklist(t.Context(), &api.IPListRequest{Subnet: subnet})
	require.NoError(t, err)

	resp, err = client.CheckAuth(t.Context(), &api.CheckAuthRequest{
		Login: "bl_" + nextID(), Password: "pass", Ip: ip,
	})
	require.NoError(t, err)
	require.True(t, resp.Ok, "IP should be allowed after blacklist removal")
}

// TestWhitelist_BypassesRateLimit adds an IP to the whitelist and verifies it is always
// allowed even after the per-IP bucket is exhausted.
func TestWhitelist_BypassesRateLimit(t *testing.T) {
	t.Parallel()
	ip := uniqueIP()
	subnet := ip + "/32"

	_, err := client.AddToWhitelist(t.Context(), &api.IPListRequest{Subnet: subnet})
	require.NoError(t, err)
	defer client.RemoveFromWhitelist(t.Context(), &api.IPListRequest{Subnet: subnet}) //nolint:errcheck

	// Send far more requests than the IP capacity (10 in integration config).
	for i := range 20 {
		resp, err := client.CheckAuth(t.Context(), &api.CheckAuthRequest{
			Login: fmt.Sprintf("wl_%s_%d", nextID(), i), Password: "pass", Ip: ip,
		})
		require.NoError(t, err)
		require.True(t, resp.Ok, "whitelisted IP should always be allowed (request %d)", i+1)
	}
}
