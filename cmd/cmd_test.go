package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"agentmux": rootCmdCustomMain,
	}))
}

func TestScripts(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(env *testscript.Env) error {
			env.Setenv("HOME", env.WorkDir)
			return nil
		},
	})
}

// rootCmdCustomMain runs the root command and returns an exit code instead of calling os.Exit.
// We execute rootCmd.Execute() directly because Execute() handles its own error printing but doesn't exit.
func rootCmdCustomMain() int {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println("Error executing rootCmd:", err)
		return 1
	}
	return 0
}
