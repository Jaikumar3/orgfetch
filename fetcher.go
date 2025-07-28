package main

import (

	"fmt"

	"os/exec"

)

func cloneRepo(url, folder string) error {

	cmd := exec.Command("git", "clone", url)

	cmd.Dir = folder

	output, err := cmd.CombinedOutput()

	if err != nil {

		return fmt.Errorf("git clone failed: %v\n%s", err, string(output))

	}

	return nil

}

