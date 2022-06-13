package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func UseDeployKit(kit Kit) (*DeployKit, error) {
	archiveUrl := fmt.Sprintf("%s/archive/main.tar.gz", kit.Url)
	log.Printf("Downloading archive %s\n", archiveUrl)
	log.Println("Setting name to deploykit.tar.gz")
	err := os.Mkdir(Workspace, 777)
	if err != nil {
		return nil, err
	}
	err = DownloadFile(fmt.Sprintf("%s/deploykit.tar.gz", Workspace), archiveUrl)
	if err != nil {
		return nil, err
	}
	deploykit := &DeployKit{
		Url:       kit.Url,
		Dir:       kit.Name,
		Variables: []Variable{},
	}
	log.Println("Extracting deploykit")
	RunCommand("tar", "-xvf", "deploykit.tar.gz").Run()
	return deploykit, nil
}

func (kit *DeployKit) Detect(request WorkRequest) (bool, string) {
	framework, err := kit.run(fmt.Sprintf("%s/scripts/detect", kit.Dir), request)
	fmt.Printf("FRAMEWORK FOUND %s\n", framework)
	if err != nil {
		return false, ""
	}
	fmt.Println("HERE2")
	return true, framework
}

func (kit *DeployKit) Deploy(request WorkRequest) (string, error) {
	log.Println("Deploying")
	stdout, err := kit.run(fmt.Sprintf("%s/scripts/deploy", kit.Dir), request)
	if err == nil {
		return stdout, err
	}
	return stdout, err
}

func (kit *DeployKit) run(scriptPath string, request WorkRequest) (string, error) {
	kit.Variables = append(kit.Variables, request.Variables...)

	scriptCommand := RunCommand(scriptPath, "app")
	scriptCommand.Env = append(scriptCommand.Env, os.Environ()...)
	detectOut, _ := scriptCommand.StdoutPipe()
	for _, val := range kit.Variables {
		variable := fmt.Sprintf("%s=%s", val.Key, val.Value)
		scriptCommand.Env = append(scriptCommand.Env, variable)
	}
	var sb strings.Builder
	scanner := bufio.NewScanner(detectOut)
	go func() {
		for scanner.Scan() {
			stdout := fmt.Sprintf("%s\n", scanner.Text())

			if strings.Contains(stdout, "##[task.setvariable]") {
				fmt.Printf("Found variable!! %s\n", stdout)
				_var := strings.Split(strings.Split(stdout, "##[task.setvariable]")[1], "=")
				kit.Variables = append(kit.Variables, Variable{
					Value: _var[1],
					Key:   _var[0],
					Type:  StringVariable,
				})
				fmt.Printf("KEY %s\n", _var[0])
				fmt.Printf("VALUE %s\n", _var[1])
			}

			for _, val := range request.Variables {
				if val.Type == SecretVariable {
					stdout = strings.ReplaceAll(stdout, val.Value, "__REDACTED_SECRET__")
				}
			}

			sb.WriteString(stdout)
			fmt.Print(stdout)
		}
	}()

	err := scriptCommand.Start()

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	err = scriptCommand.Wait()

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return sb.String(), err

}
