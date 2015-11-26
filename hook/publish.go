package hook

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

const nginxTemplate = `
server {
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

func saveNginxProfile(domain, name, ip, port string) {
	fmt.Println("\n")
	nginxConf := fmt.Sprintf(nginxTemplate, domain, port)
	fmt.Println(nginxConf)
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

	// write template to /etc/nginx/sites-available
	saveNginxProfile(proj.Domain, proj.Name, port.HostIP, port.HostPort)

	// sym link template output to /etc/nginx/sites-enabled
	// reload nginx via nginx reload

	return nil
}
