package plan_exec

import (
	"fmt"
	"log"
	"os"
	"plandex/api"
	"plandex/auth"
	"plandex/fs"
	"plandex/stream"
	streamtui "plandex/stream_tui"
	"plandex/term"

	"github.com/fatih/color"
	"github.com/plandex/plandex/shared"
)

func TellPlan(
	params ExecParams,
	prompt string,
	tellBg,
	tellStop,
	tellNoBuild,
	isUserContinue bool,
) {
	contexts, apiErr := api.Client.ListContext(params.CurrentPlanId, params.CurrentBranch)

	if apiErr != nil {
		color.New(color.FgRed).Fprintln(os.Stderr, "Error getting context:", apiErr)
		os.Exit(1)
	}

	anyOutdated, didUpdate, canceled := params.CheckOutdatedContext(true, contexts)

	if anyOutdated && !didUpdate && canceled {
		term.StopSpinner()
		if isUserContinue {
			log.Println("Plan won't continue")
		} else {
			log.Println("Prompt not sent")
		}
		os.Exit(0)
	}

	projectPaths, _, err := fs.GetProjectPaths(fs.GetBaseDirForContexts(contexts))

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting project paths:", err)
		return
	}

	var fn func() bool
	fn = func() bool {

		var buildMode shared.BuildMode
		if tellNoBuild {
			buildMode = shared.BuildModeNone
		} else {
			buildMode = shared.BuildModeAuto
		}

		if isUserContinue {
			term.StartSpinner("⚡️ Continuing plan...")
		} else {
			term.StartSpinner("💬 Sending prompt...")
		}

		apiErr := api.Client.TellPlan(params.CurrentPlanId, params.CurrentBranch, shared.TellPlanRequest{
			Prompt:         prompt,
			ConnectStream:  !tellBg,
			AutoContinue:   !tellStop,
			ProjectPaths:   projectPaths,
			BuildMode:      buildMode,
			IsUserContinue: isUserContinue,
			ApiKey:         os.Getenv("OPENAI_API_KEY"),
		}, stream.OnStreamPlan)

		term.StopSpinner()

		if apiErr != nil {
			if apiErr.Type == shared.ApiErrorTypeTrialMessagesExceeded {
				fmt.Fprintf(os.Stderr, "\n🚨 You've reached the free trial limit of %d messages per plan\n", apiErr.TrialMessagesExceededError.MaxReplies)

				res, err := term.ConfirmYesNo("Upgrade now?")

				if err != nil {
					fmt.Fprintln(os.Stderr, "Error prompting upgrade trial:", err)
					return false
				}

				if res {
					err := auth.ConvertTrial()
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error converting trial:", err)
						return false
					}
					// retry action after converting trial
					return fn()
				}
				return false
			}

			fmt.Fprintln(os.Stderr, "Prompt error:", apiErr.Msg)
			return false
		} else if apiErr != nil && isUserContinue && apiErr.Type == shared.ApiErrorTypeContinueNoMessages {
			fmt.Println("🤷‍♂️ There's no plan yet to continue")
			fmt.Println()
			term.PrintCmds("", "tell")
			os.Exit(1)
		}

		if !tellBg {
			go func() {
				err := streamtui.StartStreamUI(prompt, false)

				if err != nil {
					log.Printf("Error starting stream UI: %v\n", err)
					os.Exit(1)
				}

				fmt.Println()

				if tellStop {
					term.PrintCmds("", "continue", "convo", "changes", "log", "rewind")
				} else if tellNoBuild {
					term.PrintCmds("", "convo", "log", "rewind")
				} else {
					term.PrintCmds("", "changes", "log", "rewind")
				}
				os.Exit(0)
			}()
		}

		return true
	}

	shouldContinue := fn()
	if !shouldContinue {
		return
	}

	if tellBg {
		fmt.Println("✅ Plan is active in the background")
		fmt.Println()
		term.PrintCmds("", "ps", "connect", "stop")
	} else {
		// Wait for stream UI to quit
		select {}
	}
}
