package utils

import (
	"database/sql"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/xerrors"
)

var (
	System = InitializeSystemFunctions()
)

var PostgresFiles = []string{"postgresql.conf", "PG_VERSION"}

/*
 * SystemFunctions holds function pointers for built-in functions that will need
 * to be mocked out for unit testing.  All built-in functions manipulating the
 * filesystem, shell, or environment should ideally be called through a function
 * pointer in System (the global SystemFunctions variable) instead of being called
 * directly.
 */

type SystemFunctions struct {
	CurrentUser  func() (*user.User, error)
	Getenv       func(key string) string
	Getpid       func() int
	Hostname     func() (string, error)
	IsNotExist   func(err error) bool
	MkdirAll     func(path string, perm os.FileMode) error
	Now          func() time.Time
	Open         func(name string) (*os.File, error)
	OpenFile     func(name string, flag int, perm os.FileMode) (*os.File, error)
	Remove       func(name string) error
	RemoveAll    func(name string) error
	Rename       func(oldpath, newpath string) error
	ReadFile     func(filename string) ([]byte, error)
	WriteFile    func(filename string, data []byte, perm os.FileMode) error
	Stat         func(name string) (os.FileInfo, error)
	FilePathGlob func(pattern string) ([]string, error)
	Create       func(name string) (*os.File, error)
	Mkdir        func(name string, perm os.FileMode) error
	SqlOpen      func(driverName, dataSourceName string) (*sql.DB, error)
	Close        func(f *os.File) error
}

func InitializeSystemFunctions() *SystemFunctions {
	return &SystemFunctions{
		CurrentUser:  user.Current,
		Getenv:       os.Getenv,
		Getpid:       os.Getpid,
		Hostname:     os.Hostname,
		IsNotExist:   os.IsNotExist,
		MkdirAll:     os.MkdirAll,
		Now:          time.Now,
		Open:         os.Open,
		OpenFile:     os.OpenFile,
		Remove:       os.Remove,
		RemoveAll:    os.RemoveAll,
		Rename:       os.Rename,
		Stat:         os.Stat,
		FilePathGlob: filepath.Glob,
		ReadFile:     ioutil.ReadFile,
		WriteFile:    ioutil.WriteFile,
		Create:       os.Create,
		Mkdir:        os.Mkdir,
		SqlOpen:      sql.Open,
		Close:        func(f *os.File) error { return f.Close() },
	}
}

func TryEnv(varname string, defval string) string {
	val := System.Getenv(varname)
	if val == "" {
		return defval
	}
	return val
}

func GetUser() (string, string, error) {
	currentUser, err := System.CurrentUser()
	if err != nil {
		return "", "", err
	}
	return currentUser.Username, currentUser.HomeDir, err
}

func GetHost() (string, error) {
	hostname, err := System.Hostname()
	return hostname, err
}

func GetStateDir() string {
	stateDir := os.Getenv("GPUPGRADE_HOME")
	if stateDir == "" {
		stateDir = filepath.Join(os.Getenv("HOME"), ".gpupgrade")
	}

	return stateDir
}

// AddEmptyFileIdempotent returns nil if the file argument was created by this call or
//   if that file already existed before this call was made.  Otherwise,
//   it returns an error.
func AddEmptyFileIdempotent(file string) error {
	f, err := System.OpenFile(file, os.O_CREATE|os.O_EXCL, 0700)
	if err != nil {
		switch x := err.(type) {
		case *os.PathError:
			if xerrors.Is(x.Err, syscall.EEXIST) {
				return nil
			}
			return err
		default:
			return err
		}
	}

	return System.Close(f)
}

// DoesPathExist returns true if the path argument can be successfully accessed
//  by the caller and false otherwise.
func DoesPathExist(path string) bool {
	_, err := System.Stat(path)
	return err == nil
}
