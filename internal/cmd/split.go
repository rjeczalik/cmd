package cmd

import "strings"

// Split splits command line string into command name and command line arguments,
// as expected by the exec.Command function.
func Split(command string) (string, []string) {
	var cmd string
	var args []string
	var i = -1
	var quote rune
	var push = func(n int) {
		if i == -1 {
			return
		}
		if offset := strings.IndexAny(string(command[n-1]), `"'`) ^ -1; cmd == "" {
			cmd = command[i : n+offset]
		} else {
			args = append(args, command[i:n+offset])
		}
	}
	for j, r := range command {
		switch r {
		case '"', '\'', '\\':
			switch quote {
			case 0:
				quote = r
			case '\\', r:
				quote = 0
			}
		case ' ':
			switch quote {
			case 0:
				push(j)
				i = -1
			case '\\':
				quote = 0
			}
		default:
			if i == -1 {
				i = j
			}
		}
	}
	push(len(command))
	return cmd, args
}
