package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type User struct {
	Name     string
	Conn     *net.Conn
	PrevName string
}

var (
	users   map[string]*User
	history string
	mutex   sync.Mutex
)

func init() {
	users = map[string]*User{}
}
func main() {
	port := ":8989"

	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
	}

	li, err := net.Listen("tcp", port)
	fmt.Println("Server is listening...")
	if err != nil {
		log.Panic(err)
	}
	defer li.Close()
	for {
		conn, err := li.Accept()
		if len(users) == 10 {
			fmt.Fprintln(conn, "Sorry, chat is filled")
			conn.Close()
			continue
		}
		if err != nil {
			log.Println(err)
		}
		mutex.Lock()
		user := &User{Conn: &conn}
		mutex.Unlock()
		user.createUser(&conn)
		go user.handle()
		defer conn.Close()
	}
}
func (u *User) handle() {
	scanner := bufio.NewScanner(*u.Conn)
	for scanner.Scan() {
		ln := scanner.Text()
		raised := u.checkFlags(&ln)
		if raised {
			continue
		}
		if ln != "" {
			msg := fmt.Sprintf("[%s][%s]:%s", time.Now().Format("2006-01-02 15:04:05"), u.Name, ln)
			mutex.Lock()
			history += msg + "\n"
			for key := range users {
				if key != u.Name {
					fmt.Fprintf(*users[key].Conn, "\n"+msg)
					fmt.Fprintf(*users[key].Conn, "\n[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), users[key].Name)
				}
			}
			mutex.Unlock()
		}
		fmt.Fprintf(*u.Conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), u.Name)
	}
	u.exitMessage()
}

func (u *User) checkFlags(ln *string) bool {
	raised := false
	arr := strings.Split(strings.Trim(*ln, " "), " ")
	if len(arr) == 2 && arr[0] == "\\change_name" {
		raised = true
		mutex.Lock()
		if users[arr[1]] != nil {
			fmt.Fprintf(*u.Conn, "\nName is already taken")
		} else {
			chgMsg := fmt.Sprintf("%s has changed name to %s", u.Name, arr[1])
			history += chgMsg + "\n"
			delete(users, u.Name)
			u.Name = arr[1]
			users[arr[1]] = u
			for key := range users {
				if key != u.Name {
					fmt.Fprintf(*users[key].Conn, "\n"+chgMsg)
					fmt.Fprintf(*users[key].Conn, "\n[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), users[key].Name)
				}
			}
			fmt.Fprintf(*u.Conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), u.Name)
		}
		mutex.Unlock()
	}
	return raised
}

func (u *User) exitMessage() {
	exitMsg := fmt.Sprintf("%s has left our chat...", u.Name)
	mutex.Lock()
	history += exitMsg + "\n"
	for key := range users {
		if key != u.Name {
			fmt.Fprintf(*users[key].Conn, "\n"+exitMsg)
			fmt.Fprintf(*users[key].Conn, "\n[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), users[key].Name)
		} else {
			delete(users, key)
		}
	}
	mutex.Unlock()
}

func (u *User) createUser(conn *net.Conn) {
	welcomeMessage := `Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    '.       | ' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
	 '-'       '--'
[ENTER YOUR NAME]:`
	fmt.Fprintf(*conn, "%v", welcomeMessage)
	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		ln := scanner.Text()

		if ln == "" {
			fmt.Fprint(*conn, "[ENTER YOUR NAME]:")
			continue
		}
		mutex.Lock()
		if users[ln] != nil {
			fmt.Fprint(*conn, "[USER EXISTS]")
			fmt.Fprint(*conn, "[ENTER YOUR NAME]:")
			mutex.Unlock()
			continue
		}
		mutex.Unlock()
		u.Name = ln
		users[ln] = u
		mutex.Lock()
		fmt.Fprintf(*conn, history)

		mutex.Unlock()
		fmt.Fprintf(*conn, "[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), u.Name)
		joinMessage := fmt.Sprintf("%s has joined our chat...", ln)
		mutex.Lock()
		history += joinMessage + "\n"
		for key := range users {
			if key != ln {
				fmt.Fprintf(*users[key].Conn, "\n"+joinMessage)
				fmt.Fprintf(*users[key].Conn, "\n[%s][%s]:", time.Now().Format("2006-01-02 15:04:05"), users[key].Name)
			}
		}
		mutex.Unlock()
		break
	}

}
