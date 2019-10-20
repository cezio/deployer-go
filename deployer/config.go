package deployer

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strings"

	// "bytes"
	//"fmt"
	//"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// RequestMethod method that can be used for deployment
type RequestMethod string

const (
	// RequestGET get method
	RequestGET RequestMethod = "GET"
	// RequestPOST post method
	RequestPOST RequestMethod = "POST"
)

// RequestMethods contains list of allowed methods
var RequestMethods = []RequestMethod{RequestGET, RequestPOST}

// RequestMethodsFromStrings converts array of strings to array of request methods
func RequestMethodsFromStrings(sIn []string) []RequestMethod {
	out := make([]RequestMethod, 0)
	for _, sva := range sIn {
		for _, meth := range RequestMethods {
			if sva == string(meth) {
				out = append(out, meth)
			}
		}
	}
	return out
}

// ConfigErrorType type of error
type ConfigErrorType int

const (
	// MissingConfig no configuration file
	MissingConfig ConfigErrorType = iota
	// ReadError can't read config
	ReadError
	// PreconditionsError invalid method or invalid secret
	PreconditionsError
	// ExecutionError execution failed
	ExecutionError
)

// ConfigError keeps information on runtime config errors
type ConfigError struct {
	// ErrorType error type
	ErrorType ConfigErrorType
	// Message with error
	Message *string
}

// IsMissingConfig returns true if error is about missing config file
func (c *ConfigError) IsMissingConfig() bool {
	return c.ErrorType == MissingConfig
}

// IsReadError returns true if error is about unreadable config file
func (c *ConfigError) IsReadError() bool {
	return c.ErrorType == ReadError
}

// IsExecutionError returns true if error is about execution problem */
func (c *ConfigError) IsExecutionError() bool {
	return c.ErrorType == ExecutionError
}

func (c *ConfigError) Error() *string {
	return c.Message

}

func runConfig(deploymentPath string, method string) *ConfigError {
	//    conf, err := NewConfigFromEnv();
	conf, err := NewConfig(configurationBase)
	if err != nil {
		var pathNotFoundMessage = "Path not found"
		return &ConfigError{MissingConfig, &pathNotFoundMessage}
	}
	cerr := conf.Read(deploymentPath)
	if cerr != nil {
		var msg = cerr.Error()
		return &ConfigError{ReadError, &msg}
	}
	exerr := conf.Run(method)
	if exerr != nil {
		return exerr
	}
	return nil
}

// RuntimeConfig keeps runtime configuration
type RuntimeConfig struct {
	// port to listen on
	Port int
	// directory with configuration
	Dir string
}

// Config Deployment configuration structure
type Config struct {
	// Dir directory where config file is
	Dir string
	// Name of config file
	Name string
	// Commands to execute (or one command with cli arguments split into list)
	Commands []string
	// LogFile file to write output
	LogFile string
	// Env overrides
	Env []string
	// Ref conf parser instance
	Ref *(viper.Viper)
	// RefCmd subcommand reference
	RefCmd *(exec.Cmd)
	// PostOnly makrs that this config should be called from POST request only
	AllowedMethods []RequestMethod
	// Secret Required secret in body
	Secret string
}

// NewConfigFromEnv Create new Config from env variable
func NewConfigFromEnv() (*Config, error) {
	confDir := os.Getenv("DEPLOYER_CONFIG")
	return NewConfig(confDir)
}

// NewConfig Create new Config with dir set to dir
func NewConfig(dir string) (*Config, error) {
	if !IsDirectory(dir) {
		return nil, errors.New("Path is not a directory")
	}
	c := Config{}
	c.Dir = dir
	return &c, nil
}

// Read and parse config file from name
func (c *Config) Read(name string) error {
	c.Name = strings.Join([]string{name, "conf"}, ".")
	log.Print("Reading " + c.Name)
	c.Ref = viper.New()
	c.Ref.SetConfigType("toml")
	c.Ref.SetConfigFile(c.Name)
	c.Ref.AddConfigPath(c.Dir)
	verr := c.Ref.ReadInConfig()
	if verr != nil {
		return verr
	}
	c.Commands = c.Ref.GetStringSlice("commands")
	c.Env = c.Ref.GetStringSlice("env")
	if c.Ref.InConfig("log-to") {
		c.LogFile = c.Ref.GetString("log-to")
	}
	c.Secret = c.Ref.GetString("secret")
	c.AllowedMethods = RequestMethodsFromStrings(c.Ref.GetStringSlice("allowed-methods"))
	return nil
}

// AllowsMethod checks if given http method is allowed
func (c *Config) AllowsMethod(method string) bool {
	var methods = c.AllowedMethods
	if methods == nil {
		methods = make([]RequestMethod, 0)
	}
	for _, m := range methods {
		if string(m) == method {
			return true
		}
	}
	return false

}

// Run runs command specified in configuration, returns nil
func (c *Config) Run(method string) *ConfigError {

	if !c.AllowsMethod(method) {
		var msg = "Method not allowed"
		return &ConfigError{PreconditionsError, &msg}
	}

	log.Print("Executing ", strings.Join(c.Commands, " "))
	cmd := exec.Command(c.Commands[0], c.Commands[1:]...)

	c.RefCmd = cmd
	cmd.Env = c.Env

	out, err := cmd.CombinedOutput()
	var outStr = string(out)
	if err != nil {
		log.Printf("Error during execution: %v %v", err, outStr)
		return &ConfigError{ExecutionError, &outStr}

	}
	log.Printf("Started..\n %v", outStr)
	return nil

}
