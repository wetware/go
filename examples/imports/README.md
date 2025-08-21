# File Descriptor Passing Examples

This directory contains examples demonstrating the file descriptor passing capabilities of `ww run`.

## Building

Build the demo program:

```bash
go build -o /tmp/fd-demo .
```

The binary will be placed in `/tmp` to avoid polluting the git workspace.

## Examples

### 1. Lisp-First Capability Tables (Recommended)

The preferred way to work with capabilities is using Lisp S-expressions. This provides a clean, declarative syntax and enables programmatic validation:

```bash
# Create capability imports file
cat >imports <<'SEXP'
(
  (fd :name "stdin"  :fd 0 :mode "r" :target 0)
  (fd :name "stdout" :fd 1 :mode "w" :target 1)
  (fd :name "stderr" :fd 2 :mode "w" :target 2)
  (fd :name "db"     :fd 3 :mode "rw" :type "socket" :target 10)
  (fd :name "cache"  :fd 4 :mode "rw" :type "file" :target 11)
)
SEXP

# Use the capability imports
ww run --fd-from @imports 3<test.txt 4>output.txt /tmp/fd-demo
```

**Benefits of Lisp approach:**
- **Declarative**: Clear, readable capability definitions
- **Programmatic**: Can include validation logic and computed values
- **Reusable**: Capability tables can be shared and versioned
- **Extensible**: Easy to add new capability types and metadata

### 2. Advanced Lisp Capabilities

You can even embed Lisp functions for dynamic validation:

```bash
# Create capability imports with validation
cat >validated-imports <<'SEXP'
(
  (def validate-fd-access (name fd mode type)
    (cond
      ((= mode "rw") (and (>= fd 3) (or (= type "file") (= type "socket"))))
      ((= mode "r") (>= fd 0))
      ((= mode "w") (>= fd 1))
      (true false)))
  
  (fd :name "db" :fd 3 :mode "rw" :type "socket" :target 10)
  (fd :name "logs" :fd 4 :mode "w" :type "file" :target 11)
)
SEXP

ww run --fd-from @validated-imports 3<>db.sock 4>app.log /tmp/fd-demo
```

### 3. Alternative: Command-Line Arguments

For simple cases or scripting, you can use command-line arguments:

```bash
# Basic single capability
ww run --fd input=3,mode=r,type=file 3<test.txt /tmp/fd-demo

# Multiple capabilities
ww run \
  --fd db=3,mode=rw,type=socket,target=10 \
  --fd logs=4,mode=w,type=file,target=11 \
  3<>db.sock 4>app.log \
  /tmp/fd-demo
```

### 4. Systemd Integration

```bash
# Set up systemd environment variables
export LISTEN_FDS=2
export LISTEN_PID=$$

# Run with systemd fds
ww run \
  --use-systemd-fds=listen \
  --fd-map listen0=10 \
  --fd-map listen1=11 \
  /tmp/fd-demo
```

### 5. Inherit Existing File Descriptors

```bash
# Open a file
exec 3<test.txt

# Inherit the fd
ww run --fdctl inherit:3 /tmp/fd-demo
```

## Environment Variables in Child Process

When using fd passing, the child process receives these environment variables:

```bash
# List of all fds with metadata
WW_FDS="((db (fd 10) (mode \"rw\") (type \"socket\")) (logs (fd 11) (mode \"w\") (type \"file\")))"

# Individual fd variables
WW_FD_DB=10
WW_FD_LOGS=11
```

## Security Features

- **Least privilege**: Only specified file descriptor capabilities are available
- **Explicit grants**: No implicit access to parent resources
- **Validation**: Access modes and types are enforced
- **Isolation**: Child processes are jailed and cannot access unauthorized resources

## Testing

You can test the functionality by running the examples above. Make sure to:

1. Create the necessary test files and sockets
2. Use valid file descriptor numbers
3. Check the environment variables in the child process
4. Verify that only specified fds are accessible
