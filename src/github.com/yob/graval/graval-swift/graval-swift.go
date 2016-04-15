// An example FTP server build on top of graval. graval handles the details
// of the FTP protocol, we just provide a persistence driver for rackspace
// cloud files.
//
// If you're looking to create a custom graval driver, this example is a
// reasonable starting point. I suggest copying this file and changing the
// function bodies as required.
//
// USAGE:
//
//    go get github.com/yob/graval
//    go install github.com/yob/graval/graval-swift
//    export UserName=myusername
//    export ApiKey=myapikey
//    export AuthUrl="https://auth.api.rackspacecloud.com/v1.0"
//    export Container=my-container
//    ./bin/graval-swift
//
package main

import (
	"errors"
	"github.com/ncw/swift"
	"github.com/yob/graval"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A minimal driver for graval that stores all data on rackspace cloudfiles. The
// authentication are ignored, any username and password will work.
//
// This really just exists as a minimal demonstration of the interface graval
// drivers are required to implement.
type SwiftDriver struct{
	connection *swift.Connection
	container  string
	user       string
}

func (driver *SwiftDriver) Authenticate(user string, pass string) bool {
	log.Printf("Authenticate: %s %s", user, pass)
	driver.user = user
	return true
}
func (driver *SwiftDriver) Bytes(path string) (bytes int) {
	path = scoped_path(driver.user, path)
	log.Printf("Bytes: %s", path)
	object, _, err := driver.connection.Object(driver.container, path)
	if err != nil {
		return -1
	}
	return int(object.Bytes)
}
func (driver *SwiftDriver) ModifiedTime(path string) (time.Time, error) {
	path = scoped_path(driver.user, path)
	log.Printf("ModifiedTime: %s", path)
	object, _, err := driver.connection.Object(driver.container, path)
	if err != nil {
		return time.Now(), err
	}
	return object.LastModified, nil
}
func (driver *SwiftDriver) ChangeDir(path string) bool {
	path = scoped_path(driver.user, path)
	if path == ("/"+driver.user) {
		return true
	}
	log.Printf("ChangeDir: %s", path)
	object, _, err := driver.connection.Object(driver.container, path)
	if err != nil {
		return false
	}
	return object.ContentType == "application/directory"
}
func (driver *SwiftDriver) DirContents(path string) (files []os.FileInfo) {
	path = scoped_path_with_trailing_slash(driver.user, path)
	log.Printf("DirContents: %s", path)
	opts    := &swift.ObjectsOpts{Prefix:path, Delimiter:'/'}
	objects, err := driver.connection.ObjectsAll(driver.container, opts)
	if err != nil {
		return // error connecting to cloud files
	}
	for _, object := range objects {
		tail     := strings.Replace(object.Name, path, "", 1)
        basename := strings.Split(tail, "/")[0]
		if object.ContentType == "application/directory" && object.SubDir == "" {
			files = append(files, graval.NewDirItem(basename))
		} else if object.ContentType != "application/directory"  {
			files = append(files, graval.NewFileItem(basename, int(object.Bytes)))
		}
	}
	return
}

func (driver *SwiftDriver) DeleteDir(path string) bool {
	path = scoped_path(driver.user, path)
	log.Printf("DeleteDir: %s", path)
	err := driver.connection.ObjectDelete(driver.container, path)
	if err != nil {
		return false
	}
	return true
}
func (driver *SwiftDriver) DeleteFile(path string) bool {
	path = scoped_path(driver.user, path)
	log.Printf("DeleteFile: %s", path)
	err := driver.connection.ObjectDelete(driver.container, path)
	if err != nil {
		return false
	}
	return true
}
func (driver *SwiftDriver) Rename(fromPath string, toPath string) bool {
	fromPath = scoped_path(driver.user, fromPath)
	toPath   = scoped_path(driver.user, toPath)
	log.Printf("Rename: %s %s", fromPath, toPath)
	return false
}
func (driver *SwiftDriver) MakeDir(path string) bool {
	path = scoped_path(driver.user, path)
	log.Printf("MakeDir: %s", path)
	opts    := &swift.ObjectsOpts{Prefix:path}
	objects, err := driver.connection.ObjectNames(driver.container, opts)
	if err != nil {
		return false // error connection to cloud files
	}
	if len(objects) > 0 {
		return false // the dir already exists
	}
	driver.connection.ObjectPutString(driver.container, path, "", "application/directory")
	return true
}
func (driver *SwiftDriver) GetFile(path string) (data string, err error) {
	path = scoped_path(driver.user, path)
	log.Printf("GetFile: %s", path)
	data, err = driver.connection.ObjectGetString(driver.container, path)
	if err != nil {
		return "", errors.New("foo")
	}
	return
}
func (driver *SwiftDriver) PutFile(destPath string, data io.Reader) bool {
	destPath = scoped_path(driver.user, destPath)
	log.Printf("PutFile: %s", destPath)
	contents, err := ioutil.ReadAll(data)
	if err != nil {
		return false
	}
	err = driver.connection.ObjectPutBytes(driver.container, destPath, contents, "application/octet-stream")
	if err != nil {
		return false
	}
	return true
}

func scoped_path_with_trailing_slash(user string, path string) string {
	path = scoped_path(user, path)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	if path == "/" {
		return ""
	}
	return path
}

func scoped_path(user string, path string) string {
	if path == "/" {
		path = ""
	}
	return filepath.Join("/", user, path)
}

// graval requires a factory that will create a new driver instance for each
// client connection. Generally the factory will be fairly minimal. This is
// a good place to read any required config for your driver.
type SwiftDriverFactory struct{}

func (factory *SwiftDriverFactory) NewDriver() (graval.FTPDriver, error) {
	driver := &SwiftDriver{}
	driver.container  = os.Getenv("Container")
	if driver.container == "" {
		return nil, errors.New("Container env variable not set")
	}
	driver.connection = &swift.Connection{
		UserName: os.Getenv("UserName"),
		ApiKey:   os.Getenv("ApiKey"),
		AuthUrl:  os.Getenv("AuthUrl"),
	}
	err := driver.connection.Authenticate()
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// it's alive!
func main() {
	factory := &SwiftDriverFactory{}
	ftpServer := graval.NewFTPServer(&graval.FTPServerOpts{ Factory: factory })
	err := ftpServer.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server!")
		log.Fatal(err)
	}
}
