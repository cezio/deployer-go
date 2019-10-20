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

// ConfitErrorType type of error
type ConfigErrorType int

const (
	// no configuration file
	MissingConfig ConfigErrorType = iota
	// can't read config
	ReadError
	// execution failed
	ExecutionError
)

type ConfigError struct {
	// error type
	ErrorType ConfigErrorType
	// message with error
	Message *string
}

func (c *ConfigError) IsMissingConfig() bool {
	return c.ErrorType == MissingConfig
}

func (c *ConfigError) IsReadError() bool {
	return c.ErrorType == ReadError
}

func (c *ConfigError) IsExecutionError() bool {
	return c.ErrorType == ExecutionError
}

func (c *ConfigError) Error() *string {
	return c.Message

}

func run_config(deployment_path string) *ConfigError {
	//    conf, err := NewConfigFromEnv();
	conf, err := NewConfig(directory_path)
	if err != nil {
		var pathNotFoundMessage = "Path not found"
		return &ConfigError{MissingConfig, &pathNotFoundMessage}
	}
	cerr := conf.Read(deployment_path)
	if cerr != nil {
		var msg = cerr.Error()
		return &ConfigError{ReadError, &msg}
	}
	conf.Run()
	return nil
}

/**

 */
type RuntimeConfig struct {
	// port to listen on
	Port int
	// directory with configuration
	Dir string
}

/*
   Deployment configuration structure
*/
type Config struct {
	// directory where config file is
	Dir string
	// name of config file
	Name string
	// commands to execute
	Commands []string
	// log file to write output
	LogFile string
	// env overrides
	Env []string
	// conf parser instance
	Ref *(viper.Viper)
	// command ref
	RefCmd *(exec.Cmd)
}

/*
   Create new Config from env variable
*/
func NewConfigFromEnv() (*Config, error) {
	conf_dir := os.Getenv("DEPLOYER_CONFIG")
	return NewConfig(conf_dir)
}

/*
   Create new Config with dir set to dir
*/
func NewConfig(dir string) (*Config, error) {
	if !IsDirectory(dir) {
		return nil, errors.New("Path is not a directory")
	}
	c := Config{}
	c.Dir = dir
	return &c, nil
}

/*
   Read and parse config file from name
*/
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
	return nil
}

func (c *Config) Run() error {

	log.Print("Executing ", strings.Join(c.Commands, " "))
	cmd := exec.Command(c.Commands[0], c.Commands[1:]...)

	c.RefCmd = cmd
	cmd.Env = c.Env

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error during execution: %v %v", err, string(out))
	} else {
		log.Printf("Started..\n %v", string(out))
	}
	return nil
}
