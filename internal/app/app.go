package app

import (
	"fmt"
	"log"

	. "github.com/rakovets/doc-tools/internal/config"
)

func Run() {
	log.Printf("INFO : application started")

	config, err := ReadConfig()
	handleError(err)
	message := fmt.Sprintf("INFO : using convertor from '%s' to '%s'", config.From.PandocName(), config.To.PandocName())
	log.Printf(message)
	if config.From == Confluence && config.To == AsciiDoc {
		err = convertConfluenceToAsciiDoc(config)
		handleError(err)
	} else if config.From == AsciiDoc && config.To == Pdf {
		err = convertAsciiDocToPdf(config)
		handleError(err)
	} else if config.From == AsciiDoc && config.To == ConfluenceWiki {
		err = convertAsciiDocToConfluence(config)
		handleError(err)
	} else {
		log.Printf("WARN : coverter from '%s' to '%s' not found", config.From.PandocName(), config.To.PandocName())
	}

	log.Printf("INFO : application finished")
}

func handleError(err error) {
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
	}
}
