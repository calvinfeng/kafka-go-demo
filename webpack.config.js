const path = require('path');

/*
const HtmlWebpackPlugin = require('html-webpack-plugin');
const HtmlWebpackPluginConfig = new HtmlWebpackPlugin({
  title: 'Kafkapo',
  filename: 'index.html',
  inject: 'body',
});
*/

module.exports = {
    entry: __dirname + '/ui/index.jsx',
    output: {
         path: __dirname + '/static',
         filename: 'index.bundle.js'
    },
    resolve: {
        extensions: ['.js', '.jsx']
    },
    module: {
        loaders: [
            { test: /\.js$/, loader: 'babel-loader', exclude: /node_modules/ },
            { test: /\.jsx$/, loader: 'babel-loader', exclude: /node_modules/ }
        ]
    }
};
