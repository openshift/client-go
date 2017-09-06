package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	g "github.com/onsi/ginkgo"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage/names"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kclientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	kinternalclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/retry"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	_ "github.com/openshift/origin/pkg/api/install"
	"github.com/openshift/origin/pkg/client"
	oclientset "github.com/openshift/origin/pkg/client/clientset/clientset"
	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	"github.com/openshift/origin/pkg/oc/cli/config"
	projectapi "github.com/openshift/origin/pkg/project/apis/project"
	templateclientset "github.com/openshift/origin/pkg/template/generated/internalclientset"
	testutil "github.com/openshift/origin/test/util"
)

// CLI provides function to call the OpenShift CLI and Kubernetes and OpenShift
// clients.
type CLI struct {
	execPath         string
	verb             string
	configPath       string
	adminConfigPath  string
	username         string
	outputDir        string
	globalArgs       []string
	commandArgs      []string
	finalArgs        []string
	stdin            *bytes.Buffer
	stdout           io.Writer
	stderr           io.Writer
	verbose          bool
	withoutNamespace bool
	kubeFramework    *e2e.Framework
}

// NewCLI initialize the upstream E2E framework and set the namespace to match
// with the project name. Note that this function does not initialize the project
// role bindings for the namespace.
func NewCLI(project, adminConfigPath string) *CLI {
	// Avoid every caller needing to provide a unique project name
	// SetupProject already treats this as a baseName
	uniqueProject := names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-", project))

	client := &CLI{}
	client.kubeFramework = e2e.NewDefaultFramework(uniqueProject)
	client.outputDir = os.TempDir()
	client.username = "admin"
	client.execPath = "oc"
	if len(adminConfigPath) == 0 {
		FatalErr(fmt.Errorf("you must set the KUBECONFIG variable to admin kubeconfig"))
	}
	client.adminConfigPath = adminConfigPath

	// Register custom ns setup func
	setCreateTestingNSFunc(uniqueProject, client.SetupProject)

	return client
}

// KubeFramework returns Kubernetes framework which contains helper functions
// specific for Kubernetes resources
func (c *CLI) KubeFramework() *e2e.Framework {
	return c.kubeFramework
}

// Username returns the name of currently logged user. If there is no user assigned
// for the current session, it returns 'admin'.
func (c *CLI) Username() string {
	return c.username
}

// AsAdmin changes current config file path to the admin config.
func (c *CLI) AsAdmin() *CLI {
	nc := *c
	nc.configPath = c.adminConfigPath
	return &nc
}

// ChangeUser changes the user used by the current CLI session.
func (c *CLI) ChangeUser(name string) *CLI {
	adminClientConfig, err := testutil.GetClusterAdminClientConfig(c.adminConfigPath)
	if err != nil {
		FatalErr(err)
	}
	_, _, clientConfig, err := testutil.GetClientForUser(*adminClientConfig, name)
	if err != nil {
		FatalErr(err)
	}

	kubeConfig, err := config.CreateConfig(c.Namespace(), clientConfig)
	if err != nil {
		FatalErr(err)
	}

	c.configPath = filepath.Join(c.outputDir, name+".kubeconfig")
	err = clientcmd.WriteToFile(*kubeConfig, c.configPath)
	if err != nil {
		FatalErr(err)
	}

	c.username = name
	e2e.Logf("configPath is now %q", c.configPath)
	return c
}

// SetNamespace sets a new namespace
func (c *CLI) SetNamespace(ns string) *CLI {
	c.kubeFramework.Namespace = &kapiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}
	return c
}

// WithoutNamespace instructs the command should be invoked without adding --namespace parameter
func (c *CLI) WithoutNamespace() *CLI {
	c.withoutNamespace = true
	return c
}

// SetOutputDir change the default output directory for temporary files
func (c *CLI) SetOutputDir(dir string) *CLI {
	c.outputDir = dir
	return c
}

