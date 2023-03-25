package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// WebSocket message types constants
const (
	MessageTypeChat   = iota // MessageTypeChat for chat messages
	MessageTypeSignal        // MessageTypeSignal for signaling messages
)

// Server struct represents the main server object
type Server struct {
	engine  *gin.Engine                // Gin engine for handling HTTP requests
	sfu     *webrtc.PeerConnection     // WebRTC PeerConnection for media streaming
	clients map[string]*websocket.Conn // Map of connected clients' WebSockets
}

// WebSocketMessage struct represents the structure of a WebSocket message
type WebSocketMessage struct {
	Type int                    `json:"type"`    // MessageTypeChat, MessageTypeSignal, etc.
	Data map[string]interface{} `json:"payload"` // contains the actual data
}

// ChatMessage struct represents the structure of a chat message
type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// NewServer function initializes and returns a new Server instance
func NewServer() (*Server, error) {
	router := gin.Default()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// WebRTC configuration
	conf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:       []string{"stun:stun.l.google.com:19302"},
				Username:   "",
				Credential: "",
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, pcErr := webrtc.NewPeerConnection(conf)
	if pcErr != nil {
		return nil, pcErr
	}

	server := &Server{
		engine:  router,
		sfu:     peerConnection,
		clients: make(map[string]*websocket.Conn),
	}

	// WebSocket endpoint for handling client connections and messages
	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Generate a unique ID for the client
		clientID := uuid.NewString()

		// Add the client to server.clients
		server.clients[clientID] = conn

		defer func() {
			// Remove the client when the connection is closed
			delete(server.clients, clientID)
		}()

		var lock sync.RWMutex

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			// Unmarshal the received WebSocket message
			var wsMsg WebSocketMessage
			err = json.Unmarshal(message, &wsMsg)
			if err != nil {
				break
			}

			switch wsMsg.Type {
			case MessageTypeChat:
				// Handle chat messages
				log.Println("client online ", len(server.clients))
				handleChatMessage(server.clients, wsMsg.Data, conn)
			case MessageTypeSignal:
				// Handle signaling messages
				lock.Lock()
				err = handleSignaling(server.sfu, wsMsg.Data)
				lock.Unlock()
			}

			if err != nil {
				break
			}
		}
	})

	return server, nil
}

// Run function starts the server at the specified address
func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

// handleChatMessage function handles incoming chat messages and broadcasts them to all clients
func handleChatMessage(clients map[string]*websocket.Conn, msgData map[string]interface{}, senderConn *websocket.Conn) {
	if msgData == nil {
		log.Println("empty data recv")
		return
	}

	log.Printf("Recv chat:%v\n", msgData)
	msgBytes, err := json.Marshal(ChatMessage{
		Sender:    msgData["sender"].(string),
		Message:   msgData["message"].(string),
		Timestamp: msgData["timestamp"].(string),
	})
	if err != nil {
		return
	}
	// Send the chat message to all connected clients
	for _, client := range clients {
		if client != senderConn {
			if err := client.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
				log.Printf("Error sending chat message to client: %v", err)
			}
		}
	}
}

// handleSignaling function handles WebRTC signaling messages and updates PeerConnection state accordingly
func handleSignaling(pc *webrtc.PeerConnection, msgData map[string]interface{}) error {
	sdpTypeStr, ok := msgData["type"].(string)
	if !ok {
		return fmt.Errorf("unable to parse SDP type")
	}

	sdpType := webrtc.NewSDPType(sdpTypeStr)

	sdpStr, ok := msgData["sdp"].(string)
	if !ok {
		return fmt.Errorf("invalid SDP")
	}

	switch webrtc.SDPType(sdpType) {
	case webrtc.SDPTypeOffer:
		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  sdpStr,
		}
		err := pc.SetRemoteDescription(offer)
		if err != nil {
			return err
		}

		answer, err := pc.CreateAnswer(nil)
		if err != nil {
			return err
		}

		err = pc.SetLocalDescription(answer)
		if err != nil {
			return err
		}

	case webrtc.SDPTypeAnswer:
		answer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  sdpStr,
		}
		err := pc.SetRemoteDescription(answer)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported SDP type: %s", sdpType)
	}

	return nil
}
