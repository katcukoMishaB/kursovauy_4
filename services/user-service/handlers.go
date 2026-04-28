package main

import (
	"database/sql"
	"encoding/json"
	"kursovauy_4/internal/auth"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		sendJSONError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	var userID string
	err = db.QueryRow(
		"INSERT INTO users (first_name, last_name, email, phone, password, registration_date) VALUES ($1, $2, $3, $4, $5, CURRENT_DATE) RETURNING id",
		req.FirstName, req.LastName, req.Email, req.Phone, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		log.Printf("Error creating user: %v", err)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			sendJSONError(w, "Email already exists", http.StatusConflict)
		} else {
			sendJSONError(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_, err = db.Exec(
		"INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin) VALUES ($1, true, false, false)",
		userID,
	)
	if err != nil {
		log.Printf("Error creating user role: %v", err)
		sendJSONError(w, "Failed to create user role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": userID, "message": "User registered successfully"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	var hashedPassword string
	err := db.QueryRow(
		"SELECT id, first_name, last_name, email, phone, password, registration_date, status FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &hashedPassword, &user.RegistrationDate, &user.Status)

	if err == sql.ErrNoRows {
		sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Printf("Database error in login: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !user.Status {
		sendJSONError(w, "User account is disabled", http.StatusForbidden)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	var role UserRole
	err = db.QueryRow(
		"SELECT user_id, is_participant, is_organizer, is_admin FROM user_roles WHERE user_id = $1",
		user.ID,
	).Scan(&role.UserID, &role.IsParticipant, &role.IsOrganizer, &role.IsAdmin)

	if err != nil {
		log.Printf("Error getting user role: %v", err)
		sendJSONError(w, "Failed to get user role", http.StatusInternalServerError)
		return
	}

	roleStr := "participant"
	if role.IsAdmin {
		roleStr = "admin"
	} else if role.IsOrganizer {
		roleStr = "organizer"
	}

	token, err := auth.GenerateToken(user.ID, user.Email, roleStr)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		sendJSONError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user":  user,
		"role":  roleStr,
	})
}

func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var user User
	err := db.QueryRow(
		"SELECT id, first_name, last_name, email, phone, registration_date, status FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.RegistrationDate, &user.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting profile: %v", err)
			sendJSONError(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		Password  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec(
			"UPDATE users SET first_name = $1, last_name = $2, phone = $3, password = $4 WHERE id = $5",
			req.FirstName, req.LastName, req.Phone, string(hashedPassword), userID,
		)
		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
	} else {
		_, err := db.Exec(
			"UPDATE users SET first_name = $1, last_name = $2, phone = $3 WHERE id = $4",
			req.FirstName, req.LastName, req.Phone, userID,
		)
		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func createOrganizerRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		sendJSONError(w, "User ID not found in request", http.StatusUnauthorized)
		return
	}

	var req struct {
		ExperienceDescription string `json:"experience_description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		sendJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingRequestID string
	err := db.QueryRow(
		"SELECT id FROM organizer_requests WHERE user_id = $1 AND status = 'в рассмотрении'",
		userID,
	).Scan(&existingRequestID)

	if err == nil {
		sendJSONError(w, "У вас уже есть активная заявка на рассмотрении", http.StatusConflict)
		return
	}
	if err != sql.ErrNoRows {
		log.Printf("Error checking existing request: %v", err)
		sendJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}

	var requestID string
	err = db.QueryRow(
		"INSERT INTO organizer_requests (user_id, experience_description, submission_date, status) VALUES ($1, $2, CURRENT_DATE, 'в рассмотрении') RETURNING id",
		userID, req.ExperienceDescription,
	).Scan(&requestID)

	if err != nil {
		log.Printf("Error creating organizer request: %v", err)
		sendJSONError(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": requestID, "message": "Request created successfully"})
}

func getOrganizerRequestsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT id, user_id, experience_description, submission_date, status, admin_comment FROM organizer_requests ORDER BY submission_date DESC",
	)
	if err != nil {
		log.Printf("Error fetching organizer requests: %v", err)
		sendJSONError(w, "Failed to fetch requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	requests := []OrganizerRequest{}
	for rows.Next() {
		var req OrganizerRequest
		var adminComment sql.NullString
		err := rows.Scan(&req.ID, &req.UserID, &req.ExperienceDesc, &req.SubmissionDate, &req.Status, &adminComment)
		if err != nil {
			log.Printf("Error scanning organizer request: %v", err)
			continue
		}
		if adminComment.Valid {
			req.AdminComment = &adminComment.String
		}
		requests = append(requests, req)
	}

	if requests == nil {
		requests = []OrganizerRequest{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func approveOrganizerRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Transaction error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var userID string
	err = tx.QueryRow("SELECT user_id FROM organizer_requests WHERE id = $1", requestID).Scan(&userID)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	_, err = tx.Exec("UPDATE organizer_requests SET status = 'одобрена' WHERE id = $1", requestID)
	if err != nil {
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE user_roles SET is_organizer = true WHERE user_id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Request approved successfully"})
}

func rejectOrganizerRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]

	var req struct {
		AdminComment string `json:"admin_comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	_, err := db.Exec(
		"UPDATE organizer_requests SET status = 'отклонена', admin_comment = $1 WHERE id = $2",
		req.AdminComment, requestID,
	)
	if err != nil {
		http.Error(w, "Failed to reject request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Request rejected successfully"})
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT id, first_name, last_name, email, phone, registration_date, status FROM users ORDER BY registration_date DESC",
	)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		sendJSONError(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &user.RegistrationDate, &user.Status)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}

	if users == nil {
		users = []User{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func updateUserStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req struct {
		Status bool `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("UPDATE users SET status = $1 WHERE id = $2", req.Status, userID)
	if err != nil {
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User status updated successfully"})
}
