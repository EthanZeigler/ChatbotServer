package botserver

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//if err == io.EOF {
//fmt.Print("[!] Invalid input: Received EOF. Terminating...\n")
//os.Exit(1)
//}
func (i *Instance) debugSendTestCallback(text string, responseWait int) {
	request := httptest.NewRequest("POST",
		"http://example.com", bytes.NewBuffer([]byte(strings.TrimSpace(text))))
	writer := httptest.NewRecorder()
	fmt.Printf(
		"===============\n"+
			"\tSend SIGINT (CTRL+C) to finish or wait %d seconds\n"+
			"===============\n", responseWait/1000)
	i.requestHandler(writer, request)

	done := make(chan bool, 1)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT)
	go func() {
		select {
		case <-sigs:
			fmt.Print("Complete\n")
			done <- true
			break
		case <-time.After(time.Duration(responseWait) * time.Millisecond):
			done <- true
			break
		}
	}()
	<-done
	fmt.Printf("\t-> Callback Body: '%v'\n", writer.Body)
	fmt.Printf("\t-> RESTful adapter buffer: '%v'\n", i.restfulDebugBuffer.String())
	i.restfulDebugBuffer.Reset()
}

// Sets the amount of time the debug server will wait for a response before returning to the menu
func debugSetWaitTime(reader *bufio.Reader, time *int) {
	fmt.Print("Enter time in milliseconds -> ")
	text, err := reader.ReadString('\n')
	if err != nil {
		// input might be a script
		// we don't want to hang the process
		if err == io.EOF {
			fmt.Print("[!] Invalid input: Received EOF. Terminating...\n")
			os.Exit(1)
		}
		fmt.Printf("[!] Invalid input: '%s' (%v)\n", text, err.Error())
		return
	}

	num, err := strconv.Atoi(strings.Trim(text, "\n"))
	if err != nil {
		fmt.Printf("[!] Invalid input: An integer is required\n")
	}

	// input is valid, set time
	*time = num
}

// Set the callback body to be "sent" as a test
func (i *Instance) debugSetString(reader *bufio.Reader, time int) {
	fmt.Print("Paste callback, end with backtick ` -> ")
	// I really can't think of a better easy terminator that's 1 character
	text, err := reader.ReadString('`')
	if err != nil {
		if err == io.EOF {
			fmt.Print("[!] Invalid input: Received EOF. Terminating...\n")
			os.Exit(0)
		}
		fmt.Printf("[!] Invalid Input\n")
		reader.Reset(os.Stdin)
	} else {
		i.debugSendTestCallback(text, time)
	}
}

//
func (i *Instance) debugSetInputPrompted(reader *bufio.Reader, time int) {
	var senderName string
	var senderID string
	var groupID string
	var text string
	// group id, sender name, sender id, text, sender id (again)
	const format = "{\"attachments\":[],\"avatar_url\":\"https://i.groupme.com/142x142.jpeg.c93f451e461b4432af" +
		"c5e1fd7f0eeb73\",\"created_at\":1546082761,\"group_id\":\"%s\",\"id\":\"154608276165808318\",\"name\"" +
		":\"%s\",\"sender_id\":\"%s\",\"sender_type\":\"user\",\"source_guid\":\"3B1120AA-8CC3-48C6-AE62-21717" +
		"458131F\",\"system\":false,\"text\":\"%s\",\"user_id\":\"%s\"}"

	getValue := func(valName string, valDefault string) (string, error) {
		errCnt := 0
		var val string
		for {
			fmt.Printf("%s [empty=%s] -> ", valName, valDefault)
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("[!] Invalid Input\n")
				if errCnt > 3 {
					return "", errors.New("too many failed inputs")
				} else {
					errCnt++
				}
			} else {
				text = strings.TrimSpace(text)
				if len(text) == 0 {
					val = valDefault
				} else {
					val = text
				}
				return val, nil
			}
		}
	}

	val, err := getValue("Sender name", "Q_A")
	if err != nil {
		return
	} else {
		senderName = val
	}

	val, err = getValue("Sender ID", "0C1A58F")
	if err != nil {
		return
	} else {
		senderID = val
	}

	// keep the sender id and group id different, just in case
	val, err = getValue("Group ID", "01234")
	if err != nil {
		return
	} else {
		groupID = val
	}

	// keep the sender id and group id different, just in case
	val, err = getValue("Text", "/help")
	if err != nil {
		return
	} else {
		text = val
	}

	msg := fmt.Sprintf(format, groupID, senderName, senderID, text, senderID)
	i.debugSendTestCallback(msg, time)

}

func (i *Instance) StartDebug(input *os.File) error {
	time.Sleep(500 * time.Millisecond)
	reader := bufio.NewReader(input)
	i.outputToBuffer = true
	responseWait := 15000

	for {
		fmt.Printf("Debug Control Menu:\n"+
			"\t(1) Send custom input (prompt)\n"+
			"\t(2) Send custom input (string)\n"+
			"\t(3) Toggle redirect response to console (%v)\n"+
			"\t(4) Set wait time for test input (%d ms)\n"+
			"\t(5) Quit\n"+
			"-> ", i.outputToBuffer, responseWait)
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Print("[!] Invalid input: Received EOF. Terminating...\n")
				os.Exit(1)
			}
			fmt.Printf("[!] Invalid input: %v\n", err)
		} else {
			ordinal, err := strconv.Atoi(strings.Trim(text, "\n"))
			if err != nil {
				fmt.Printf("[!] Invalid input: '%s' (%v)\n", text, err)
				continue
			}
			switch ordinal {
			case 1:
				i.debugSetInputPrompted(reader, responseWait)
				break
			case 2:
				i.debugSetString(reader, responseWait)
				break
			case 3:
				i.outputToBuffer = !i.outputToBuffer
				break
			case 4:
				debugSetWaitTime(reader, &responseWait)
				break
			case 5:
				return nil
			}
		}
	}
}
