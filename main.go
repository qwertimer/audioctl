package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/google/shlex"
)

func handleLine(line string) (string, string) {
	var desc, name string
	description, err := regexp.MatchString("Description", line)
	if err != nil {
		log.Fatal(err)
	} else if description {
		desc = line
		desc = strings.TrimSpace(desc)
	}
	device, err := regexp.MatchString("Name", line)
	if err != nil {
		log.Fatal(err)
	} else if device {
		name = line
		name = strings.TrimSpace(name)
	}
	return name, desc
}

func main() {
	cmdin := "pactl"
	input := "list sinks"
	strout, _ := shlex.Split(input)
	cmd := exec.Command(cmdin, strout...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("whoops PIPE")
	}
	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()
	if err != nil {
		log.Fatal("DIDNT START")
	}
	var sinks []string
	var dessinks []string
	for scanner.Scan() {
		// Do something with the line here.
		name, desc := handleLine(scanner.Text())
		if name != "" {
			s := strings.Split(name, " ")
			words := s[1]
			// msg := strings.Join(words, " ")
			sinks = append(sinks, words)
		}
		if desc != "" {
			s := strings.Split(desc, " ")
			word := s[1:]
			msg := strings.Join(word, " ")
			dessinks = append(dessinks, msg)
		}

	}
	if scanner.Err() != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}
	cmd.Wait()
	audioMap := make(map[string]string)
	for count, s := range dessinks {
		audioMap[s] = sinks[count]
	}
	subprocess := exec.Command("rofi", "-dmenu")
	stdin, err := subprocess.StdinPipe()
	if err != nil {
		log.Fatal("Broke pipe here")
	}
	defer stdin.Close()
	var out, serr bytes.Buffer
	subprocess.Stdout = &out
	subprocess.Stderr = &serr
	if err = subprocess.Start(); err != nil { // Use start, not run
		fmt.Println("An error occured: ", err) // replace with logger, or anything you want
	}
	for i := range dessinks {
		input := dessinks[i] + "\n"
		io.WriteString(stdin, input)
	}

	subprocess.Wait()

	val := out.String()
	val = strings.TrimSuffix(val, "\n")
	fmt.Println(val)
	value, ok := audioMap[val]
	fmt.Println(ok)
	// fmt.Println(value)
	c2 := exec.Command("pactl", "set-default-sink", value)
	var out1, serr1 bytes.Buffer
	c2.Stdout = &out1
	c2.Stderr = &serr1
	c2.Run()
	fmt.Println("out: ", out1.String(), "err:", serr1.String())
}
