package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	flag.Parse()

	whomst := flag.Args()[0]
	port := flag.Args()[1]

	server := newChatServer(whomst, net.JoinHostPort("", port))

	if len(flag.Args()) > 2 {
		connectAddr := flag.Args()[2]

		go func() {
			server.Dial(fmt.Sprintf("ws://%s/chat", connectAddr))
		}()
	}

	go func() {
		fmt.Println(server.ListenAndServe())
	}()

	go func() {
		var lastRead clock
		for {
			updates := server.chat.since(lastRead)
			lastRead = lastRead.update(updates.lastMessageAt())

			if len(updates.Messages) > 0 {
				for _, msg := range updates.Messages {
					fmt.Println(msg.Whomst, "::", msg.Txt)
				}
			}

			time.Sleep(20 * time.Millisecond)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if len(txt) > 0 {
			server.chat.send(whomst, txt)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
