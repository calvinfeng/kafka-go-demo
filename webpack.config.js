const path = require('path');
const ExtractTextPlugin = require('extract-text-webpack-plugin');

/*
const HtmlWebpackPlugin = require('html-webpack-plugin');
const HtmlWebpackPluginConfig = new HtmlWebpackPlugin({
  title: 'Kafkapo',
  filename: 'index.html',
  inject: 'body',
});
*/

const extractSCSS = new ExtractTextPlugin('index.bundle.css');

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
        { test: /\.jsx$/, loader: 'babel-loader', exclude: /node_modules/ },
        { test: /\.scss$/, loaders: ['style-loader', 'css-loader', 'sass-loader'], exclude: /node_modules/}
        // { test: /\.sass$/, loader: ExtractTextPlugin.extract("style-loader", "css-loader!sass-loader"), exclude: /node_modules/ }
        /* using ExtractTextPlugin.extract(["css","sass"]) works too */
      ],
      rules: [
        {
          test: /\.jsx$/,
          exclude: /(node_modules|bower_components)/,
          use: { loader: 'babel-loader' }
        },
        {
          test: /\.scss$/,
          use: ExtractTextPlugin.extract({
            fallback: 'style-loader',
            //resolve-url-loader may be chained before sass-loader if necessary
            use: ['css-loader', 'sass-loader']
          })
        }
      ]
    },
    plugins: [extractSCSS],
    devtool: 'source-maps'
};


/*
rules: [

  {
    test: /\.scss$/,
    use: ExtractTextPlugin.extract({
      fallback: 'style-loader',
      use: 'css-loader'
    })
  }
]
*/
