package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/aleh/docode-waf/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("=== WAF Admin Password Reset Tool ===")
	fmt.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sqlx.Connect(cfg.Database.Driver, cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Get username/email
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter admin username or email: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		log.Fatal("Username/email cannot be empty")
	}

	// Check if admin exists
	var adminID string
	var adminEmail string
	err = db.QueryRow("SELECT id, email FROM admins WHERE username = $1 OR email = $1", username).Scan(&adminID, &adminEmail)
	if err != nil {
		log.Fatalf("Admin not found with username/email: %s", username)
	}

	fmt.Printf("Found admin: %s (ID: %s)\n", adminEmail, adminID)

	// Get new password
	fmt.Print("Enter new password: ")
	password1, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	fmt.Println()

	fmt.Print("Confirm new password: ")
	password2, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	fmt.Println()

	if string(password1) != string(password2) {
		log.Fatal("Passwords do not match")
	}

	if len(password1) < 6 {
		log.Fatal("Password must be at least 6 characters")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword(password1, bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Update password
	_, err = db.Exec("UPDATE admins SET password = $1, updated_at = NOW() WHERE id = $2", string(hashedPassword), adminID)
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Password updated successfully!")
}
