package hook

import (
	"fmt"
	"os"
	"os/exec"

	docker "github.com/fsouza/go-dockerclient"
)

const nginxTemplate = `
server {
    listen 80;
    server_name %s

    location / {
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header Host $http_host;
        proxy_set_header X-NginX-Proxy true;

        proxy_pass http://localhost:%s/;
    }
}
`

func saveNginxProfile(domain, name, ip, port string) error {
	fmt.Println("\n")
	nginxConf := fmt.Sprintf(nginxTemplate, domain, port)
	fmt.Println(nginxConf)

	siteAvailablePath := fmt.Sprintf("/etc/nginx/sites-available/%s", name)
	fout, err := os.Create(siteAvailablePath)
	if err != nil {
		return err
	}
	defer fout.Close()

	if _, err = fout.WriteString(nginxConf); err != nil {
		return err
	}

	siteEnabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", name)

	if _, err := os.Stat(siteEnabledPath); os.IsNotExist(err) && os.Symlink(siteAvailablePath, siteEnabledPath) != nil {
		fmt.Println("sym link failed")
		return err
	}

	reload := exec.Command("nginx", "reload")
	return reload.Run()
}

func publish(proj repository, container *docker.Container) error {

	ports := container.NetworkSettings.Ports

	var port docker.PortBinding
	for p, binding := range ports {

		if p.Port() == "80" {
			port = binding[0]
			break
		}
	}

	return saveNginxProfile(proj.Domain, proj.Name, port.HostIP, port.HostPort)
}
