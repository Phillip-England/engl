package shell

type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// AllowedCommands - Add new commands here
var AllowedCommands = []Command{
	{
		Name:        "tree",
		Description: "Display directory tree structure",
		Example:     "tree /path -L 2",
	},
	{
		Name:        "cat",
		Description: "Display file contents",
		Example:     "cat file.txt",
	},
	{
		Name:        "ls",
		Description: "List directory contents",
		Example:     "ls -la /path",
	},
	{
		Name:        "pwd",
		Description: "Print working directory",
		Example:     "pwd",
	},
}

// commandAllowed checks if a command is in the allowed list
func commandAllowed(name string) bool {
	for _, cmd := range AllowedCommands {
		if cmd.Name == name {
			return true
		}
	}
	return false
}
