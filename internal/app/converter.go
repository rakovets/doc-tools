package app

import (
	"bufio"
	"bytes"
	"crypto/tls"
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
		cmd := prepareCommand(global.From, global.To)
		convertedContent, err := convert(cmd, global, []byte(content.Body.Storage.Value), asciiDocHeader)
		if err != nil {
			return err
		}

		filename := normalizeFilename(content.Title) + global.To.FileExtension()
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

		cmd := prepareCommand(global.From, global.To)
		convertedContent, err := convert(cmd, global, content, "")
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

// todo: migrate to https://github.com/confluence-publisher/confluence-publisher
func convertAsciiDocToConfluence(global *Global) error {
	for _, importPage := range global.ImportPages {
		filepath := strings.Join([]string{global.Input, importPage.Source}, "/")
		content, err := os.ReadFile(filepath)
		if err != nil {
			return err
		}

		cmd := prepareCommand(global.From, DocBook)
		tmpContent, err := convert(cmd, global, content, "")
		cmd = prepareCommand(DocBook, Markdown)
		tmpContent, err = convert(cmd, global, tmpContent, "")
		cmd = prepareCommand(Markdown, global.To)
		convertedContent, err := convert(cmd, global, tmpContent, "")

		if err != nil {
			return err
		}

		err = writeContentToConfluence(global.ConfluenceConfig, importPage, convertedContent)
		if err != nil {
			return err
		}
	}
	return nil
}

// todo: move to package `rest_client`
func readContent(config ConfluenceConfig, page string) (*ConfluenceContent, error) {
	baseUrl := strings.Join([]string{config.Url, "rest", "api", "content"}, "/")
	url := fmt.Sprintf("%s/%s?expand=body.storage,version", baseUrl, page)
	log.Printf("DEBUG: start to read content from '%s'", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.Username, config.Password)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
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

	log.Printf("DEBUG: finish to read content from '%s'", url)
	return &content, nil
}

// todo: move to package `rest_client`
func writeContentToConfluence(config ConfluenceConfig, importPage ImportPage, content []byte) error {
	baseUrl := strings.Join([]string{config.Url, "rest", "api", "content"}, "/")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("%s/%s", baseUrl, importPage.Id)

	currentContent, _ := readContent(config, importPage.Id)

	log.Printf("DEBUG: start to upload content to '%s'", url)

	confluenceContent := ConfluenceContentRequest{}
	confluenceContent.Id = importPage.Id
	confluenceContent.Title = importPage.Title
	confluenceContent.Type = "page"
	confluenceContent.Version.Number = currentContent.Version.Number + 1
	confluenceContent.Body.Wiki.Value = string(content)

	preparedContent, _ := json.Marshal(confluenceContent)
	reader := bytes.NewReader(preparedContent)
	req, err := http.NewRequest("PUT", url, reader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}
	req.SetBasicAuth(config.Username, config.Password)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		responseBody, _ := io.ReadAll(resp.Body)
		log.Printf("DEBUG: confluence return '%s' with body '%s'", resp.Status, responseBody)
		return fmt.Errorf("confluence return '%s'", resp.Status)
	}

	log.Printf("DEBUG: finish to upload content to '%s'", url)
	return nil
}

func convert(cmd *exec.Cmd, config *Global, input []byte, header string) ([]byte, error) {
	log.Printf("DEBUG: start to convert content from '%s' to '%s'", config.From.PandocName(), config.To.PandocName())
	log.Printf("TRACE: use command '%s'", cmd)
	stdin, _ := cmd.StdinPipe()
	go write(stdin, input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: verbose error\n%s", out)
		return nil, err
	}

	log.Printf("DEBUG: finish to convert content from '%s' to '%s'", config.From.PandocName(), config.To.PandocName())
	if header == "" {
		return out, nil
	}
	return append([]byte(header), out...), nil
}

func write(stdin io.WriteCloser, input []byte) {
	defer stdin.Close()
	var n int
	if len(input) > 4096 {
		writer := bufio.NewWriterSize(stdin, 4096)
		n, _ = writer.Write(input)
	} else {
		n, _ = stdin.Write(input)
	}
	log.Printf("TRACE: wrote %d byte", n)
}

func prepareCommand(from ConversionType, to ConversionType) *exec.Cmd {
	if from == AsciiDoc {
		switch to {
		case Pdf:
			return exec.Command("asciidoctor", "-r", "asciidoctor-pdf", "-b", "pdf", "-o", "-", "-")
		case DocBook:
			return exec.Command("asciidoctor", "-b", "docbook", "-o", "-", "-")
		default:
			return exec.Command("asciidoctor", "-o", "-", "-")
		}
	}
	toArg := "--to=" + to.PandocName()
	fromArg := "--from=" + from.PandocName()
	return exec.Command("pandoc", "--wrap=none", fromArg, toArg)
}

func writeContent(path string, content []byte) error {
	log.Printf("DEBUG: start to write content to '%s'", path)
	ensureDir(path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.Write(content)
	if err != nil {
		return err
	}

	log.Printf("DEBUG: finish to write content to '%s'", path)
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

func normalizeFilename(filename string) string {
	tmp := strings.ReplaceAll(filename, "/", "_")
	return tmp
}
