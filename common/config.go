package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Version     uint32
	DataDir     string
	LogDir      string
	RemotePeer  string
	Listener    string
	LogLevel    string
}

var (
	defaultDataDir = "./data"
	defaultLogDir  = "./log"
)

func defaultConfig() Config {
	cfg := Config{
		Version: 1,
		DataDir: defaultDataDir,
		LogDir: defaultLogDir,
	}
	return cfg
}

func LoadConfig(filepath string) Config {
	defaultCfg := defaultConfig()
	fileCfg := Config{}
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	cfgData, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(cfgData, &fileCfg); err != nil {
		panic(err)
	}

	//怎样用配置文件里的内容覆盖默认配置
	fileCfg = mergeConfig(defaultCfg, fileCfg)


	return fileCfg

}

//合并两个配置
func mergeConfig(defaultCfg , fileCfg Config) Config {
	return fileCfg
}