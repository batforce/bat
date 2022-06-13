package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func UseBuildKit(kit Kit) (*Buildkit, error) {
	archiveUrl := fmt.Sprintf("%s/archive/main.tar.gz", kit.Url)
	log.Printf("Downloading archive %s\n", archiveUrl)
	log.Println("Setting name to buildkit.tar.gz")
	err := DownloadFile(fmt.Sprintf("%s/buildkit.tar.gz", Workspace), archiveUrl)
	if err != nil {
		return nil, err
	}
	buildkit := &Buildkit{
		Url:       kit.Url,
		Dir:       kit.Name,
		Variables: []Variable{},
	}
	log.Println("Extracting buildkit")
	RunCommand("tar", "-xvf", "buildkit.tar.gz").Run()

	return buildkit, nil
}

func (kit *Buildkit) Detect(request WorkRequest) (bool, string) {
	framework, err := kit.run(fmt.Sprintf("%s/scripts/detect", kit.Dir), request)
	fmt.Printf("FRAMEWORK FOUND %s\n", framework)
	if err != nil {
		return false, ""
	}
	return true, framework
}

func (kit *Buildkit) Compile(request WorkRequest) (string, error) {
	stdout, err := kit.run(fmt.Sprintf("%s/scripts/compile", kit.Dir), request)

	if err == nil {
		return stdout, err
	}
	return stdout, err
}

func (kit *Buildkit) PreCompile(request WorkRequest) (string, error) {
	stdout, err := kit.run(fmt.Sprintf("%s/scripts/pre-compile", kit.Dir), request)
	if err == nil {
		return "", err
	}
	return stdout, err
}

func (kit *Buildkit) Release(request WorkRequest) (string, error) {
	stdout, err := kit.run(fmt.Sprintf("%s/scripts/release", kit.Dir), request)
	if err == nil {
		return "", err
	}
	return stdout, err
}

func (kit *Buildkit) run(scriptPath string, request WorkRequest) (string, error) {
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
				_var := strings.Split(strings.Split(stdout, "##[task.setvariable]")[1], "=")
				kit.Variables = append(kit.Variables, Variable{
					Value: _var[1],
					Key:   _var[0],
					Type:  StringVariable,
				})
			}

			for _, val := range request.Variables {
				if val.Type == SecretVariable {
					stdout = strings.ReplaceAll(stdout, val.Value, "__REDACTED_SECRET__")
				}
			}

			sb.WriteString(stdout)
			log.Println(stdout)
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
