package api

type Definition struct {
	Nodes map[string]Node `yaml:"nodes"`
}

type Node struct {
	Uses      string                 `yaml:"uses" json:"uses"`
	DependsOn []string               `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Depends   []string               `yaml:"depends,omitempty" json:"depends,omitempty"`
	With      map[string]interface{} `yaml:"with,omitempty" json:"with,omitempty"`
	Inputs    map[string]interface{} `yaml:"inputs,omitempty" json:"inputs,omitempty"`
}
