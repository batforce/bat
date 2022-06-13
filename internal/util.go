package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var Workspace = "workspace"

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func DownloadFile(filepath string, url string) (err error) {
	log.Printf("Creating file")
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Println("error here")
		return err
	}
	defer out.Close()

	log.Printf("Fetching binary")
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Printf("Copying file")
	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func RunCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Dir = Workspace
	return cmd
}

func CreateWorkspace() error {
	return os.Mkdir(Workspace, 777)
}

func CleanWorkspace() error {
	log.Println("Cleaning workspace")
	return exec.Command("rm", "-rf", Workspace).Run()
}
