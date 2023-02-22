package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/rakovets/doc-tools/internal/config"
)

func convertConfluenceToAsciiDoc(global *Global) error {
	currentConfig := global.ConfluenceConfig

	for _, page := range currentConfig.Pages {
		content, err := readContent(currentConfig, page)
		if err != nil {
			return err
		}

		asciiDocHeader := fmt.Sprintf("= %s\n\n", content.Title)
		convertedContent, err := convert(global, []byte(content.Body.Storage.Value), asciiDocHeader)
		if err != nil {
			return err
		}

		filename := content.Title + global.To.FileExtension()
		path := strings.Join([]string{global.Output, filename}, string(os.PathSeparator))
		err = writeContent(path, convertedContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertAsciiDocToPdf(global *Global) error {
	for _, filename := range find(global.Input, global.From.FileExtension()) {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		convertedContent, err := convert(global, content, "")
		if err != nil {
			return err
		}

		filename := cleanFilename(filename) + global.To.FileExtension()
		path := strings.Join([]string{global.Output, filename}, string(os.PathSeparator))
		err = writeContent(path, convertedContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func readContent(config ConfluenceConfig, page string) (*ConfluenceContent, error) {
	baseUrl := strings.Join([]string{config.Url, "rest", "api", "content"}, "/")
	url := fmt.Sprintf("%s/%s?expand=body.storage", baseUrl, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.Username, config.Password)
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("confluence return '%s'", resp.Status)
	}

	content := ConfluenceContent{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &content)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG: read content from '%s'", url)
	return &content, nil
}

func convert(config *Global, input []byte, header string) ([]byte, error) {
	cmd := prepareCommand(config)
	stdin, err := cmd.StdinPipe()
	if nil != err {
		return nil, err
	}
	_, err = stdin.Write(input)
	if err != nil {
		return nil, err
	}
	stdin.Close()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG: convert content from '%s' to '%s'", config.From.PandocName(), config.To.PandocName())
	if header == "" {
		return out, nil
	}
	return append([]byte(header), out...), nil
}

func prepareCommand(config *Global) *exec.Cmd {
	if config.From == AsciiDoc {
		if config.To == Pdf {
			return exec.Command("asciidoctor", "--verbose", "-r", "asciidoctor-pdf", "-b", "pdf", "-o", "-", "-")
		}
		return exec.Command("asciidoctor")
	}
	toArg := "--to=" + config.To.PandocName()
	fromArg := "--from=" + config.From.PandocName()
	return exec.Command("pandoc", "--verbose", "--wrap=none", fromArg, toArg)
}

func writeContent(path string, content []byte) error {
	ensureDir(path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.Write(content)
	if err != nil {
		return err
	}

	log.Printf("DEBUG: write content to '%s'", path)
	return nil
}

func ensureDir(filename string) {
	dirName := filepath.Dir(filename)
	if _, err := os.Stat(dirName); err != nil {
		err = os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func find(dir, ext string) []string {
	var a []string
	err := filepath.WalkDir(dir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return a
}

func cleanFilename(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}
