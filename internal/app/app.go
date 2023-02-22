package app

import (
	. "github.com/rakovets/doc-tools/internal/config"

	"log"
)

func Run() {
	log.Printf("INFO : Application started")

	config, err := ReadConfig()
	handleError(err)
	if config.From == Confluence && config.To == AsciiDoc {
		err = convertConfluenceToAsciiDoc(config)
		handleError(err)
	}

	log.Printf("INFO : Application finished")
}

func handleError(err error) {
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}
