package main

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

//This is the user model
type User struct {
	SocialId       string
	Name           string
	Email          string
	Birthday       string
	Gender         string
	Link           string
	Timezone       int
	SocialUserName string
	Picture        string
}

//Finds the user in mongodb by email and sets the provided with the result
func (u *User) find(result *User) error {
	session, c, err := initDB(*dbUrl, *dbName, *colName)
	if err != nil {
		return errors.New("Database error")
	}
	defer session.Close()

	return c.Find(bson.M{"email": u.Email}).One(result)
}

//Helper func used to sing in the user
func (u *User) login() (string, error) {

	session, c, err := initDB(*dbUrl, *dbName, *colName)
	if err != nil {
		return "", errors.New("Database error")
	}
	defer session.Close()

	result := User{}
	if !isEmailValid(u.Email) {
		return "", errors.New("Invalid email")
	}

	if err := c.Find(bson.M{"email": u.Email}).One(&result); err == nil {
		token := getNewJwt(result.Email)
		return token, nil
	}

	c.Remove(bson.M{"email": u.Email})
	if err := c.Insert(u); err != nil {
		return "", err
	}

	token := getNewJwt(result.Email)
	return token, nil
}

func (user *User) getUserFromHeader(r *http.Request) {
	user.Email = r.Header.Get("email")
	user.Name = r.Header.Get("name")
}

func (user *User) addUserToHeader(r *http.Request) {
	r.Header.Add("email", user.Email)
	r.Header.Add("name", user.Name)
}

type LoginBinding struct {
	Token    string
	Provider string
}
