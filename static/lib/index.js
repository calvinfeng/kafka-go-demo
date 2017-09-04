const socket = new WebSocket('ws://localhost:8000/chat/streams');

socket.onopen = (e) => {
    console.log("Joined streams");
    socket.send(
        JSON.stringify({
            email: "",
            username: "",
            message: "Hi, this is Client!!!"
        })
    );
};
