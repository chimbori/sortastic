package conf

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var AppName = "Sortastic"

var Config AppConfig

type AppConfigWeb struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type AppConfigDirectory struct {
	Slug        string `yaml:"slug"`
	Source      string `yaml:"source"`
	Mode        string `yaml:"mode"`
	Destination string `yaml:"destination"`
	Trash       string `yaml:"trash"`
}

type AppConfig struct {
	Web         AppConfigWeb         `yaml:"web"`
	Directories []AppConfigDirectory `yaml:"directories"`
}

func ReadConfig() AppConfig {
	buf, err := ioutil.ReadFile("sortastic.yml")
	if err != nil {
		log.Fatal(err)
	}

	c := &AppConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		log.Fatal(err)
	}

	if c.Web.Host == "" {
		c.Web.Host = fmt.Sprintf("%s", GetOutboundIP())
	}

	for i := range c.Directories {
		if c.Directories[i].Source != "" {
			c.Directories[i].Source, _ = filepath.Abs(c.Directories[i].Source)
		}
		if c.Directories[i].Trash != "" {
			c.Directories[i].Trash, _ = filepath.Abs(c.Directories[i].Trash)
		}
	}

	// json, _ := json.MarshalIndent(*c, "", "\t")
	// fmt.Print(string(json))

	return *c
}
