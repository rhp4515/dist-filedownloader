package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yob/graval"
)

/*const (
	fileOne = "This is the first file available for download.\n\nBy Jàmes"
	fileTwo = "This is file number two.\n\n2012-12-04"
)*/

type MemDriver struct {
	fIndex map[string]string
}

func (driver *MemDriver) LoadFIndex() (err error) {
	fData, err := ioutil.ReadFile("fs")
	if err != nil {
		return err
	}
	driver.fIndex = make(map[string]string)

	files := strings.Split(string(fData), "\n")
	for index, _ := range files {
		parts := strings.Split(files[index], "\t")
		parts[1] = strings.TrimSuffix(parts[1], "\n")
		driver.fIndex[parts[0]] = parts[1]
	}
	return nil
}

func (driver *MemDriver) Authenticate(user string, pass string) bool {
	return user == "test" && pass == "1234"
}

func (driver *MemDriver) Bytes(path string) (bytes int) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(fi.Size())
	return int(fi.Size())
}
func (driver *MemDriver) ModifiedTime(path string) (time.Time, error) {
	return time.Now(), nil
}
func (driver *MemDriver) ChangeDir(path string) bool {
	return path == "/" || path == "/files"
}
func (driver *MemDriver) DirContents(path string) (files []os.FileInfo) {
	// files = []os.FileInfo{}
	// switch path {
	// case "/":
	// 	files = append(files, graval.NewDirItem("files"))
	// 	files = append(files, graval.NewFileItem("one.txt", len(fileOne)))
	// case "/files":
	// 	files = append(files, graval.NewFileItem("two.txt", len(fileOne)))
	// }
	// return files
	return nil
}

func (driver *MemDriver) DeleteDir(path string) bool {
	return false
}
func (driver *MemDriver) DeleteFile(path string) bool {
	return false
}
func (driver *MemDriver) Rename(fromPath string, toPath string) bool {
	return false
}
func (driver *MemDriver) MakeDir(path string) bool {
	return false
}
func (driver *MemDriver) GetFile(path string) (data string, err error) {
	// fmt.Println("getfile", path)
	fData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(fData), nil
}

func (driver *MemDriver) LocateFile(path string) (port string) {
	// log.Println(path)
	// log.Printf("%+v", driver.fIndex)
	port, exist := driver.fIndex[path]
	if !exist {
		port = "-1"
	}
	return
}

func (driver *MemDriver) PutFile(destPath string, data io.Reader) bool {
	return false
}

type MemDriverFactory struct{}

func (factory *MemDriverFactory) NewDriver() (graval.FTPDriver, error) {
	memDriver := &MemDriver{}
	err := memDriver.LoadFIndex()

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return memDriver, nil
}

func main() {
	factory := &MemDriverFactory{}
	if port, err := strconv.Atoi(os.Args[1]); err == nil {
		ftpServer := graval.NewFTPServer(&graval.FTPServerOpts{Factory: factory, Port: port})
		err := ftpServer.ListenAndServe()
		if err != nil {
			log.Print(err)
			log.Fatal("Error starting server!")
		}
	}
}
