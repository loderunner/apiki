---
name: CLI and TUI testing
overview: 'Introduce testing for cobra commands and bubbletea TUI at two levels: (1) integration-style tests with side effects captured/isolated, and (2) I/O-focused tests that assert correct file writes and shell command output.'
todos:
  - id: export-fs-setters
    content: Export UseFs in entries and config packages for test filesystem injection
    status: pending
  - id: ctx-keychain-injection
    content: Define Keychain interface in keychain package, add context injection, refactor to delegate through it
    status: pending
  - id: ctx-prompt-injection
    content: Define Prompter interface in prompt package, add context injection, refactor to delegate through it
    status: pending
  - id: thread-ctx-commands
    content: Add ctx parameter to command Run functions and Unlock; pass cmd.Context() from main.go
    status: pending
  - id: dotenv-startdir
    content: Parameterize LoadDotEnvEntries with startDir instead of calling os.Getwd internally
    status: pending
  - id: pure-model-tests
    content: Write model_test.go with state-machine tests for all TUI modes and transitions
    status: pending
  - id: teatest-integration
    content: Add teatest dependency and write tui_test.go with full interactive TUI tests
    status: pending
  - id: cli-command-tests
    content: Write tests for encrypt, decrypt, rotate, restore, and unlock commands with mock context
    status: pending
  - id: shell-output-tests
    content: Write shell_test.go testing generateShellCommands with various entry/env combinations
    status: pending
isProject: false
---

# CLI and TUI Testing Plan

## Approach

Two different injection strategies for two different situations:

- **TUI Model** -- needs `afero.Fs` for `persistEntries`/`persistSelection`, but never touches keychain or prompt. The `entries` and `config` packages already use a package-level `var fs`. Export a `UseFs` setter; tests swap it globally. No DI on the model at all.
- **Cobra commands** -- need keychain and prompt mocks. Cobra has built-in `cmd.Context()` / `cmd.SetContext(ctx)`. Define interfaces in the `keychain` and `prompt` packages with context-based injection. Command `Run` functions accept `ctx context.Context` as first param, threaded from `cmd.Context()`. Production code passes the default context (no setup needed). Tests build a context with mocks.
- **Environment variables** -- tests use `t.Setenv`. No abstraction.

---

## Phase 1: Refactors for Testability

### 1a. Export filesystem setters from `entries` and `config`

Both packages already have `var fs = afero.NewOsFs()` and swap it in their own `TestMain`. Export a setter so test code in other packages can also swap it.

In [internal/entries/entries.go](internal/entries/entries.go) and [internal/config/config.go](internal/config/config.go), add:

```go
func UseFs(f afero.Fs) { fs = f }
```

Since these are `internal/` packages, only code within the module can call this -- no API pollution concern.

### 1b. Keychain interface and context injection

In [internal/keychain/keychain.go](internal/keychain/keychain.go), define the interface, a real implementation, and context helpers:

```go
type Keychain interface {
    Store(key []byte) error
    Retrieve() ([]byte, error)
    Delete() error
}

type osKeychain struct{}
// osKeychain implements Keychain using go-keyring (existing code)

type contextKey struct{}

func WithKeychain(ctx context.Context, kc Keychain) context.Context {
    return context.WithValue(ctx, contextKey{}, kc)
}

func fromContext(ctx context.Context) Keychain {
    if kc, ok := ctx.Value(contextKey{}).(Keychain); ok {
        return kc
    }
    return osKeychain{}
}
```

Package-level functions become context-aware delegates:

```go
func Store(ctx context.Context, key []byte) error  { return fromContext(ctx).Store(key) }
func Retrieve(ctx context.Context) ([]byte, error) { return fromContext(ctx).Retrieve() }
func Delete(ctx context.Context) error              { return fromContext(ctx).Delete() }
```

### 1c. Prompter interface and context injection

Same pattern in [internal/prompt/prompt.go](internal/prompt/prompt.go):

```go
type Prompter interface {
    ReadPassword(prompt string) (string, error)
    ReadChoice(prompt string, choices map[rune]string) (string, error)
    ReadChoiceWithDefault(prompt string, choices map[rune]string, defaultValue string) (string, error)
}

type terminalPrompter struct{}
// terminalPrompter implements Prompter using os.Stdin/term (existing code)
```

Context helpers follow the same `WithPrompter` / `fromContext` pattern. Package-level functions become `ReadPassword(ctx, ...)`, `ReadChoice(ctx, ...)`, etc.

### 1d. Thread context through commands

All command `Run` functions and `Unlock` gain `ctx context.Context` as first parameter:

- `commands.Unlock(ctx, file)` -- uses `prompt.ReadPassword(ctx, ...)` and `keychain.Retrieve(ctx)`
- `encrypt.Run(ctx, path)`, `decrypt.Run(ctx, path)`, `rotate.Run(ctx, path)`, `restore.Run(ctx, variablesPath, configPath)`
- `apiki.Run(ctx, variablesPath, configPath)`

In [main.go](main.go), each cobra `RunE` passes `cmd.Context()` -- which is `context.Background()` by default. **No mock wiring in production code.**

### 1e. Parameterize `.env` discovery

[commands/apiki/dotenv.go](commands/apiki/dotenv.go): `LoadDotEnvEntries` currently calls `os.Getwd()` internally. Change it to accept `startDir` as a parameter. The caller in `Run()` passes `os.Getwd()`, tests pass a temp directory. `FindDotEnvFiles` already takes `startDir`, so only `LoadDotEnvEntries` needs the change.

---

