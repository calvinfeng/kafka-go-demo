import React from 'react';
import hash from 'object-hash';
import randomSentence from 'random-sentence';


class Kafkapo extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            hasUserInfo: false,
            isConnected: false,
            historicalMessages: [],
            username: "",
            email: "",
            message: ""
        };
        this.handleMsgChange = this.handleMsgChange.bind(this);
        this.handleSend = this.handleSend.bind(this);
        this.handleUserInfoSubmit = this.handleUserInfoSubmit.bind(this);
        this.handleUserInfoChange = this.handleUserInfoChange.bind(this);
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
                historicalMessages: this.state.historicalMessages.concat([JSON.parse(msg.data)])
            })
        };
    }

    handleMsgChange(e) {
        e.preventDefault();
        this.setState({ message: e.target.value });
    }

    handleUserInfoChange(e) {
        e.preventDefault();
        switch(e.target.id) {
            case 'username':
                this.setState({ username: e.target.value });
                break;
            case 'email':
                this.setState({ email: e.target.value });
                break;
            default:
                console.log("username and email html elements do not exist");
        }
    }

    handleSend(e) {
        if (e) {
            e.preventDefault();
        }

        const payload = {
            email: this.state.email,
            username: this.state.username,
            message: this.state.message
        };

        payload.hash = hash(payload);
        this.state.websocket.send(JSON.stringify(payload));
        this.setState({message: ""});
    }

    handleUserInfoSubmit(e) {
        e.preventDefault();
        if (this.state.username.length > 0 && this.state.email.length > 0) {
            this.setState({ hasUserInfo: true });
        }

        setInterval(() => {
            this.setState({ message: randomSentence() });
            this.handleSend();
        }, 500);
    }

    get connectionState() {
        if (this.state.isConnected) {
            return <div>Stream is connected</div>;
        }
    }

    get historicalMessages() {
        const messages = this.state.historicalMessages.map((msgData) => {
            const msgId = msgData.hash ? msgData.hash : hash(msgData);
            return <li key={msgId}><strong>{msgData.username}</strong> says "{msgData.message}"</li>;
        });

        return <ul>{messages}</ul>;
    }

    get userInfoInputForm() {
        if (this.state.hasUserInfo) {
            return (
                <div>
                    <label>Current user: {this.state.username}</label><br/>
                    <label>{this.state.email}</label>
                </div>
            );
        }
        return (
            <form onSubmit={this.handleUserInfoSubmit}>
                <label>
                    Username: <input type="text" id="username" value={this.state.username} onChange={this.handleUserInfoChange} />
                </label>
                <label>
                    Email: <input type="text" id="email" value={this.state.email} onChange={this.handleUserInfoChange} />
                </label>
                <input type="submit" value="Submit" />
            </form>
        )
    }

    get messageInputForm() {
        if (this.state.hasUserInfo) {
            return (
                <form onSubmit={this.handleSend}>
                    <label>Message: <input type="text" id="message" value={this.state.message} onChange={this.handleMsgChange} /></label>
                    <input type="submit" value="Send" />
                </form>
            )
        }
        return <div>Please enter your username and email</div>;
    }

    render() {
        return (
            <div>
                <h1>Hello World</h1>
                {this.userInfoInputForm}
                {this.connectionState}
                <h2>Messages</h2>
                {this.historicalMessages}
                <h2>Form</h2>
                {this.messageInputForm}
            </div>
        );
    }
}

export default Kafkapo;
