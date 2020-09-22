package certmagic_storage_crdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/maragudk/certmagic-storage-crdb/internal/storagetest"
	"github.com/stretchr/testify/require"
)

func TestCRDBStorage_Connect(t *testing.T) {
	t.Run("connects with no error", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Connect(context.Background())
		require.NoError(t, err)
	})
}

func TestCRDBStorage_Lock(t *testing.T) {
	t.Run("locks on a key", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Lock(context.Background(), "test")
		require.NoError(t, err)
	})

	t.Run("errors on duplicate keys", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Lock(context.Background(), "test")
		require.NoError(t, err)

		err = s.Lock(context.Background(), "test")
		require.Error(t, err)
	})

	t.Run("does not error on duplicate keys if expired", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorageWithLockTimeout(time.Microsecond)
		defer cleanup()

		err := s.Lock(context.Background(), "test")
		require.NoError(t, err)

		err = s.Lock(context.Background(), "test")
		require.NoError(t, err)
	})
}

func TestCRDBStorage_Unlock(t *testing.T) {
	t.Run("unlocks", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Lock(context.Background(), "test")
		require.NoError(t, err)

		err = s.Unlock("test")
		require.NoError(t, err)
	})
}

func TestCRDBStorage_Load(t *testing.T) {
	t.Run("returns nil if no such key", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		value, err := s.Load("test")
		require.NoError(t, err)
		require.Nil(t, value)
	})

	t.Run("returns value if stored", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)

		value, err := s.Load("test")
		require.NoError(t, err)
		require.Equal(t, []byte("hat"), value)
	})
}

func TestCRDBStorage_Store(t *testing.T) {
	t.Run("stores a key-value pair", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)
	})

	t.Run("stores a key-value pair on subsequent requests", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)

		err = s.Store("test", []byte("partyhat"))
		require.NoError(t, err)

		value, err := s.Load("test")
		require.NoError(t, err)
		require.Equal(t, []byte("partyhat"), value)
	})
}

func TestCRDBStorage_Delete(t *testing.T) {
	t.Run("does not error on no such key", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Delete("test")
		require.NoError(t, err)
	})

	t.Run("deletes the value at key", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)

		err = s.Delete("test")
		require.NoError(t, err)

		value, err := s.Load("test")
		require.NoError(t, err)
		require.Nil(t, value)
	})
}

func TestCRDBStorage_Exists(t *testing.T) {
	t.Run("returns false if key does not exist", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		exists := s.Exists("test")
		require.False(t, exists)
	})

	t.Run("returns true if key does exist", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)

		exists := s.Exists("test")
		require.True(t, exists)
	})
}

func TestCRDBStorage_List(t *testing.T) {

}

func TestCRDBStorage_Stat(t *testing.T) {
	t.Run("returns info about the value stored at key", func(t *testing.T) {
		s, cleanup := storagetest.CreateStorage()
		defer cleanup()

		err := s.Store("test", []byte("hat"))
		require.NoError(t, err)

		info, err := s.Stat("test")
		require.NoError(t, err)
		require.Equal(t, "test", info.Key)
		require.WithinDuration(t, time.Now(), info.Modified, time.Second)
		require.Equal(t, int64(3), info.Size)
		require.True(t, info.IsTerminal)
	})
}
