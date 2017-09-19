import React from 'react';
import ReactDOM from 'react-dom';
import Kafkapo from './components/Kafkapo';

document.addEventListener("DOMContentLoaded", () => {
  ReactDOM.render(<Kafkapo />, document.getElementById('kafkapo-application'));
});
