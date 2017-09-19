import React from 'react';
import hash from 'object-hash';
import randomSentence from 'random-sentence';
import AppBar from 'material-ui/AppBar';
import TextField from 'material-ui/TextField';
import RaisedButton from 'material-ui/RaisedButton';
import Paper from 'material-ui/Paper';
import Divider from 'material-ui/Divider';


class Kafkapo extends React.Component {
    state = {
        hasUserInfo: false,
        isConnected: false,
        historicalMessages: [],
        username: "",
        email: "",
        message: ""
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

    handleMsgChange = (e) => {
        e.preventDefault();
        this.setState({ message: e.target.value });
    }

    handleUserInfoChange = (field) => {
        switch(field) {
            case 'username':
                return (e) => {
                    e.preventDefault();
                    this.setState({ username: e.target.value });
                };
            case 'email':
                return (e) => {
                    e.preventDefault();
                    this.setState({ email: e.target.value });
                };
            default:
                console.warn("Unknown field:", field);
        }
    }

    handleSend = (e) => {
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

    handleUserInfoSubmit = (e) => {
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
            return (
                <div>
                    <h2>Socket Connection</h2>
                    <p>Stream is connected</p>
                </div>
            );
        }
        return (
            <div>
                <h2>Socket Connection</h2>
                <p>Attempting to connect</p>
            </div>
        );
    }

    get historicalMessages() {
        const listItems = [];
        for (let i = 0; i < this.state.historicalMessages.length; i++) {
            const msgData = this.state.historicalMessages[i];
            listItems.push(<li key={i}><strong>{msgData.username}</strong> says "{msgData.message}"</li>);
        }
        return <ul>{listItems}</ul>;
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
                <TextField
                    hintText="Any name will do"
                    floatingLabelText="Username"
                    value={this.state.username}
                    onChange={this.handleUserInfoChange('username')} />
                <br />
                <TextField
                    hintText="Any email will do"
                    floatingLabelText="Email"
                    value={this.state.email}
                    onChange={this.handleUserInfoChange('email')} />
                <br />
                <RaisedButton type="submit" label="Submit" />
            </form>
        )
    }

    get messageInputForm() {
        if (this.state.hasUserInfo) {
            return (
                <form onSubmit={this.handleSend}>
                    <TextField
                        hintText="Send a message to Kafka cluster"
                        floatingLabelText="Message"
                        value={this.state.message}
                        onChange={this.handleMsgChange}
                        fullWidth={true} />
                    <RaisedButton type="submit" label="Send" />
                </form>
            )
        }
        return <p>Please enter your username and email</p>;
    }

    render() {
        return (
            <div>
                <AppBar title="Kafkapo" />
                {this.userInfoInputForm}
                <Paper zDepth={1}>
                    {this.connectionState}
                    <Divider />
                    <h2>Messages</h2>
                    {this.historicalMessages}
                    <Divider />
                    <h2>Form</h2>
                    {this.messageInputForm}
                </Paper>
            </div>
        );
    }
}

export default Kafkapo;
