gowebso
=======

Simple chat room implementation using go

This project is based on http://gary.burd.info/go-websocket-chat with a lot of additional features.

Some of the new features
------------------------
1. JWT authorization method
2. User management
3. Rooms management
4. Mongo db to store the users data

Example
-------------

prerequisites:-
  - Go distribution from https://golang.org/
  - MongoDb from http://www.mongodb.org/

1. Open cmd and run this command 'go get github.com/KasemJaffer/gowebso' this will download the project in the 
2. Make sure mongodb is running (cmd: mongod)
3. Open cmd in the project folder
4. Type 'go get' and press enter this will download the library used in the project
5. To run the project type 'go build && gowebso.exe'
6. Go to the browser and open 'localhost:8080'



Work is still in progress
