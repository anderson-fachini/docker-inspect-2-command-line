package main

import (
	"fmt"
    "os"
	"io/ioutil"
	"encoding/json"
	"encoding/hex"
	"strings"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("You have to specify the json file with the docker ispect of a container.");
        fmt.Println("Usage: ./docker2go docker_inspect_of_a_container.json");
        os.Exit(1)
    }

    inspectFile := "./" + os.Args[1]
    dat, err := ioutil.ReadFile(inspectFile)
    if err != nil {
        fmt.Println("Error reading file:", err)
        return
    }

    type Config struct {
        Hostname string
        Env []string
        Image string
    }

    type PortBinding struct {
        HostIp string
        HostPort string
    }

    type LogConfig struct {
        Config map[string] string
    }

    type RestartPolicy struct {
        Name string
    }

    type HostConfig struct {
        PortBindings map[string] []PortBinding
        Links []string
        Binds []string
        LogConfig LogConfig
        RestartPolicy RestartPolicy
    }

    type DockerConfig struct {
        Name string
        Config Config
        HostConfig HostConfig
    }

    var dockerConfigs []DockerConfig    
    err = json.Unmarshal(dat, &dockerConfigs)
    if err != nil {
        fmt.Println("Error parsing json:", err)
        return
    }

    var dockerConfig = dockerConfigs[0]

    var command string
    command += "docker run -d"
    command += " --name=" + dockerConfig.Name[1:]

    if dockerConfig.HostConfig.RestartPolicy.Name != "no" {
        command += " --restart " + dockerConfig.HostConfig.RestartPolicy.Name
    }

    _, err = hex.DecodeString(dockerConfig.Config.Hostname)
    if err != nil {
        command += " --hostname=" + dockerConfig.Config.Hostname
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
        }
    }

    if len(dockerConfig.HostConfig.LogConfig.Config) > 0 {
        for configName, configValue := range dockerConfig.HostConfig.LogConfig.Config {
            command += " --log-opt " + configName + "=" + configValue 
        }
    }

    if dockerConfig.HostConfig.Binds != nil {
        for _, bind := range dockerConfig.HostConfig.Binds {
            command += " -v "
            if strings.IndexAny(bind, " &;") != -1 {
                bind = "\"" + bind + "\""
            }
            command += bind
        }
    }

    if dockerConfig.HostConfig.PortBindings != nil {
        for port, bindigs := range dockerConfig.HostConfig.PortBindings {
            for _, binding := range bindigs {
                command += " -p " + binding.HostPort + ":" + strings.Split(port, "/")[0]
            }
        }
    }

    command += " " + dockerConfig.Config.Image

    fmt.Println(command)
}