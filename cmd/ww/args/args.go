package args

const separator = "--"

const GuestArgs = "guestArgs"

// split host and guest arguments and remove the separator.
// The function will always return initialized slices.
func SplitArgs(args []string) ([]string, []string) {
	if len(args) == 0 {
		return []string{}, []string{}
	}

	separatorIndex := -1
	for i, arg := range args {
		if arg == separator {
			separatorIndex = i
			break
		}
	}

	if separatorIndex == -1 {
		return args, []string{}
	}

	hostArgs := args[:separatorIndex]
	guestArgs := args[separatorIndex+1:]
	return hostArgs, guestArgs
}
