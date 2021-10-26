// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"io"
	"os/exec"
	"regexp"
	"strconv"
)

// TODO support other ports besides 80
var (
	portForwardRegexp = regexp.MustCompile(`Forwarding from (127.0.0.1|\\[::1\\]):([0-9]+) -> 80`)
)

// StartCmdAndStreamOutput returns stdout and stderr after starting the given cmd.
// The function was inspired from kubernetes e2e framework
func StartCmdAndStreamOutput(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	return
}

// runPortForward runs port-forward, warning, this may need root functionality on some systems.
// The function was inspired from kubernetes e2e framework
func RunPortForward(ns string, svcName string, port int) (string, error) {
	/* #nosec G204 */
	// Avoid "Subprocess launched with variable" error
	cmd := exec.Command("kubectl", "-n", ns, "port-forward", "svc/"+svcName, ":"+strconv.Itoa(port))
	// This is somewhat ugly but is the only way to retrieve the port that was picked
	// by the port-forward command. We don't want to hard code the port as we have no
	// way of guaranteeing we can pick one that isn't in use, particularly on Jenkins.
	portOutput, _, err := StartCmdAndStreamOutput(cmd)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 128)
	var n int
	if n, err = portOutput.Read(buf); err != nil {
		return "", err
	}
	portForwardOutput := string(buf[:n])
	match := portForwardRegexp.FindStringSubmatch(portForwardOutput)
	if len(match) != 3 {
		return "", err
	}

	_, err = strconv.Atoi(match[2])
	if err != nil {
		return "", err
	}

	return match[2], nil
}
