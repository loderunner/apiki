//go:build acceptance

package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	nginxContainer   testcontainers.Container
	testNetwork      *testcontainers.DockerNetwork
	nginxHost        string
	nginxPort        string
	distDir          string
	testArtifactsDir string
	snapshotVersion  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Run goreleaser snapshot build
	distDir = filepath.Join("..", "dist")
	if err := runGoreleaser(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build snapshot artifacts: %v\n", err)
		os.Exit(1)
	}

	// Get the snapshot version from goreleaser's metadata.json
	snapshotVersion = findSnapshotVersionForSetup()
	if snapshotVersion == "" {
		fmt.Fprintf(
			os.Stderr,
			"failed to determine snapshot version from dist/metadata.json\n",
		)
		os.Exit(1)
	}

	// Create a temporary directory for test artifacts (don't modify dist/)
	var err error
	testArtifactsDir, err = os.MkdirTemp("", "apiki-acceptance-*")
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed to create test artifacts directory: %v\n",
			err,
		)
		os.Exit(1)
	}

	//nolint:lll,golines
	// Create version directory structure to match GitHub releases URL structure
	// GitHub releases:
	// https://github.com/.../releases/download/v1.0.0/install.sh
	// https://github.com/.../releases/download/v1.0.0/apiki_1.0.0_linux_amd64.tar.gz
	// We serve:        http://nginx:80/v0.0.0-SNAPSHOT-none/install.sh
	//                  http://nginx:80/v0.0.0-SNAPSHOT-none/apiki_0.0.0-SNAPSHOT-none_linux_amd64.tar.gz
	versionDir := filepath.Join(testArtifactsDir, snapshotVersion)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create version directory: %v\n", err)
		os.Exit(1)
	}

	// Copy install.sh (versioned) and release artifacts to the version
	// directory
	if err := copyFile(
		filepath.Join(distDir, "install.sh"),
		filepath.Join(versionDir, "install.sh"),
	); err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy install.sh: %v\n", err)
		os.Exit(1)
	}

	entries, readErr := os.ReadDir(distDir)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "failed to read dist directory: %v\n", readErr)
		os.Exit(1)
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if strings.HasPrefix(entry.Name(), "apiki_") {
			src := filepath.Join(distDir, entry.Name())
			dst := filepath.Join(versionDir, entry.Name())
			if err := copyFile(src, dst); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"failed to copy %s: %v\n",
					entry.Name(),
					err,
				)
				os.Exit(1)
			}
		}
	}

	// Create Docker network for containers to communicate
	var networkErr error
	testNetwork, networkErr = network.New(ctx)
	if networkErr != nil {
		fmt.Fprintf(os.Stderr, "failed to create network: %v\n", networkErr)
		os.Exit(1)
	}

	// Start nginx container to serve test artifacts
	absTestArtifactsDir, err := filepath.Abs(testArtifactsDir)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed to get absolute test artifacts path: %v\n",
			err,
		)
		os.Exit(1)
	}

	absNginxConf, err := filepath.Abs("nginx.conf")
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed to get absolute nginx.conf path: %v\n",
			err,
		)
		os.Exit(1)
	}

	nginxReq := testcontainers.ContainerRequest{
		Image:        "nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absNginxConf,
				ContainerFilePath: "/etc/nginx/nginx.conf",
			},
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Mounts = append(hc.Mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: absTestArtifactsDir,
				Target: "/usr/share/nginx/html",
			})
		},
		Networks: []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"nginx"},
		},
		WaitingFor: wait.ForHTTP("/").WithPort("80/tcp"),
	}

	nginxContainer, err = testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: nginxReq,
			Started:          true,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start nginx container: %v\n", err)
		os.Exit(1)
	}

	nginxHost, err = nginxContainer.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get nginx host: %v\n", err)
		os.Exit(1)
	}

	mappedPort, err := nginxContainer.MappedPort(ctx, "80")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get nginx port: %v\n", err)
		os.Exit(1)
	}
	nginxPort = mappedPort.Port()

	code := m.Run()

	// Cleanup
	if nginxContainer != nil {
		_ = nginxContainer.Terminate(ctx)
	}
	if testNetwork != nil {
		_ = (*testNetwork).Remove(ctx)
	}
	if testArtifactsDir != "" {
		_ = os.RemoveAll(testArtifactsDir)
	}

	os.Exit(code)
}