// SetupProject creates a new project and assign a random user to the project.
// All resources will be then created within this project and Kubernetes E2E
// suite will destroy the project after test case finish.
func (c *CLI) SetupProject(name string, kubeClient kclientset.Interface, _ map[string]string) (*kapiv1.Namespace, error) {
	newNamespace := names.SimpleNameGenerator.GenerateName(fmt.Sprintf("extended-test-%s-", name))
	c.SetNamespace(newNamespace).ChangeUser(fmt.Sprintf("%s-user", c.Namespace()))
	e2e.Logf("The user is now %q", c.Username())

	e2e.Logf("Creating project %q", c.Namespace())
	_, err := c.Client().ProjectRequests().Create(&projectapi.ProjectRequest{
		ObjectMeta: metav1.ObjectMeta{Name: c.Namespace()},
	})
	if err != nil {
		e2e.Logf("Failed to create a project and namespace %q: %v", c.Namespace(), err)
		return nil, err
	}
	if err := wait.ExponentialBackoff(retry.DefaultBackoff, func() (bool, error) {
		if _, err := c.KubeClient().Core().Pods(c.Namespace()).List(metav1.ListOptions{}); err != nil {
			if apierrs.IsForbidden(err) {
				e2e.Logf("Waiting for user to have access to the namespace")
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return &kapiv1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: c.Namespace()}}, err
}

// Verbose turns on printing verbose messages when executing OpenShift commands
func (c *CLI) Verbose() *CLI {
	c.verbose = true
	return c
}

// Client provides an OpenShift client for the current user. If the user is not
// set, then it provides client for the cluster admin user
func (c *CLI) Client() *client.Client {
	_, clientConfig, err := configapi.GetInternalKubeClient(c.configPath, nil)
	osClient, err := client.New(clientConfig)
	if err != nil {
		FatalErr(err)
	}
	return osClient
}

// Clientset provides the new OpenShift clientset for the current user. If the user is not
// set, then it provides client for the cluster admin user
func (c *CLI) Clientset() *oclientset.Clientset {
	_, clientConfig, err := configapi.GetInternalKubeClient(c.configPath, nil)
	if err != nil {
		FatalErr(err)
	}

	cs, err := oclientset.NewForConfig(clientConfig)
	if err != nil {
		FatalErr(err)
	}

	return cs
}

// AdminClient provides an OpenShift client for the cluster admin user.
func (c *CLI) AdminClient() *client.Client {
	_, clientConfig, err := configapi.GetInternalKubeClient(c.adminConfigPath, nil)
	osClient, err := client.New(clientConfig)
	if err != nil {
		FatalErr(err)
	}
	return osClient
}

// Client provides an OpenShift client for the current user. If the user is not
// set, then it provides client for the cluster admin user
func (c *CLI) TemplateClient() templateclientset.Interface {
	_, clientConfig, err := configapi.GetInternalKubeClient(c.configPath, nil)
	client, err := templateclientset.NewForConfig(clientConfig)
	if err != nil {
		FatalErr(err)
	}
	return client
}

// AdminClient provides an OpenShift client for the cluster admin user.
func (c *CLI) AdminTemplateClient() templateclientset.Interface {
	_, clientConfig, err := configapi.GetInternalKubeClient(c.adminConfigPath, nil)
	client, err := templateclientset.NewForConfig(clientConfig)
	if err != nil {
		FatalErr(err)
	}
	return client
}

// KubeClient provides a Kubernetes client for the current namespace
func (c *CLI) KubeClient() kclientset.Interface {
	kubeClient, _, err := configapi.GetExternalKubeClient(c.configPath, nil)
	if err != nil {
		FatalErr(err)
	}
	return kubeClient
}

// KubeClient provides a Kubernetes client for the current namespace
func (c *CLI) InternalKubeClient() kinternalclientset.Interface {
	kubeClient, _, err := configapi.GetInternalKubeClient(c.configPath, nil)
	if err != nil {
		FatalErr(err)
	}
	return kubeClient
}

// AdminKubeClient provides a Kubernetes client for the cluster admin user.
func (c *CLI) AdminKubeClient() kclientset.Interface {
	kubeClient, _, err := configapi.GetExternalKubeClient(c.adminConfigPath, nil)
	if err != nil {
		FatalErr(err)
	}
	return kubeClient
}

// AdminKubeClient provides a Kubernetes client for the cluster admin user.
func (c *CLI) InternalAdminKubeClient() kinternalclientset.Interface {
	kubeClient, _, err := configapi.GetInternalKubeClient(c.adminConfigPath, nil)
	if err != nil {
		FatalErr(err)
	}
	return kubeClient
}

// Namespace returns the name of the namespace used in the current test case.
// If the namespace is not set, an empty string is returned.
func (c *CLI) Namespace() string {
	if c.kubeFramework.Namespace == nil {
		return ""
	}
	return c.kubeFramework.Namespace.Name
}

// setOutput allows to override the default command output
func (c *CLI) setOutput(out io.Writer) *CLI {
	c.stdout = out
	return c
}

// Run executes given OpenShift CLI command verb (iow. "oc <verb>").
// This function also override the default 'stdout' to redirect all output
// to a buffer and prepare the global flags such as namespace and config path.
func (c *CLI) Run(commands ...string) *CLI {
	in, out, errout := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	nc := &CLI{
		execPath:        c.execPath,
		verb:            commands[0],
		kubeFramework:   c.KubeFramework(),
		adminConfigPath: c.adminConfigPath,
		configPath:      c.configPath,
		username:        c.username,
		outputDir:       c.outputDir,
		globalArgs: append(commands, []string{
			fmt.Sprintf("--config=%s", c.configPath),
		}...),
	}
	if !c.withoutNamespace {
		nc.globalArgs = append(nc.globalArgs, fmt.Sprintf("--namespace=%s", c.Namespace()))
	}
	nc.stdin, nc.stdout, nc.stderr = in, out, errout
	return nc.setOutput(c.stdout)
}

// Template sets a Go template for the OpenShift CLI command.
// This is equivalent of running "oc get foo -o template --template='{{ .spec }}'"
func (c *CLI) Template(t string) *CLI {
	if c.verb != "get" {
		FatalErr("Cannot use Template() for non-get verbs.")
	}
	templateArgs := []string{"--output=template", fmt.Sprintf("--template=%s", t)}
	commandArgs := append(c.commandArgs, templateArgs...)
	c.finalArgs = append(c.globalArgs, commandArgs...)
	return c
}

// InputString adds expected input to the command
func (c *CLI) InputString(input string) *CLI {
	c.stdin.WriteString(input)
	return c
}

// Args sets the additional arguments for the OpenShift CLI command
func (c *CLI) Args(args ...string) *CLI {
	c.commandArgs = args
	c.finalArgs = append(c.globalArgs, c.commandArgs...)
	return c
}

func (c *CLI) printCmd() string {
	return strings.Join(c.finalArgs, " ")
}

type ExitError struct {
	Cmd    string
	StdErr string
	*exec.ExitError
}

// Output executes the command and returns stdout/stderr combined into one string
func (c *CLI) Output() (string, error) {
	if c.verbose {
		fmt.Printf("DEBUG: oc %s\n", c.printCmd())
	}
	cmd := exec.Command(c.execPath, c.finalArgs...)
	cmd.Stdin = c.stdin
	e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	switch err.(type) {
	case nil:
		c.stdout = bytes.NewBuffer(out)
		return trimmed, nil
	case *exec.ExitError:
		e2e.Logf("Error running %v:\n%s", cmd, trimmed)
		return trimmed, &ExitError{ExitError: err.(*exec.ExitError), Cmd: c.execPath + " " + strings.Join(c.finalArgs, " "), StdErr: trimmed}
	default:
		FatalErr(fmt.Errorf("unable to execute %q: %v", c.execPath, err))
		// unreachable code
		return "", nil
	}
}

// Outputs executes the command and returns the stdout/stderr output as separate strings
func (c *CLI) Outputs() (string, string, error) {
	if c.verbose {
		fmt.Printf("DEBUG: oc %s\n", c.printCmd())
	}
	cmd := exec.Command(c.execPath, c.finalArgs...)
	cmd.Stdin = c.stdin
	e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))
	//out, err := cmd.CombinedOutput()
	var stdErrBuff, stdOutBuff bytes.Buffer
	cmd.Stdout = &stdOutBuff
	cmd.Stderr = &stdErrBuff
	err := cmd.Run()

	stdOutBytes := stdOutBuff.Bytes()
	stdErrBytes := stdErrBuff.Bytes()
	stdOut := strings.TrimSpace(string(stdOutBytes))
	stdErr := strings.TrimSpace(string(stdErrBytes))
	switch err.(type) {
	case nil:
		c.stdout = bytes.NewBuffer(stdOutBytes)
		c.stderr = bytes.NewBuffer(stdErrBytes)
		return stdOut, stdErr, nil
	case *exec.ExitError:
		e2e.Logf("Error running %v:\nStdOut>\n%s\nStdErr>\n%s\n", cmd, stdOut, stdErr)
		return stdOut, stdErr, err
	default:
		FatalErr(fmt.Errorf("unable to execute %q: %v", c.execPath, err))
		// unreachable code
		return "", "", nil
	}
}

