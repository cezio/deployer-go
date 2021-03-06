package deployer

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/viper"
)

// RequestMethod method that can be used for deployment
type RequestMethod string

const (
	// RequestGET get method
	RequestGET RequestMethod = "GET"
	// RequestPOST post method
	RequestPOST RequestMethod = "POST"
	// MaxSecretSize max size of secret body
	MaxSecretSize int = 64
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
	// SetupError cannot prepare execution environment
	SetupError
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

// IsExecutionError returns true if error is about execution problem
func (c *ConfigError) IsExecutionError() bool {
	return c.ErrorType == ExecutionError
}

// IsPreconditionsError returns true if secret is invalid or mehtod is invalid
func (c *ConfigError) IsPreconditionsError() bool {
	return c.ErrorType == PreconditionsError
}

// IsSetupError returns true if preparation of execution has failed
func (c *ConfigError) IsSetupError() bool {
	return c.ErrorType == SetupError
}

func (c *ConfigError) Error() *string {
	return c.Message
}

func runConfig(deploymentPath string, r *http.Request) *ConfigError {
	// configurationBase from http.go
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
	var method = r.Method
	var secret, serr = readSecret(conf, r)
	// non-empty error
	if serr != nil {
		var msg = serr.Error()
		return &ConfigError{ReadError, msg}

	}
	// check if the secret is correct, if it's provided
	if secret != nil {
		check := conf.Check(method, *secret)
		if check != nil {
			return check
		}
	}
	exerr := conf.Run()
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

// DeploymentConfig Deployment configuration structure
type DeploymentConfig struct {
	// DirName directory where config file is
	DirName string
	// RunDirName name of a directory where the command should be executed, default: DirName
	RunDirName string
	// ConfigName of config file
	ConfigName string
	// Commands to execute (or one command with cli arguments split into list)
	Commands []string
	// LogFile file to write output
	LogFile string
	// Env overrides
	Env []string
	// Ref conf parser instance
	RefCfg *(viper.Viper)
	// RefCmd subcommand reference
	RefCmd *(exec.Cmd)
	// PostOnly makrs that this config should be called from POST request only
	AllowedMethods []RequestMethod
	// Secret Required secret in body (if SecretHeader is omited)
	Secret *string
	// Secret stored in this header instead of body
	SecretHeader *string
}

// NewConfigFromEnv Create new Config from env variable
func NewConfigFromEnv() (*DeploymentConfig, error) {
	confDir := os.Getenv("DEPLOYER_CONFIG")
	return NewConfig(confDir)
}

// NewConfig Create new Config with dir set to dir
func NewConfig(dir string) (*DeploymentConfig, error) {
	if !IsDirectory(dir) {
		return nil, errors.New("Path is not a directory")
	}
	c := DeploymentConfig{}
	c.DirName = dir
	return &c, nil
}

// Read and parse config file from name
func (c *DeploymentConfig) Read(name string) error {
	c.ConfigName = strings.Join([]string{name, "conf"}, ".")
	log.Print("Reading " + c.ConfigName)
	c.RefCfg = viper.New()
	c.RefCfg.SetConfigType("toml")
	c.RefCfg.SetConfigName(c.ConfigName)
	c.RefCfg.AddConfigPath(c.DirName)
	verr := c.RefCfg.ReadInConfig()
	if verr != nil {
		log.Printf("Error when reading config %v: %v", c.ConfigName, verr)
		return verr
	}
	c.Commands = c.RefCfg.GetStringSlice("commands")
	c.Env = c.RefCfg.GetStringSlice("env")
	if c.RefCfg.IsSet("dir") {
		c.RunDirName = c.RefCfg.GetString("dir")
	} else {
		c.RunDirName = c.DirName
	}
	if c.RefCfg.InConfig("log-to") {
		c.LogFile = c.RefCfg.GetString("log-to")
	}
	if c.RefCfg.IsSet("secret") {
		secret := c.RefCfg.GetString("secret")
		c.Secret = &secret
	}
	if c.RefCfg.IsSet("secret-header") {
		secretHeader := c.RefCfg.GetString("secret-header")
		c.SecretHeader = &secretHeader
	}
	if c.RefCfg.IsSet("allowed-methods") {
		c.AllowedMethods = RequestMethodsFromStrings(c.RefCfg.GetStringSlice("allowed-methods"))
	}
	return nil
}

// AllowsMethod checks if given http method is allowed
func (c *DeploymentConfig) AllowsMethod(method string) bool {
	var methods = c.AllowedMethods
	if methods == nil {
		methods = RequestMethods
	}
	for _, m := range methods {
		if string(m) == method {
			return true
		}
	}
	return false
}

// AllowedSecret checks if provided secret is correct
func (c *DeploymentConfig) AllowedSecret(secret string) bool {
	var sec = c.Secret
	// null secret means no secret
	if sec == nil {
		return true
	}
	return *sec == secret
}

// Check runs pre-execution checks: allowed method check and secret check
func (c *DeploymentConfig) Check(method string, secret string) *ConfigError {
	if !c.AllowsMethod(method) {
		var mmsg = "Method not allowed"
		return &ConfigError{PreconditionsError, &mmsg}
	}
	if !c.AllowedSecret(secret) {
		var smsg = "Secret mismatched"
		return &ConfigError{PreconditionsError, &smsg}
	}
	return nil
}

// DoLock Ensures that only one call will be executed at one time
func (c *DeploymentConfig) DoLock() (*os.File, *ConfigError) {
	lockfPath := fmt.Sprintf("%v.lock", c.ConfigName)
	lockf, lockfErr := os.OpenFile(lockfPath, os.O_RDONLY|os.O_CREATE, 0755)
	if lockfErr != nil {
		lockfErrMsg := fmt.Sprintf("Cannot open lock file: %v: %v", lockfPath, lockfErr.Error())
		log.Printf(lockfErrMsg)
		return nil, &ConfigError{ExecutionError, &lockfErrMsg}
	}
	flockErr := syscall.Flock(int(lockf.Fd()), syscall.LOCK_EX)
	if flockErr != nil {
		flockfErrMsg := fmt.Sprintf("Cannot lock file: %v: %v", lockfPath, flockErr.Error())
		log.Printf(flockfErrMsg)
		return nil, &ConfigError{SetupError, &flockfErrMsg}
	}
	return lockf, nil
}

// Run runs command specified in configuration, returns nil
func (c *DeploymentConfig) Run() *ConfigError {
	lockf, lerr := c.DoLock()
	if lerr != nil {
		return lerr
	}
	defer syscall.Flock(int(lockf.Fd()), syscall.LOCK_UN)

	log.Print("Executing ", strings.Join(c.Commands, " "))
	cmd := exec.Command(c.Commands[0], c.Commands[1:]...)
	c.RefCmd = cmd
	// use config's dir as a workdir
	cmd.Dir = c.RunDirName
	// prepare env based on current + config
	cmd.Env = *(c.PrepareEnv())
	out, err := cmd.CombinedOutput()
	syscall.Flock(int(lockf.Fd()), syscall.LOCK_UN)
	outStr := string(out)
	if err != nil {
		log.Printf("Error during execution:\n %v \n %v", err, outStr)
		return &ConfigError{ExecutionError, &outStr}
	}
	log.Printf("Executed..:\n %v", outStr)

	return nil
}

// PrepareEnv returns an array of env entries: local env + entries from configuration
func (c *DeploymentConfig) PrepareEnv() *([]string) {
	env := make([]string, 0)
	copy(env, c.Env)
	env = append(env, c.Env...)
	return &env
}
