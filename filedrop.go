package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/studio-b12/gowebdav"
	"golang.org/x/term"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DavConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var c *gowebdav.Client

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randString(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func connectDav() {
	var davConfig DavConfig
	configDir, err := os.UserHomeDir()
	if err != nil {
		panic("read home dir failed")
	}
	configFile, err := os.ReadFile(filepath.Join(configDir, ".filedrop.json"))
	if err != nil {
		panic("config file ~/.filedrop.json not exist")
	}
	err = json.Unmarshal(configFile, &davConfig)
	if err != nil {
		panic("unmarshal config file failed")
	}
	c = gowebdav.NewClient(davConfig.URL, davConfig.Username, davConfig.Password)
	err = c.Connect()
	if err != nil {
		panic("connect to dav server failed")
	}
	makeStorage()
}

func makeStorage() {
	_, err := c.Stat("/filedrop")
	if err != nil {
		c.Mkdir("/filedrop", 0644)
	}
}

func createShareCode() string {
	files, _ := c.ReadDir("/filedrop")
	if len(files) == 0 {
		return randString(6)
	}
	for {
		code := randString(6)
		for _, file := range files {
			if !strings.HasPrefix(file.Name(), code) {
				return code
			}
		}
	}
}

func uploadFile(filename string) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic("file not exist")
	}
	code := createShareCode()
	filenameEncoded := base64.RawURLEncoding.EncodeToString([]byte(filepath.Base(filename)))
	path := "/filedrop/" + code + "," + filenameEncoded
	fmt.Printf("uploading file %s, share code %s\n", filepath.Base(filename), code)
	err = c.WriteStream(path, file, 0644)
	if err != nil {
		panic("upload file failed")
	}
	fmt.Printf("uploaded\n")
}

func downloadFile(filename string) {
	stream, _ := c.ReadStream("/filedrop/" + filename)
	defer stream.Close()
	filenameEncoded := filename[7:]
	filenameBytes, _ := base64.RawURLEncoding.DecodeString(filenameEncoded)
	filename = path.Base(string(filenameBytes))
	fmt.Printf("downloading file %s\n", filename)
	file, err := os.Create(filename)
	if err != nil {
		panic("create file failed")
	}
	defer file.Close()
	_, err = io.Copy(file, stream)
	if err != nil {
		panic("write file failed")
	}
	fmt.Printf("downloaded\n")
}

func downloadLatestFile() {
	files, _ := c.ReadDir("/filedrop")
	latest := time.Time{}
	latestFilename := ""
	for _, file := range files {
		if latest.Before(file.ModTime()) {
			latest = file.ModTime()
			latestFilename = file.Name()
		}
	}
	downloadFile(latestFilename)
}

func downloadFileWithCode(code string) {
	files, _ := c.ReadDir("/filedrop")
	for _, file := range files {
		if file.Name()[:6] == code {
			downloadFile(file.Name())
			return
		}
	}
	fmt.Printf("file not found with code %s\n", code)
}

func makeConfig() {
	configDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	davConfig := DavConfig{}
	fmt.Printf("dav server url: ")
	fmt.Scanln(&davConfig.URL)
	fmt.Printf("username: ")
	fmt.Scanln(&davConfig.Username)
	fmt.Printf("password: ")
	term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Scanln(&davConfig.Password)
	configFile, _ := json.Marshal(davConfig)
	err = os.WriteFile(filepath.Join(configDir, ".filedrop.json"), configFile, 0644)
	if err != nil {
		panic("write config file failed")
	}
	fmt.Printf("saved at %s\n", filepath.Join(configDir, ".filedrop.json"))
}

func list() {
	files, _ := c.ReadDir("/filedrop")
	fmt.Printf("code\tdate\tfilename\n")
	for _, file := range files {
		rawFilename := file.Name()
		filenameEncoded := rawFilename[7:]
		filenameBytes, _ := base64.RawURLEncoding.DecodeString(filenameEncoded)
		filename := path.Base(string(filenameBytes))
		uploadDate := file.ModTime().Format("01/02")
		fmt.Printf("%s\t%s\t%s\n", rawFilename[:6], uploadDate, filename)
	}
}
func prune() {
	pruneTime := time.Now().Add(-time.Hour * 24)
	files, _ := c.ReadDir("/filedrop")
	for _, file := range files {
		if file.ModTime().Before(pruneTime) {
			c.Remove("/filedrop/" + file.Name())
			fmt.Printf("remove %s\n", file.Name()[:6])
		}
	}
}

func usage() {
	println("usage:")
	println(os.Args[0] + " up filename")
	println(os.Args[0] + " down [code]")
	println(os.Args[0] + " list")
	println(os.Args[0] + " prune")
	println(os.Args[0] + " config")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR: ", r)
		}
	}()

	if len(os.Args) == 2 {
		if os.Args[1] == "config" {
			makeConfig()
		} else if os.Args[1] == "down" {
			connectDav()
			downloadLatestFile()
		} else if os.Args[1] == "list" {
			connectDav()
			list()
		} else if os.Args[1] == "prune" {
			choice := ""
			fmt.Printf("are you sure you want to remove files 24 hours ago? [y/N]")
			fmt.Scanln(&choice)
			if choice == "y" || choice == "Y" {
				connectDav()
				prune()
			}
		} else {
			usage()
		}
		return
	}

	if len(os.Args) == 3 {
		if os.Args[1] == "up" {
			connectDav()
			uploadFile(os.Args[2])
		} else if os.Args[1] == "down" {
			connectDav()
			downloadFileWithCode(os.Args[2])
		} else {
			usage()
		}
		return
	}

	usage()
}
