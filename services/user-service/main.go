package main

import (
	"database/sql"
	"encoding/json"
	"kursovauy_4/internal/database"
	"kursovauy_4/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

var db *sql.DB

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func main() {
	gormDB, err := database.ConnectGORM()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db, err = gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get sql db:", err)
	}
	defer db.Close()

	r := gin.Default()
	r.POST("/register", adapt(registerHandler))
	r.POST("/login", adapt(loginHandler))
	r.GET("/profile", adapt(middleware.AuthMiddleware(getProfileHandler)))
	r.PUT("/profile", adapt(middleware.AuthMiddleware(updateProfileHandler)))
	r.POST("/organizer-request", adapt(middleware.AuthMiddleware(createOrganizerRequestHandler)))
	r.GET("/organizer-requests", adapt(middleware.AdminMiddleware(getOrganizerRequestsHandler)))
	r.POST("/organizer-requests/:id/approve", adaptWithParams(middleware.AdminMiddleware(approveOrganizerRequestHandler), "id"))
	r.POST("/organizer-requests/:id/reject", adaptWithParams(middleware.AdminMiddleware(rejectOrganizerRequestHandler), "id"))
	r.GET("/users", adapt(middleware.AdminMiddleware(getUsersHandler)))
	r.PUT("/users/:id", adaptWithParams(middleware.AdminMiddleware(updateUserStatusHandler), "id"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("User service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func adapt(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c.Writer, c.Request)
	}
}

func adaptWithParams(handler http.HandlerFunc, params ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		vars := map[string]string{}
		for _, param := range params {
			vars[param] = c.Param(param)
		}
		req := mux.SetURLVars(c.Request, vars)
		handler(c.Writer, req)
	}
}
