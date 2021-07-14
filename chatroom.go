package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type channel chan<- string

var messages = make(chan string)

var members = make(map[string]channel)

func main() {
	chatRoom, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcast()

	for {
		conn, err := chatRoom.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go enterChatRoom(conn)
	}
}

func broadcast() {
	for {
		select {
		case msg := <-messages:
			for _, ch := range members {
				ch <- msg
			}
		}
	}
}

func enterChatRoom(conn net.Conn) {
	ch, name := setupChatting(conn)

	sendWelcomeMessage(name, ch)

	input := bufio.NewScanner(conn)
	for {
		for input.Scan() {
			typed := input.Text()

			if typed == "exit()" {
				messages <- name + " has left"

				delete(members, name)
				close(ch)

				conn.Close()
				return
			}

			messages <- name + ": " + typed
		}
	}
}

func setupChatting(conn net.Conn) (chan<- string, string) {
	ch := make(chan string)

	go clientWriter(conn, ch)

	name := askName(conn)
	for {
		if name == "exit()" {
			fmt.Fprintln(conn, "Can't use name 'exit()'")
			continue
		}

		if _, ok := members[name]; ok {
			fmt.Fprintln(conn, "Name is already exist")
			name = askName(conn)
		} else {
			members[name] = ch
			break
		}
	}

	return ch, name
}

func clientWriter(conn net.Conn, ch <- chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func askName(conn net.Conn) string {
	fmt.Fprintln(conn, "Enter your name: ")
	name, _ := bufio.NewReader(conn).ReadString('\n')
	return name[:len(name) - 1]
}

func sendWelcomeMessage(name string, ch chan<- string) {
	ch <- "If you want to exit, please type 'exit()'"
	ch <- "You are " + name
	messages <- name + " has arrived"
}
