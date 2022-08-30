package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempt" bson:"lastname,omitempty"`
}

var client *mongo.Client
var url, dbname, collectionname string

func CreatePerson(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	var person Person

	json.NewDecoder(r.Body).Decode(&person)

	log.Println(person.Firstname)
	log.Println(person.Lastname)
	if len(person.Firstname) <= 0 || len(person.Lastname) <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": + "Please provide valid credentials"`))
		return
	}
	coll := client.Database(dbname).Collection(collectionname)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	result, _ := coll.InsertOne(ctx, person)
	json.NewEncoder(w).Encode(result)
}

func GetPeople(w http.ResponseWriter, r *http.Request) {
	log.Println("handling GetPeopleEndpoint...")
	w.Header().Add("content-type", "application/json")
	var people []Person
	coll := client.Database(dbname).Collection(collectionname)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	cursor, err := coll.Find(ctx, bson.D{{}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":` + err.Error() + `"}"`))
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var p Person
		cursor.Decode(&p)
		people = append(people, p)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":` + err.Error() + `"}"`))
		return
	}
	json.NewEncoder(w).Encode(people)

}
func main() {
	fmt.Println("Starting the application...")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("some error loding the .env files")
	}
	url = os.Getenv("DB_URL")
	dbname = os.Getenv("DB_NAME")
	collectionname = os.Getenv("DB_COLLECTION_NAME")

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(url).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePerson).Methods("GET")
	router.HandleFunc("/people", GetPeople).Methods("GET")
	http.ListenAndServe(":8080", router)
}
