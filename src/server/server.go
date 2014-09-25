package server

import (
	"net/http"
	"encoding/json"
	"log"
	"github.com/docker/libswarm"
	"github.com/docker/libswarm/backends"
	"os/exec"

	yaml "gopkg.in/yaml.v2"

	model "../model"
	"fmt"
	"github.com/gorilla/mux"
)

func Start(host string) {
	r := mux.NewRouter()
	r.HandleFunc("/", listContainers).Methods("GET")
	r.HandleFunc("/new", createContainer).Methods("POST")
	http.Handle("/", r)
	log.Printf("Listening on %v...", host)
	err := http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatalf("An error has occured : %v", err)
	}
}

func sendJsonResponse(payload interface{}, w http.ResponseWriter, status int) {
	result, err := json.Marshal(payload)
	if err != nil {
		sendError(err, w)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(result)
}

func sendError(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	log.Println(err)
}

func getDockerClient() (*libswarm.Client, error) {
	_, back, err := backends.New().Attach("dockerclient")
	if err != nil {
		return nil, err
	}
	instance, err := back.Spawn("tcp://127.0.0.1:4243")
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func getContainers() ([]string, error) {
	instance, err := getDockerClient()
	if err != nil {
		return nil, err
	}
	containers, err := instance.Ls()
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func listContainers(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL)
	containers, err := getContainers()
	if err != nil {
		sendError(err, w)
		return
	}
	sendJsonResponse(containers, w, 200)
}

func removeContainer(name string) error {
	cmd := exec.Command("docker", "rm", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func stopContainer(name string) error {
	dockerClient, err := getDockerClient()
	if err != nil {
		return err
	}
	_, container, err := dockerClient.Attach(name)
	if err != nil {
		return err
	}
	container.Stop()
	return nil
}

func addArg(args []string, value string) []string {
	if value != "" {
		args = append(args, value)
	}
	return args
}

func addArgKV(args []string, key, value string) []string {
	if key != "" && value != "" {
		args = append(args, key)
		args = append(args, value)
	}
	return args
}

func addArgsKV(args []string, key string, values []string) []string {
	for _, value := range values {
		args = addArgKV(args, key, value)
	}
	return args
}

func buildWeaveArgs(config model.ContainerData) []string {
	args := make([]string, 0, 256)

	// Weave part : positional arguments
	args = addArg(args, "weave")
	args = addArg(args, "run")
	args = addArg(args, config.Ip)

	// Docker run options
	args = addArgKV(args, "--name", config.Name)
	args = addArgsKV(args, "--link", config.Links)
	args = addArgsKV(args, "--publish", config.Ports)
	args = addArgsKV(args, "--expose", config.Expose)
	args = addArgsKV(args, "--volume", config.Volumes)
	args = addArgsKV(args, "--volume-from", config.Volumes_from)
	environments := make([]string, len(config.Environment))
	for key, value := range config.Environment {
		environments = append(environments, key + "=" + value)
	}
	args = addArgsKV(args, "--env", environments)
	args = addArgKV(args, "--net", config.Net)
	args = addArgsKV(args, "--dns", config.Dns)
	args = addArgKV(args, "--workdir", config.Working_dir)
	args = addArgKV(args, "--entrypoint", config.Entrypoint)
	args = addArgKV(args, "--user", config.User)
	args = addArgKV(args, "--dns-search", config.Domainname)
	args = addArgKV(args, "--memory", config.Mem_limit)
	args = addArgKV(args, "--privileged", config.Privileged)

	// Image name : positional argument
	args = addArg(args, config.Image)

	log.Println(args)

	return args
}

func createContainer(w http.ResponseWriter, r *http.Request) {
	p := make([]byte, r.ContentLength)
	r.Body.Read(p)
	config := model.ContainerData{}

	err := yaml.Unmarshal(p, &config)
	if err != nil {
		sendError(err, w)
		return
	}

	containers, err := getContainers()
	if err != nil {
		sendError(err, w)
		return
	}

	replace := false
	for _, val := range containers {
		if config.Name == val {
			replace = true
			break
		}
	}
	if replace {
		err := stopContainer(config.Name)
		if err != nil {
			sendError(err, w)
			return
		}
		err = removeContainer(config.Name)
		if err != nil {
			sendError(err, w)
			return
		}
	}

	fmt.Println(string(p))
	args := buildWeaveArgs(config)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		sendError(err, w)
		return
	}
	sendJsonResponse(out, w, 200)
}
