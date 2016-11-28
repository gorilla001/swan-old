package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"os"
)

// NewRunCommand returns the CLI command for "run"
func NewRunCommand() cli.Command {
	return cli.Command{
		Name:  "run",
		Usage: "run new application",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file",
				Usage: "Run application from `FILE`",
			},
			cli.StringFlag{
				Name:  "name",
				Usage: "Set application name",
			},
			cli.StringFlag{
				Name:  "image",
				Usage: "Image to run",
			},
			cli.IntFlag{
				Name:  "instances",
				Usage: "Instances to be run",
				Value: 1,
			},
			cli.Float64Flag{
				Name:  "cpus",
				Usage: "Cpu limit for instance",
				Value: 0.1,
			},
			cli.Float64Flag{
				Name:  "mem",
				Usage: "Memory limit for instance",
				Value: 5,
			},
			cli.Float64Flag{
				Name:  "disk",
				Usage: "Disk limit for instance",
				Value: 0,
			},
			cli.StringFlag{
				Name:  "run-as",
				Usage: "Run app as some role",
			},
			cli.StringFlag{
				Name:  "network",
				Usage: "Container network mode",
				Value: "BRIDGE",
			},
			cli.IntFlag{
				Name:  "port",
				Usage: "Container service port",
			},
			cli.IntFlag{
				Name:  "port-name",
				Usage: "Container named port",
			},
			cli.StringFlag{
				Name:  "port-protocol",
				Usage: "Container port protocol",
			},
			cli.IntFlag{
				Name:  "kill-duration",
				Usage: "Duration before sending SIGKILL to container",
				Value: 3,
			},
			cli.BoolFlag{
				Name:  "privileged",
				Usage: "Give extended privileges to this container",
			},
			cli.BoolFlag{
				Name:  "force-pull-image",
				Usage: "Force pull image or not, if it is not exists",
			},
			cli.StringFlag{
				Name:  "label",
				Usage: "Set meta data on a container",
			},
			cli.StringFlag{
				Name:  "ip",
				Usage: "Container IPv4 address",
			},
			cli.StringFlag{
				Name:  "env",
				Usage: "Set environment variables",
			},
			cli.StringFlag{
				Name:  "volume",
				Usage: "Mount volume",
			},
		},
		Action: func(c *cli.Context) error {
			if err := runApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// runApplication executes the "run" command.
func runApplication(c *cli.Context) error {
	var version types.Version

	if c.String("from-file") != "" {
		file, err := ioutil.ReadFile(c.String("from-file"))
		if err != nil {
			return fmt.Errorf("Read json file failed: %s", err.Error())
		}

		if err := json.Unmarshal(file, &version); err != nil {
			return fmt.Errorf("Unmarshal error: %s", err.Error())
		}

		if c.IsSet("name") {
			name := c.String("name")
			if name != "" {
				version.ID = name
			}
		}

		if c.IsSet("image") {
			image := c.String("image")
			if image != "" {
				version.Container.Docker.Image = &image
			}
		}

		if c.IsSet("instances") {
			instances := c.Int("instances")
			if instances > 0 {
				version.Instances = instances
			}
		}

		if c.IsSet("cpus") {
			cpus := c.Float64("cpus")
			if cpus > 0 {
				version.Cpus = cpus
			}
		}

		if c.IsSet("mem") {
			mem := c.Float64("mem")
			if mem > 0 {
				version.Mem = mem
			}
		}

		if c.IsSet("disk") {
			disk := c.Float64("disk")
			if disk > 0 {
				version.Disk = disk
			}
		}

		if c.IsSet("run-as") {
			runas := c.String("run-as")
			if runas != "" {
				version.RunAS = runas
			}
		}

		if c.IsSet("network") {
			network := c.String("network")
			if network != "" {
				version.Container.Docker.Network = network
			}
		}

		port := c.Int("port")
		portProtocol := c.String("port-protocol")
		if port > 0 && portProtocol == "" {
			return errors.New("--port-protocol must be specified with --port")
		}

		if c.IsSet("privileged") {
			p := true
			version.Container.Docker.Privileged = &p
		}

		if c.IsSet("force-pull-image") {
			f := true
			version.Container.Docker.ForcePullImage = &f
		}
	} else {

		if !c.IsSet("name") {
			return errors.New("--name must be specified")
		}

		name := c.String("name")
		if name == "" {
			return errors.New("name can't be empty")
		}

		if !c.IsSet("image") {
			return errors.New("--image must be specified")
		}

		image := c.String("image")
		if image == "" {
			return errors.New("image can't be empty")
		}

		runas := c.String("run-as")
		if runas == "" {
			runas = "defaultGroup"
		}

		version.ID = name
		version.RunAS = runas

		forcePullImage := c.IsSet("force-pull-image")
		privileged := c.IsSet("privileged")

		version.Container = &types.Container{
			Type: "DOCKER",
			Docker: &types.Docker{
				Image:          &image,
				ForcePullImage: &forcePullImage,
				Privileged:     &privileged,
				Network:        c.String("network"),
			},
		}

		if c.IsSet("port") {
			if !c.IsSet("port-protocol") {
				return errors.New("--port-protocol must be specified with --port")
			}
			version.Container.Docker.PortMappings = &[]types.PortMapping{
				{
					ContainerPort: c.Int("port"),
					Protocol:      c.String("port-protocol"),
				},
			}
		}

		version.Cpus = c.Float64("cpus")
		version.Mem = c.Float64("mem")
		version.Disk = c.Float64("disk")
		version.Instances = c.Int("instances")
	}

	b, err := json.Marshal(&version)
	if err != nil {
		return fmt.Errorf("Marsh failed: %s", err.Error())
	}

	httpClient := NewHTTPClient("/v1/apps")
	_, err = httpClient.Post(b)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	handleEvents(version.ID)

	httpClient = NewHTTPClient(fmt.Sprintf("/v1/apps/%s/tasks", version.ID))
	resp, err := httpClient.Get()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}
	defer resp.Body.Close()

	var tasks []*types.Task
	if err = json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return err
	}

	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader([]string{
		"APPNAME",
		"TASKNAME",
		"ADDRESS",
	})

	versionId := ""
	versionID := version.ID

	for _, task := range tasks {
		tb.Append([]string{
			versionID,
			task.Name,
			fmt.Sprintf("%s:%d", *task.AgentHostname, task.Port),
		})
		if versionId == "" {
			versionId = versionID
		}

		if versionId != "" {
			versionID = ""
		}
	}

	fmt.Println("")
	tb.Render()
	fmt.Println("")

	return nil
}

func handleEvents(name string) {
	httpClient := NewHTTPClient("/v1/event")
	resp, _ := httpClient.Get()
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for {
		event := new(types.Event)
		if err := dec.Decode(event); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		//fmt.Fprintln(os.Stdout, event.Message)
		if event.ID == name {
			fmt.Fprintln(os.Stdout, event.Message)

			if event.Type == "FINISHED" {
				return
			}
		}
	}
}
