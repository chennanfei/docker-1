package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-check/check"
)

// This is a heisen-test.  Because the created timestamp of images and the behavior of
// sort is not predictable it doesn't always fail.
func (s *DockerSuite) TestBuildHistory(c *check.C) {
	name := "testbuildhistory"
	defer deleteImages(name)
	_, err := buildImage(name, `FROM busybox
RUN echo "A"
RUN echo "B"
RUN echo "C"
RUN echo "D"
RUN echo "E"
RUN echo "F"
RUN echo "G"
RUN echo "H"
RUN echo "I"
RUN echo "J"
RUN echo "K"
RUN echo "L"
RUN echo "M"
RUN echo "N"
RUN echo "O"
RUN echo "P"
RUN echo "Q"
RUN echo "R"
RUN echo "S"
RUN echo "T"
RUN echo "U"
RUN echo "V"
RUN echo "W"
RUN echo "X"
RUN echo "Y"
RUN echo "Z"`,
		true)

	if err != nil {
		c.Fatal(err)
	}

	out, exitCode, err := runCommandWithOutput(exec.Command(dockerBinary, "history", "testbuildhistory"))
	if err != nil || exitCode != 0 {
		c.Fatalf("failed to get image history: %s, %v", out, err)
	}

	actualValues := strings.Split(out, "\n")[1:27]
	expectedValues := [26]string{"Z", "Y", "X", "W", "V", "U", "T", "S", "R", "Q", "P", "O", "N", "M", "L", "K", "J", "I", "H", "G", "F", "E", "D", "C", "B", "A"}

	for i := 0; i < 26; i++ {
		echoValue := fmt.Sprintf("echo \"%s\"", expectedValues[i])
		actualValue := actualValues[i]

		if !strings.Contains(actualValue, echoValue) {
			c.Fatalf("Expected layer \"%s\", but was: %s", expectedValues[i], actualValue)
		}
	}

}

func (s *DockerSuite) TestHistoryExistentImage(c *check.C) {
	historyCmd := exec.Command(dockerBinary, "history", "busybox")
	_, exitCode, err := runCommandWithOutput(historyCmd)
	if err != nil || exitCode != 0 {
		c.Fatal("failed to get image history")
	}
}

func (s *DockerSuite) TestHistoryNonExistentImage(c *check.C) {
	historyCmd := exec.Command(dockerBinary, "history", "testHistoryNonExistentImage")
	_, exitCode, err := runCommandWithOutput(historyCmd)
	if err == nil || exitCode == 0 {
		c.Fatal("history on a non-existent image didn't result in a non-zero exit status")
	}
}

func (s *DockerSuite) TestHistoryImageWithComment(c *check.C) {
	name := "testhistoryimagewithcomment"
	defer deleteImages(name)

	// make a image through docker commit <container id> [ -m messages ]
	//runCmd := exec.Command(dockerBinary, "run", "-i", "-a", "stdin", "busybox", "echo", "foo")
	runCmd := exec.Command(dockerBinary, "run", "--name", name, "busybox", "true")
	out, _, err := runCommandWithOutput(runCmd)
	if err != nil {
		c.Fatalf("failed to run container: %s, %v", out, err)
	}

	waitCmd := exec.Command(dockerBinary, "wait", name)
	if out, _, err := runCommandWithOutput(waitCmd); err != nil {
		c.Fatalf("error thrown while waiting for container: %s, %v", out, err)
	}

	comment := "This_is_a_comment"

	commitCmd := exec.Command(dockerBinary, "commit", "-m="+comment, name, name)
	if out, _, err := runCommandWithOutput(commitCmd); err != nil {
		c.Fatalf("failed to commit container to image: %s, %v", out, err)
	}

	// test docker history <image id> to check comment messages
	historyCmd := exec.Command(dockerBinary, "history", name)
	out, exitCode, err := runCommandWithOutput(historyCmd)
	if err != nil || exitCode != 0 {
		c.Fatalf("failed to get image history: %s, %v", out, err)
	}

	outputTabs := strings.Fields(strings.Split(out, "\n")[1])
	//outputTabs := regexp.MustCompile("  +").Split(outputLine, -1)
	actualValue := outputTabs[len(outputTabs)-1]

	if !strings.Contains(actualValue, comment) {
		c.Fatalf("Expected comments %q, but found %q", comment, actualValue)
	}

}

func (s *DockerSuite) TestHistoryHumanOptionFalse(c *check.C) {
	out, _, _ := runCommandWithOutput(exec.Command(dockerBinary, "history", "--human=false", "busybox"))
	lines := strings.Split(out, "\n")
	sizeColumnRegex, _ := regexp.Compile("SIZE +")
	indices := sizeColumnRegex.FindStringIndex(lines[0])
	startIndex := indices[0]
	endIndex := indices[1]
	for i := 1; i < len(lines)-1; i++ {
		if endIndex > len(lines[i]) {
			endIndex = len(lines[i])
		}
		sizeString := lines[i][startIndex:endIndex]
		if _, err := strconv.Atoi(strings.TrimSpace(sizeString)); err != nil {
			c.Fatalf("The size '%s' was not an Integer", sizeString)
		}
	}
}

func (s *DockerSuite) TestHistoryHumanOptionTrue(c *check.C) {
	out, _, _ := runCommandWithOutput(exec.Command(dockerBinary, "history", "--human=true", "busybox"))
	lines := strings.Split(out, "\n")
	sizeColumnRegex, _ := regexp.Compile("SIZE +")
	humanSizeRegex, _ := regexp.Compile("^\\d+.*B$") // Matches human sizes like 10 MB, 3.2 KB, etc
	indices := sizeColumnRegex.FindStringIndex(lines[0])
	startIndex := indices[0]
	endIndex := indices[1]
	for i := 1; i < len(lines)-1; i++ {
		if endIndex > len(lines[i]) {
			endIndex = len(lines[i])
		}
		sizeString := lines[i][startIndex:endIndex]
		if matchSuccess := humanSizeRegex.MatchString(strings.TrimSpace(sizeString)); !matchSuccess {
			c.Fatalf("The size '%s' was not in human format", sizeString)
		}
	}
}
