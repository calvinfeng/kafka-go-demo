import React from 'react';

class Kafkapo extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            isConnected: false
        };
    }

    componentDidMount() {
        this.state.websocket = new WebSocket('ws://localhost:8000/chat/streams');
        this.state.websocket.onopen = (e) => {
            // this.sendMessage('I am connected!');
            this.setState({
                isConnected: true
            });
        }

        this.state.websocket.onmessage = (msg) => {
            console.log(msg);
        }
    }

    sendMessage(msg) {
        const payload = {
            email: "kaka@example.com",
            username: "kaka",
            message: msg
        };
        this.state.websocket.send(JSON.stringify(payload));
    }

    get connectionState() {
        if (this.state.isConnected) {
            return <div>Stream is connected</div>;
        }
    }

    render() {
        return (
            <div>
                <h1>Hello World</h1>
                {this.connectionState}
            </div>
        );
    }
}

export default Kafkapo;
