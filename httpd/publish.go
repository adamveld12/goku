package httpd

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/adamveld12/goku"
	docker "github.com/fsouza/go-dockerclient"
)

const nginxTemplate = `
server {
		listen 80;

    server_name %s;

    location / {
        proxy_set_header Host $http_host;
        proxy_set_header X-Real-IP $remote_addr;

        proxy_pass http://localhost:%s/;
    }
}
`

// publish publishes a container via nginx
func publish(proj Project, container *docker.Container) error {
	ports := container.NetworkSettings.Ports

	var port docker.PortBinding
	for p, binding := range ports {
		if p.Port() == "80" {
			port = binding[0]
			break
		}
	}

	return saveNginxProfile(proj.Domain, proj.Name, port.HostPort)
}

func saveNginxProfile(domain, name, port string) error {
	l := NewLog("[publish processor]", true)

	siteAvailablePath := fmt.Sprintf("/etc/nginx/sites-available/%s", name)
	fout, err := os.Create(siteAvailablePath)
	if err != nil {
		l.Tracef("could not create nginx configuration file for %s", name)
		return err
	}

	defer fout.Close()

	nginxConf := fmt.Sprintf(nginxTemplate, domain, port)
	l.Trace(nginxConf)

	if _, err = fout.WriteString(nginxConf); err != nil {
		l.Tracef("could not write nginx configuration for %s", name)
		return err
	}

	siteEnabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", name)

	if _, err := os.Stat(siteEnabledPath); os.IsNotExist(err) && os.Symlink(siteAvailablePath, siteEnabledPath) != nil {
		l.Trace("sym link failed")
		return err
	}

	reloadCmd := exec.Command("service", "nginx", "reload")
	if err := reloadCmd.Run(); err != nil {
		l.Tracef("could not reload nginx profile\n%s", err.Error())
		return err
	}

	return nil
}
