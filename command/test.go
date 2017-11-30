package command

import (
	"fmt"
	"strconv"

	"github.com/heroku/force/desktop"
	. "github.com/heroku/force/error"
	. "github.com/heroku/force/lib"
)

var cmdTest = &Command{
	Usage: "test (all | classname... | classname.method...)",
	Short: "Run apex tests",
	Long: `
Run apex tests

Test Options
  -namespace=<namespace>     Select namespace to run test from
  -class=class               Select class to run tests from
  -v                         Verbose logging

Examples:

  force test all
  force test Test1 Test2 Test3
  force test Test1.method1 Test1.method2
  force test -namespace=ns Test4
  force test -class=Test1 method1 method2
  force test -v Test1
`,
}

func init() {
	cmdTest.Flag.BoolVar(&verboselogging, "v", false, "set verbose logging")
	cmdTest.Run = runTests
}

var (
	namespaceTestFlag = cmdTest.Flag.String("namespace", "", "namespace to run tests in")
	classFlag         = cmdTest.Flag.String("class", "", "class to run tests from")
	verboselogging    bool
)

func RunTests(testRunner TestRunner, tests []string, namespace string) (output TestCoverage, err error) {
	output, err = testRunner.RunTests(tests, namespace)
	if err != nil {
		return
	}
	if output.NumberRun == 0 && output.NumberFailures == 0 {
		err = fmt.Errorf("Test classes specified not found: %v", tests)
		return
	}
	return
}

func QualifyMethods(class string, methods []string) []string {
	if len(methods) == 0 {
		return []string{class}
	}
	var qualified []string
	for _, method := range methods {
		qualified = append(qualified, fmt.Sprintf("%s.%s", class, method))
	}
	return qualified
}

func runTests(cmd *Command, args []string) {
	if len(args) < 1 && *classFlag == "" {
		ErrorAndExit("must specify tests to run")
	}
	force, _ := ActiveForce()
	if *classFlag != "" {
		args = QualifyMethods(*classFlag, args)
	}
	output, err := RunTests(force.Partner, args, *namespaceTestFlag)
	success := false
	if err != nil {
		ErrorAndExit(err.Error())
	}
	if verboselogging {
		fmt.Println(output.Log)
		fmt.Println()
	}
	var percent string
	fmt.Println("Coverage:")
	fmt.Println()
	for index := range output.NumberLocations {
		if output.NumberLocations[index] != 0 {
			locations := float64(output.NumberLocations[index])
			notCovered := float64(output.NumberLocationsNotCovered[index])
			percent = strconv.Itoa(int((locations-notCovered)/locations*100)) + "%"
		} else {
			percent = "0%"
		}
		fmt.Println("  " + percent + "\t" + output.Name[index])
	}
	fmt.Println()
	fmt.Println()
	fmt.Println("Results:")
	fmt.Println()
	for index := range output.SMethodNames {
		fmt.Println("  [PASS]  " + output.SClassNames[index] + "::" + output.SMethodNames[index])
	}

	for index := range output.FMethodNames {
		fmt.Println("  [FAIL]  " + output.FClassNames[index] + "::" + output.FMethodNames[index] + ": " + output.FMessage[index])
		fmt.Println("    " + output.FStackTrace[index])
	}
	fmt.Println()
	fmt.Println()

	success = len(output.FMethodNames) == 0

	// Handle notifications
	desktop.NotifySuccess("test", success)
	if !success {
		ErrorAndExit("Tests Failed")
	}
}