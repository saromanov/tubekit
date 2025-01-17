package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/reconquest/karma-go"
	clientcmdapi "k8s.io/client-go/tools/clientcmd"
)

type Resource struct {
	Name      string
	Namespace string
}

func parseKubernetesContexts() ([]string, error) {
	config, err := clientcmdapi.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to load kube config",
		)
	}

	contexts := []string{}
	for name := range config.Contexts {
		contexts = append(contexts, name)
	}

	sort.Strings(contexts)

	return contexts, nil
}

func requestNamespaces(ctlPath string, params *Params) ([]string, error) {
	// omit namespace argument because requesting list of them
	cmd, args := getCommand(
		ctlPath, buildArgContext(params.Context), "", "",
		"get", "namespaces", "-o", "json",
	)

	debugcmd(args)

	ctx := karma.Describe(
		"cmdline",
		fmt.Sprintf("%q", args),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	contents, err := cmd.Output()
	if err != nil {
		return nil, ctx.Format(
			err,
			"tubectl command failed",
		)
	}

	resources, err := unmarshalResources(contents)
	if err != nil {
		return nil, ctx.Reason(err)
	}

	namespaces := []string{}
	for _, resource := range resources {
		namespaces = append(namespaces, resource.Name)
	}

	return namespaces, nil
}

func requestResources(ctlPath string, params *Params) ([]Resource, error) {
	cmd, args := getCommand(
		ctlPath,
		buildArgContext(params.Context),
		buildArgNamespace(params.Namespace),
		buildArgAllNamespaces(params.AllNamespaces),
		"get", params.Match.Resource, "-o", "json",
	)

	debugcmd(args)

	ctx := karma.Describe(
		"cmdline",
		fmt.Sprintf("%q", args),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	contents, err := cmd.Output()
	if err != nil {
		return nil, ctx.Format(
			err,
			"tubectl command failed",
		)
	}

	resources, err := unmarshalResources(contents)
	if err != nil {
		return nil, ctx.Reason(err)
	}

	return resources, nil
}

func unmarshalResources(contents []byte) ([]Resource, error) {
	var answer struct {
		Items []struct {
			Metadata Resource
		}
	}

	err := json.Unmarshal(contents, &answer)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to unmarshal JSON output",
		)
	}

	resources := []Resource{}
	for _, item := range answer.Items {
		resources = append(resources, item.Metadata)
	}

	return resources, nil
}

func getCommand(
	ctlPath string,
	argContext,
	argNamespace string,
	argAllNamespaces string,
	value ...string,
) (*exec.Cmd, []string) {
	args := []string{}
	if argContext != "" {
		args = append(args, argContext)
	}
	if argNamespace != "" {
		args = append(args, argNamespace)
	}

	args = append(args, value...)

	if argAllNamespaces != "" {
		args = append(args, argAllNamespaces)
	}

	return exec.Command(ctlPath, args...), append([]string{ctlPath}, args...)
}

func buildArgContext(value string) string {
	if value != "" {
		return "--context=" + value
	}

	return ""
}

func buildArgNamespace(value string) string {
	if value != "" {
		return "--namespace=" + value
	}

	return ""
}

func buildArgAllNamespaces(value bool) string {
	if value {
		return "--all-namespaces"
	}

	return ""
}
