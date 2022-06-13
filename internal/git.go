package internal

import (
	_ "embed"
	"fmt"
	"io/ioutil"
)

func CheckoutGit(script string, request WorkRequest) error {
	ioutil.WriteFile(fmt.Sprintf("%s/temp.sh", Workspace), []byte(script), 777)
	c := RunCommand("bash", "temp.sh", request.RepositoryUrl, request.Hash)
	b, e := c.Output()
	if e != nil {
		fmt.Println(string(b))
		fmt.Println("=======")
		fmt.Println(e)
		return e
	}
	fmt.Println(string(b))
	RunCommand("rm", "-rf", "temp.sh").Run()
	return nil
}