func runGoreleaser() error {
	cmd := exec.Command("goreleaser",
		"release",
		"--snapshot",
		"--clean",
		"--skip=publish",
	)
	cmd.Dir = ".."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func TestInstallScript(t *testing.T) {
	ctx := context.Background()
	// nginx serves files from testArtifactsDir (mounted at
	// /usr/share/nginx/html)
	// Structure:
	//   - /<version>/install.sh
	//   - /<version>/apiki_<version>_<os>_<arch>.tar.gz
	baseURL := "http://nginx:80"

	t.Run("SuccessWithCurl", func(t *testing.T) {
		t.Parallel()
		container := startTestContainer(t, ctx, "curl-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testSuccessInstall(
			t,
			ctx,
			container,
			baseURL,
			map[string]string{},
			"curl",
		)
		testShellProfile(t, ctx, container, ".bashrc")
	})

	t.Run("SuccessWithWget", func(t *testing.T) {
		t.Parallel()
		container := startTestContainer(t, ctx, "wget-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testSuccessInstall(
			t,
			ctx,
			container,
			baseURL,
			map[string]string{},
			"wget",
		)
		testShellProfile(t, ctx, container, ".bashrc")
	})

	t.Run("SuccessWithZsh", func(t *testing.T) {
		t.Parallel()
		container := startTestContainer(t, ctx, "curl-zsh")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testSuccessInstall(
			t,
			ctx,
			container,
			baseURL,
			map[string]string{},
			"curl",
		)
		testShellProfile(t, ctx, container, ".zshrc")
	})

	t.Run("SuccessWithFish", func(t *testing.T) {
		t.Parallel()
		container := startTestContainer(t, ctx, "fish")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testSuccessInstall(
			t,
			ctx,
			container,
			baseURL,
			map[string]string{},
			"curl",
		)
		testShellProfile(t, ctx, container, ".config/fish/config.fish")
	})

	t.Run("SuccessCustomDir", func(t *testing.T) {
		t.Parallel()
		customDir := "/tmp/apiki-custom"
		// Pre-create the custom directory
		// (install script only auto-creates the default dir)
		container := startTestContainer(t, ctx, "curl-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		_, _, err := container.Exec(ctx, []string{"mkdir", "-p", customDir})
		require.NoError(t, err)
		testSuccessInstall(t, ctx, container, baseURL, map[string]string{
			"APIKI_DIR": customDir,
		}, "curl")

		// Verify binary exists at the custom path
		binaryPath := customDir + "/apiki"
		cmd := []string{"test", "-x", binaryPath}
		exitCode, _, err := container.Exec(ctx, cmd)
		require.NoError(t, err)
		require.Equal(t, 0, exitCode, "binary should be at %s", binaryPath)
	})

	t.Run("FailureDownloadError", func(t *testing.T) {
		t.Parallel()
		// Point to a non-existent version directory so artifact download fails
		// The install script is fetched OK, but the archive download will fail
		container := startTestContainer(t, ctx, "curl-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testFailureInstall(t, ctx, container, baseURL, map[string]string{
			"APIKI_INSTALL_BASE_URL": "http://nginx:80/nonexistent-version",
		}, "curl", "Failed to download")
	})

	t.Run("FailureChecksumMismatch", func(t *testing.T) {
		t.Parallel()
		// This test would require tampering with the checksum file
		// For now, we'll skip it as it's complex to set up
		t.Skip("checksum mismatch test requires tampering with artifacts")
	})

	t.Run("FailureDirIsFile", func(t *testing.T) {
		t.Parallel()
		// Create a file where the install directory would be
		container := startTestContainer(t, ctx, "curl-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		fileAsDir := "/tmp/apiki-is-a-file"
		_, _, err := container.Exec(ctx, []string{"touch", fileAsDir})
		require.NoError(t, err)
		testFailureInstall(t, ctx, container, baseURL, map[string]string{
			"APIKI_DIR": fileAsDir,
		}, "curl", "has the same name as installation directory")
	})

	t.Run("FailureDirMissing", func(t *testing.T) {
		t.Parallel()
		container := startTestContainer(t, ctx, "curl-bash")
		t.Cleanup(func() { _ = container.Terminate(ctx) })
		testFailureInstall(t, ctx, container, baseURL, map[string]string{
			"APIKI_DIR": "/nonexistent/dir",
		}, "curl", "that directory does not exist")
	})
}

func testSuccessInstall(
	t *testing.T,
	ctx context.Context,
	container testcontainers.Container,
	baseURL string,
	envVars map[string]string,
	downloader string,
) {
	t.Helper()
	installScriptURL := fmt.Sprintf(
		"%s/%s/install.sh",
		baseURL,
		snapshotVersion,
	)
	var cmd []string

	if downloader == "curl" {
		cmd = []string{
			"sh",
			"-c",
			fmt.Sprintf("curl -sSL %s | sh", installScriptURL),
		}
	} else {
		cmd = []string{
			"sh",
			"-c",
			fmt.Sprintf("wget -qO- %s | sh", installScriptURL),
		}
	}

	// Set version environment variables if not already set
	if _, ok := envVars["APIKI_INSTALL_BASE_URL"]; !ok {
		envVars["APIKI_INSTALL_BASE_URL"] = fmt.Sprintf(
			"%s/%s",
			baseURL,
			snapshotVersion,
		)
	}
	if _, ok := envVars["APIKI_VERSION"]; !ok {
		envVars["APIKI_VERSION"] = snapshotVersion
	}

	env := []string{}
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	exitCode, _, err := container.Exec(ctx, cmd, tcexec.WithEnv(env))
	require.NoError(t, err)
	require.Equal(t, 0, exitCode, "install script should succeed")

	// Verify installation by running `apiki version`
	// We need to manually source profile files because non-interactive shells
	// (shell -c) don't source profiles automatically
	getShellCmd := []string{"sh", "-c", "echo $SHELL"}
	_, shellReader, err := container.Exec(
		ctx,
		getShellCmd,
		tcexec.Multiplexed(),
	)
	require.NoError(t, err)
	shellBytes, err := io.ReadAll(shellReader)
	require.NoError(t, err)
	shell := strings.TrimSpace(string(shellBytes))

	var versionCmd []string
	if strings.Contains(shell, "fish") {
		versionCmd = []string{
			shell, "-c",
			"source ~/.config/fish/config.fish; apiki version",
		}
	} else {
		versionCmd = []string{
			shell, "-c",
			`[ -f ~/.bashrc ] && . ~/.bashrc \
				|| [ -f ~/.zshrc ] && . ~/.zshrc \
				|| [ -f ~/.profile ] && . ~/.profile; \
				apiki version`,
		}
	}

	exitCode, reader, err := container.Exec(
		ctx,
		versionCmd,
		tcexec.Multiplexed(),
	)
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, 0, exitCode, "apiki version should succeed")
	// The version output should contain the snapshot version (without the 'v'
	// prefix)
	expectedVersion := strings.TrimPrefix(snapshotVersion, "v")
	assert.Contains(
		t,
		string(output),
		expectedVersion,
		"apiki version output should contain expected version %q, got: %s",
		expectedVersion,
		string(output),
	)
}

func testFailureInstall(
	t *testing.T,
	ctx context.Context,
	container testcontainers.Container,
	baseURL string,
	envVars map[string]string,
	downloader string,
	expectedError string,
) {
	t.Helper()
	installScriptURL := fmt.Sprintf(
		"%s/%s/install.sh",
		baseURL,
		snapshotVersion,
	)
	var cmd []string

	if downloader == "curl" {
		cmd = []string{
			"sh",
			"-c",
			fmt.Sprintf("curl -sSL %s 2>&1 | sh 2>&1", installScriptURL),
		}
	} else {
		cmd = []string{
			"sh",
			"-c",
			fmt.Sprintf("wget -qO- %s 2>&1 | sh 2>&1", installScriptURL),
		}
	}

	// Set version environment variables if not already set
	if _, ok := envVars["APIKI_INSTALL_BASE_URL"]; !ok {
		envVars["APIKI_INSTALL_BASE_URL"] = fmt.Sprintf(
			"%s/%s",
			baseURL,
			snapshotVersion,
		)
	}
	if _, ok := envVars["APIKI_VERSION"]; !ok {
		envVars["APIKI_VERSION"] = snapshotVersion
	}

	env := []string{}
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Execute and capture output
	exitCode, reader, err := container.Exec(
		ctx,
		cmd,
		tcexec.WithEnv(env),
		tcexec.Multiplexed(),
	)
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.NotEqual(t, 0, exitCode, "install script should fail")
	assert.Contains(
		t,
		string(output),
		expectedError,
		"error output should contain expected message, got: %s",
		string(output),
	)
}

func testShellProfile(
	t *testing.T,
	ctx context.Context,
	container testcontainers.Container,
	profilePath string,
) {
	t.Helper()
	// Check if profile was updated
	fullPath := fmt.Sprintf("/home/testuser/%s", profilePath)
	cmd := []string{"grep", "-q", "apiki", fullPath}
	exitCode, _, err := container.Exec(ctx, cmd)
	require.NoError(t, err)
	require.Equal(
		t,
		0,
		exitCode,
		"shell profile should contain apiki configuration",
	)
}

// goreleaserMetadata mirrors the structure of goreleaser's metadata.json
// output.
// See: github.com/goreleaser/goreleaser/v2/internal/pipe/metadata
type goreleaserMetadata struct {
	Version string `json:"version"`
}

func findSnapshotVersionForSetup() string {
	metadataPath := filepath.Join(distDir, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return ""
	}

	var meta goreleaserMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return ""
	}

	version := meta.Version
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}

func startTestContainer(
	t *testing.T,
	ctx context.Context,
	dockerfile string,
) testcontainers.Container {
	t.Helper()
	dockerfilePath := filepath.Join(
		"dockerfiles",
		fmt.Sprintf("Dockerfile.%s", dockerfile),
	)
	absAcceptanceDir, err := filepath.Abs(".")
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    absAcceptanceDir,
			Dockerfile: dockerfilePath,
		},
		Networks: []string{testNetwork.Name},
	}

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.NoError(t, err, "failed to start test container")

	return container
}
