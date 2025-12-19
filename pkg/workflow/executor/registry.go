package executor

import "fmt"

var executors = map[string]Executor{}

func Register(name string, executor Executor) {
	executors[name] = executor
}

func Get(name string) (Executor, error) {
	executor, ok := executors[name]
	if !ok {
		return nil, fmt.Errorf("executor not found: %s", name)
	}
	return executor, nil
}

func All() map[string]Executor {
	return executors
}
