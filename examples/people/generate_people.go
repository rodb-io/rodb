package main

import (
	faker "github.com/bxcodec/faker/v3"
	"encoding/json"
	"log"
	"os"
)

func main() {
	file, err := os.Create("/people.json")
	if err != nil {
		log.Fatal("Error create the file: ", err)
	}
	defer file.Close()

	for i := 1; i <= 1000000; i++ {
		jsonDocument, err := json.Marshal(map[string]interface{} {
			"id":          i,
			"username":    faker.Username(),
			"email":       faker.Email(),
			"gender":      faker.Gender(),
			"phoneNumber": faker.Phonenumber(),
			"firstName":   faker.FirstName(),
			"lastName":    faker.LastName(),
		})
		if err != nil {
			log.Fatal("Error writing to the file: ", err)
		}

		_, err = file.Write(jsonDocument)
		if err != nil {
			log.Fatal("Error writing to the file: ", err)
		}
		_, err = file.WriteString("\n")
		if err != nil {
			log.Fatal("Error writing to the file: ", err)
		}
	}
}
