package client

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/lzambarda/hbt/internal"
)

func send(msg string, read bool) (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:"+internal.Port)
	if err != nil {
		return "", err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	if err != nil {
		return "", err
	}
	err = conn.CloseWrite()
	if err != nil {
		return "", err
	}
	if !read {
		return "", nil
	}
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	return string(buffer), err
}

func SendStop() error {
	return exec.Command("sh", "-c", "kill -TERM $(pgrep hbtsrv)").Run()
}

func SendTrack(args []string) error {
	_, err := send(strings.Join(append([]string{"track"}, args...), "\n"), false)
	return err
}

func SendHint(args []string) error {
	msg, err := send(strings.Join(append([]string{"hint"}, args...), "\n"), true)
	if err != nil {
		return err
	}
	fmt.Print(msg)
	return nil
}

func SendEnd(args []string) error {
	_, err := send(strings.Join(append([]string{"end"}, args...), "\n"), false)
	return err
}
