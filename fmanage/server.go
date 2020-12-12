package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type user struct {
	ID       string   `bson:"_id"`
	Email    string   `bson:"email" json:"Email,omitempty"`
	Username string   `bson:"username" json:"Username,omitempty"`
	Password string   `bson:"password" json:"Password,omitempty"`
	File     []string `bson:"file"`
}

var (
	jwtSecret = []byte("thesecretthatsnotasecretanymore")
)

var userCtxKey = &contextKey{"username"}

type contextKey struct {
	username string
}

type filesAvailable struct {
	ID    string   `json:"id"`
	Names []string `json:"names"`
}

func getUserIDByEmail(u string) (string, error) {
	var newlogin *user
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
	collection := client.Database("uploadService").Collection("users")
	err = collection.FindOne(context.TODO(), bson.M{"Email": u}).Decode(&newlogin)
	if err != nil {
		return "", err
	}
	if newlogin.Username != "" && newlogin.Email != "" {
		return string(newlogin.ID), nil
	}
	return "", nil
}

func ParseToken(tokenStr string) (string, string, error) {
	if tokenStr == "" {
		return "", "", errors.New("Authorization token must be present")
	}
	token, damn := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("There was an error")
		}
		return jwtSecret, nil
	})
	if damn != nil {
		return "", "", errors.New("Authorization token must be present")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, err := claims["username"].(string)
		email, err1 := claims["email"].(string)
		if !err || !err1 {
			return "", "", nil
		}
		return username, email, nil
	} else {
		return "", "", errors.New("Invalid authorization token")
	}
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var key string

		cookie, err := r.Cookie("auth")
		if err != nil {
			http.Error(w, "{\"data\"}: {\"Not Logged In\"}", http.StatusBadRequest)
			return
		}

		var cookievalue = cookie.Value
		token := cookievalue
		username, email, err := ParseToken(token)

		if err != nil {
			http.Error(w, "{\"data\"}: {\"Not Logged In\"}", http.StatusBadRequest)
			return
		}
		if username == "" || email == "" {
			http.Error(w, "{\"data\"}: {\"Not Logged In\"}", http.StatusBadRequest)
			return
		}

		key, _ = getUserIDByEmail(email)
		ctx := context.WithValue(r.Context(), userCtxKey, key)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

func upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userid := ForContext(ctx)
	docid, err := primitive.ObjectIDFromHex(userid)
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Could not get the file" }`))
		return
	}
	fileName := r.FormValue("file_name")
	fileName = handler.Filename

	defer file.Close()

	folder := "files/"
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
		f, err := os.OpenFile("files/index.html", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Panic(err)
			f.Close()
			w.Header().Set("content-type", "application/json")
			w.Write([]byte(`{ "Status": "Error on creating folder" }`))
			return
		}
		f.Close()
	}
	folder = "files/" + userid + "/"
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
	}

	f, err := os.OpenFile(folder+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Unable to save the file" }`))
		log.Panic(err)
		return

	}
	defer f.Close()
	_, _ = io.Copy(f, file)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://az3z3l:"+os.Getenv("mongopwd")+"@cloudcomputing.4rrk6.mongodb.net/uploadService?retryWrites=true&w=majority",
	))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("uploadService").Collection("users")
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Could'nt connect to db" }`))
		log.Panic(err)
		return
	}
	filter := bson.D{{"_id", docid}}
	var chall *user
	err = collection.FindOne(context.Background(), filter).Decode(&chall)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Unable to find the user" }`))
	}

	if chall.File == nil {
		update := bson.M{"$set": bson.M{"file": []string{fileName}}}
		_, erro := collection.UpdateOne(
			context.Background(),
			filter,
			update,
		)
		if erro != nil {
			w.Header().Set("content-type", "application/json")
			w.Write([]byte(`{ "Status": "Failed to add file" }`))
			return
		}
	} else {
		for _, i := range chall.File {
			if i == fileName {
				w.Header().Set("content-type", "application/json")
				w.Write([]byte(`{ "Status": "OK" }`))
				return
			}
		}
	}

	if chall.File != nil {
		update := bson.M{"$push": bson.M{"file": fileName}}
		_, erro := collection.UpdateOne(
			context.Background(),
			filter,
			update,
		)
		if erro != nil {
			w.Header().Set("content-type", "application/json")
			w.Write([]byte(`{ "Status": "Unable to add chall to db" }`))
			return
		}
	}
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(`{ "Status": "OK" }`))
	return
}

func delete(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userid := ForContext(ctx)
	docid, err := primitive.ObjectIDFromHex(userid)

	type filee struct {
		Filename string `bson:"filename"`
	}

	decoder := json.NewDecoder(r.Body)
	var t filee
	err = decoder.Decode(&t)
	fmt.Println(t.Filename)
	if err != nil || t.Filename == "" {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "No data" }`))
		return
	}

	filey := "files/" + userid + "/" + t.Filename
	if _, err := os.Stat(filey); os.IsNotExist(err) {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "File does not exist" }`))
		return
	}

	err = os.Remove(filey)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Unable to delete file" }`))
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
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("uploadService").Collection("users")

	filter := bson.D{{"_id", docid}}
	update := bson.M{"$pull": bson.M{"file": t.Filename}}
	_, erro := collection.UpdateOne(
		context.Background(),
		filter,
		update,
	)
	if erro != nil {
		w.Header().Set("content-type", "application/json")
		w.Write([]byte(`{ "Status": "Unable to delete filename from db" }`))
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write([]byte(`{ "Status": "OK" }`))
	return
}

func available(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userid := ForContext(ctx)
	docid, err := primitive.ObjectIDFromHex(userid)

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
	filter := bson.M{"_id": docid}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)

	jsondat := filesAvailable{Names: result.File, ID: userid}
	encjson, _ := json.Marshal(jsondat)

	w.Header().Set("content-type", "application/json")
	w.Write([]byte(string(encjson)))
	return
}

// ForContext get values passed using the request context
func ForContext(ctx context.Context) string {
	raw, _ := ctx.Value(userCtxKey).(string)
	return raw
}

func main() {
	var dir string

	flag.StringVar(&dir, "dir", "./files", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	r := mux.NewRouter()

	r.PathPrefix("/download/").Handler(http.StripPrefix("/download/", http.FileServer(http.Dir(dir))))

	r.Handle("/upload", middleware(http.HandlerFunc(upload))).Methods("POST")
	r.Handle("/delete", middleware(http.HandlerFunc(delete))).Methods("POST")
	r.Handle("/available", middleware(http.HandlerFunc(available))).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:80",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Platform runs at http://fmanage:80/")
	log.Fatal(srv.ListenAndServe())

}
