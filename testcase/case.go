package testcase

import (
	"github.com/crazycs520/load/cmd"
)

func init() {
	cmd.RegisterCaseCmd(NewIndexLookUpWrongPlan)
}
