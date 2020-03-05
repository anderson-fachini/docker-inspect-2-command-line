package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DockerConfig struct {
	Name   string
	Config struct {
		Env      []string
		Hostname string
		Image    string
	}
	HostConfig struct {
		AutoRemove bool
		Binds      []string
		Dns        []string
		Links      []string
		LogConfig  struct {
			Config map[string]string
		}
		Memory       int
		NanoCpus     int
		PortBindings map[string][]struct {
			HostIp   string
			HostPort string
		}
		RestartPolicy struct {
			Name string
		}
	}
}

var dockerConfigs []DockerConfig
var format = false

func jsonInspectToCommand(data []byte) (string, error) {
	dockerConfigs = nil
	var fmtEnding = ""
	if format {
		fmtEnding = " \\\n"
	}

	err := json.Unmarshal(data, &dockerConfigs)
	if err != nil {
		fmt.Println("Error parsing json:", err)
		return "", err
	}

	var dockerConfig = dockerConfigs[0]

	var command string
	command += "docker run -d"
	command += " --name=" + dockerConfig.Name[1:]
	command += fmtEnding

	if dockerConfig.HostConfig.AutoRemove {
		command += " --rm"
	}

	if dockerConfig.HostConfig.RestartPolicy.Name != "no" && dockerConfig.HostConfig.RestartPolicy.Name != "" {
		command += " --restart " + dockerConfig.HostConfig.RestartPolicy.Name
		command += fmtEnding
	}

	_, err = hex.DecodeString(dockerConfig.Config.Hostname)
	if err != nil {
		command += " --hostname=" + dockerConfig.Config.Hostname
		command += fmtEnding
	}

	if dockerConfig.HostConfig.Memory > 0 {
		// memory is stored in bytes, minumum is 4M
		command += " -m "

		var kbMemory = float64(dockerConfig.HostConfig.Memory) / 1024
		var mbMemory = kbMemory / 1024
		var gbMemory = mbMemory / 1024

		if gbMemory >= 1 && math.Round(gbMemory) == gbMemory {
			command += strconv.Itoa(int(gbMemory)) + "g"
		} else if mbMemory >= 1 && math.Round(mbMemory) == mbMemory {
			command += strconv.Itoa(int(mbMemory)) + "m"
		} else if kbMemory >= 1 && math.Round(kbMemory) == kbMemory {
			command += strconv.Itoa(int(kbMemory)) + "k"
		} else {
			command += strconv.Itoa(dockerConfig.HostConfig.Memory) + "b"
		}
		command += fmtEnding
	}

	if dockerConfig.HostConfig.NanoCpus > 0 {
		var cpus = float64(dockerConfig.HostConfig.NanoCpus) / 1000000000
		command += " --cpus=" + strconv.FormatFloat(cpus, 'f', -1, 64)
	}

	if len(dockerConfig.HostConfig.Dns) > 0 {
		for _, dns := range dockerConfig.HostConfig.Dns {
			command += " --dns=" + dns
		}
		command += fmtEnding
	}

	if dockerConfig.Config.Env != nil {
		for _, env := range dockerConfig.Config.Env {
			if !(strings.HasPrefix(env, "PATH") || strings.HasPrefix(env, "LANG") || strings.HasPrefix(env, "LC_ALL")) {
				command += " -e "
				if strings.IndexAny(env, " &;") != -1 {
					command += env[0:strings.Index(env, "=")] + "=\"" + env[strings.Index(env, "=")+1:] + "\""
				} else {
					command += env
				}
				command += fmtEnding
			}
		}
	}

	if dockerConfig.HostConfig.Binds != nil {
		for _, bind := range dockerConfig.HostConfig.Binds {
			command += " -v "
			if strings.IndexAny(bind, " &;") != -1 {
				bind = "\"" + bind + "\""
			}
			command += bind
			command += fmtEnding
		}
	}

	if dockerConfig.HostConfig.PortBindings != nil {
		for port, bindigs := range dockerConfig.HostConfig.PortBindings {
			for _, binding := range bindigs {
				command += " -p " + binding.HostPort + ":" + strings.Split(port, "/")[0]
				command += fmtEnding
			}
		}
	}

	if dockerConfig.HostConfig.Links != nil {
		for _, link := range dockerConfig.HostConfig.Links {
			splitted := strings.Split(link, "/")
			preLink := splitted[1][:len(splitted[1])-1]
			command += " --link " + preLink
			if preLink != splitted[3] {
				command += ":" + splitted[3]
			}
			command += fmtEnding
		}
	}

	if len(dockerConfig.HostConfig.LogConfig.Config) > 0 {
		for configName, configValue := range dockerConfig.HostConfig.LogConfig.Config {
			command += " --log-opt " + configName + "=" + configValue
		}
		command += fmtEnding
	}

	command += " " + dockerConfig.Config.Image
	return command, nil
}

func main() {
	var data []byte

	fmtFlag := flag.Bool("format", false, "format the output")

	flag.Parse()

	format = *fmtFlag
	args := flag.Args()

	if len(args) > 0 {
		for i := 0; i < len(args); i++ {
			fileName, err := filepath.Abs(args[i])

			fileInfo, err := os.Stat(fileName)
			if err != nil {
				fmt.Println("Error checking file", fileName, err)
				continue
			}

			if !fileInfo.IsDir() {
				fileContent, err := ioutil.ReadFile(fileName)
				if err != nil {
					fmt.Println("Error reading file", fileName, err)
					continue
				}

				command, err := jsonInspectToCommand(fileContent)
				if err != nil {
					fmt.Println("Error obtaining command from file", fileName, err)
					continue
				}

				if len(args) > 1 {
					fmt.Println("File:", filepath.Base(fileName))
					fmt.Println(command)
					if i < len(args)-1 {
						fmt.Println()
					}
				} else {
					fmt.Println(command)
				}
			}
		}
		os.Exit(0)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		var sdtinData string
		for scanner.Scan() {
			sdtinData += scanner.Text()
		}

		if scanner.Err() != nil {
			fmt.Println("Error reading stdin", scanner.Err())
			os.Exit(1)
		}
		data = []byte(sdtinData)

		command, err := jsonInspectToCommand(data)
		if err != nil {
			fmt.Println("Error obtaining command from json", err)
			os.Exit(1)
		}

		fmt.Println(command)
	}

	if len(data) == 0 {
		fmt.Println("Usage:")
		fmt.Println("./docker2go docker_inspect_of_a_container.txt OR")
		fmt.Println("docker inspect <container> | docker2go")
		os.Exit(0)
	}
}