// Background executes the command in the background and returns the Cmd object
// which may be killed later via cmd.Process.Kill().  It also returns buffers
// holding the stdout & stderr of the command, which may be read from only after
// calling cmd.Wait().
func (c *CLI) Background() (*exec.Cmd, *bytes.Buffer, *bytes.Buffer, error) {
	if c.verbose {
		fmt.Printf("DEBUG: oc %s\n", c.printCmd())
	}
	cmd := exec.Command(c.execPath, c.finalArgs...)
	cmd.Stdin = c.stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout = bufio.NewWriter(&stdout)
	cmd.Stderr = bufio.NewWriter(&stderr)

	e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))

	err := cmd.Start()
	return cmd, &stdout, &stderr, err
}

// BackgroundRC executes the command in the background and returns the Cmd
// object which may be killed later via cmd.Process.Kill().  It returns a
// ReadCloser for stdout.  If in doubt, use Background().  Consult the os/exec
// documentation.
func (c *CLI) BackgroundRC() (*exec.Cmd, io.ReadCloser, error) {
	if c.verbose {
		fmt.Printf("DEBUG: oc %s\n", c.printCmd())
	}
	cmd := exec.Command(c.execPath, c.finalArgs...)
	cmd.Stdin = c.stdin
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))

	err = cmd.Start()
	return cmd, stdout, err
}

// Stdout returns the current stdout writer
func (c *CLI) Stdout() io.Writer {
	return c.stdout
}

// OutputToFile executes the command and store output to a file
func (c *CLI) OutputToFile(filename string) (string, error) {
	content, err := c.Output()
	if err != nil {
		return "", err
	}
	path := filepath.Join(c.outputDir, c.Namespace()+"-"+filename)
	return path, ioutil.WriteFile(path, []byte(content), 0644)
}

// Execute executes the current command and return error if the execution failed
// This function will set the default output to Ginkgo writer.
func (c *CLI) Execute() error {
	out, err := c.Output()
	if _, err := io.Copy(g.GinkgoWriter, strings.NewReader(out+"\n")); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to copy the output to ginkgo writer")
	}
	os.Stdout.Sync()
	return err
}

// FatalErr exits the test in case a fatal error has occurred.
func FatalErr(msg interface{}) {
	e2e.Failf("%v", msg)
}
