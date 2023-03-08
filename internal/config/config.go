package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Global struct {
	From             ConversionType `yaml:"from"`
	To               ConversionType `yaml:"to"`
	Input            string         `yaml:"inputDir"`
	Output           string         `yaml:"outputDir"`
	ConfluenceConfig `yaml:"config"`
}

type ConfluenceConfig struct {
	Url         string       `yaml:"url"`
	Username    string       `yaml:"username"`
	Password    string       `yaml:"password"`
	Pages       []string     `yaml:"pages"`
	ImportPages []ImportPage `yaml:"importPages"`
}

type ImportPage struct {
	Id       string `yaml:"id"`
	ParentId string `yaml:"parentId"`
	Title    string `yaml:"title"`
	Source   string `yaml:"source"`
}

func ReadConfig() (*Global, error) {
	configPath := os.Getenv("CONFIG_PATH")
	filename := "configs/config.yaml"
	if configPath != "" {
		filename = configPath
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("ERROR: File '%s' doesn't exist", filename)
		return nil, err
	}
	global := Global{}
	err = yaml.Unmarshal(file, &global)
	if err != nil {
		return nil, err
	}

	if global.From == Confluence || global.To == ConfluenceWiki {
		confluenceUrl := os.Getenv("CONFLUENCE_URL")
		confluenceUsername := os.Getenv("CONFLUENCE_USERNAME")
		confluencePassword := os.Getenv("CONFLUENCE_PASSWORD")
		if confluenceUsername != "" && confluencePassword != "" {
			log.Printf("DEBUG: use Confluence config from environment variables")
			global.Url = confluenceUrl
			global.Username = confluenceUsername
			global.Password = confluencePassword
		} else {
			log.Printf("DEBUG: use Confluence config from from config file")
		}
	}

	return &global, nil
}
