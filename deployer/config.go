package deployer

import (
    "log"
    "errors"
    "os"
    "os/exec"
    "strings"
    // "bytes"
    //"fmt"
	//"github.com/spf13/pflag"
    "github.com/spf13/viper"
    )


type ConfigErrorType int

const (
    MissingConfig ConfigErrorType = iota
    ReadError
    ExecutionError
)

type ConfigError struct {
    // error type
    ErrorType ConfigErrorType
    // message with error
    Message string
}


func (c *ConfigError) IsMissingConfig() (bool) {
    return c.ErrorType == MissingConfig;
}

func (c *ConfigError) IsReadError() (bool) {
    return c.ErrorType == ReadError;
}

func (c *ConfigError) IsExecutionError() (bool) {
    return c.ErrorType == ExecutionError;
}

func (c *ConfigError) Error() string {
    return c.Message;

}


func run_config(deployment_path string) (*ConfigError){
    conf, err := NewConfigFromEnv();
    if (err != nil){
        return &ConfigError{MissingConfig, "Path not found"};
    }
    cerr := conf.Read(deployment_path);
    if (cerr != nil){
        return &ConfigError{ReadError, cerr.Error()};
        }
    conf.Run();
    return nil;
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
func NewConfigFromEnv() (*Config, error){
    conf_dir := os.Getenv("DEPLOYER_CONFIG");
    if (conf_dir == ""){
        return nil, errors.New("no config dir");
    }
    c := NewConfig(conf_dir);
    return c, nil;
}


/*
    Create new Config with dir set to dir
*/
func NewConfig(dir string) (*Config) {
    c := Config{};
    c.Dir = dir;
    return &c;
}


/*
    Read and parse config file from name
*/
func (c *Config) Read(name string) (error) {
    c.Name = strings.Join([]string{name, "conf"}, ".");
    log.Print("Reading " + c.Name);
    c.Ref = viper.New();
    c.Ref.SetConfigType("toml");
    c.Ref.SetConfigFile(c.Name);
    c.Ref.AddConfigPath(c.Dir);
    verr := c.Ref.ReadInConfig();
    if (verr != nil){
        return verr;
    }
    c.Commands = c.Ref.GetStringSlice("commands");
    c.Env = c.Ref.GetStringSlice("env");
    if (c.Ref.InConfig("log-to")){
        c.LogFile = c.Ref.GetString("log-to");
    }
    return nil;
}


func (c *Config) Run() (error) {

    log.Print("Executing ", strings.Join(c.Commands, " "));
    cmd := exec.Command(c.Commands[0], c.Commands[1:]...);

    c.RefCmd = cmd;
    cmd.Env = c.Env;

    out, err := cmd.CombinedOutput();
    if (err != nil){
        log.Printf("Error during execution: %v %v", err, string(out));
    } else {
        log.Printf("Started..\n %v", string(out));
    }
    return nil;
}
