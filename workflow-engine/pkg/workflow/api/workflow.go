package api

type Definition struct {
	Nodes map[string]Node `yaml:"nodes"`
}

type Node struct {
	Uses      string                 `yaml:"uses" json:"uses"`
	DependsOn []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	With      map[string]interface{} `yaml:"with,omitempty" json:"with,omitempty"`
}
