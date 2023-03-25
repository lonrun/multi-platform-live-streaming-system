package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// WebSocket message types
const (
	MessageTypeChat = iota
	MessageTypeSignal
)

type Server struct {
	engine  *gin.Engine
	sfu     *webrtc.PeerConnection
	clients map[string]*websocket.Conn
}

type WebSocketMessage struct {
	Type int                    `json:"type"`    // MessageTypeChat, MessageTypeSignal, etc.
	Data map[string]interface{} `json:"payload"` // contains the actual data
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

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

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

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
				handleChatMessage(server.clients, wsMsg.Data)
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

func (s *Server) Run(addr string) error {
	return s.engine.Run(addr)
}

func handleChatMessage(clients map[string]*websocket.Conn, msgData map[string]interface{}) {
	log.Printf("Recv chat:%v\n", msgData)

	if msgData == nil {
		log.Println("empty data recv")
		return
	}

	msgBytes, err := json.Marshal(ChatMessage{
		Sender:    msgData["sender"].(string),
		Message:   msgData["message"].(string),
		Timestamp: msgData["timestamp"].(string),
	})
	if err != nil {
		return
	}

	for _, client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
			log.Printf("Error sending chat message to client: %v", err)
		}
	}
}

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
