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

func List(host string) {
	resp, err := http.Get("http://" + host)
	if err != nil {
		log.Fatalln(err)
	}
	var containers []string
	err = json.NewDecoder(resp.Body).Decode(&containers)
	if err != nil {
		log.Fatalln(err)
	}
	sep := "\n  "
	fmt.Printf("Containers :" + sep + "%v\n", strings.Join(containers, sep))
}

func Deploy(command []string, templateFile, varsFile, host string) {
	component := command[1]
	targetEnv := command[2]

	templateRaw, err := ioutil.ReadFile(templateFile)
	if err != nil {
		log.Fatalln(err)
	}
	template := string(templateRaw)

	varsString, err := ioutil.ReadFile(varsFile)
	if err != nil {
		log.Fatalln(err)
	}
	m := make(map[string]map[string]map[string]interface {})
	err = yaml.Unmarshal(varsString, &m)
	if err != nil {
		log.Fatalln(err)
	}
	vars := make(map[string]string, len(m["global"]) + len(m[targetEnv]))
	vars = extractVars(vars, m, "global", component)
	vars = extractVars(vars, m, targetEnv, component)
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%v}}", key)
		template = strings.Replace(template, placeholder, value, -1)
	}
	configs := make(map[string]map[string]interface {})
	err = yaml.Unmarshal([]byte(template), &configs)
	config := configs[component]
	config["name"] = fmt.Sprintf("%v_%v_%v", vars["project"], targetEnv, component)
	payload, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.Post("http://"+string(host)+"/new", "application/x-yaml", strings.NewReader(string(payload)))
	if err != nil {
		log.Fatalln(err)
	}
	p := make([]byte, resp.ContentLength)
	resp.Body.Read(p)
	body := string(p)
	body = strings.Replace(body, "\"", "", -1)
	id, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		log.Fatalln(err)
	}
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
