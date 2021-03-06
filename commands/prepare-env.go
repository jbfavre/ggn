package commands

import (
	"github.com/blablacar/ggn/work"
	"github.com/spf13/cobra"
)

func prepareEnvCommands(env *work.Env) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   env.GetName(),
		Short: "Run command for " + env.GetName(),
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check of " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Check()
		},
	}

	fleetctlCmd := &cobra.Command{
		Use:   "fleetctl",
		Short: "Run fleetctl command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Fleetctl(args)
		},
	}

	listUnitsCmd := &cobra.Command{
		Use:   "list-units",
		Short: "Run list-units command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.FleetctlListUnits()
		},
	}

	listMachinesCmd := &cobra.Command{
		Use:   "list-machines",
		Short: "Run list-machines command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.FleetctlListMachines()
		},
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate units for " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Generate()
		},
	}
	envCmd.AddCommand(generateCmd, fleetctlCmd, checkCmd, listUnitsCmd, listMachinesCmd)

	for _, serviceName := range env.ListServices() {
		service := env.LoadService(serviceName)
		envCmd.AddCommand(prepareServiceCommands(service))
	}

	return envCmd
}
