package commands

import "fmt"

var Registry = map[string]Command{}

func register(command Command) {
	_, ok := Registry[command.Name]
	if ok {
		panic(fmt.Sprintf("Command %s already registered", command.Name))
	}
	if command.Handler == nil {
		panic(fmt.Sprintf("Command %s has no handler", command.Name))
	}
	Registry[command.Name] = command
}
