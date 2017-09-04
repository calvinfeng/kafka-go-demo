const socket = new WebSocket('ws://localhost:8000/chat/streams');

console.log("Hello World");

socket.onopen = (e) => {
    console.log("Joined streams");
    console.log(e);
};
