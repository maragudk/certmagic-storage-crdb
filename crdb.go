package certmagic_storage_crdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/caddyserver/certmagic"
	_ "github.com/lib/pq"
)

const (
	defaultQueryTimeout = 3 * time.Second
)

// CRDBStorage implements certmagic.Storage.
type CRDBStorage struct {
	DB          *sql.DB
	user        string
	host        string
	port        int
	database    string
	cert        string
	key         string
	rootCert    string
	lockTimeout time.Duration
}

var _ certmagic.Storage = (*CRDBStorage)(nil)

// Options for New.
type Options struct {
	User        string
	Host        string
	Port        int
	Database    string
	Cert        string
	Key         string
	RootCert    string
	LockTimeout time.Duration
}

// New CRDBStorage.
func New(options Options) *CRDBStorage {
	if options.LockTimeout == 0 {
		options.LockTimeout = time.Minute
	}
	return &CRDBStorage{
		user:        options.User,
		host:        options.Host,
		port:        options.Port,
		database:    options.Database,
		cert:        options.Cert,
		key:         options.Key,
		rootCert:    options.RootCert,
		lockTimeout: options.LockTimeout,
	}
}

func (s *CRDBStorage) Connect(ctx context.Context) error {
	dataSourceName := fmt.Sprintf("postgresql://%v@%v:%v/%v?", s.user, s.host, s.port, s.database)
	if s.cert != "" && s.key != "" && s.rootCert != "" {
		dataSourceName += fmt.Sprintf("sslmode=verify-full&sslcert=%v&sslkey=%v&sslrootcert=%v", s.cert, s.key, s.rootCert)
	} else {
		dataSourceName += "sslmode=disable"
	}

	var err error
	s.DB, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("could not open connection: %w", err)
	}

	if err := s.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("could not ping: %w", err)
	}

	return nil
}

func (s *CRDBStorage) Lock(ctx context.Context, key string) error {
	err := s.inTransaction(ctx, func(tx *sql.Tx) error {
		var locked bool
		row := tx.QueryRowContext(ctx, `select exists (select 1 from certmagic_locks where "key" = $1 and expires > now());`, key)
		if err := row.Scan(&locked); err != nil {
			return fmt.Errorf("could not check lock for key %v: %v", key, err)
		}

		if locked {
			return fmt.Errorf("a lock exists for key %v", key)
		}

		if _, err := tx.ExecContext(ctx, `upsert into certmagic_locks ("key", expires) values ($1, (now() + $2)::timestamp)`, key, s.lockTimeout.String()); err != nil {
			return fmt.Errorf("could not acquire lock for key %v: %v", key, err)
		}
		return nil
	})
	return err
}

func (s *CRDBStorage) Unlock(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `delete from certmagic_locks where key = $1`, key)
	return err
}

func (s *CRDBStorage) Store(key string, value []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `upsert into certmagic_values ("key", "value", updated) values ($1, $2, now())`, key, value)
	return err
}

func (s *CRDBStorage) Load(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	var v []byte
	row := s.DB.QueryRowContext(ctx, `select "value" from certmagic_values where "key" = $1`, key)
	if err := row.Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// TODO is this correct, or should it be an error?
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

func (s *CRDBStorage) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	_, err := s.DB.ExecContext(ctx, `delete from certmagic_values where "key" = $1`, key)
	return err
}

func (s *CRDBStorage) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	var exists bool
	row := s.DB.QueryRowContext(ctx, `select exists (select 1 from certmagic_values where "key" = $1)`, key)
	if err := row.Scan(&exists); err != nil {
		return false
	}
	return exists
}

func (s *CRDBStorage) List(prefix string, recursive bool) ([]string, error) {
	panic("implement me")
}

func (s *CRDBStorage) Stat(key string) (certmagic.KeyInfo, error) {
	// TODO if the key doesn't exist, is that an error?

	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout)
	defer cancel()
	row := s.DB.QueryRowContext(ctx, `select length(value), updated from certmagic_values where "key" = $1`, key)
	info := certmagic.KeyInfo{
		Key:        key,
		IsTerminal: true, // TODO
	}
	if err := row.Scan(&info.Size, &info.Modified); err != nil {
		return info, err
	}
	return info, nil
}

// inTransaction runs callback in a transaction, and makes sure to handle rollbacks, commits etc.
func (s *CRDBStorage) inTransaction(ctx context.Context, callback func(tx *sql.Tx) error) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	if err := callback(tx); err != nil {
		return rollback(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

// rollback a transaction, handling both the original error and any transaction rollback errors.
func rollback(tx *sql.Tx, err error) error {
	if txErr := tx.Rollback(); txErr != nil {
		return fmt.Errorf("could not rollback transaction after error (transaction error: %v)", txErr)
	}
	return err
}
