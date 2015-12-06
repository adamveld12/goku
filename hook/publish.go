package hook

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/adamveld12/goku/log"
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
func publish(proj repository, container *docker.Container) error {

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
	fmt.Println("\n")

	siteAvailablePath := fmt.Sprintf("/etc/nginx/sites-available/%s", name)
	fout, err := os.Create(siteAvailablePath)
	if err != nil {
		log.Debugf("could not create nginx configuration file for %s", name)
		return err
	}

	defer fout.Close()

	nginxConf := fmt.Sprintf(nginxTemplate, domain, port)
	log.Debug(nginxConf)

	if _, err = fout.WriteString(nginxConf); err != nil {
		log.Debugf("could not write nginx configuration for %s", name)
		return err
	}

	siteEnabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", name)

	if _, err := os.Stat(siteEnabledPath); os.IsNotExist(err) && os.Symlink(siteAvailablePath, siteEnabledPath) != nil {
		log.Debug("sym link failed")
		return err
	}

	reloadCmd := exec.Command("service", "nginx", "reload")
	if err := reloadCmd.Start(); err != nil {
		log.Debugf("could not start nginx reload\n%s", err.Error())
		return err
	}

	if err := reloadCmd.Wait(); err != nil {
		log.Debugf("nginx reload failed\n%s", err.Error())
		return err
	}

	return nil
}
