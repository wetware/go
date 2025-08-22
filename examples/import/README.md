# File Descriptor Passing Examples

Examples demonstrating file descriptor passing capabilities of `ww run`.

## Building

```bash
go build -o /tmp/fd-demo .
```

Binary is placed in `/tmp` to avoid polluting the git workspace.

## API

Uses `name=fdnum` format for file descriptor passing:

```bash
ww run --with-fd <name>=<fdnum> <binary>
```

## Examples

### Single File Descriptor

```bash
# Create test file
echo 'Hello World' > test.txt

# Pass as fd 3, named "input"
ww run --with-fd input=3 3<test.txt /tmp/fd-demo
```

### Multiple File Descriptors

```bash
# Create test files
echo 'input data' > input.txt
touch output.txt

# Pass both files
ww run --with-fd input=3 --with-fd output=4 3<input.txt 4>output.txt /tmp/fd-demo
```

### Database and Cache

```bash
# Create mock files
echo 'db data' > db.txt
echo 'cache data' > cache.txt

# Pass to child process
ww run --with-fd db=3 --with-fd cache=4 3<db.txt 4<cache.txt /tmp/fd-demo
```

## Environment Variables

Child process receives environment variables for each named file descriptor:

```bash
# Target FDs auto-assigned starting at 3
WW_FD_INPUT=3
WW_FD_OUTPUT=4
WW_FD_DB=5
WW_FD_CACHE=6
```

## Implementation

- **Syntax**: `--with-fd name=fdnum`
- **Target Assignment**: Auto-assigned starting at fd 3
- **Environment**: `WW_FD_<NAME>=<target_fd>` variables
- **Validation**: Prevents duplicate names, validates FD numbers

## Limitations

Current implementation does not include:
- Access mode specifications
- File type specifications
- S-expression configuration files
- Systemd socket activation
- Custom target FD assignment
- Symlink creation

## Testing

```bash
# Show usage
/tmp/fd-demo

# Test with fd passing
echo 'test data' > test.txt
ww run --with-fd input=3 3<test.txt /tmp/fd-demo
```

Child process displays environment variables and attempts to access passed file descriptors.
