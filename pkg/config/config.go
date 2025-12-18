package serviceconfig

// import (
// 	"fmt"
// 	"os"

// 	"gopkg.in/yaml.v2"
// )

// type configType struct {
// 	// Add your service-specific configuration here
// 	Service struct {
// 		// Example configuration fields
// 		// Host string `yaml:"host"`
// 		// Port string `yaml:"port"`
// 	} `yaml:"service"`
// }

// var config configType

// // Add getter functions for your configuration
// // Example:
// // func GetServiceHost() string {
// // 	if config.Service.Host == "" {
// // 		return "localhost"
// // 	}
// // 	return config.Service.Host
// // }

// // validateConfigPath just makes sure, that the path provided is a file,
// // that can be read
// func validateConfigPath(path string) error {
// 	s, err := os.Stat(path)
// 	if err != nil {
// 		return err
// 	}
// 	if s.IsDir() {
// 		return fmt.Errorf("'%s' is a directory, not a normal file", path)
// 	}
// 	return nil
// }

// // ParseConfig returns a new decoded Config struct
// func ParseConfig(configPath string) error {
// 	// validate config path before decoding
// 	if err := validateConfigPath(configPath); err != nil {
// 		return err
// 	}

// 	// Open config file
// 	file, err := os.Open(configPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()

// 	// Init new YAML decode
// 	d := yaml.NewDecoder(file)

// 	// Start YAML decoding from file
// 	if err := d.Decode(&config); err != nil {
// 		return err
// 	}

// 	return nil
// }
