package connection_test

import (
	"os"
	"os/exec"
	"testing"

	"auth/internal/config"
	"auth/internal/connection"
)

// TestGetDatabase_ConnectionFailure tests that the GetDatabase function will call log.Fatal
// when it cannot ping the database.
// We use a subprocess to prevent log.Fatal from killing the main test runner.
func TestGetDatabase_ConnectionFailure(t *testing.T) {
	if os.Getenv("TEST_GETDATABASE_CRASHER") == "1" {
		conf := config.Database{
			HOST: "localhost",
			PORT: "99999",
			USER: "dummy",
			PASS: "dummy",
			NAME: "dummy",
			TZ:   "UTC",
		}
		connection.GetDatabase(conf)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestGetDatabase_ConnectionFailure")
	cmd.Env = append(os.Environ(), "TEST_GETDATABASE_CRASHER=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("Subprocess berjalan dengan error %v, ekspektasi program keluar/crash (exit status 1) akibat log.Fatal", err)
}

// TestGetDatabase_Integration is integration test that will actually connect to a real PostgreSQL database. It is skipped unless RUN_INTEGRATION_TESTS is set to 1.
func TestGetDatabase_Integration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set environment variable RUN_INTEGRATION_TESTS=1 untuk menjalankan.")
	}

	// Make sure you have set these environment variables before running the integration test
	conf := config.Database{
		HOST: os.Getenv("DB_HOST"),
		PORT: os.Getenv("DB_PORT"),
		USER: os.Getenv("DB_USER"),
		PASS: os.Getenv("DB_PASS"),
		NAME: os.Getenv("DB_NAME"),
		TZ:   "UTC",
	}

	goquDB, sqlDB := connection.GetDatabase(conf)
	defer sqlDB.Close()

	if goquDB == nil {
		t.Error("Expected goquDB tidak nil")
	}
	if sqlDB == nil {
		t.Error("Expected sqlDB tidak nil")
	}
}
