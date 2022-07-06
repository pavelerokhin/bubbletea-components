package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

var (
	level1regex = `([0-9]{2})000000`
	level2regex = `([0-9]{2})([0-9]{3})000`
	level3regex = `([0-9]{2})([0-9]{3})([0-9]{3})`

	r1 = regexp.MustCompile(level1regex)
	r2 = regexp.MustCompile(level2regex)
	r3 = regexp.MustCompile(level3regex)
)

func ReadCSV(filename string) *item {
	file := readFile(filename)
	defer file.Close()

	csvLines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var t item

	for _, line := range csvLines {
		code := line[0]
		value := line[1]

		parseRow(code, value, &t)
	}

	return &t
}

func parseRow(code, value string, t *item) {
	if r1.MatchString(code) {
		if j := fineByID(code, t); j == -1 {
			t.Ancestors = append(t.Ancestors, item{
				ID:        code,
				Title:     value,
				Ancestors: []item{},
			})
		} else {
			t.Ancestors[j].Title = value
		}

		return
	}

	if r2.MatchString(code) {
		i := item{
			ID:        code,
			Title:     value,
			Ancestors: []item{},
		}

		codeParts := r2.FindAllStringSubmatch(code, -1)
		level1part := fmt.Sprintf("%s000000", codeParts[0][1])
		var parentHasBeenFound bool

		for j1, i1 := range t.Ancestors {
			if i1.ID == level1part {
				if j := fineByID(code, t); j == -1 {
					t.Ancestors[j1].Ancestors = append(t.Ancestors[j1].Ancestors, i)
				} else {
					t.Ancestors[j1].Ancestors[j].Title = value
				}
				parentHasBeenFound = true
			}
		}

		if !parentHasBeenFound {
			i1 := item{
				ID:        level1part,
				Title:     "",
				Ancestors: []item{i},
			}
			t.Ancestors = append(t.Ancestors, i1)
		}

		return
	}

	if r3.MatchString(code) {
		i := item{
			ID:    code,
			Title: value,
		}

		codeParts := r3.FindAllStringSubmatch(code, -1)
		level1part := fmt.Sprintf("%s000000", codeParts[0][1])
		level2part := fmt.Sprintf("%s%s000", codeParts[0][1], codeParts[0][2])
		var parent1HasBeenFound, parent2HasBeenFound bool
		var parent1Level *item

		for j1, i1 := range t.Ancestors {
			if i1.ID != level1part {
				continue
			}
			parent1HasBeenFound = true

			for j2, i2 := range i1.Ancestors {
				if i2.ID == level2part {
					t.Ancestors[j1].Ancestors[j2].Ancestors = append(t.Ancestors[j1].Ancestors[j2].Ancestors, i)
					parent2HasBeenFound = true
				}
			}
		}

		if !parent1HasBeenFound {
			i1 := item{
				ID:    level1part,
				Title: "",
				Ancestors: []item{
					{
						ID:        level2part,
						Title:     "",
						Ancestors: []item{i},
					}},
			}
			t.Ancestors = append(t.Ancestors, i1)
		}

		if !parent2HasBeenFound {
			i2 := item{
				ID:        level2part,
				Title:     "",
				Ancestors: []item{i},
			}
			parent1Level.Ancestors = append(parent1Level.Ancestors, i2)
		}
		return
	}
}

func fineByID(id string, t *item) int {
	for j, i := range t.Ancestors {
		if i.ID == id {
			return j
		}
	}
	return -1
}

func readFile(filename string) *os.File {
	file, err := os.Open(filename)

	if err != nil {
		log.Printf("cannot open file %s", filename)
		os.Exit(1)
	}

	return file
}

func WriteJSON(t *item) {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create("taxonomy.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err2 := f.WriteString(string(b))

	if err2 != nil {
		log.Fatal(err2)
	}
}
