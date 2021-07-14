package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type member struct {
	channel chan<- string
	name    string
}

var (
	entering = make(chan member)
	leaving  = make(chan member)
	messages = make(chan string)
	checking = make(chan bool)
)

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
	members := make(map[string]member)

	for {
		select {
		case msg := <-messages:
			for _, m := range members {
				m.channel <- msg
			}
		case newbie := <-entering:
			if _, ok := members[newbie.name]; ok {
				checking <- false
			} else {
				members[newbie.name] = newbie
				checking <- true
			}
		case leaved := <-leaving:
			delete(members, leaved.name)
			close(leaved.channel)
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
				leaving <- member{ch, name}

				messages <- name + " has left"

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

		newbie := member{ch, name}

		entering <- newbie
		isExist := <-checking
		if isExist {
			break
		} else {
			fmt.Fprintln(conn, "Name is already exist")
			name = askName(conn)
		}
	}

	return ch, name
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func askName(conn net.Conn) string {
	fmt.Fprintln(conn, "Enter your name: ")
	name, _ := bufio.NewReader(conn).ReadString('\n')
	return name[:len(name)-1]
}

func sendWelcomeMessage(name string, ch chan<- string) {
	ch <- "If you want to exit, please type 'exit()'"
	ch <- "You are " + name
	messages <- name + " has arrived"
}
