package main

import (
	"errors"
	"fmt"
	"os"

	buildahcli "github.com/containers/buildah/pkg/cli"
	"github.com/containers/buildah/util"
	"github.com/spf13/cobra"
)

type rmResults struct {
	all bool
}

func init() {
	var (
		rmDescription = "\n  Removes one or more working containers, unmounting them if necessary."
		opts          rmResults
	)
	rmCommand := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"delete"},
		Short:   "Remove one or more working containers",
		Long:    rmDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rmCmd(cmd, args, opts)
		},
		Example: `buildah rm containerID
  buildah rm containerID1 containerID2 containerID3
  buildah rm --all`,
	}
	rmCommand.SetUsageTemplate(UsageTemplate())

	flags := rmCommand.Flags()
	flags.SetInterspersed(false)
	flags.BoolVarP(&opts.all, "all", "a", false, "remove all containers")
	rootCmd.AddCommand(rmCommand)
}

func rmCmd(c *cobra.Command, args []string, iopts rmResults) error {
	delContainerErrStr := "error removing container"
	if len(args) == 0 && !iopts.all {
		return errors.New("container ID must be specified")
	}
	if len(args) > 0 && iopts.all {
		return errors.New("when using the --all switch, you may not pass any containers names or IDs")
	}

	if err := buildahcli.VerifyFlagsArgsOrder(args); err != nil {
		return err
	}

	store, err := getStore(c)
	if err != nil {
		return err
	}

	var lastError error
	if iopts.all {
		builders, err := openBuilders(store)
		if err != nil {
			return fmt.Errorf("error reading build containers: %w", err)
		}

		for _, builder := range builders {
			id := builder.ContainerID
			if err = builder.Delete(); err != nil {
				lastError = util.WriteError(os.Stderr, fmt.Errorf("%s %q: %w", delContainerErrStr, builder.Container, err), lastError)
				continue
			}
			fmt.Printf("%s\n", id)
		}
	} else {
		for _, name := range args {
			builder, err := openBuilder(getContext(), store, name)
			if err != nil {
				lastError = util.WriteError(os.Stderr, fmt.Errorf("%s %q: %w", delContainerErrStr, name, err), lastError)
				continue
			}
			id := builder.ContainerID
			if err = builder.Delete(); err != nil {
				lastError = util.WriteError(os.Stderr, fmt.Errorf("%s %q: %w", delContainerErrStr, name, err), lastError)
				continue
			}
			fmt.Printf("%s\n", id)
		}

	}
	return lastError
}
