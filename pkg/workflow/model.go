package workflow

type Definition struct {
	Nodes map[string]Node `yaml:"nodes"`
}

type Node struct {
	Uses      string                 `yaml:"uses"`
	DependsOn []string               `yaml:"depends_on,omitempty"`
	With      map[string]interface{} `yaml:"with,omitempty"`
}
