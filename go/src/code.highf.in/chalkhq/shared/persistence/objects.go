package persistence

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// persist data to disk every 2 seconds.
// todo: optimize into goroutines and move to shared lib, etc.
// todo: add function to mkdir of parents if not exists
func PersistData(object interface{}, filename string, changed *bool) {
	for {
		tmpFilename := filename + ".tmp"
		time.Sleep(2 * time.Second)
		if *changed == true {
			*changed = false
			file, err := os.Create(tmpFilename)
			if err != nil {
				fmt.Println("Error creating file " + tmpFilename)
				return
			}
			defer file.Close()

			objectJson, err := json.MarshalIndent(object, "", " ")
			if err != nil {
				fmt.Println("Error Marshalling object")
				return
			}

			file.Write(objectJson)

			// remove the old file
			os.Remove(filename)

			// rename the new file
			os.Rename(tmpFilename, filename)
		}
	}
}

// get persisted data
func GetData(object interface{}, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file " + filename)
		fmt.Println(err.Error())
		return
	}

	defer file.Close()

	jsonParser := json.NewDecoder(file)

	if err = jsonParser.Decode(&object); err == io.EOF || err == nil {
	} else {
		fmt.Println(err)
	}
}
