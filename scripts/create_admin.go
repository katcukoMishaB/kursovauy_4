package main

import (
	"fmt"
	"kursovauy_4/internal/database"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Использование: go run create_admin.go <email> <password> <first_name> [last_name]")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	firstName := os.Args[3]
	lastName := "Администратор"
	if len(os.Args) > 4 {
		lastName = os.Args[4]
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	var userID string
	err = db.QueryRow(
		"INSERT INTO users (first_name, last_name, email, password, registration_date, status) VALUES ($1, $2, $3, $4, CURRENT_DATE, true) ON CONFLICT (email) DO UPDATE SET password = $4 RETURNING id",
		firstName, lastName, email, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	_, err = db.Exec(
		"INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin) VALUES ($1, true, true, true) ON CONFLICT (user_id) DO UPDATE SET is_admin = true, is_organizer = true",
		userID,
	)
	if err != nil {
		log.Fatal("Failed to create admin role:", err)
	}

	fmt.Printf("Администратор создан успешно!\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("User ID: %s\n", userID)
}

