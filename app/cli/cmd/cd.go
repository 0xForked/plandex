package cmd

import (
	"fmt"
	"os"
	"plandex/api"
	"plandex/auth"
	"plandex/lib"
	"plandex/term"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/plandex/plandex/shared"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cdCmd)
}

var cdCmd = &cobra.Command{
	Use:     "cd [name-or-index]",
	Aliases: []string{"set-plan"},
	Short:   "Set current plan by name or index",
	Args:    cobra.MaximumNArgs(1),
	Run:     cd,
}

func cd(cmd *cobra.Command, args []string) {
	auth.MustResolveAuthWithOrg()
	lib.MustResolveProject()

	var nameOrIdx string
	if len(args) > 0 {
		nameOrIdx = strings.TrimSpace(args[0])
	}

	var plan *shared.Plan

	plans, apiErr := api.Client.ListPlans([]string{lib.CurrentProjectId})

	if apiErr != nil {
		fmt.Fprintln(os.Stderr, "Error getting plans:", apiErr.Msg)
		os.Exit(1)
	}

	if len(plans) == 0 {
		fmt.Println("🤷‍♂️ No plans")
		fmt.Println()
		term.PrintCmds("", "new")
		return
	}

	if nameOrIdx == "" {

		opts := make([]string, len(plans))
		for i, plan := range plans {
			opts[i] = plan.Name
		}

		selected, err := term.SelectFromList("Select a plan", opts)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error selecting plan:", err)
			return
		}

		for _, p := range plans {
			if p.Name == selected {
				plan = p
				break
			}
		}
	} else {
		// see if it's an index
		idx, err := strconv.Atoi(nameOrIdx)

		if err == nil {
			if idx > 0 && idx <= len(plans) {
				plan = plans[idx-1]
			} else {
				fmt.Fprintln(os.Stderr, "Error: index out of range")
				os.Exit(1)
			}
		} else {
			for _, p := range plans {
				if p.Name == nameOrIdx {
					plan = p
					break
				}
			}
		}
	}

	if plan == nil {
		fmt.Fprintln(os.Stderr, "🚨 Plan not found")
		os.Exit(1)
	}

	err := lib.WriteCurrentPlan(plan.Id)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error setting current plan:", err)
		os.Exit(1)
	}

	// reload current plan, which will also handle setting the right branch
	lib.MustLoadCurrentPlan()

	// fire and forget SetProjectPlan request (we don't care about the response or errors)
	// this only matters for setting the current plan on a new device (i.e. when the current plan is not set)
	go api.Client.SetProjectPlan(lib.CurrentProjectId, shared.SetProjectPlanRequest{PlanId: plan.Id})

	// give the SetProjectPlan request some time to be sent before exiting
	time.Sleep(50 * time.Millisecond)

	fmt.Fprintln(os.Stderr, "✅ Changed current plan to "+color.New(color.FgGreen, color.Bold).Sprint(plan.Name))

	fmt.Println()
	term.PrintCmds("", "current")
}
