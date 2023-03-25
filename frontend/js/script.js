const messageForm = document.getElementById("message-form");
const messageInput = document.getElementById("message-input");
const messages = document.getElementById("messages");

// Connect to WebSocket
const ws = new WebSocket("ws://localhost:8000/ws");
const pc = new RTCPeerConnection({
  iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
});

// WebSocket message types
const MessageTypeChat = 0;
const MessageTypeSignal = 1;

ws.addEventListener("open", (event) => {
  console.log("WebSocket connection opened:", event);
});

// Add an event listener for when a message is received from the server
ws.addEventListener("message", (event) => {
  const data = JSON.parse(event.data);
  switch (data.type) {
    case MessageTypeChat: // MessageTypeChat
      const chatMessage = JSON.parse(event.payload);
      displayMessage(
        chatMessage.sender,
        chatMessage.message,
        chatMessage.timestamp
      );
      break;
    case MessageTypeSignal: // MessageTypeSignal
      // Handle signaling messages here
      handleSignalingMessage(data.payload);
      break;
  }
});

function displayMessage(sender, message, timestamp) {
  const messageContainer = document.createElement("div");
  messageContainer.innerHTML = `<b>${sender}:</b> ${message} <i>(${timestamp})</i>`;
  messages.appendChild(messageContainer);
}

async function handleSignalingMessage(data) {
  if (data.sdp) {
    await pc.setRemoteDescription(new RTCSessionDescription(data));
    if (data.type === "offer") {
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      signalingChannel.send(
        JSON.stringify({ type: "signal", payload: pc.localDescription })
      );
    }
  } else if (data.candidate) {
    try {
      await pc.addIceCandidate(data);
    } catch (err) {
      console.error("Error adding ICE candidate:", err);
    }
  }
}

messageForm.addEventListener("submit", (event) => {
  event.preventDefault();

  const messageText = messageInput.value.trim();
  if (!messageText) return;

  const messageData = {
    type: MessageTypeChat,
    payload: {
      sender: "Your_Sender_Name",
      message: messageText,
      timestamp: new Date().toISOString(),
    },
  };

  ws.send(JSON.stringify(messageData));

  displayMessage("You", messageText, new Date().toLocaleTimeString());

  messageInput.value = "";
});
