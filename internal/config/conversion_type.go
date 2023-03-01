package config

type ConversionType struct {
	fileExtension string
	pandocName    string
}

var (
	AsciiDoc       = ConversionType{".adoc", "asciidoc"}
	Confluence     = ConversionType{".html", "html"}
	ConfluenceWiki = ConversionType{".jira", "jira"}
	Markdown       = ConversionType{".md", "markdown"}
	DocBook        = ConversionType{".dbk", "docbook"}
	Pdf            = ConversionType{".pdf", "pdf"}
)

func (c *ConversionType) PandocName() string {
	return c.pandocName
}

func (c *ConversionType) FileExtension() string {
	return c.fileExtension
}

func (c *ConversionType) ConversionType(code string) ConversionType {
	switch code {
	case "asciidoc":
		return AsciiDoc
	case "confluence":
		return Confluence
	case "confluence-wiki":
		return ConfluenceWiki
	case "markdown":
		return Markdown
	case "docbook":
		return DocBook
	case "pdf":
		return Pdf
	}
	return ConversionType{"unknown", code}
}

func (c *ConversionType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var typeAsString string
	err := unmarshal(&typeAsString)
	if err != nil {
		return err
	}
	*c = c.ConversionType(typeAsString)
	return nil
}
