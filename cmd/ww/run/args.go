package run

const (
	argSeparator    = "--"
	indexExecutable = 0 // The executable file is the first argument.
)

// Args represents a set of arguments for the ww run command.
// Arguments are meant for the run command (host process) by default.
// Arguments after a `--` delimiter are meant for the guest process.
// E.g. ww run /ipfs/... -- guest-arg1 guest-arg2
// will pass guest-arg1 and guest-arg2 to the guest process.
type Args struct {
	Host  []string
	Guest []string
}

// ParseArgs will build an Args struct.
// It will always initialize values, ensuring Args.Host and Args.Guest are not nil.
func ParseArgs(args []string) *Args {
	if len(args) == 0 {
		return &Args{
			Host:  []string{},
			Guest: []string{},
		}
	}

	hostArgs := make([]string, 0, len(args))
	i := 0
	for ; i < len(args); i++ {
		arg := args[i]
		if arg == argSeparator {
			i++
			break
		}
		hostArgs = append(hostArgs, arg)
	}

	// No guest arguments provided.
	if i >= len(args) {
		return &Args{
			Host:  hostArgs,
			Guest: []string{},
		}
	}

	guestArgs := make([]string, 0, len(args)-i)
	for ; i < len(args); i++ {
		guestArgs = append(guestArgs, args[i])
	}

	return &Args{
		Host:  hostArgs,
		Guest: guestArgs,
	}
}
