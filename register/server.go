package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type reg struct {
	Email     string `bson:"Email" json:"Email,omitempty"`
	Username  string `bson:"Username" json:"Username,omitempty"`
	Password1 string `bson:"Password1" json:"Password1,omitempty"`
	Password2 string `bson:"Password2" json:"Password2,omitempty"`
}

type user struct {
	ID       string    `bson:"_id"`
	Email    string    `bson:"email" json:"Email,omitempty"`
	Username string    `bson:"username" json:"Username,omitempty"`
	Password string    `bson:"password" json:"Password,omitempty"`
	Files    []*string `bson:"files"`
}

func isEmail(str string) bool {
	var email string
	email = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	rx := regexp.MustCompile(email)
	return rx.MatchString(str)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func register(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var t reg
	err := decoder.Decode(&t)
	fmt.Println(t)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "No data" }`))
		return
	}
	if t.Password1 != t.Password2 {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "Passwords not matching" }`))
		return
	}

	if len(t.Password1) < 8 {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "Length of password less than 8" }`))
		return
	}

	if !isEmail(t.Email) {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "Invalid email" }`))
		return
	}
	pass, _ := hashPassword(t.Password1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://az3z3l:"+os.Getenv("mongopwd")+"@cloudcomputing.4rrk6.mongodb.net/uploadService?retryWrites=true&w=majority",
	))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	var result user
	collection := client.Database("uploadService").Collection("users")
	err = collection.FindOne(context.TODO(), bson.M{"Email": t.Email}).Decode(&result)
	if err != nil {
		ins := map[string]string{
			"Email":    t.Email,
			"Username": t.Username,
			"Password": pass,
		}
		_, err = collection.InsertOne(context.TODO(), ins)
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.Write([]byte(`{ "Error": "Error inserting data. Please try again later" }`))
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "OK" }`))
	} else {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "Email exists already" }`))
		return
	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/register", register).Methods("POST")
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:80",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Platform runs at http://register:80/")
	log.Fatal(srv.ListenAndServe())

}
