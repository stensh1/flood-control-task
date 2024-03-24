// Package server implements the methods and data structures
// responsible for the operation of the HTTP server
// All data structures are defined in a separate models.go file
package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Start constructor creates S logger, fills in configuration and database structures,
// and creates a server instance for the internal *http structure.Server
func (server *S) Start() {
	// New logger object
	server.logger()

	// Loading and parsing configuration data from yaml file
	if err := server.loadConfig(); err != nil {
		server.LogFatal.Panic("Server's configuration uploaded status:", err)
	} else {
		server.LogInfo.Println("Server's configuration uploaded status: OK")
	}

	// Creating a server w/ config
	server.S = &http.Server{
		Addr:    server.c.Server.Host + ":" + server.c.Server.Port,
		Handler: server.newRouter(),
	}

	// Trying to connect Redis
	if err := server.dbConnect(); err != nil {
		server.LogFatal.Panic("Database connection status:", err)
	} else {
		server.LogInfo.Println("Database connection status: OK")
	}

	// Start listening in goroutine
	server.LogInfo.Println("Server is starting on:", server.S.Addr)
	go func() {
		if err := server.S.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				server.LogErr.Panic("Server failed to start:", err)
			}
		}
	}()
}

// logger is a private S method that implements three levels of logging
func (server *S) logger() {
	// Creating a new log file just for each server launch
	f, err := os.OpenFile("logs/"+fmt.Sprint(time.Now().Date())+" "+
		fmt.Sprint(time.Now().Clock())+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to create log file:", err)
		server.LogInfo = log.New(os.Stdout, "INFO:", log.Ldate|log.Ltime)
		server.LogErr = log.New(os.Stdout, "ERROR:", log.Ldate|log.Ltime)
		server.LogFatal = log.New(os.Stdout, "FATAL ERROR:", log.Ldate|log.Ltime)
	} else {
		server.LogInfo = log.New(f, "INFO:\\t", log.Ldate|log.Ltime)
		server.LogErr = log.New(f, "ERROR:\\t", log.Ldate|log.Ltime)
		server.LogFatal = log.New(f, "FATAL ERROR:\\t", log.Ldate|log.Ltime)
	}
}

// loadConfig is a private S method opens the yaml configuration file
// and parse the data from it into the internal Config structure
func (server *S) loadConfig() error {
	// Initialising config object
	server.c = Config{}

	// Reading the configuration .yaml file
	if err := server.c.newConfig("cfg/config.yaml"); err != nil {
		return err
	}

	return nil
}

// dbConnect is a private S method takes values
// from environment variables and yaml config and uses them to create
// a new database connection client
func (server *S) dbConnect() error {
	dbPswd, ok := os.LookupEnv("REDIS_PSWD")
	if !ok {
		err := errors.New("unable to get database password")
		return err
	}

	server.db = redis.NewClient(&redis.Options{
		Addr:     server.c.DB.Host + ":" + server.c.DB.Port,
		Password: dbPswd,
	})

	return nil
}

// Check method of the S increases the counter of the number
// of requests to the server on each call and checks whether
// the limit of requests has been exceeded according to the configuration.
// The check method also uses the Expire method to reset the counters for each user
// after the number of seconds specified in the configuration
func (server *S) Check(ctx context.Context, userID int64) (bool, error) {
	// Increasing the request counter
	server.db.Incr(ctx, strconv.Itoa(int(userID)))

	// Receive the current number of requests to the server from the user
	cnt, err := server.db.Get(ctx, strconv.Itoa(int(userID))).Int()
	if err != nil {
		return false, err
	}

	server.LogInfo.Println("The new request: ", userID, "Total requests count: ", cnt)

	// If this is the first request, then we start the reset counter on the timer
	if cnt == 1 {
		server.db.Expire(ctx, strconv.Itoa(int(userID)), server.c.Server.FloodControl.Time)
	}

	// Checking the condition for exceeding the number of requests
	if cnt > server.c.Server.FloodControl.Requests {
		return false, nil
	}

	return true, nil
}

// newRouter create router and define routes and return that router
func (server *S) newRouter() *http.ServeMux {
	router := http.NewServeMux()

	// root path
	router.HandleFunc("/", server.handler)
	// working path
	router.HandleFunc("/check", server.checker)

	return router
}

// handler handles events by path "/".
// Does nothing, tells the user that you need to go to the work path "/check"
func (server *S) handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Nothing here, go to: /check\n"))
}

// checker is a private S method handles events by path "/check".
// If the client sent the request without cookies, then checker generates cookies and sends them to the client.
// If the client request contains cookies, the Check method is executed.
func (server *S) checker(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	c, err := r.Cookie("cookie")
	if err != nil {
		// Receiving generated cookies
		c = server.giveCookie(w, r)

		// Registering a new connection in Redis
		err := server.db.Set(ctx, c.Value, 0, 0).Err()
		if err != nil {
			server.LogErr.Println("Writing in database status:", err)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Please try to reload page :("))
		}

		server.LogInfo.Println("The new connection registered in redis: ", c.Value)

		return
	}

	// Converting the cookie value to int to match the input parameter of the Check method
	id, err := strconv.Atoi(c.Value)
	if err != nil {
		server.LogErr.Println("Cookie value cannot be translated into int: ", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Please try to reload page :("))
		return
	}

	// Calling a check on the number of requests to the server
	ok, err := server.Check(ctx, int64(id))
	if err != nil {
		server.LogErr.Println("Cannot get data from database, ", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Please try to reload page :("))
		return
	}
	// If the number of requests does not exceed the set value
	if ok {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User cookie: " + c.Value + "\nStatus: OK"))
	} else {
		w.WriteHeader(429)
		w.Write([]byte("Too many requests"))
		server.LogInfo.Println("Too many requests by: ", c.Value)
	}
}

// giveCookie is a S method that just generates some cookie and sen it to client
func (server *S) giveCookie(w http.ResponseWriter, r *http.Request) *http.Cookie {
	c := &http.Cookie{
		Name:   "cookie",
		Value:  r.RemoteAddr[10:],
		MaxAge: 300,
	}

	server.LogInfo.Println("The new cookie sent: ", c)

	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cookies have been successfully sent\n"))

	return c
}

// newConfig is a Config private method that downloads an yaml file and decodes it into a Config structure
func (c *Config) newConfig(configPath string) error {
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("bbb")
		return err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(c); err != nil {
		return err
	}

	return nil
}
