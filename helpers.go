package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
)

type ResponseMessage struct {
	Message string
	Token   string
	User    User
}

func initDB(dburl string, dbname string, collectionName string) (*mgo.Session, *mgo.Collection, error) {
	session, err := mgo.Dial(dburl)
	if err != nil {
		log.Println("initDB: " + err.Error())
		return nil, nil, err
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB(dbname).C(collectionName)

	return session, c, nil
}

func parseBody(body io.ReadCloser, data interface{}) error {
	reqBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(reqBytes, &data)
	if err != nil {
		return err
	}
	return nil
}

func getUserFromQueryString(req *http.Request, user *User) {
	user.Email = req.FormValue("Email")
	return
}

func respondWithJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json ; charset=utf-8")
	fmt.Fprintf(w, toJsonString(data))
}

func respondFailWithJson(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json ; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintf(w, toJsonString(data))
}

func toJsonString(data interface{}) (jsonString string) {
	jsonBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return
	}
	return string(jsonBytes)
}

func printIps() {
	name, err := os.Hostname()
	if err != nil {
		log.Printf("Oops: %v\n", err)
		return
	}
	log.Printf("Machine Name: %v\n", name)
	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Printf("Oops: %v\n", err)
		return
	}

	for _, a := range addrs {
		log.Println("Address: " + a)
	}
}

func isEmailValid(email string) bool {
	exp, err := regexp.Compile(emailRegex)
	if err != nil {
		panic(err)
	}

	if exp.MatchString(email) {
		log.Println("Valid email: " + email)
		return true
	}
	log.Println("Invalid email: " + email)
	return false
}

// const emailRegex = "[a-zA-Z0-9\\+\\.\\_\\%\\-\\+]{1,256}" +
// 	"\\@" +
// 	"[a-zA-Z0-9][a-zA-Z0-9\\-]{0,64}" +
// 	"(" +
// 	"\\." +
// 	"[a-zA-Z0-9][a-zA-Z0-9\\-]{0,25}" +
// 	")+"
const emailRegex = "(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])"
