const path = require('path');

module.exports = {
    entry: __dirname + '/static/lib/index.js',
    output: {
         path: __dirname + '/static',
         filename: 'index.bundle.js'
    }
};
