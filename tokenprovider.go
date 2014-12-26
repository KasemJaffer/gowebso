package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"time"
)

//Creates a new signed JWT with the provided email as claim
func getNewJwt(email string) string {
	key := []byte(jwtSigningKey)
	// Create the token
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	// Set some claims
	token.Claims["email"] = email
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	// Sign and get the complete encoded token as a string
	tokenString, _ := token.SignedString(key)
	log.Println("createJwt: " + tokenString)
	return tokenString
}

//Checks wether the JWT provided is valid or not and sets the user
func parseJwt(myToken string, user *User) bool {
	token, err := jwt.Parse(myToken, keyFn)
	if err == nil && token.Valid {
		user.Email = token.Claims["email"].(string)
		user.find(user)
		return true
	}
	return false
}

//A func to return the JWT signing key as []byte
func keyFn(token *jwt.Token) (interface{}, error) {
	return []byte(jwtSigningKey), nil
}

//Makes api call to the provider with the access token and retrieves the user data
func getUserFromSocialTokens(token string, provider string, user *User) error {

	switch provider {

	case "facebook":
		url := "https://graph.facebook.com/me?access_token=" + token
		//log.Println("Calling: " + url)
		res, err := http.Get(url)

		if err != nil || res.StatusCode != http.StatusOK {
			log.Println("Calling Error: " + err.Error())
			return errors.New("invalid token")
		}

		log.Println("Calling Success")

		var data map[string]interface{}
		parseBody(res.Body, &data)
		log.Println("Parse Success")
		user.SocialId = getStringProperty(data, "id")
		user.Email = getStringProperty(data, "email")
		user.Name = getStringProperty(data, "name")
		user.Link = getStringProperty(data, "link")
		user.Birthday = getStringProperty(data, "birthday")
		user.Gender = getStringProperty(data, "gender")
		user.Timezone = getIntProperty(data, "timezone")
		user.SocialUserName = getStringProperty(data, "username")
		log.Println("User Success")
		return nil

	case "google":
		url := "https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=" + token
		//log.Println("Calling: " + url)
		res, err := http.Get(url)

		if err != nil || res.StatusCode != http.StatusOK {
			log.Println("Calling Error: " + err.Error())
			return errors.New("invalid token")
		}

		log.Println("Calling Success")

		var data map[string]interface{}
		parseBody(res.Body, &data)
		log.Println("Parse Success")
		user.SocialId = getStringProperty(data, "id")
		user.Email = getStringProperty(data, "email")
		user.Name = getStringProperty(data, "name")
		user.Link = getStringProperty(data, "link")
		user.Birthday = getStringProperty(data, "birthday")
		user.Gender = getStringProperty(data, "gender")
		user.Picture = getStringProperty(data, "picture")
		log.Println("User Success")
		return nil

	default:
		return errors.New("unsupported provider")
	}
}

func getIntProperty(data map[string]interface{}, name string) int {
	if val := data[name]; val != nil {
		return int(val.(float64))
	}
	return 0
}

func getStringProperty(data map[string]interface{}, name string) string {
	if val := data[name]; val != nil {
		return val.(string)
	}
	return ""
}

const jwtSigningKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA4f5wg5l2hKsTeNem/V41fGnJm6gOdrj8ym3rFkEU/wT8RDtn
SgFEZOQpHEgQ7JL38xUfU0Y3g6aYw9QT0hJ7mCpz9Er5qLaMXJwZxzHzAahlfA0i
cqabvJOMvQtzD6uQv6wPEyZtDTWiQi9AXwBpHssPnpYGIn20ZZuNlX2BrClciHhC
PUIIZOQn/MmqTD31jSyjoQoV7MhhMTATKJx2XrHhR+1DcKJzQBSTAGnpYVaqpsAR
ap+nwRipr3nUTuxyGohBTSmjJ2usSeQXHI3bODIRe1AuTyHceAbewn8b462yEWKA
Rdpd9AjQW5SIVPfdsz5B6GlYQ5LdYKtznTuy7wIDAQABAoIBAQCwia1k7+2oZ2d3
n6agCAbqIE1QXfCmh41ZqJHbOY3oRQG3X1wpcGH4Gk+O+zDVTV2JszdcOt7E5dAy
MaomETAhRxB7hlIOnEN7WKm+dGNrKRvV0wDU5ReFMRHg31/Lnu8c+5BvGjZX+ky9
POIhFFYJqwCRlopGSUIxmVj5rSgtzk3iWOQXr+ah1bjEXvlxDOWkHN6YfpV5ThdE
KdBIPGEVqa63r9n2h+qazKrtiRqJqGnOrHzOECYbRFYhexsNFz7YT02xdfSHn7gM
IvabDDP/Qp0PjE1jdouiMaFHYnLBbgvlnZW9yuVf/rpXTUq/njxIXMmvmEyyvSDn
FcFikB8pAoGBAPF77hK4m3/rdGT7X8a/gwvZ2R121aBcdPwEaUhvj/36dx596zvY
mEOjrWfZhF083/nYWE2kVquj2wjs+otCLfifEEgXcVPTnEOPO9Zg3uNSL0nNQghj
FuD3iGLTUBCtM66oTe0jLSslHe8gLGEQqyMzHOzYxNqibxcOZIe8Qt0NAoGBAO+U
I5+XWjWEgDmvyC3TrOSf/KCGjtu0TSv30ipv27bDLMrpvPmD/5lpptTFwcxvVhCs
2b+chCjlghFSWFbBULBrfci2FtliClOVMYrlNBdUSJhf3aYSG2Doe6Bgt1n2CpNn
/iu37Y3NfemZBJA7hNl4dYe+f+uzM87cdQ214+jrAoGAXA0XxX8ll2+ToOLJsaNT
OvNB9h9Uc5qK5X5w+7G7O998BN2PC/MWp8H+2fVqpXgNENpNXttkRm1hk1dych86
EunfdPuqsX+as44oCyJGFHVBnWpm33eWQw9YqANRI+pCJzP08I5WK3osnPiwshd+
hR54yjgfYhBFNI7B95PmEQkCgYBzFSz7h1+s34Ycr8SvxsOBWxymG5zaCsUbPsL0
4aCgLScCHb9J+E86aVbbVFdglYa5Id7DPTL61ixhl7WZjujspeXZGSbmq0Kcnckb
mDgqkLECiOJW2NHP/j0McAkDLL4tysF8TLDO8gvuvzNC+WQ6drO2ThrypLVZQ+ry
eBIPmwKBgEZxhqa0gVvHQG/7Od69KWj4eJP28kq13RhKay8JOoN0vPmspXJo1HY3
CKuHRG+AP579dncdUnOMvfXOtkdM4vk0+hWASBQzM9xzVcztCa+koAugjVaLS9A+
9uQoqEeVNTckxx0S2bYevRy7hGQmUJTyQm3j1zEUR5jpdbL83Fbq
-----END RSA PRIVATE KEY-----`
