package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var server = flag.String("s", "", "server")
var file = flag.String("f", "", "file")

func main() {
	flag.Parse()
	err := m()
	if err != nil {
		log.Println(err)
	}
}

func m() error {
	cfg, err := getClientconfig()
	if err != nil {
		return fmt.Errorf("get client config: %w", err)
	}
	conn, err := ssh.Dial("tcp", *server, cfg)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	version := conn.ServerVersion()
	log.Println(string(version))

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("sftp new client: %w", err)
	}
	defer client.Close()
	filename := "/tmp/" + uuid.New().String()
	src, err := os.Open(*file)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer src.Close()
	dst, err := client.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	err = dst.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()
	session.Stderr = os.Stderr
	session.Stdout = os.Stdout
	log.Println(filename)
	err = session.Run("DISPLAY=:0 xdg-open " + filename)
	if err != nil {
		return fmt.Errorf("exec open: %w", err)
	}
	return nil
}

func getClientconfig() (*ssh.ClientConfig, error) {
	privateKey, err := getPrivateKey()
	if err != nil {

		return nil, fmt.Errorf("get private: %w", err)
	}
	cfg := &ssh.ClientConfig{
		User: "wurongjie",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return cfg, nil
}
func getPrivateKey() (ssh.Signer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}
	b, err := ioutil.ReadFile(filepath.Join(home, ".ssh/id_rsa"))
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}
	privateKey, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return privateKey, nil
}
