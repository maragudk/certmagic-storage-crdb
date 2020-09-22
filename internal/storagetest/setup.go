package storagetest

import (
	"bufio"
	"context"
	"database/sql"
	"os"
	"strings"
	"time"

	crdb "github.com/maragudk/certmagic-storage-crdb"
)

const port = 26257

func createStorage(user string, timeout time.Duration) *crdb.CRDBStorage {
	return crdb.New(crdb.Options{
		User:        user,
		Host:        "localhost",
		Port:        port,
		Database:    "certmagic",
		LockTimeout: timeout,
	})
}

func CreateStorage() (*crdb.CRDBStorage, func()) {
	rootStorer := setupDB()
	s := createStorage("certmagic", 0)
	if err := s.Connect(context.Background()); err != nil {
		panic(err)
	}

	return s, func() {
		dropDB(rootStorer)
	}
}

func CreateStorageWithLockTimeout(t time.Duration) (*crdb.CRDBStorage, func()) {
	rootStorer := setupDB()
	s := createStorage("certmagic", t)
	if err := s.Connect(context.Background()); err != nil {
		panic(err)
	}

	return s, func() {
		dropDB(rootStorer)
	}
}

// setupDB with root privileges.
func setupDB() *crdb.CRDBStorage {
	s := createStorage("root", 0)
	if err := s.Connect(context.Background()); err != nil {
		panic(err)
	}

	executeSQLFromFile(s.DB, "internal/storagetest/testdata/drop-database.sql")
	executeSQLFromFile(s.DB, "internal/storagetest/testdata/create-database.sql")
	executeSQLFromFile(s.DB, "tables.sql")

	return s
}

func dropDB(s *crdb.CRDBStorage) {
	executeSQLFromFile(s.DB, "internal/storagetest/testdata/drop-database.sql")
}

func executeSQLFromFile(db *sql.DB, path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := ""
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments
		if strings.HasPrefix(line, "--") {
			continue
		}
		query += line + " "
		if !strings.HasSuffix(query, "; ") {
			continue
		}
		_, err := db.ExecContext(ctx, query)
		query = ""
		if err != nil {
			panic(err)
		}
	}
}