## Phase 2: Pure TUI Model Tests

**File: `commands/apiki/model_test.go`**

Test the model as a state machine -- construct a `Model`, send messages, assert state transitions and view output. `persistEntries`/`persistSelection` write to the in-memory filesystem swapped via `UseFs`.

Test cases:

- **List navigation**: up/down/wrap-around, viewport scrolling
- **Selection**: space toggles, radio-button deselection of same-name entries
- **Add entry**: `+` enters add mode, form fields, Enter saves, Esc cancels
- **Edit entry**: `=` enters edit mode, form pre-filled, Enter saves
- **Delete entry**: `-` shows confirmation, `y` deletes, `n` cancels
- **Promote .env entry**: `=` on a `.env` entry shows promote dialog
- **Filter**: `/` enters filter mode, typing filters, Esc clears
- **Import**: `i` loads env entries, space selects, Enter confirms import
- **Quit**: `q` sets `cancelled`, Enter calls `persistSelection` then `quitting`
- **View output**: assert rendered strings contain expected elements (title, entries, checkboxes, help bar)

Setup pattern:

```go
func TestMain(m *testing.M) {
    entries.UseFs(afero.NewMemMapFs())
    config.UseFs(afero.NewMemMapFs())
    os.Exit(m.Run())
}

func newTestModel(t *testing.T, testEntries []Entry) Model {
    t.Helper()
    file := &entries.File{Entries: /* ... */}
    return NewModel(file, "/tmp/test/variables.json", "/tmp/test/config.json", nil, testEntries)
}
```

No context, no DI on the model -- `entries.Save` and `config.Save` use the swapped package-level fs automatically.

---

## Phase 3: TUI Integration Tests with `teatest`

**File: `commands/apiki/tui_test.go`**

Add `github.com/charmbracelet/x/exp/teatest` as a dependency. Use it for full interactive TUI tests. Same `TestMain` fs swap as Phase 2 (shared test file).

```go
func TestTUISelectAndApply(t *testing.T) {
    model := newTestModel(t, sampleEntries())
    tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))

    // Select first entry
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
    // Apply and quit
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    fm := tm.FinalModel(t)
    m := fm.(Model)
    require.True(t, m.Quitting())
    require.True(t, m.Entries()[0].Selected)
}
```

Test scenarios:

- Select entries and quit -- assert config file written, assert `generateShellCommands` output
- Add entry via form -- assert variables file updated
- Delete entry -- assert variables file updated
- Filter and select -- assert correct entry selected
- Ctrl-C -- assert cancelled, no file writes
- Assert intermediate output with `teatest.WaitFor`

Add `lipgloss.SetColorProfile(termenv.Ascii)` in test `init()` to normalize output across environments.

---

## Phase 4: CLI Command Tests

One test file per command package. Each test builds a context with the mocks it needs, and swaps the fs via `UseFs`.

### `commands/encrypt/encrypt_test.go`

```go
func TestMain(m *testing.M) {
    entries.UseFs(afero.NewMemMapFs())
    os.Exit(m.Run())
}

func TestEncryptPasswordMode(t *testing.T) {
    ctx := context.Background()
    ctx = prompt.WithPrompter(ctx, &mockPrompter{
        passwords: []string{"secret", "secret"},
        choices:   []string{"password"},
    })
    ctx = keychain.WithKeychain(ctx, &mockKeychain{})

    seedVariablesFile(t, "/tmp/vars.json", testEntries())

    err := encrypt.Run(ctx, "/tmp/vars.json")
    require.NoError(t, err)

    file, err := entries.Load("/tmp/vars.json")
    require.NoError(t, err)
    require.True(t, file.Encrypted())
    require.Equal(t, "password", file.Encryption.Mode)
}
```

Same pattern for decrypt, rotate, restore commands.

### `commands/unlock_test.go`

Test `Unlock` with mock prompter and keychain on context (password attempts, wrong password, keychain mode).

---

## Phase 5: Shell Command Output Tests

Test `generateShellCommands` in [commands/apiki/main.go](commands/apiki/main.go) directly -- it's a pure function that takes entries and an env snapshot, returning a string of `export`/`unset` statements.

**File: `commands/apiki/shell_test.go`**

```go
func TestGenerateShellCommands(t *testing.T) {
    es := []Entry{
        {Entry: entries.Entry{Name: "FOO", Value: "bar"}, Selected: true},
        {Entry: entries.Entry{Name: "BAZ", Value: "qux"}, Selected: false},
    }
    envSnapshot := map[string]string{"FOO": "old", "BAZ": "original"}

    output := generateShellCommands(es, envSnapshot)
    require.Contains(t, output, "export FOO='bar'")
    require.Contains(t, output, "unset BAZ")
}
```

Test edge cases: single-quote escaping, no-op when value unchanged, empty selection.

---

## Review Order

Since this plan touches many files, review in this order:

1. `internal/entries/entries.go`, `internal/config/config.go` -- `UseFs` export
2. `internal/keychain/keychain.go` -- interface + context injection
3. `internal/prompt/prompt.go` -- interface + context injection
4. `commands/unlock.go` -- add ctx parameter
5. `commands/encrypt/encrypt.go` (and decrypt, rotate, restore) -- add ctx parameter
6. `main.go` -- pass `cmd.Context()` at call sites
7. `commands/apiki/dotenv.go` -- parameterized `LoadDotEnvEntries`
8. `commands/apiki/main.go` -- thread ctx through `Run`
9. Test files: `model_test.go`, `tui_test.go`, `shell_test.go`, command tests
