const messageForm = document.getElementById("message-form");
const messageInput = document.getElementById("message-input");
const messages = document.getElementById("messages");

// Connect to WebSocket
const ws = new WebSocket("ws://localhost:8000/ws");

// WebSocket message types
const MessageTypeChat = 0;
const MessageTypeSignal = 1;

ws.addEventListener("open", (event) => {
  console.log("WebSocket connection opened:", event);
});

ws.addEventListener("message", (event) => {
  const chatMessage = JSON.parse(event.data);
  displayMessage(chatMessage.sender, chatMessage.message, chatMessage.timestamp);
});

function displayMessage(sender, message, timestamp) {
  const messageContainer = document.createElement("div");
  messageContainer.innerHTML = `<b>${sender}:</b> ${message} <i>(${timestamp})</i>`;
  messages.appendChild(messageContainer);
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
