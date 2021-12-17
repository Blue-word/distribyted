package loader

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const m1 = "magnet:?xt=urn:btih:a88fda5954e89178c372716a6a78b8180ed4dad3"

func TestAddUser(t *testing.T) {
	require := require.New(t)

	db, err := NewDB(filepath.Join("../../distribyted-data/metadata", "magnetdb"))

	//tmpService, err := os.MkdirTemp("", "test-service")
	//require.NoError(err)
	//db, err := NewDB(tmpService)
	require.NoError(err)
	defer db.Close()

	user1 := "user1"
	err = db.AddUser(user1, "123456")
	require.NoError(err)
	pass, err := db.GetUserPassword(user1)
	require.NoError(err)
	fmt.Println(pass)

	user2 := "user2"
	err = db.AddUser(user2, "123456")
	require.NoError(err)
	pass, err = db.GetUserPassword(user2)
	require.NoError(err)
	fmt.Println(pass)
}

func TestAddMagnet(t *testing.T) {
	require := require.New(t)

	db, err := NewDB(filepath.Join("../../distribyted-data/metadata", "magnetdb"))
	//tmpService, err := os.MkdirTemp("", "test-service")
	//require.NoError(err)
	//tmpStorage, err := os.MkdirTemp("", "test-service")
	//require.NoError(err)
	//cs := storage.NewFile(tmpStorage)
	//defer cs.Close()
	//db, err := NewDB(tmpService)
	require.NoError(err)
	defer db.Close()

	user1 := "user1"
	err = db.AddMagnet("dir1", m1, user1)
	require.NoError(err)
	err = db.AddMagnet("dirCommon", m1, user1)
	require.NoError(err)
	l, err := db.ListMagnets(user1)
	require.NoError(err)
	require.Len(l, 2)
	require.Len(l["dir1"], 1)
	require.Equal(l["dir1"][0], m1)
	require.Len(l["dirCommon"], 1)
	require.Equal(l["dirCommon"][0], m1)

	user2 := "user2"
	err = db.AddMagnet("dir2", m1, user2)
	require.NoError(err)
	err = db.AddMagnet("dirCommon", m1, user2)
	require.NoError(err)
	l, err = db.ListMagnets(user2)
	require.NoError(err)
	require.Len(l, 2)
	require.Len(l["dir2"], 1)
	require.Equal(l["dir2"][0], m1)
	require.Len(l["dirCommon"], 1)
	require.Equal(l["dirCommon"][0], m1)

}
