package lifecycle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteDockerPrepareScriptRetainedCheckout(t *testing.T) {
	_, origin := setupPostludeRepo(t, "main")
	seedOriginBranch(t, origin, "task-branch")
	for _, profile := range []struct {
		name     string
		script   string
		selected bool
	}{
		{name: "blank profile"},
		{name: "stored default with rewritten URL and selected branch", script: defaultDockerPrepareScript, selected: true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			root := t.TempDir()
			workspace := filepath.Join(root, "workspace")
			hookLog := filepath.Join(root, "hook.log")
			trace := filepath.Join(root, "git.trace")
			metadata := map[string]interface{}{
				MetadataKeySetupScript:     profile.script,
				MetadataKeyRepoSetupScript: "echo setup >> " + shellQuote(hookLog),
				MetadataKeyBaseBranch:      "main",
				MetadataKeyWorktreeBranch:  "task-branch",
				"repository_clone_url":     "file://" + origin,
			}
			if profile.selected {
				const configuredURL = "https://example.invalid/repository.git"
				metadata["repository_clone_url"] = configuredURL
				metadata[metadataCheckoutBranch] = "task-branch"
				metadata[metadataCheckoutRef] = "refs/heads/task-branch"
				runIn(t, root, "git", "config", "--file", filepath.Join(root, ".gitconfig"),
					"url.file://"+origin+".insteadOf", configuredURL)
			}
			prepare := resolveDockerPrepareForTest(t, "remote_docker", workspace, metadata)
			runPrepare := func() {
				t.Helper()
				cmd := exec.Command("sh", "-c", prepare)
				cmd.Env = append(os.Environ(), "HOME="+root, "GIT_CONFIG_NOSYSTEM=1", "GIT_TRACE="+trace)
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", output)
			}

			runPrepare()
			require.Equal(t, "task-branch", strings.TrimSpace(string(runIn(t, workspace, "git", "branch", "--show-current"))))
			hooks, err := os.ReadFile(hookLog)
			require.NoError(t, err)
			require.Equal(t, "setup\n", string(hooks))

			runIn(t, workspace, "git", "-c", "user.name=Test", "-c", "user.email=test@example.com",
				"commit", "--allow-empty", "-m", "local only")
			head := runIn(t, workspace, "git", "rev-parse", "HEAD")
			marker := filepath.Join(workspace, "local.txt")
			require.NoError(t, os.WriteFile(marker, []byte("local edits"), 0o600))

			runPrepare()
			require.Equal(t, head, runIn(t, workspace, "git", "rev-parse", "HEAD"))
			require.Equal(t, "task-branch", strings.TrimSpace(string(runIn(t, workspace, "git", "branch", "--show-current"))))
			content, err := os.ReadFile(marker)
			require.NoError(t, err)
			require.Equal(t, "local edits", string(content))
			hooks, err = os.ReadFile(hookLog)
			require.NoError(t, err)
			require.Equal(t, "setup\nsetup\n", string(hooks))
			gitTrace, err := os.ReadFile(trace)
			require.NoError(t, err)
			require.Equal(t, 1, strings.Count(string(gitTrace), "built-in: git clone "))
		})
	}

	t.Run("rejects unrelated and invalid checkouts", func(t *testing.T) {
		testRemoteDockerPrepareRejectsOccupiedWorkspace(t, origin)
	})

	t.Run("keeps other script resolution", func(t *testing.T) {
		workspace := t.TempDir()
		for _, executorType := range []string{"remote_docker", "local_docker"} {
			const custom = defaultDockerPrepareScript + "\n# profile-specific edit\n"
			script := resolveDockerPrepareForTest(t, executorType, workspace, map[string]interface{}{
				MetadataKeySetupScript: custom,
				"repository_clone_url": origin,
			})
			require.Contains(t, script, "# profile-specific edit")
			require.NotContains(t, script, "normalize_repository_origin")
		}
		script := resolveDockerPrepareForTest(t, "local_docker", workspace, map[string]interface{}{
			"repository_clone_url": origin,
		})
		require.NotContains(t, script, "normalize_repository_origin")
	})
}

func resolveDockerPrepareForTest(t *testing.T, executorType, workspace string, metadata map[string]interface{}) string {
	t.Helper()
	script, err := resolveDockerPrepareScript(&ExecutorCreateRequest{Metadata: metadata}, executorType)
	require.NoError(t, err)
	return strings.ReplaceAll(script, shellQuote(dockerWorkspacePath), shellQuote(workspace))
}

func testRemoteDockerPrepareRejectsOccupiedWorkspace(t *testing.T, origin string) {
	unrelated, _ := setupPostludeRepo(t, "main")
	for _, workspace := range []string{unrelated, t.TempDir()} {
		marker := filepath.Join(workspace, "keep.txt")
		require.NoError(t, os.WriteFile(marker, []byte("keep"), 0o600))
		script := resolveDockerPrepareForTest(t, "remote_docker", workspace, map[string]interface{}{
			"repository_clone_url": "file://" + origin,
		})
		cmd := exec.Command("sh", "-c", script)
		cmd.Env = append(os.Environ(), "HOME="+t.TempDir(), "GIT_CONFIG_NOSYSTEM=1")
		output, err := cmd.CombinedOutput()
		require.Error(t, err, "%s", output)
		content, err := os.ReadFile(marker)
		require.NoError(t, err)
		require.Equal(t, "keep", string(content))
	}
}
