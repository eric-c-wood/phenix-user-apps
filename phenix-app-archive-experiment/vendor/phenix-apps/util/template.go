package util

import (
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func GenerateFromTemplate(name string, tmpl []byte, data interface{}, w io.Writer) error {
	t := template.Must(template.New(name).Parse(string(tmpl)))

	if err := t.Execute(w, data); err != nil {
		return fmt.Errorf("executing %s template: %w", name, err)
	}

	return nil
}

func CreateFileFromTemplate(name string, tmpl []byte, data interface{}, filename string) error {
	dir := filepath.Dir(filename)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating template path: %w", err)
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating template file: %w", err)
	}

	defer f.Close()

	return GenerateFromTemplate(name, tmpl, data, f)
}

func RestoreAsset(templates embed.FS, path, name string) error {
	data, err := templates.ReadFile(name)
	if err != nil {
		return err
	}

	file, err := templates.Open(name)
	if err != nil {
		return err
	}

	info, err := file.Stat()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, info.Mode())
	if err != nil {
		return err
	}

	err = os.Chtimes(path, info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}

	return nil
}
