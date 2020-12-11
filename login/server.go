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

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type login struct {
	Email    string `bson:"Email" json:"Email,omitempty"`
	Password string `bson:"Password" json:"Password,omitempty"`
}

type user struct {
	ID       string    `bson:"_id"`
	Email    string    `bson:"email" json:"Email,omitempty"`
	Username string    `bson:"username" json:"Username,omitempty"`
	Password string    `bson:"password" json:"Password,omitempty"`
	Files    []*string `bson:"files"`
}

var (
	jwtSecret = []byte("thesecretthatsnotasecretanymore")
)

// CreateTokenEndpoint new jwt token
func CreateTokenEndpoint(s *user) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": s.Username,
		"email":    s.Email,
	})
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

func isEmail(str string) bool {
	var email string
	email = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	rx := regexp.MustCompile(email)
	return rx.MatchString(str)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func register(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var t login
	err := decoder.Decode(&t)
	fmt.Println(t)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "No data" }`))
		return
	}

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
	filter := bson.M{"Email": t.Email}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "wrong username/password" }`))
		return
	}
	if !checkPasswordHash(t.Password, result.Password) {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Error": "wrong username/password" }`))
		return
	}
	if result.Username != "" && result.Email != "" {
		tokenString := CreateTokenEndpoint(&result)
		// TODO return the API key

		cookie := http.Cookie{
			Name:     "auth",
			Value:    tokenString,
			SameSite: 2,
			Path:     "/",
		}

		http.SetCookie(w, &cookie)

		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status":  "OK"}`))
		return
	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", register).Methods("POST")
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:80",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Platform runs at http://login:80/")
	log.Fatal(srv.ListenAndServe())

}
