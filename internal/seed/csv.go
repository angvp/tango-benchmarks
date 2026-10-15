package seed

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// WriteCSV writes categories.csv, pets.csv and customers.csv into dir.
func WriteCSV(dir string) error {
	if err := writeFile(filepath.Join(dir, "categories.csv"), []string{"id", "name"}, categoryRows()); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(dir, "pets.csv"), []string{"id", "name", "price_cents", "status", "category_id"}, petRows()); err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "customers.csv"), []string{"id", "name", "email"}, customerRows())
}

func categoryRows() [][]string {
	var rows [][]string
	for _, c := range Categories() {
		rows = append(rows, []string{strconv.Itoa(c.ID), c.Name})
	}
	return rows
}

func petRows() [][]string {
	var rows [][]string
	for _, p := range Pets() {
		rows = append(rows, []string{strconv.Itoa(p.ID), p.Name, strconv.Itoa(p.PriceCents), p.Status, strconv.Itoa(p.CategoryID)})
	}
	return rows
}

func customerRows() [][]string {
	var rows [][]string
	for _, c := range Customers() {
		rows = append(rows, []string{strconv.Itoa(c.ID), c.Name, c.Email})
	}
	return rows
}

func writeFile(path string, header []string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	w := csv.NewWriter(f)
	w.UseCRLF = false
	if err := w.Write(header); err != nil {
		f.Close()
		return err
	}
	if err := w.WriteAll(rows); err != nil {
		f.Close()
		return err
	}
	if err := w.Error(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
