package command

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/swan/types"
	"github.com/urfave/cli"
	"os"
)

// NewShowCommand returns the CLI command for "show"
func NewShowCommand() cli.Command {
	return cli.Command{
		Name:      "show",
		Usage:     "show application or task info",
		ArgsUsage: "[name]",
		Action: func(c *cli.Context) error {
			if err := showApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// showApplication executes the "show" command.
func showApplication(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("Task or App ID required")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("/v1/apps/%s/tasks", c.Args()[0]))
	resp, err := httpClient.Get()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}
	defer resp.Body.Close()

	var tasks []*types.Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return err
	}

	data, err := json.Marshal(&tasks)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(data))

	return nil
}
