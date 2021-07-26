package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"fmt"

	// log "github.com/sirupsen/logrus"
    // "github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	// "github.com/newrelic/go-agent/v3/newrelic"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v3"
	"github.com/newrelic/go-agent/v3/newrelic"
)

var (
	// ErrHttpGenericMessage that is returned in general case, details should be logged in such case
	ErrHttpGenericMessage = echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")

	// ErrWrongCredentials indicates that login attempt failed because of incorrect login or password
	ErrWrongCredentials = echo.NewHTTPError(http.StatusUnauthorized, "username or password is invalid")

	jwtSecret = "myfancysecret"
)

func main() {

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("auth-api-todo-app"),
		newrelic.ConfigLicense("dd4f396711488259fbc92a1a5af076b8a661NRAL"),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		log.Printf("New Relic error: %s", err.Error())
	}

	hostport := ":" + os.Getenv("AUTH_API_PORT")
	userAPIAddress := os.Getenv("USERS_API_ADDRESS")

	envJwtSecret := os.Getenv("JWT_SECRET")
	if len(envJwtSecret) != 0 {
		jwtSecret = envJwtSecret
	}

	userService := UserService{
		Client:         http.DefaultClient,
		UserAPIAddress: userAPIAddress,
		AllowedUserHashes: map[string]interface{}{
			"admin_admin": nil,
			"johnd_foo":   nil,
			"janed_ddd":   nil,
		},
	}

	// Echo instance
	e := echo.New()

	// The New Relic Middleware should be the first middleware registered
	e.Use(nrecho.Middleware(app))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Route => handler
	e.GET("/version", func(c echo.Context) error {
		return c.String(http.StatusOK, "Auth API, written in Go\n")
	})

	e.POST("/login", getLoginHandler(userService))

	// Start server
	e.Logger.Fatal(e.Start(hostport))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getLoginHandler(userService UserService) echo.HandlerFunc {
	f := func(c echo.Context) error {

		log.Printf("Create log")
		// logger := log.New()
		// logger.SetFormatter(nrlogrusplugin.ContextFormatter{})
		// ctx := newrelic.NewContext(context.Background(), txn)
		// testing := logger.WithContext(ctx).Info("request-login!")
		// log.Printf("%+v\n", testing)

		txn := nrecho.FromContext(c)
		segment := txn.StartSegment("request-login")
		requestData := LoginRequest{}
		decoder := json.NewDecoder(c.Request().Body)
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("could not read credentials from POST body: %s", err.Error())
			return ErrHttpGenericMessage
		}
		segment.End()

		txn.AddAttribute("username", requestData.Username)
		segment2 := txn.StartSegment("login")
		ctx := c.Request().Context()

		// log.Printf("%+v\n", ctx)
		user, err := userService.Login(ctx, requestData.Username, requestData.Password, txn)
		if err != nil {
			if err != ErrWrongCredentials {
				log.Printf("could not authorize user '%s': %s", requestData.Username, err.Error())
				return ErrHttpGenericMessage
			}

			return ErrWrongCredentials
		}
		token := jwt.New(jwt.SigningMethodHS256)
		segment2.End()

		segment3 := txn.StartSegment("generate-sent-token")
		claims := token.Claims.(jwt.MapClaims)
		claims["username"] = user.Username
		claims["firstname"] = user.FirstName
		claims["lastname"] = user.LastName
		claims["role"] = user.Role
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			log.Printf("could not generate a JWT token: %s", err.Error())
			return ErrHttpGenericMessage
		}
		segment3.End()

		return c.JSON(http.StatusOK, map[string]string{
			"accessToken": t,
		})
	}

	return echo.HandlerFunc(f)
}
