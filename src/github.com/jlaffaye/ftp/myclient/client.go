package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/jlaffaye/ftp"
)

func retrieveFile(file string, serverID string) {
	fname := "files/" + file
	c, err := ftp.Connect(serverID)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Quit()

	c.Login("test", "1234")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()
	r, err := c.Retr(fname)
	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(os.Args[2], buf, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
func main() {
	serverID := "127.0.0.1:" + os.Args[1]
	fname := "files/" + os.Args[2]
	c, err := ftp.Connect(serverID)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Quit()

	c.Login("test", "1234")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()
	r, err := c.Loc(fname)
	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(buf))
	serverID = "127.0.0.1:" + string(buf)
	retrieveFile(os.Args[2], serverID)

}
