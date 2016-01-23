package main
import (
	"github.com/gorilla/websocket"
	"net/http"
	"log"
	"github.com/rabbitcount/goblueprints/chapter1/trace"
	"github.com/stretchr/objx"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan *message
	// join is a channel for clients wishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in the room.
	// hold the pointer of the client (reference)
	clients map[*client] bool
	// tracer will receive trace information of activity in the room
	tracer trace.Tracer
}

// newRoom makes a new room hat is ready to go.
// the users of our code need only call the newRoom function
// instead of the more verbose six lines of code.
func newRoom() *room {
	return &room{
		forward:	make(chan *message),
		join:		make(chan *client),
		leave:		make(chan *client),
		clients:	make(map[*client]bool),
		tracer:		trace.Off(),
	}
}

func (r *room) run()  {
	// indicates that this method will run forever, until the program is terminated
	// if run this code as a Go routine, it will run in the background,
	// which won't bloc the rest of our application.
	for {
		// keep watching the three channels inside the room
		select {
		case client := <- r.join:
			// joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <- r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <- r.forward:
			// If we receive a message on the forward channel,
			// we iterate over all the clients and send the message down each client's send channel.
			// forward message to all clients
			for client := range r.clients {
				r.tracer.Trace("Message received: ", string(msg.Message))
				select {
				case client.send <- msg:
					// send the message
					r.tracer.Trace(" -- sent to client")
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize	= 1024
	messageBufferSize	= 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// means a room can now act as a handler
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// In order to use web sockets, we must upgrade the HTTP connection using
	// the websocket.Upgrader type, which is reusable so we need only create one.
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &client {
		socket: 	socket,
		send:		make(chan *message, messageBufferSize),
		room:		r,
		userData:	objx.MustFromBase64(authCookie.Value),
	}
	r.join <- client
	// will call after the operation finished
	defer func() {
		r.leave <- client
	}()
	// The write method for the client is then called as a Go routine
	// the word go followed by a space character
	go client.write()
	client.read()
}