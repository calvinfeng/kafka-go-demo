import React from 'react';
import ReactDOM from 'react-dom';
import Kafkapo from './components/Kafkapo';

import lightBaseTheme from 'material-ui/styles/baseThemes/lightBaseTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import getMuiTheme from 'material-ui/styles/getMuiTheme';

import '../styles/application.scss';

const Application = () => (
  <MuiThemeProvider muiTheme={getMuiTheme(lightBaseTheme)}>
    <Kafkapo />
  </MuiThemeProvider>
);

document.addEventListener("DOMContentLoaded", () => {
  ReactDOM.render(<Application />, document.getElementById('kafkapo-application'));
});
