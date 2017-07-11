package deployer

import (
    "log"
    "errors"
    "os"
    "os/exec"
    "strings"
	//"github.com/spf13/pflag"
    "github.com/spf13/viper"
    )



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
    c.Ref.SetConfigName(c.Name);
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
    c.Commands = append([]string{"-c"}, c.Commands...)
    cmd := exec.Command("/bin/bash", c.Commands...);
    c.RefCmd = cmd;
    cmd.Env = c.Env;
    cmd.Start();
    return nil;
}
