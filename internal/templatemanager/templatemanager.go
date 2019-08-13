package templatemanager

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/oxtoacart/bpool"
)

var templates map[string]*template.Template
var bufferPool *bpool.BufferPool
var mainTemplateDef = `{{define "main" }} {{ template "base" . }} {{ end }}`

// create a buffer pool
func init() {
	bufferPool = bpool.NewBufferPool(64)
}

type TemplateConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
}

var templateConfig *TemplateConfig

func SetTemplateConfig(layoutPath, includePath string) {
	templateConfig = &TemplateConfig{layoutPath, includePath}
}

func LoadTemplates() (err error) {

	if templateConfig == nil {
		err = NewError("TemplateConfig not initialized")
		return err
	}
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	layoutFiles, err := filepath.Glob(templateConfig.TemplateLayoutPath + "*.html")
	if err != nil {
		return err
	}
	log.Println("layoutFiles:", layoutFiles)

	includeFiles, err := filepath.Glob(templateConfig.TemplateIncludePath + "*.html")
	if err != nil {
		return err
	}
	log.Println("includeFiles:", includeFiles)

	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(mainTemplateDef)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			return err
		}
		templates[fileName] = template.Must(templates[fileName].ParseFiles(files...))
	}

	log.Println("templates loading successful")
	return nil

}

func RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name),
			http.StatusInternalServerError)
		err := NewError("Template doesn't exist")
		return err
	}

	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	err := tmpl.Execute(buf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		err := NewError("Template execution failed")
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	return err
}
