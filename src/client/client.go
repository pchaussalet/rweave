package client

import (
	"log"
	"io/ioutil"
	"reflect"
	"fmt"
	"strings"
	"net/http"
	"encoding/base64"

	yaml "gopkg.in/yaml.v2"
	"encoding/json"
)

var debugEnabled = false

func debug(message string, args ...interface{}) {
	if debugEnabled {
		fmt.Printf(message, args...)
	}
}

func fail() {
	debug("[FAIL]\n")
}

func success() {
	debug("[OK]\n")
}

func List(host string, verbose bool) {
	debugEnabled = verbose
	debug("Calling http://%v endpoint...", host)
	resp, err := http.Get("http://" + host)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()
	debug("Parsing remote response...")
	var containers []string
	err = json.NewDecoder(resp.Body).Decode(&containers)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()
	if len(containers) > 0 {
		sep := "\n  "
		fmt.Printf("Containers :" + sep + "%v\n", strings.Join(containers, sep))
	} else {
		fmt.Println("No container running on remote host.")
	}
}

func Deploy(command []string, templateFile, varsFile, host string, verbose bool) {
	debugEnabled = verbose
	component := command[1]
	targetEnv := command[2]
	version := ""
	if len(command) == 4 {
		version = command[3]
	}

	if (verbose) {
		debug("Deploy agruments : component=%v - environment=%v - version=%v\n", component, targetEnv, version)
	}

	debug("Reading template file %v...", templateFile)
	templateRaw, err := ioutil.ReadFile(templateFile)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	template := string(templateRaw)
	success()

	debug("Reading values file %v...", varsFile)
	varsString, err := ioutil.ReadFile(varsFile)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()

	debug("Parsing values file...")
	m := make(map[string]map[string]map[string]interface {})
	err = yaml.Unmarshal(varsString, &m)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()

	debug("Extracting values for environment...")
	vars := make(map[string]string, len(m["global"]) + len(m[targetEnv]))
	vars["VERSION"] = version
	vars = extractVars(vars, m, "global", component)
	vars = extractVars(vars, m, targetEnv, component)
	success()

	debug("Replacing values in template...")
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%v}}", key)
		template = strings.Replace(template, placeholder, value, -1)
	}
	success()

	debug("Parsing generated config file...")
	configs := make(map[string]map[string]interface {})
	err = yaml.Unmarshal([]byte(template), &configs)
	config := configs[component]
	config["name"] = fmt.Sprintf("%v_%v_%v", vars["project"], targetEnv, component)
	payload, err := yaml.Marshal(config)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()
	debug("Generated config :\n%v", string(payload))

	debug("Calling remote server http://%v/new...", string(host))
	resp, err := http.Post("http://"+string(host)+"/new", "application/x-yaml", strings.NewReader(string(payload)))
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()

	debug("Extracting container id...")
	p := make([]byte, resp.ContentLength)
	resp.Body.Read(p)
	body := string(p)
	body = strings.Replace(body, "\"", "", -1)
	id, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		fail()
		log.Fatalln(err)
	}
	success()
	fmt.Print(string(id))
}

func extractVars(vars map[string]string, m map[string]map[string]map[string]interface {}, env, component string) map[string]string {
	for part, values := range m[env] {
		if part == "vars" || part == component {
			for key, value := range values {
				if reflect.TypeOf(value).Kind() == reflect.String {
					vars[key] = reflect.ValueOf(value).String()
				}
			}
		}
	}
	return vars
}
