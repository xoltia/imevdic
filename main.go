package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"sync"
)

type Entry struct {
	Term           string
	Reading        string
	ContentStrings []string
}

func (e *Entry) UnmarshalJSON(b []byte) (err error) {
	var content []interface{}

	err = json.Unmarshal(b, &[]interface{}{
		&e.Term,
		&e.Reading,
		nil,
		nil,
		nil,
		&content,
	})

	if err != nil {
		return
	}

	for _, v := range content {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		if m["type"] != "structured-content" {
			continue
		}

		cm, ok := m["content"].([]interface{})

		if !ok {
			continue
		}

		var contentMaps []map[string]interface{}

		for _, v := range cm {
			m, ok := v.(map[string]interface{})
			if ok {
				contentMaps = append(contentMaps, m)
			}
		}

		for len(contentMaps) > 0 {
			contentMap := contentMaps[0]

			if contentString, okType := contentMap["content"].(string); okType {
				e.ContentStrings = append(e.ContentStrings, contentString)
				contentMaps = contentMaps[1:]
				continue
			}

			var contentMaps2 []map[string]interface{}
			var cm2 []interface{}

			if cm2, ok = contentMap["content"].([]interface{}); ok {
				for _, v := range cm2 {
					m, ok := v.(map[string]interface{})
					if ok {
						contentMaps2 = append(contentMaps2, m)
					}
				}
			}

			contentMaps = append(contentMaps[1:], contentMaps2...)
		}
	}

	return nil
}

func main() {
	files, err := os.ReadDir("./data/Pixiv")

	if err != nil {
		panic(err)
	}

	entryChunks := make(chan []Entry)
	wg := sync.WaitGroup{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "term_bank_") {
			continue
		}

		f, err := os.Open("./data/Pixiv/" + file.Name())

		if err != nil {
			panic(err)
		}

		wg.Add(1)

		go func() {
			defer f.Close()
			defer wg.Done()

			var data []Entry
			err = json.NewDecoder(f).Decode(&data)

			if err != nil {
				panic(err)
			}

			entryChunks <- data
		}()
	}

	outFile, err := os.Create("dict.txt")

	if err != nil {
		panic(err)
	}

	defer outFile.Close()

	go func() {
		wg.Wait()
		close(entryChunks)
	}()

	for entries := range entryChunks {
		for _, entry := range entries {
			if !slices.Contains(entry.ContentStrings, "バーチャルYouTuber") {
				continue
			}

			outFile.WriteString(entry.Reading + "\t" + entry.Term + "\t名詞\n")
		}
	}
}
