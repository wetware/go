# Wetware Go

A Go implementation of the Wetware system for distributed computing and cellular automata.

## Getting Started

```bash
go install github.com/wetware/go/cmd/ww
```

Then try starting the shell:
```
ww run /ipfs/QmcCFV1Vure9qs2nSQaNH9WjQdhbN4NgoQA5FoDazjyMK9
```

## File Descriptor Passing

Wetware provides secure file descriptor passing capabilities through the `ww run` command, enabling least-authority semantics for child processes. File descriptors are explicitly granted to child processes, ensuring they only have access to the resources they need.

**We recommend using Lisp S-expressions for capability tables** as they provide a clean, declarative syntax and enable programmatic validation. Command-line arguments are available for simple cases and scripting.

### Overview

File descriptor passing in Wetware is capability-safe and follows the Principle of Least Authority (POLA) model. By default, all non-specified file descriptors are closed (except stdio unless overridden), ensuring that child processes cannot access resources they haven't been explicitly granted.

### CLI Flags

#### `--fd <name>[=<fdnum>][,<k>=<v>...]`

Maps an existing parent file descriptor to a name with optional configuration:

- **Required**: `name=fdnum` - Maps fd number to a named reference
- **Options**:
  - `mode=` - Access mode: `r` (read), `w` (write), `rw` (read-write)
  - `type=` - Type: `stream`, `file`, or `socket`
  - `target=` - Target fd number in child process (auto-assigned if not specified)
  - `move=` - Whether to move (close original) or duplicate: `true`/`false`
  - `cloexec=` - Set close-on-exec flag: `true`/`false` (default: `true`)
  - `pathlink=` - Create symlink in jail: `true`/`false`

**Examples:**
```bash
# Basic fd mapping
ww run --fd db=3 /ipfs/foo

# With options
ww run --fd db=3,mode=rw,type=socket,target=10 /ipfs/foo

# Multiple fds
ww run --fd db=3,mode=rw --fd cache=4,mode=r,type=file /ipfs/foo
```

#### `--fd-map <name>=<targetfd>`

Overrides the numeric target for a named file descriptor:

```bash
ww run --fd db=3 --fd-map db=10 /ipfs/foo
```

#### `--fdctl <unix-sock-path>[,<k>=<v>...]` or `inherit:<fdnum>`

Accepts file descriptors via SCM_RIGHTS messages or inherits existing ones:

```bash
# Inherit existing fd
ww run --fdctl inherit:5 /ipfs/foo

# Unix socket (future implementation)
ww run --fdctl /path/to/socket /ipfs/foo
```

#### `--use-systemd-fds[=<prefix>]`

Imports file descriptors from systemd socket activation:

```bash
ww run --use-systemd-fds=listen --fd-map listen0=10 /ipfs/foo
```

#### `--fd-from @<path>|-` (Recommended)

Bulk capability specification from file or stdin using Lisp S-expressions. This is the preferred approach for complex capability tables:

```bash
# From file
ww run --fd-from @imports /ipfs/foo

# From stdin
cat >caps.sexp <<'SEXP'
(
  (fd :name "stdin"  :fd 0 :mode "r" :target 0)
  (fd :name "stdout" :fd 1 :mode "w" :target 1)
  (fd :name "logs"   :fd 5 :mode "w" :type "file" :target 9)
)
SEXP
ww run --fd-from - /ipfs/foo
```

**Benefits of Lisp approach:**
- **Declarative**: Clear, readable capability definitions
- **Programmatic**: Can include validation logic and computed values
- **Reusable**: Capability tables can be shared and versioned
- **Extensible**: Easy to add new capability types and metadata

#### `--fd-verbose`

Enables verbose logging of file descriptor grants:

```bash
ww run --fd-verbose --fd db=3 /ipfs/foo
# Output: grant fd name=db num=10 type=file mode=rw move=false cloexec=true
```

### Child Process ABI

Each mapped file descriptor has a stable number in the child process (either specified via `target` or auto-assigned starting from 10).

#### Environment Variables

The following environment variables are set in the child process:

- **`WW_FDS`** - S-expression mapping of names to fd metadata:
  ```lisp
  (
    (db (fd 10) (mode "rw") (type "stream"))
    (cache (fd 11) (mode "rw") (type "file"))
  )
  ```

- **`WW_FD_<NAME>=<fdnum>`** - Individual variables for convenience:
  ```bash
  WW_FD_DB=10
  WW_FD_CACHE=11
  ```

#### Symlink Creation

When `pathlink=true` is specified, a symlink is created inside the jail directory pointing to `/proc/self/fd/<target>`. This allows the child process to access the file descriptor through a named path.

### Security Model

- **Default behavior**: All non-specified file descriptors are closed
- **No implicit authority**: `--with-all` does not grant additional file descriptors
- **Explicit grants**: File descriptors must be explicitly specified
- **CLOEXEC by default**: File descriptors are marked close-on-exec unless overridden
- **Validation**: Access modes and types are validated against actual capabilities

### Examples

#### Database Connection and Log File

```bash
# Pass database socket and log file
ww run \
  --fd db=3,mode=rw,type=socket,target=10 \
  --fd logs=5,mode=w,type=file,target=11 \
  /ipfs/foo
```

#### Systemd Socket Activation

```bash
# Use systemd-provided sockets
ww run \
  --use-systemd-fds=listen \
  --fd-map listen0=10 \
  --fd-map listen1=11 \
  /ipfs/foo
```

#### Bulk Configuration

```bash
# Create fd specification file
cat >fds.sexp <<'SEXP'
(
  (fd (name "stdin")  (fd 0) (mode "r") (target 0))
  (fd (name "stdout") (fd 1) (mode "w") (target 1))
  (fd (name "stderr") (fd 2) (mode "w") (target 2))
  (fd (name "db")     (fd 3) (mode "rw") (type "socket") (target 10))
  (fd (name "cache")  (fd 4) (mode "rw") (type "file") (target 11))
)
SEXP

# Use the specification
ww run --fd-from @fds.sexp /ipfs/foo
```

### Error Handling

Errors are surfaced with clear messages and non-zero exit codes:

- **Validation errors**: Invalid modes, types, or formats
- **Duplicate mappings**: Same name or target fd used multiple times
- **File access errors**: Cannot access specified file descriptors
- **Parse errors**: Invalid S-expression format in bulk specifications

The command fails fast with descriptive error messages to help users quickly identify and fix configuration issues.
