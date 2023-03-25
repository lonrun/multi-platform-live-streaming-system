const messages = document.getElementById('messages');
const messageInput = document.getElementById('message-input');
const messageForm = document.getElementById('message-form');

const socket = new WebSocket('ws://localhost:8000/ws');

// WebSocket message types
const MessageTypeChat = 0;
const MessageTypeSignal = 1;

socket.onmessage = (event) => {
    const msgData = JSON.parse(event.data);
    const messageElement = document.createElement('div');
    messageElement.textContent = `${msgData.sender} (${msgData.timestamp}): ${msgData.message}`;
    messages.appendChild(messageElement);
    messages.scrollTop = messages.scrollHeight;
};

socket.onopen = () => {
    console.log('Connected to the server.');
};

socket.onclose = () => {
    console.log('Disconnected from the server.');
};

messageForm.addEventListener('submit', (event) => {
    event.preventDefault();

    if (messageInput.value.trim() === '') {
        return;
    }

    socket.send(JSON.stringify({ type: MessageTypeChat, payload: {
        sender: "Your_Sender_Name",
        message: messageInput.value,
        timestamp: new Date().toISOString()
      } }));

    console.log("msg:"+ JSON.stringify({ type: MessageTypeChat, payload: {
        sender: "Your_Sender_Name",
        message: messageInput.value,
        timestamp: new Date().toISOString()
      } }));

    messageInput.value = '';
});
