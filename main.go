package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
)

type Animal struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var animals []Animal
var nextAnimalID = 1

const animalsFile = "animals.json"

var storagePath string

func saveAnimalsToFile(animals []Animal) error {
	data, err := json.Marshal(animals)
	if err != nil {
		return err
	}
	return os.WriteFile(storagePath+"/"+animalsFile, data, 0644)
}

func loadAnimalsFromFile() ([]Animal, error) {
	data, err := os.ReadFile(storagePath + "/" + animalsFile)
	if err != nil {
		return nil, err
	}
	var animals []Animal
	if err := json.Unmarshal(data, &animals); err != nil {
		return nil, err
	}
	return animals, nil
}

func GetAnimals(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(animals)
}

func AddAnimal(w http.ResponseWriter, r *http.Request) {
	var newAnimal Animal
	json.NewDecoder(r.Body).Decode(&newAnimal)

	newAnimal.ID = nextAnimalID
	nextAnimalID++

	animals = append(animals, newAnimal)
	err := saveAnimalsToFile(animals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(newAnimal)
}

func RemoveAnimal(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	animalID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	for index, animal := range animals {
		if animal.ID == animalID {
			animals = append(animals[:index], animals[index+1:]...)
			err := saveAnimalsToFile(animals)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}

	json.NewEncoder(w).Encode(animals)
}

func cleanup() {
	fmt.Println("Saving animals to file before exit...")
	if err := saveAnimalsToFile(animals); err != nil {
		fmt.Println("Error saving animals:", err)
	}
}

func main() {
	// Define a flag named "animals-storage" with a default value
	animalsStorage := flag.String("animals-storage", "/animals-storage", "Path to the animals storage directory")
	flag.Parse()
	storagePath = *animalsStorage

	// Check if the directory exists, and if not, create it
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		err := os.MkdirAll(storagePath, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/animals", GetAnimals).Methods("GET")
	router.HandleFunc("/animals", AddAnimal).Methods("POST")
	router.HandleFunc("/animals/{id:[0-9]+}", RemoveAnimal).Methods("DELETE")

	loadedAnimals, err := loadAnimalsFromFile()
	if err != nil {
		fmt.Println("Error loading animals from file:", err)
	} else {
		animals = loadedAnimals
		for _, animal := range animals {
			if animal.ID >= nextAnimalID {
				nextAnimalID = animal.ID + 1
			}
		}
	}

	// Handle cleanup on application exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	http.ListenAndServe(":8080", router)
}
