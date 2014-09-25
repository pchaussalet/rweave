package model

type ContainerData struct {
	Name string
	Ip string
	Image string
	Build string
	Command string
	Links []string
	Ports []string
	Expose []string
	Volumes []string
	Volumes_from []string
	Environment map[string]string
	Net string
	Dns []string
	Working_dir string
	Entrypoint string
	User string
	Domainname string
	Mem_limit string
	Privileged string
}

