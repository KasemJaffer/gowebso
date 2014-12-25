package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"
)

func main() {
	//printIps()
	flag.Parse()
	http.Handle("/", homeHandler())
	http.Handle("/ws", authorize(wsHandler()))
	http.Handle("/login", loginHandler())
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

//This handler is used to sign in the user with access token retrieved from the social provider.
//By making api call to the provider to verify the access token and retrieve the data.
//If the access token was verified it responds with a new JWT and the user retrieved.
func loginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		socialToken := r.Header.Get("token")
		provider := r.Header.Get("provider")
		var user User

		//api call to verify and get the user data
		if err := getUserFromSocialTokens(socialToken, provider, &user); err != nil {
			respondFailWithJson(w, &ResponseMessage{Message: err.Error()}, http.StatusBadRequest)
			return
		}

		token, err := user.login()
		if err != nil {
			log.Println(err.Error())
			respondFailWithJson(w, &ResponseMessage{Message: err.Error()}, http.StatusBadRequest)
			return
		}

		log.Println("created")
		respondWithJson(w, &ResponseMessage{Message: "Created", Token: token, User: user})
		return
	})
}

//A dummy home page
func homeHandler() http.Handler {
	homeTempl = template.Must(template.New("").Parse(homeText))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		homeTempl.Execute(w, r.Host)
	})
}

//The websocket communication handler.
//Once called it creates a new websocket connection for the user.
func wsHandler() http.Handler {
	go h.run()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user User
		user.getUserFromHeader(r)

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("wsHandler err: " + err.Error() + " while creating connection with: " + user.Email)
			return
		}
		c := &connection{send: make(chan outgoing), ws: ws, userName: user.Email}
		log.Println("wsHandler: new connection " + user.Email)

		defer func() { h.disconnectChannel <- c }()
		go c.writer()
		c.reader(ws)
	})
}

//Midleware to authorize the requests by checking the validity of the JWT.
//Once authorized it adds the user data to header
func authorize(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user User
		if !parseJwt(r.FormValue("token"), &user) {
			log.Println("Invalid token")
			respondFailWithJson(w, &ResponseMessage{Message: "Invalid token"}, http.StatusBadRequest)
			return
		}
		user.addUserToHeader(r)
		log.Println("User authenticated: " + user.Email)
		h.ServeHTTP(w, r)
	})
}

var (
	addr      = flag.String("addr", ":8080", "http service address")
	dbUrl     = flag.String("dbUrl", "localhost:27017", "mongo db address")
	dbName    = flag.String("dbName", "gowebso", "database name in mongo db")
	colName   = flag.String("colName", "Users", "collection name in the provided db")
	homeTempl *template.Template
)

const homeText = `<html>
<head>
<title>Chat Example</title>
<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
<script type="text/javascript">
    $(function() {

        var conn;
        var accessToken = $("#accessToken");
        var msg = $("#msg");
        var room = $("#room");
        var log = $("#log");

        function appendLog(msg) {
            var d = log[0]
            var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
            msg.appendTo(log)
            if (doScroll) {
                d.scrollTop = d.scrollHeight - d.clientHeight;
            }
        }

        $("#form").submit(function() {
            if (!conn) {
                return false;
            }
            if (!msg.val()) {
                return false;
            }

            conn.send(JSON.stringify({C:"S", R:room.val(), M: msg.val(), T:"M" }));
            msg.val("");
            return false
        });

        $("#form2").submit(function() {
            if (!conn) {
                return false;
            }
            if (!room.val()) {
                return false;
            }
            conn.send(JSON.stringify({C:"J", R:room.val(), M: msg.val()}));
            msg.val("");
            return false
        });

        $("#form3").submit(function() {
            if (!conn) {
                return false;
            }
            if (!room.val()) {
                return false;
            }
            conn.send(JSON.stringify({C:"L", R:room.val(), M: msg.val()}));
            msg.val("");
            return false
        });

        $("#form4").submit(function() {
            if (window["WebSocket"]) {
                conn = new WebSocket("ws://{{$}}/ws?token=" + encodeURIComponent(accessToken.val()));
                conn.onclose = function(evt) {
                    appendLog($("<div><b>Connection closed.</b></div>"))
                }
                conn.onmessage = function(evt) {
                    appendLog($("<div/>").text(evt.data))
                }
            } else {
                appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
            }
            return false
        });
    });
</script>
<style type="text/css">
html {
    overflow: hidden;
}

body {
    overflow: hidden;
    overflow-y: scroll;
    padding: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    background: gray;
}

#log {
    background: white;
    margin: 0;
    padding: 0.5em 0.5em 0.5em 0.5em;
    overflow: auto;
    height: 75%;
    margin: .5em .5em;
}

#wrap-forms{
    padding: 0.5em 0.5em 0.5em 0.5em;
}

form{
    margin-bottom: .5em;
}

#form {
    width: 100%;
    overflow: hidden;
}
#form2 {
    width: 100%;
    overflow: hidden;
}

#form3 {
    width: 100%;
    overflow: hidden;
}
input[type="submit"]{
    width:70px;
}

</style>
</head>
<body>
<div id="log"></div>
<div id="wrap-forms">
    <form id="form">
        <input type="submit" value="Send" />
        <input type="text" id="msg" />
    </form>
    <form id="form2">
        <input type="submit" value="Join" />
        <input type="text" id="room", value="default"/>
    </form>
    <form id="form3">
        <input type="submit" value="Leave"  />
    </form>
    <form id="form4">
        <input type="submit" value="Connect" />
        <input type="text" id="accessToken", value=""/>
    </form>
</div>
</body>
</html>
`
