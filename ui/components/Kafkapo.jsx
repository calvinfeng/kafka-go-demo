import React from 'react';

class Kafkapo extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            isConnected: false,
            messageData: [],
            username: "",
            email: "",
            userMessage: ""
        };
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    componentDidMount() {
        this.state.websocket = new WebSocket('ws://sw33:8000/chat/streams');
        this.state.websocket.onopen = (e) => {
            this.setState({
                isConnected: true
            });
        };

        this.state.websocket.onmessage = (msg) => {
            this.setState({
                messageData: this.state.messageData.concat([JSON.parse(msg.data)])
            })
        };
    }

    sendMessage(msg) {
        const payload = {
            email: "kaka@example.com",
            username: "kaka",
            message: msg
        };
        this.state.websocket.send(JSON.stringify(payload));
    }

    handleChange(e) {
        e.preventDefault();
        switch(e.target.id) {
            case 'username':
                this.setState({ username: e.target.value });
                break;
            case 'email':
                this.setState({ email: e.target.value });
                break;
            case 'message':
                this.setState({ userMessage: e.target.value });
                break;
            default:
                console.log("Do nothing");
        }
    }

    handleSubmit(e) {
        e.preventDefault();
        const payload = {
            email: this.state.email,
            username: this.state.username,
            message: this.state.userMessage
        };
        this.state.websocket.send(JSON.stringify(payload));
    }

    get connectionState() {
        if (this.state.isConnected) {
            return <div>Stream is connected</div>;
        }
    }

    get chatMessages() {
        const messages = this.state.messageData.map((msgData) => {
            return <li>{msgData.message} from {msgData.username}</li>;
        });
        return <ul>{messages}</ul>;
    }

    get inputForm() {
        return (
            <form onSubmit={this.handleSubmit}>
                <label>Username: <input type="text" id="username" value={this.state.username} onChange={this.handleChange} /></label>
                <label>Email: <input type="text" id="email" value={this.state.email} onChange={this.handleChange} /></label>
                <label>Message: <input type="text" id="message" value={this.state.userMessage} onChange={this.handleChange} /></label>
                <input type="submit" value="Submit" />
            </form>
        )
    }

    render() {
        return (
            <div>
                <h1>Hello World</h1>
                {this.connectionState}
                <h2>Messages</h2>
                {this.chatMessages}
                <h2>Form</h2>
                {this.inputForm}
            </div>
        );
    }
}

export default Kafkapo;
