package generator

import (
	"path/filepath"
	"testing"

	"github.com/mvanhorn/cli-printing-press/v4/internal/naming"
	"github.com/mvanhorn/cli-printing-press/v4/internal/spec"
	"github.com/stretchr/testify/require"
)

func TestEmitsConcurrentTokenRaceTest(t *testing.T) {
	t.Parallel()

	cc := ccRemintSpec()
	require.True(t, New(cc, t.TempDir()).emitsConcurrentTokenRaceTest())

	refresh := minimalSpec("oauth2-refresh-race")
	refresh.Auth = spec.AuthConfig{
		Type:     spec.AuthTypeOAuth2Refresh,
		Header:   "Authorization",
		TokenURL: "https://auth.example.com/oauth/token",
	}
	require.True(t, New(refresh, t.TempDir()).emitsConcurrentTokenRaceTest())

	apiKey := minimalSpec("api-key-race")
	require.False(t, New(apiKey, t.TempDir()).emitsConcurrentTokenRaceTest())

	google := minimalSpec("google-race")
	google.Auth = spec.AuthConfig{
		Type:     "bearer_token",
		Subtype:  spec.AuthSubtypeGoogleServiceAccount,
		Header:   "Authorization",
		TokenURL: "https://oauth2.googleapis.com/token",
	}
	require.False(t, New(google, t.TempDir()).emitsConcurrentTokenRaceTest())
}

func TestConcurrentClientCredentialsMintIsRaceFree(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("generated CLI race tests run in the full generated-test CI lane")
	}

	apiSpec := ccRemintSpec()
	apiSpec.Name = "ccrace"
	apiSpec.Auth.EnvVars = []string{"CCRACE_API_KEY", "CCRACE_SECRET_KEY"}
	apiSpec.Config.Path = "~/.config/ccrace-pp-cli/config.toml"
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	body := readGeneratedFile(t, outputDir, "internal", "client", "token_race_test.go")
	require.Contains(t, body, "func TestConcurrentClientCredentialsMint")

	out, err := runGoCommandOutput(t, outputDir, "test", "-race", "-count=1", "-v", "./internal/client", "-run", "^TestConcurrentClientCredentialsMint$")
	require.NoError(t, err, out)
	require.Contains(t, out, "--- PASS: TestConcurrentClientCredentialsMint")
}

func TestConcurrentRefreshTokenMintIsRaceFree(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("generated CLI race tests run in the full generated-test CI lane")
	}

	apiSpec := minimalSpec("refreshrace")
	apiSpec.Auth = spec.AuthConfig{
		Type:     spec.AuthTypeOAuth2Refresh,
		Header:   "Authorization",
		TokenURL: "https://auth.example.com/oauth/token",
		EnvVars: []string{
			"REFRESHRACE_CLIENT_ID",
			"REFRESHRACE_CLIENT_SECRET",
			"REFRESHRACE_REFRESH_TOKEN",
		},
	}
	outputDir := filepath.Join(t.TempDir(), naming.CLI(apiSpec.Name))
	require.NoError(t, New(apiSpec, outputDir).Generate())

	body := readGeneratedFile(t, outputDir, "internal", "client", "token_race_test.go")
	require.Contains(t, body, "func TestConcurrentRefreshTokenMint")

	out, err := runGoCommandOutput(t, outputDir, "test", "-race", "-count=1", "-v", "./internal/client", "-run", "^TestConcurrentRefreshTokenMint$")
	require.NoError(t, err, out)
	require.Contains(t, out, "--- PASS: TestConcurrentRefreshTokenMint")
}
